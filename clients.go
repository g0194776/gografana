package gografana

var (
	clients map[string]func(string, string) GrafanaClienter
)

//根据Grafana的版本号来获取指定的Client
func GetClientByVersion(version, apiAddress, apikey string) (GrafanaClienter, error) {
	if v, ok := clients[version]; ok {
		return v(apiAddress, apikey), nil
	}
	return nil, ErrNoSpecifiedVerClient{}
}

func init() {
	clients = make(map[string]func(string, string) GrafanaClienter)
	clients["5.x"] = func(apiAddress, token string) GrafanaClienter {
		return &GrafanaClient_5_0{basicAddress: apiAddress, token: token}
	}
}

type GrafanaClienter interface {
	GetAllDashboards() ([]Board, error)
	IsBoardExists(title string) (bool, *Board, error)
	NewDashboard(board *Board, folderId uint, overwrite bool) (*Board, error)
	DeleteDashboard(uid string) (bool, error)
	GetDashboardDetails(uid string) (*Board, error)
}
