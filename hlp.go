package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/tenox7/stc/api"
)

func cfg(apiKey, target, homeDir string) (string, string, error) {
	if apiKey == "" {
		apiKey = os.Getenv("APIKEY")
	}
	if apiKey != "" && target != "" {
		return apiKey, target, nil
	}

	if homeDir == "" {
		homeDir = filepath.Dir(os.Args[0])
	}

	var err error
	var f []byte
	f, err = ioutil.ReadFile(homeDir + string(os.PathSeparator) + "/config.xml")
	if err != nil {
		return "", "", err
	}

	x := struct {
		XMLName xml.Name `xml:"configuration"`
		GUI     struct {
			ApiKey  string `xml:"apikey,omitempty"`
			Address string `xml:"address" default:"127.0.0.1:8384"`
			UseTLS  bool   `xml:"tls,attr"`
		} `xml:"gui"`
	}{}

	err = xml.Unmarshal(f, &x)
	if err != nil {
		return "", "", err
	}

	p := "http://"
	if x.GUI.UseTLS {
		p = "https://"
	}

	return x.GUI.ApiKey, p + x.GUI.Address, nil
}

func isConn(paused, conn bool, ID, myID string) string {
	if ID == myID {
		return "Myself"
	}
	if paused {
		return "Paused"
	}
	if conn {
		return "OK"
	}
	return "Offline"
}

func fStatus(paused bool, status string, err, loChg uint64) string {
	if paused {
		return "Paused"
	}
	if err > 0 {
		return "Errors"
	}
	if loChg > 0 {
		return "LocAdds"
	}
	return status
}

func folderID(fName string) (string, error) {
	cfg, err := api.GetConfig()
	if err != nil {
		return "", err
	}
	fID := ""
	for _, f := range cfg.Folders {
		if f.Label != fName {
			continue
		}
		fID = f.ID
	}
	return fID, nil
}
