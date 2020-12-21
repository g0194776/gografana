package gografana

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const grafanaOK string = "success"

type GrafanaClient_5_0 struct {
	basicAddress  string
	client        *http.Client
	authenticator Authenticator
}

type NewDashboardError struct {
	Err    error
	Status string
}

func (e NewDashboardError) Error() string {
	return fmt.Sprintf("Internal Error: %s, Status: %s", e.Err.Error(), e.Status)
}

func (gc *GrafanaClient_5_0) initClient() {
	if gc.client != nil {
		return
	}
	gc.client = &http.Client{}
}

func (gc *GrafanaClient_5_0) GetAllDashboards() ([]Board, error) {
	urlPath := fmt.Sprintf("%s/api/search?type=dash-db", gc.basicAddress)
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}
	bodyData, err := gc.getHTTPResponse(req, "GetAllDashboards(api/search?type=dash-db)")
	if err != nil {
		return nil, err
	}
	var boards []Board
	err = json.Unmarshal(bodyData, &boards)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal response body failed while calling to API GetAllDashboards(api/search?type=dash-db), error: %s", err.Error())
	}
	return boards, nil
}


func (gc *GrafanaClient_5_0) GetDashboardsByTitleAndFolderId(title string, folderId int) ([]Board, error) {
	urlPath := fmt.Sprintf("%s/api/search?query=%s&folderIds=%s", gc.basicAddress, title, strconv.Itoa(folderId))
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}
	bodyData, err := gc.getHTTPResponse(req, "GetDashboardsByTitleAndFolderId(api/search?query=&folderIds=)")
	if err != nil {
		return nil, err
	}
	var boards []Board
	err = json.Unmarshal(bodyData, &boards)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal response body failed while calling to API GetDashboardsByTitleAndFolderId(api/search?query=&folerIds=), error: %s", err.Error())
	}
	return boards, nil
}

func (gc *GrafanaClient_5_0) GetDashboardsByFolderId(folderId int) ([]Board, error) {
	urlPath := fmt.Sprintf("%s/api/search?folderIds=%d", gc.basicAddress, folderId)
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}
	bodyData, err := gc.getHTTPResponse(req, "GetAllDashboards(api/search?type=folderIds)")
	if err != nil {
		return nil, err
	}
	var boards []Board
	err = json.Unmarshal(bodyData, &boards)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal response body failed while calling to API GetAllDashboards(api/search?type=folderIds), error: %s", err.Error())
	}
	return boards, nil
}

func (gc *GrafanaClient_5_0) IsBoardExists(title string) (bool, *Board, error) {
	boards, err := gc.GetAllDashboards()
	if err != nil {
		return false, nil, err
	}
	for _, v := range boards {
		if v.Title == title {
			return true, &v, nil
		}
	}
	return false, nil, nil
}

