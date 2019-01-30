package gografana

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const grafanaOK string = "success"

type GrafanaClient_5_0 struct {
	basicAddress string
	token        string
	client       *http.Client
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
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", gc.token))
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
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", gc.token))
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
