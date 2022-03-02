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

	fmt.Fprintf(t, "\nFolder\tStatus\tSync\tGlobal\tLocal\tOoSync\n")

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
			fStatus(f.Paused, fs.State),
			co.Completion,
			humanize.Bytes(fs.GlobalBytes),
			humanize.Bytes(fs.LocalBytes),
			humanize.Bytes(fs.NeedBytes),
		)
	}

	t.Flush()

	fmt.Fprintf(t, "\nDevice\tStatus\tSync\tDownload\tUpload\tOoSync\n")

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
			isConn(d.Paused, cons[d.DeviceID].Connected),
			co.Completion,
			humanize.Bytes(cons[d.DeviceID].InBytesTotal),
			humanize.Bytes(cons[d.DeviceID].OutBytesTotal),
			humanize.Bytes(co.NeedBytes),
		)
	}

	t.Flush()

	return nil
}

func main() {
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

	err = dash()
	if err != nil {
		log.Fatal(err)
	}
}