func (gc *GrafanaClient_5_0) NewDashboard(board *Board, folderId uint, overwrite bool) (*Board, error) {
	if board.Timezone == "" {
		board.Timezone = "browser"
	}
	bodyReq := CreateDashboardRequest{Board: *board, Overwrite: overwrite, FolderId: folderId}
	bodyStr, err := json.Marshal(bodyReq)
	if err != nil {
		return board, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/dashboards/db", gc.basicAddress), strings.NewReader(string(bodyStr)))
	if err != nil {
		return board, err
	}
	rspBody, err := gc.getHTTPResponse(req, "NewDashboard(api/dashboards/db)")
	if err != nil {
		return board, err
	}
	var rsp CreateDashboardResponse
	err = json.Unmarshal(rspBody, &rsp)
	if err != nil {
		return board, fmt.Errorf("Unmarshal response body failed while calling to API NewDashboard(api/dashboards/db), error: %s", err.Error())
	}
	if rsp.Status != grafanaOK {
		return board, &NewDashboardError{Err: fmt.Errorf("Grafana operation failed while calling to API NewDashboard(api/dashboards/db), error: %s", rsp.Message), Status: rsp.Status}

	}
	board.ID = rsp.ID
	board.UID = rsp.UID
	board.Url = rsp.Url //TODO: ??
	board.Version = rsp.Version
	return board, nil
}

func (gc *GrafanaClient_5_0) CreateAPIKey(name string, role string, secondsToLive int) (string, error) {
	body := map[string]interface{}{
		"name":          name,
		"role":          role,
		"secondsToLive": secondsToLive,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/auth/keys", gc.basicAddress), strings.NewReader(string(b)))
	rspBody, err := gc.getHTTPResponse(req, "CreateAPIKey(api/auth/keys)")
	if err != nil {
		return "", err
	}

	var rsp CreateAPIKeyResponse
	err = json.Unmarshal(rspBody, &rsp)
	if err != nil {
		return "", fmt.Errorf("unmarshal response body failed while calling to API CreateAPIKey(api/auth/keys), error: %s", err.Error())
	}

	return rsp.Key, nil
}

func (gc *GrafanaClient_5_0) FindAllAPIKeys() ([]APIKey, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/auth/keys", gc.basicAddress), nil)
	rspBody, err := gc.getHTTPResponse(req, "GetAllAPIKeys(api/auth/keys)")
	if err != nil {
		return nil, err
	}
	var apiKeys []APIKey
	err = json.Unmarshal(rspBody, &apiKeys)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal response body failed while calling to API GetAllAPIKeys(api/auth/keys), error: %s", err.Error())
	}
	return apiKeys, nil
}

func (gc *GrafanaClient_5_0) DeleteAPIKey(id int) (bool, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/auth/keys/%d", gc.basicAddress, id), nil)
	if err != nil {
		return false, err
	}
	_, err = gc.getHTTPResponse(req, "DeleteAPIKey(api/auth/keys/[ID])")
	if err != nil {
		return false, err
	}
	return true, nil
}

// Status Codes:
//-------------------
// 200 – Deleted
// 401 – Unauthorized
// 403 – Access denied
// 404 – Not found
func (gc *GrafanaClient_5_0) DeleteDashboard(uid string) (bool, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/dashboards/uid/%s", gc.basicAddress, uid), nil)
	if err != nil {
		return false, err
	}
	_, err = gc.getHTTPResponse(req, "DeleteDashboard(/api/dashboards/uid/[UID])")
	if err != nil {
		return false, err
	}
	return true, nil
}

// Status Codes:
//-------------------
// 200 – Found
// 401 – Unauthorized
// 403 – Access denied
// 404 – Not found
func (gc *GrafanaClient_5_0) GetDashboardDetails(uid string) (*Board, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/dashboards/uid/%s", gc.basicAddress, uid), nil)
	if err != nil {
		return nil, err
	}
	bodyData, err := gc.getHTTPResponse(req, "GetDashboardDetails(/api/dashboards/uid/[UID])")
	if err != nil {
		return nil, err
	}
	var rsp GetDashboardByUIdResponse
	err = json.Unmarshal(bodyData, &rsp)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal response body failed while calling to API GetDashboardDetails(/api/dashboards/uid/[UID]), error: %s", err.Error())
	}
	return &rsp.Dashboard, nil
}

func (gc *GrafanaClient_5_0) EnsureFolderExists(folderId int, uid, title string) (int, bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/folders/id/%d", gc.basicAddress, folderId), nil)
	if err != nil {
		return -1, false, err
	}
	_, statusCode, err := gc.getHTTPResponseWithStatusCode(req, "GetFolderById(/api/folders/id/[FOLDER-ID])")
	if err != nil {
		return -1, false, err
	}
	//folder existed.
	if statusCode == 200 {
		return -1, false, nil
	}
	//try to create a new folder.
	bodyReq := CreateFolderRequest{UID: uid, Title: title}
	bodyStr, err := json.Marshal(bodyReq)
	if err != nil {
		return -1, false, err
	}
	var bodyData []byte
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/folders", gc.basicAddress), strings.NewReader(string(bodyStr)))
	if err != nil {
		return -1, false, err
	}
	bodyData, err = gc.getHTTPResponse(req, "CreateFolder(/api/folders)")
	if err != nil {
		return -1, false, err
	}
	var rsp CreateFolderResponse
	err = json.Unmarshal(bodyData, &rsp)
	if err != nil {
		return -1, false, fmt.Errorf("Unmarshal response body failed while calling to API CreateFolder(/api/folders), error: %s", err.Error())
	}
	return rsp.ID, true, nil
}

func (gc *GrafanaClient_5_0) getHTTPResponse(req *http.Request, flag string) ([]byte, error) {
	gc.initClient()
	//加入统一授权
	gc.authenticator.SetAuthentication(req)
	req.Header.Add("Content-Type", "application/json")
	rsp, err := gc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	bodyData, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read response body failed while calling to API %s, error: %s", flag, err.Error())
	}
	if rsp.StatusCode != 200 {
		return nil, fmt.Errorf("Remote API returned Non 200/OK status code in the %s response(%d), body: %s", flag, rsp.StatusCode, string(bodyData))
	}
	return bodyData, nil
}

func (gc *GrafanaClient_5_0) getHTTPResponseWithStatusCode(req *http.Request, flag string) ([]byte, int, error) {
	gc.initClient()
	//加入统一授权
	gc.authenticator.SetAuthentication(req)
	req.Header.Add("Content-Type", "application/json")
	rsp, err := gc.client.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer rsp.Body.Close()
	bodyData, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("Read response body failed while calling to API %s, error: %s", flag, err.Error())
	}
	if rsp.StatusCode == 404 {
		return nil, 404, nil
	}
	if rsp.StatusCode != 200 {
		return nil, rsp.StatusCode, fmt.Errorf("Remote API returned Non 200/OK status code in the %s response(%d), body: %s", flag, rsp.StatusCode, string(bodyData))
	}
	return bodyData, rsp.StatusCode, nil
}

