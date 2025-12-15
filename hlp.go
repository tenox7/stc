// syncthing cli tool - helper functions
package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tenox7/stc/api"
)

func usage() {
	o := flag.CommandLine.Output()
	fmt.Fprintf(o, "stc [flags] [commands]\n\nflags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(o, "\ncommands:\n"+
		"  log            - print syncthing 'recent' log\n"+
		"  restart        - restart syncthing daemon\n"+
		"  shutdown       - shutdown syncthing daemon\n"+
		"  errors         - print errors visible in web UI\n"+
		"  clear_errors   - clear errors in the web UI\n"+
		"  post_error     - posts a custom error message in the web UI\n"+
		"  folder_errors  - prints folder errors from scan or pull\n"+
		"  folder_pause   - pause specified folder\n"+
		"  folder_resume  - unpause specified folder\n"+
		"  id             - print ID of this node\n"+
		"  reset_db       - reset the database / file index\n"+
		"  rescan         - rescan a folder or 'all'\n"+
		"  override       - override remote changed for a send-only folder (OoSync)\n"+
		"  revert         - revert local changes for a receive-only folder (LocAdds)\n"+
		"  events [types] - prints a json list of latest events, [types] is a comma-delimited list of events\n"+
		"                   see https://docs.syncthing.net/dev/events.html#event-types for a list of event types\n"+
		"  json_dump      - prints a json object with device and folder info, for easier parsing in scripts\n",
	)
}

func printVer() {
	fmt.Printf("stc version %v\n", GitTag)
}

// cfg determines the final API key and target URL.
// The order of precedence for each value is:
// 1. Command-line flag (e.g., -apikey, -target)
// 2. Environment variables
// 3. Value from config.xml
func cfg(apiKeyFlag, targetFlag, homeDirFlag string) (string, string, error) {
	// 1. Start by loading values from the config file to establish a baseline.
	apiKeyFromFile, targetFromFile, cfgErr := readConfigXML(homeDirFlag)

	// 2. Set the final values, starting with the baseline from the file.
	finalAPIKey := apiKeyFromFile
	finalTarget := targetFromFile

	// 3. Override with environment variables if they are set.
	if envKey := os.Getenv("STC_APIKEY"); envKey != "" {
		finalAPIKey = envKey
	} else if envKey := os.Getenv("APIKEY"); envKey != "" {
		finalAPIKey = envKey
	}
	if envTarget := os.Getenv("STC_TARGET"); envTarget != "" {
		finalTarget = envTarget
	}

	// 4. Override with command-line flags, as they have the highest precedence.
	if apiKeyFlag != "" {
		finalAPIKey = apiKeyFlag
	}
	if targetFlag != "" {
		finalTarget = targetFlag
	}

	// 5. After checking all sources, validate that we have the required values.
	if finalAPIKey == "" {
		if cfgErr != nil && !errors.Is(cfgErr, os.ErrNotExist) {
			return "", "", fmt.Errorf("API key not found and config file invalid: %w", cfgErr)
		}
		return "", "", fmt.Errorf("API key not found. Provide it via -apikey flag, STC_APIKEY env var, or config.xml")
	}

	if finalTarget == "" {
		if cfgErr != nil && !errors.Is(cfgErr, os.ErrNotExist) {
			return "", "", fmt.Errorf("target not found and config file invalid: %w", cfgErr)
		}
		return "", "", fmt.Errorf("target URL not found. Provide it via -target flag, STC_TARGET env var, or config.xml")
	}

	return finalAPIKey, finalTarget, nil
}

// readConfigXML finds and parses the Syncthing config.xml to extract the API
// key and target URL. It returns os.ErrNotExist if the file cannot be found.
func readConfigXML(homeDir string) (string, string, error) {
	cfgFile, err := findCfgFile(homeDir)
	if err != nil {
		// Pass the error up, which will likely be os.ErrNotExist
		return "", "", err
	}

	f, err := os.ReadFile(cfgFile)
	if err != nil {
		return "", "", fmt.Errorf("could not read %s: %w", cfgFile, err)
	}

	x := struct {
		XMLName xml.Name `xml:"configuration"`
		GUI     struct {
			ApiKey  string `xml:"apikey"`
			Address string `xml:"address"`
			UseTLS  bool   `xml:"tls,attr"`
		} `xml:"gui"`
	}{}

	if err = xml.Unmarshal(f, &x); err != nil {
		return "", "", fmt.Errorf("could not parse %s: %w", cfgFile, err)
	}

	// If address is empty in XML return empty
	if x.GUI.Address == "" {
		return x.GUI.ApiKey, "", nil
	}

	scheme := "http://"
	if x.GUI.UseTLS {
		scheme = "https://"
	}

	return x.GUI.ApiKey, scheme + x.GUI.Address, nil
}

// findCfgFile searches for config.xml in standard Syncthing locations.
// It returns a path and nil on success, or an empty path and an error on failure.
// It specifically returns os.ErrNotExist if the file isn't found in any searched location.
func findCfgFile(homeDir string) (string, error) {
	// If a specific home directory is provided, only check there.
	if homeDir != "" {
		path := filepath.Join(homeDir, "config.xml")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		return "", fmt.Errorf("config.xml not found in specified homedir %s: %w", homeDir, os.ErrNotExist)
	}

	// Create a list of potential directories to check.
	dirsToCheck := []string{}
	userHomeDir, err := os.UserHomeDir()
	if err == nil {
		// New default config dir (as of syncthing v1.29.3)
		dirsToCheck = append(dirsToCheck, filepath.Join(userHomeDir, ".local", "state", "syncthing"))
		// Try Windows Scoop persist
		dirsToCheck = append(dirsToCheck, filepath.Join(userHomeDir, "scoop", "persist", "syncthing", "config"))
	}
	userCfgDir, err := os.UserConfigDir()
	if err == nil {
		// Older default config dir
		dirsToCheck = append(dirsToCheck, filepath.Join(userCfgDir, "syncthing"))
	}
	userCacheDir, err := os.UserCacheDir()
	if err == nil {
		// try AppData/Local/syncthing (windows)
		dirsToCheck = append(dirsToCheck, filepath.Join(userCacheDir, "syncthing"))
	}
	execDir, err := os.Executable()
	if err == nil {
		// Fall back to path of executable
		dirsToCheck = append(dirsToCheck, filepath.Dir(execDir))
	}

	// Iterate through the potential directories and return the first match.
	for _, dir := range dirsToCheck {
		path := filepath.Join(dir, "config.xml")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("config.xml not found in any standard location: %w", os.ErrNotExist)
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

func fStatus(paused bool, ty, st string, err, loChg, needItms uint64) string {
	if paused {
		return "Paused"
	}
	if err > 0 {
		return "Errors"
	}
	if ty == "sendonly" && needItms > 0 {
		return "OoSync"
	}
	if ty == "receiveonly" && loChg > 0 {
		return "LocAdds"
	}
	return st
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
