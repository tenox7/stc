package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"
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

func fStatus(paused bool, status string) string {
	if paused {
		return "Paused"
	}
	return status
}