func (gc *GrafanaClient_5_0) GetAllDataSources() ([]*DataSource, error) {
	urlPath := fmt.Sprintf("%s/api/datasources", gc.basicAddress)
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}
	bodyData, err := gc.getHTTPResponse(req, "GetAllDataSources(api/datasources)")
	if err != nil {
		return nil, err
	}
	var ds []*DataSource
	err = json.Unmarshal(bodyData, &ds)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal response body failed while calling to API GetAllDataSources(api/datasources), error: %s", err.Error())
	}
	return ds, nil
}

func (gc *GrafanaClient_5_0) GetDashSourceById(id int) (*DataSource, error) {
	urlPath := fmt.Sprintf("%s/api/datasources/%d", gc.basicAddress, id)
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}
	bodyData, err := gc.getHTTPResponse(req, "GetDashSourceById(api/datasources)")
	if err != nil {
		return nil, err
	}
	var ds DataSource
	err = json.Unmarshal(bodyData, &ds)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal response body failed while calling to API GetDashSourceById(api/datasources), error: %s", err.Error())
	}
	return &ds, nil
}

func (gc *GrafanaClient_5_0) DeleteDashSource(id int) error {
	urlPath := fmt.Sprintf("%s/api/datasources/%d", gc.basicAddress, id)
	req, err := http.NewRequest("DELETE", urlPath, nil)
	if err != nil {
		return err
	}
	_, err = gc.getHTTPResponse(req, "DeleteDashSource(api/datasources)")
	return err
}

func (gc *GrafanaClient_5_0) CreateDashSource(ds *DataSource) error {
	bodyStr, err := json.Marshal(ds)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/datasources", gc.basicAddress), strings.NewReader(string(bodyStr)))
	if err != nil {
		return err
	}
	rspBody, statusCode, err := gc.getHTTPResponseWithStatusCode(req, "CreateDashSource(api/datasources)")
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return fmt.Errorf("HTTP Response != 200 while calling to API CreateDashSource(api/datasources), error: %s", err.Error())
	}
	var rsp CreateDataSourceResponse
	err = json.Unmarshal(rspBody, &rsp)
	if err != nil {
		return fmt.Errorf("Unmarshal response body failed while calling to API CreateDashSource(api/datasources), error: %s", err.Error())
	}
	ds.ID = rsp.ID
	return nil
}

func (gc *GrafanaClient_5_0) GetAllFolders() ([]Folder, error) {
	urlPath := fmt.Sprintf("%s/api/folders?limit=10000", gc.basicAddress)
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}
	bodyData, err := gc.getHTTPResponse(req, "GetAllFolders(api/folders?limit=10000)")
	if err != nil {
		return nil, err
	}
	var folders []Folder
	err = json.Unmarshal(bodyData, &folders)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal response body failed while calling to API GetAllFolders(api/folders?limit=10000), error: %s", err.Error())
	}
	return folders, nil
}

func (gc *GrafanaClient_5_0) GetAllNotificationChannels() ([]NotificationChannel, error) {
	urlPath := fmt.Sprintf("%s/api/alert-notifications", gc.basicAddress)
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}
	bodyData, err := gc.getHTTPResponse(req, "GetAllNotificationChannels(api/alert-notifications)")
	if err != nil {
		return nil, err
	}
	var channels []NotificationChannel
	err = json.Unmarshal(bodyData, &channels)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal response body failed while calling to API GetAllNotificationChannels(api/alert-notifications), error: %s", err.Error())
	}
	return channels, nil
}

func (gc *GrafanaClient_5_0) CreateNotificationChannel(nc *NotificationChannel) error {
	bodyStr, err := json.Marshal(nc)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/alert-notifications", gc.basicAddress), strings.NewReader(string(bodyStr)))
	if err != nil {
		return err
	}
	rspBody, statusCode, err := gc.getHTTPResponseWithStatusCode(req, "CreateNotificationChannel(api/alert-notifications)")
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return fmt.Errorf("HTTP Response != 200 while calling to API CreateNotificationChannel(api/alert-notifications), error: %s", err.Error())
	}
	var rsp CreateNotificationChannelResponse
	err = json.Unmarshal(rspBody, &rsp)
	if err != nil {
		return fmt.Errorf("Unmarshal response body failed while calling to API CreateNotificationChannel(api/alert-notifications), error: %s", err.Error())
	}
	nc.ID = rsp.ID
	return nil
}
