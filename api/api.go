package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/go-resty/resty/v2"
)

var (
	c = resty.New()
)

type StConfig struct {
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

type SysConn map[string]struct {
	Connected     bool   `json:"connected"`
	InBytesTotal  uint64 `json:"inBytesTotal"`
	OutBytesTotal uint64 `json:"outBytesTotal"`
}

type SysConnections struct {
	Connections SysConn
}

type SysStatus struct {
	MyID   string `json:"myID"`
	Uptime int64  `json:"uptime"`
}

type SysVersion struct {
	Version string `json:"version"`
}

type DbStatus struct {
	GlobalBytes           uint64 `json:"globalBytes"`
	GlobalFiles           uint64 `json:"globalFiles"`
	LocalBytes            uint64 `json:"localBytes"`
	LocalFiles            uint64 `json:"localFiles"`
	NeedBytes             uint64 `json:"needBytes"`
	State                 string `json:"state"`
	Errors                uint64 `json:"errors"`
	ReceiveOnlyTotalItems uint64 `json:"receiveOnlyTotalItems"`
}

type DbCompletion struct {
	Completion float64 `json:"completion"`
	NeedBytes  uint64  `json:"needBytes"`
}

type SysErrors struct {
	Errors []struct {
		When    string `json:"when"`
		Message string `json:"message"`
	} `json:"errors"`
}

type FolderErrors struct {
	Errors []struct {
		Path  string `json:"path"`
		Error string `json:"error"`
	} `json:"errors"`
}

func IgnoreCertErrors() {
	c.SetTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}})
}

func apiError(e string) error {
	pc, _, _, _ := runtime.Caller(1)
	return fmt.Errorf("%s: %s", runtime.FuncForPC(pc).Name(), e)
}

func SetApiKeyTarget(a, t string) error {
	if a == "" || t == "" {
		return fmt.Errorf("apikey and target must be specified")
	}
	c.SetHeader("X-API-Key", a)
	c.SetBaseURL(t + "/rest/")
	return nil
}

func GetConfig() (StConfig, error) {
	r, err := c.R().Get("config")
	if err != nil {
		return StConfig{}, err
	}
	if r.IsError() {
		return StConfig{}, apiError(r.Status())
	}

	cfg := StConfig{}
	err = json.Unmarshal(r.Body(), &cfg)
	if err != nil {
		return StConfig{}, err
	}

	return cfg, nil
}

func GetFolderStatus(f string) (DbStatus, error) {
	r, err := c.R().SetQueryString("folder=" + f).Get("db/status")
	if err != nil {
		return DbStatus{}, err
	}
	if r.IsError() {
		return DbStatus{}, apiError(r.Status())
	}

	dbs := DbStatus{}
	err = json.Unmarshal(r.Body(), &dbs)
	if err != nil {
		return DbStatus{}, err
	}

	return dbs, nil
}

func GetCompletion(qStr string) (DbCompletion, error) {
	r, err := c.R().SetQueryString(qStr).Get("db/completion")
	if err != nil {
		return DbCompletion{}, err
	}
	if r.StatusCode() == 404 {
		return DbCompletion{}, nil
	}
	if r.IsError() {
		return DbCompletion{}, apiError(r.Status())
	}

	dbc := DbCompletion{}
	err = json.Unmarshal(r.Body(), &dbc)
	if err != nil {
		return DbCompletion{}, err
	}

	return dbc, nil
}

func GetConnection() (SysConn, error) {
	r, err := c.R().Get("system/connections")
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, apiError(r.Status())
	}

	co := SysConnections{}
	err = json.Unmarshal(r.Body(), &co)
	if err != nil {
		return nil, err
	}

	return co.Connections, nil
}

func GetSysStatus() (SysStatus, error) {
	r, err := c.R().Get("system/status")
	if err != nil {
		return SysStatus{}, err
	}
	if r.IsError() {
		return SysStatus{}, apiError(r.Status())
	}

	st := SysStatus{}
	err = json.Unmarshal(r.Body(), &st)
	if err != nil {
		return SysStatus{}, err
	}

	return st, nil
}

func GetSysVersion() (SysVersion, error) {
	r, err := c.R().Get("system/version")
	if err != nil {
		return SysVersion{}, err
	}
	if r.IsError() {
		return SysVersion{}, apiError(r.Status())
	}

	ve := SysVersion{}
	err = json.Unmarshal(r.Body(), &ve)
	if err != nil {
		return SysVersion{}, err
	}

	return ve, nil
}

func GetLogTxt() (string, error) {
	r, err := c.R().Get("system/log.txt")
	if err != nil {
		return "", err
	}
	if r.IsError() {
		return "", apiError(r.Status())
	}
	return r.String(), nil
}

func Shutdown() error {
	_, err := c.R().Post("system/shutdown")
	return err
}

func Restart() error {
	_, err := c.R().Post("system/restart")
	return err
}

func ResetDB() error {
	_, err := c.R().Post("system/reset")
	return err
}

func GetSysErrors() (SysErrors, error) {
	r, err := c.R().Get("system/error")
	if err != nil {
		return SysErrors{}, err
	}
	if r.IsError() {
		return SysErrors{}, apiError(r.Status())
	}

	se := SysErrors{}
	err = json.Unmarshal(r.Body(), &se)
	if err != nil {
		return SysErrors{}, err
	}
	return se, nil
}

func ClearErrors() error {
	_, err := c.R().Post("system/error/clear")
	return err
}

func PostError(msg string) error {
	_, err := c.R().SetBody(msg).Post("system/error")
	return err
}

func GetFolderErrors(folderID string) (FolderErrors, error) {
	r, err := c.R().SetQueryString("folder=" + folderID).Get("folder/errors")
	if err != nil {
		return FolderErrors{}, err
	}
	if r.IsError() {
		return FolderErrors{}, apiError(r.Status())
	}

	fe := FolderErrors{}
	err = json.Unmarshal(r.Body(), &fe)
	if err != nil {
		return FolderErrors{}, err
	}
	return fe, nil
}

func Rescan(folderID string) error {
	r, err := c.R().SetQueryString("folder=" + folderID).Post("db/scan")
	if err != nil {
		return nil
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func Override(folderID string) error {
	r, err := c.R().SetQueryString("folder=" + folderID).Post("db/override")
	if err != nil {
		return nil
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func Revert(folderID string) error {
	r, err := c.R().SetQueryString("folder=" + folderID).Post("db/revert")
	if err != nil {
		return nil
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}
