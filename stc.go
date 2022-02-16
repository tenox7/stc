// sync thing client
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"text/tabwriter"

	humanize "github.com/dustin/go-humanize"
	"github.com/go-resty/resty/v2"
)

var (
	apiKey = flag.String("api_key", "", "Syncthing API Key") // TODO: also check env var
	target = flag.String("target", "http://127.0.0.1:8384", "Syncthing Target")

	c = resty.New()
)

type stConfig struct {
	Folders []struct {
		ID     string `json:"id"`
		Label  string `json:"label"`
		Paused bool   `json:"paused"`
	} `json:"folders"`

	Devices []struct {
		DeviceID string `json:"deviceID"`
		Name     string `json:"name"`
		Paused   bool   `json:"paused"`
	}
}

type dbStatus struct {
	GlobalBytes uint64 `json:"globalBytes"`
	GlobalFiles uint64 `json:"globalFiles"`
	LocalBytes  uint64 `json:"localBytes"`
	LocalFiles  uint64 `json:"localFiles"`
	State       string `json:"state"`
}

type dbCompletion struct {
	Completion float64 `json:"completion"`
}

func config() (stConfig, error) {
	r, err := c.R().SetHeader("X-API-Key", *apiKey).Get(*target + "/rest/config")
	if err != nil {
		return stConfig{}, err
	}

	cfg := stConfig{}
	err = json.Unmarshal(r.Body(), &cfg)
	if err != nil {
		return stConfig{}, err
	}

	return cfg, nil
}

func folderStatus(f string) (dbStatus, error) {
	r, err := c.R().SetHeader("X-API-Key", *apiKey).SetQueryString("folder=" + f).Get(*target + "/rest/db/status")
	if err != nil {
		return dbStatus{}, err
	}

	dbs := dbStatus{}
	err = json.Unmarshal(r.Body(), &dbs)
	if err != nil {
		return dbStatus{}, err
	}

	return dbs, nil
}

func completion(d string) (dbCompletion, error) {
	r, err := c.R().SetHeader("X-API-Key", *apiKey).SetQueryString("device=" + d).Get(*target + "/rest/db/completion")
	if err != nil {
		return dbCompletion{}, err
	}

	dbc := dbCompletion{}
	err = json.Unmarshal(r.Body(), &dbc)
	if err != nil {
		return dbCompletion{}, err
	}

	return dbc, nil
}

func connection() (map[string]bool, error) {
	r, err := c.R().SetHeader("X-API-Key", *apiKey).Get(*target + "/rest/system/connections")
	if err != nil {
		return nil, err
	}

	var ci map[string]interface{}
	err = json.Unmarshal(r.Body(), &ci)
	if err != nil {
		return nil, err
	}
	ccon := make(map[string]bool)
	for i, j := range ci["connections"].(map[string]interface{}) {
		//fmt.Printf(">> %v = %v\n", i, j.(map[string]interface{})["connected"])
		ccon[i] = j.(map[string]interface{})["connected"].(bool)
	}
	return ccon, nil
}

func status() error {
	cfg, err := config()
	if err != nil {
		return err
	}

	t := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintf(t, "Folder\tPaused\tState\tGlobal\tLocal\n")

	for _, f := range cfg.Folders {
		st, err := folderStatus(f.ID)
		if err != nil {
			return err
		}
		fmt.Fprintf(t, "%v\t%v\t%v\t%v\t%v\n",
			f.Label,
			f.Paused,
			st.State,
			humanize.Bytes(st.GlobalBytes),
			humanize.Bytes(st.LocalBytes),
		)
	}

	t.Flush()

	cons, err := connection()
	if err != nil {
		return err
	}
	fmt.Fprintf(t, "\nDevice\tPaused\tConnected\tCompletion\n")

	for _, d := range cfg.Devices {
		st, err := completion(d.DeviceID)
		if err != nil {
			return err
		}

		fmt.Fprintf(t, "%v\t%v\t%v\t%5.1f%%\n",
			d.Name,
			d.Paused,
			cons[d.DeviceID],
			st.Completion,
		)
	}

	t.Flush()

	return nil
}

func main() {
	flag.Parse()

	err := status()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
