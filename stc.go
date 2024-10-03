// syncthing cli tool
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"text/tabwriter"

	humanize "github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
	"github.com/tenox7/stc/api"
)

var (
	apiKey  = flag.String("apikey", "", "Syncthing API Key")
	target  = flag.String("target", "", "Syncthing Target URL")
	homeDir = flag.String("homedir", "", "Syncthing Home Directory, used to get API Key and Target")
	igCert  = flag.Bool("ignore_cert_errors", false, "ignore https/ssl/tls cert errors")
	verFlag = flag.Bool("version", false, "print version")
	GitTag  string
)

type SyncFolder struct {
	Name   string  `json:"folderName"`
	Status string  `json:"status"`
	Sync   float64 `json:"syncPercentDone"`
	Global uint64  `json:"globalBytes"`
	Local  uint64  `json:"localBytes"`
	Needs  uint64  `json:"missingBytes"`
}

type SyncDevice struct {
	Name     string  `json:"deviceName"`
	Status   string  `json:"status"`
	Sync     float64 `json:"syncPercentDone"`
	Download uint64  `json:"downloadedBytes"`
	Upload   uint64  `json:"uploadedBytes"`
	Needs    uint64  `json:"missingBytes"`
}

func dash() error {
	dumpErrors(true)

	cfg, err := api.GetConfig()
	if err != nil {
		return err
	}

	st, err := api.GetSysStatus()
	if err != nil {
		return err
	}

	sv, err := api.GetSysVersion()
	if err != nil {
		return err
	}

	myName := ""
	for _, n := range cfg.Devices {
		if n.DeviceID != st.MyID {
			continue
		}
		myName = n.Name
	}
	if myName == "" {
		return fmt.Errorf("unable to find this device name")
	}

	cons, err := api.GetConnection()
	if err != nil {
		return err
	}

	t := tabwriter.NewWriter(os.Stdout, 9, 0, 2, ' ', tabwriter.TabIndent)

	fmt.Fprintf(t, "Host\tUptime\tVersion\n")
	fmt.Fprintf(t, "%v\t%v\t%v\n",
		myName,
		durafmt.ParseShort(time.Duration(st.Uptime*1000000000)),
		sv.Version,
	)

	fmt.Fprintf(t, "\nFolder\tStatus\tSync\tGlobal\tLocal\tNeeds\n")

	for _, f := range cfg.Folders {
		fs, err := api.GetFolderStatus(f.ID)
		if err != nil {
			return err
		}
		co, err := api.GetCompletion("folder=" + f.ID)
		if err != nil {
			return err
		}
		fmt.Fprintf(t, "%v\t%v\t%5.1f%%\t%v\t%v\t%v\n",
			f.Label,
			fStatus(f.Paused, f.Type, fs.State, fs.Errors, fs.ReceiveOnlyTotalItems, fs.NeedTotalItems),
			co.Completion,
			humanize.Bytes(fs.GlobalBytes),
			humanize.Bytes(fs.LocalBytes),
			humanize.Bytes(fs.NeedBytes),
		)
	}

	t.Flush()

	fmt.Fprintf(t, "\nDevice\tStatus\tSync\tDownload\tUpload\tNeeds\n")

	for _, d := range cfg.Devices {
		co, err := api.GetCompletion("device=" + d.DeviceID)
		if err != nil {
			return err
		}

		if d.Name == myName {
			d.Name = "*" + d.Name
		}

		fmt.Fprintf(t, "%v\t%v\t%5.1f%%\t%v\t%v\t%v\n",
			d.Name,
			isConn(d.Paused, cons[d.DeviceID].Connected, d.DeviceID, st.MyID),
			co.Completion,
			humanize.Bytes(cons[d.DeviceID].InBytesTotal),
			humanize.Bytes(cons[d.DeviceID].OutBytesTotal),
			humanize.Bytes(co.NeedBytes),
		)
	}

	t.Flush()

	return nil
}

func getFolderInfoAsStruct(cfg api.StConfig) ([]SyncFolder, error) {
	dumpErrors(true)

	st, err := api.GetSysStatus()
	if err != nil {
		return nil, err
	}

	myName := ""
	for _, n := range cfg.Devices {
		if n.DeviceID != st.MyID {
			continue
		}
		myName = n.Name
	}
	if myName == "" {
		return nil, fmt.Errorf("unable to find this device name")
	}

	folders := []SyncFolder{}

	for _, f := range cfg.Folders {
		fs, err := api.GetFolderStatus(f.ID)
		if err != nil {
			return nil, err
		}
		co, err := api.GetCompletion("folder=" + f.ID)
		if err != nil {
			return nil, err
		}
		folders = append(folders,
			SyncFolder{
				Name:   f.Label,
				Status: fStatus(f.Paused, f.Type, fs.State, fs.Errors, fs.ReceiveOnlyTotalItems, fs.NeedTotalItems),
				Sync:   co.Completion,
				Global: fs.GlobalBytes,
				Local:  fs.LocalBytes,
				Needs:  fs.NeedBytes,
			})
	}

	return folders, nil
}

