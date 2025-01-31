package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"

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
		Type   string `json:"type"`
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
	Ram    uint64 `json:"sys"`
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
	NeedTotalItems        uint64 `json:"needTotalItems"`
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

func apiError(e interface{}) error {
	pc, fi, li, _ := runtime.Caller(1)
	return fmt.Errorf("%v (%v:%v): %v", runtime.FuncForPC(pc).Name(), filepath.Base(fi), li, e)
}

func IgnoreCertErrors() {
	c.SetTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}})
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
		return StConfig{}, apiError(err)
	}
	if r.IsError() {
		return StConfig{}, apiError(r.Status())
	}

	cfg := StConfig{}
	err = json.Unmarshal(r.Body(), &cfg)
	if err != nil {
		return StConfig{}, apiError(err)
	}

	return cfg, nil
}

func GetFolderStatus(f string) (DbStatus, error) {
	r, err := c.R().SetQueryString("folder=" + f).Get("db/status")
	if err != nil {
		return DbStatus{}, apiError(err)
	}
	if r.IsError() {
		return DbStatus{}, apiError(r.Status())
	}

	dbs := DbStatus{}
	err = json.Unmarshal(r.Body(), &dbs)
	if err != nil {
		return DbStatus{}, apiError(err)
	}

	return dbs, nil
}

func PauseFolder(f string, p bool) error {
	r, err := c.R().SetBody(`{ "paused": ` + strconv.FormatBool(p) + `}`).Patch("config/folders/" + f)
	if err != nil {
		return apiError(err)
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func GetCompletion(qStr string) (DbCompletion, error) {
	r, err := c.R().SetQueryString(qStr).Get("db/completion")
	if err != nil {
		return DbCompletion{}, apiError(err)
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
		return DbCompletion{}, apiError(err)
	}

	return dbc, nil
}

func GetConnection() (SysConn, error) {
	r, err := c.R().Get("system/connections")
	if err != nil {
		return nil, apiError(err)
	}
	if r.IsError() {
		return nil, apiError(r.Status())
	}

	co := SysConnections{}
	err = json.Unmarshal(r.Body(), &co)
	if err != nil {
		return nil, apiError(err)
	}

	return co.Connections, nil
}

func GetSysStatus() (SysStatus, error) {
	r, err := c.R().Get("system/status")
	if err != nil {
		return SysStatus{}, apiError(err)
	}
	if r.IsError() {
		return SysStatus{}, apiError(r.Status())
	}

	st := SysStatus{}
	err = json.Unmarshal(r.Body(), &st)
	if err != nil {
		return SysStatus{}, apiError(err)
	}

	return st, nil
}

func GetSysVersion() (SysVersion, error) {
	r, err := c.R().Get("system/version")
	if err != nil {
		return SysVersion{}, apiError(err)
	}
	if r.IsError() {
		return SysVersion{}, apiError(r.Status())
	}

	ve := SysVersion{}
	err = json.Unmarshal(r.Body(), &ve)
	if err != nil {
		return SysVersion{}, apiError(err)
	}

	return ve, nil
}

func GetLogTxt() (string, error) {
	r, err := c.R().Get("system/log.txt")
	if err != nil {
		return "", apiError(err)
	}
	if r.IsError() {
		return "", apiError(r.Status())
	}
	return r.String(), nil
}

func Shutdown() error {
	r, err := c.R().Post("system/shutdown")
	if err != nil {
		return apiError(err)
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func Restart() error {
	r, err := c.R().Post("system/restart")
	if err != nil {
		return apiError(err)
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func ResetDB() error {
	r, err := c.R().Post("system/reset")
	if err != nil {
		return apiError(err)
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func GetSysErrors() (SysErrors, error) {
	r, err := c.R().Get("system/error")
	if err != nil {
		return SysErrors{}, apiError(err)
	}
	if r.IsError() {
		return SysErrors{}, apiError(r.Status())
	}

	se := SysErrors{}
	err = json.Unmarshal(r.Body(), &se)
	if err != nil {
		return SysErrors{}, apiError(err)
	}
	return se, nil
}

func ClearErrors() error {
	r, err := c.R().Post("system/error/clear")
	if err != nil {
		return apiError(err)
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func PostError(msg string) error {
	r, err := c.R().SetBody(msg).Post("system/error")
	if err != nil {
		return apiError(err)
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func GetFolderErrors(folderID string) (FolderErrors, error) {
	r, err := c.R().SetQueryString("folder=" + folderID).Get("folder/errors")
	if err != nil {
		return FolderErrors{}, apiError(err)
	}
	if r.IsError() {
		return FolderErrors{}, apiError(r.Status())
	}

	fe := FolderErrors{}
	err = json.Unmarshal(r.Body(), &fe)
	if err != nil {
		return FolderErrors{}, apiError(err)
	}
	return fe, nil
}

func Rescan(folderID string) error {
	r, err := c.R().SetQueryString("folder=" + folderID).Post("db/scan")
	if err != nil {
		return apiError(err)
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func Override(folderID string) error {
	r, err := c.R().SetQueryString("folder=" + folderID).Post("db/override")
	if err != nil {
		return apiError(err)
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func Revert(folderID string) error {
	r, err := c.R().SetQueryString("folder=" + folderID).Post("db/revert")
	if err != nil {
		return apiError(err)
	}
	if r.IsError() {
		return apiError(r.Status())
	}
	return nil
}

func Events(event_types string, limit int, since int) (string, error) {
	r, err := c.R().
		SetQueryString("events=" + event_types).
		SetQueryString(fmt.Sprintf("since=%d", since)).
		SetQueryString(fmt.Sprintf("limit=%d", limit)).
		Get("events")
	if err != nil {
		return "", apiError(err)
	}
	if r.IsError() {
		return "", apiError(r.Status())
	}

	return r.String(), nil
}
