/*
 * Copyright 2021 Google LLC
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

func dash() error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}

	t := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintf(t, "Folder\tPaused\tState\tGlobal\tLocal\n")

	for _, f := range cfg.Folders {
		st, err := getFolderStatus(f.ID)
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

	cons, err := getConnection()
	if err != nil {
		return err
	}
	fmt.Fprintf(t, "\nDevice\tPaused\tConnected\tCompletion\tDownload\tUpload\n")

	for _, d := range cfg.Devices {
		st, err := getCompletion(d.DeviceID)
		if err != nil {
			return err
		}

		fmt.Fprintf(t, "%v\t%v\t%v\t%5.1f%%\t%v\t%v\n",
			d.Name,
			d.Paused,
			cons[d.DeviceID].Connected,
			st.Completion,
			humanize.Bytes(cons[d.DeviceID].InBytesTotal),
			humanize.Bytes(cons[d.DeviceID].OutBytesTotal),
		)
	}

	t.Flush()

	return nil
}

func main() {
	flag.Parse()

	err := dash()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
