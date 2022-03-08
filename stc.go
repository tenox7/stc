/*
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// syncthing cli tool
package main

import (
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
)

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
	flag.Usage = usage
	flag.Parse()

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
	default:
		err = dash()
	}

	if err != nil {
		log.Fatal(err)
	}
}
