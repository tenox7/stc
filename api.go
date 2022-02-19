package main

import "encoding/json"

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
type sysConn map[string]struct {
	Connected     bool   `json:"connected"`
	InBytesTotal  uint64 `json:"inBytesTotal"`
	OutBytesTotal uint64 `json:"outBytesTotal"`
}

type sysConnections struct {
	Connections sysConn
}

type sysStatus struct {
	CpuPercent float64 `json:"cpuPercent"`
	MyID       string  `json:"myID"`
	Uptime     int64   `json:"uptime"`
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

func getConfig() (stConfig, error) {
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

func getFolderStatus(f string) (dbStatus, error) {
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

func getCompletion(d string) (dbCompletion, error) {
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

func getConnection() (sysConn, error) {
	r, err := c.R().SetHeader("X-API-Key", *apiKey).Get(*target + "/rest/system/connections")
	if err != nil {
		return nil, err
	}

	co := sysConnections{}
	err = json.Unmarshal(r.Body(), &co)
	if err != nil {
		return nil, err
	}

	return co.Connections, nil
}

func getSysStatus() (sysStatus, error) {
	r, err := c.R().SetHeader("X-API-Key", *apiKey).Get(*target + "/rest/system/status")
	if err != nil {
		return sysStatus{}, err
	}

	st := sysStatus{}
	err = json.Unmarshal(r.Body(), &st)
	if err != nil {
		return sysStatus{}, err
	}

	return st, nil
}