func getDeviceInfoAsStruct(cfg api.StConfig) ([]SyncDevice, error) {

	st, err := api.GetSysStatus()
	if err != nil {
		return nil, err
	}

	myName := ""
	for _, n := range cfg.Devices {
		if n.DeviceID != st.MyID {
			continue
		}
		myName = n.Name
	}
	if myName == "" {
		return nil, fmt.Errorf("unable to find this device name")
	}

	cons, err := api.GetConnection()
	if err != nil {
		return nil, err
	}

	devices := []SyncDevice{}

	for _, d := range cfg.Devices {
		co, err := api.GetCompletion("device=" + d.DeviceID)
		if err != nil {
			return nil, err
		}

		if d.Name == myName {
			d.Name = "*" + d.Name
		}
		devices = append(devices,
			SyncDevice{
				Name:     d.Name,
				Status:   isConn(d.Paused, cons[d.DeviceID].Connected, d.DeviceID, st.MyID),
				Sync:     co.Completion,
				Download: cons[d.DeviceID].InBytesTotal,
				Upload:   cons[d.DeviceID].OutBytesTotal,
				Needs:    co.NeedBytes,
			})
	}

	return devices, nil
}

func dumpDashAsJson() error {

	cfg, err := api.GetConfig()
	if err != nil {
		return err
	}

	devices, err := getDeviceInfoAsStruct(cfg)
	if err != nil {
		return err
	}
	folders, err := getFolderInfoAsStruct(cfg)
	if err != nil {
		return err
	}

	output := struct {
		Folders []SyncFolder `json:"folders"`
		Devices []SyncDevice `json:"devices"`
	}{
		Folders: folders,
		Devices: devices,
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	fmt.Println(string(jsonData))
	return nil
}

func dumpLogTxt() error {
	s, err := api.GetLogTxt()
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func dumpErrors(eLn bool) error {
	e, err := api.GetSysErrors()
	if err != nil {
		return err
	}
	for _, er := range e.Errors {
		fmt.Println(er.When, er.Message)
	}
	if eLn && len(e.Errors) > 0 {
		fmt.Println()
	}
	return nil
}

func dumpMyID() error {
	st, err := api.GetSysStatus()
	if err != nil {
		return err
	}
	fmt.Println(st.MyID)
	return nil
}

func rescan(fName string) error {
	if fName == "all" {
		return api.Rescan("")
	}
	fID, err := folderID(fName)
	if err != nil {
		return err
	}
	if fID == "" {
		return fmt.Errorf("folder %q not found, use 'all' to rescan all folders", fName)
	}
	return api.Rescan(fID)
}

func override(fName string) error {
	fID, err := folderID(fName)
	if err != nil {
		return err
	}
	if fID == "" {
		return fmt.Errorf("folder %q not found", fName)
	}
	return api.Override(fID)
}

func revert(fName string) error {
	fID, err := folderID(fName)
	if err != nil {
		return err
	}
	if fID == "" {
		return fmt.Errorf("folder %q not found", fName)
	}
	return api.Revert(fID)
}

func folderErrors(fName string) error {
	fID, err := folderID(fName)
	if err != nil {
		return err
	}
	if fID == "" {
		return fmt.Errorf("folder %q not found", fName)
	}
	fe, err := api.GetFolderErrors(fID)
	if err != nil {
		return err
	}

	for _, e := range fe.Errors {
		fmt.Printf("Error: %v : %v\n", e.Path, e.Error)
	}
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Usage = usage
	flag.Parse()
	if *verFlag {
		printVer()
		os.Exit(0)
	}

	a, t, err := cfg(*apiKey, *target, *homeDir)
	if err != nil {
		log.Fatal("apikey and target flags not specified, config file: ", err)
	}

	err = api.SetApiKeyTarget(a, t)
	if err != nil {
		log.Fatal(err)
	}

	if *igCert {
		api.IgnoreCertErrors()
	}

	switch flag.Arg(0) {
	case "log":
		err = dumpLogTxt()
	case "shutdown":
		err = api.Shutdown()
	case "restart":
		err = api.Restart()
	case "reset_db":
		err = api.ResetDB()
	case "errors":
		err = dumpErrors(false)
	case "clear_errors":
		err = api.ClearErrors()
	case "post_error":
		err = api.PostError(flag.Arg(1))
	case "folder_errors":
		err = folderErrors(flag.Arg(1))
	case "id":
		err = dumpMyID()
	case "rescan":
		err = rescan(flag.Arg(1))
	case "override":
		err = override(flag.Arg(1))
	case "revert":
		err = revert(flag.Arg(1))
	case "folder_pause":
		err = api.PauseFolder(flag.Arg(1), true)
	case "folder_resume":
		err = api.PauseFolder(flag.Arg(1), false)
	case "json_dump":
		err = dumpDashAsJson()
	default:
		err = dash()
	}

	if err != nil {
		log.Fatal(err)
	}
}
