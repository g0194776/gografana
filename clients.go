package gografana

var (
	clients map[string]func(string, Authenticator) GrafanaClienter
)

//根据Grafana的版本号来获取指定的Client
func GetClientByVersion(version, apiAddress string, auth Authenticator) (GrafanaClienter, error) {
	if v, ok := clients[version]; ok {
		return v(apiAddress, auth), nil
	}
	return nil, ErrNoSpecifiedVerClient{}
}

func init() {
	clients = make(map[string]func(string, Authenticator) GrafanaClienter)
	clients["5.x"] = func(apiAddress string, auth Authenticator) GrafanaClienter {
		return &GrafanaClient_5_0{basicAddress: apiAddress, authenticator: auth}
	}
}

type GrafanaClienter interface {
	GetAllDashboards() ([]Board, error)
	GetAllFolders() ([]Folder, error)
	GetDashboardsByFolderId(folderId int) ([]Board, error)
	IsBoardExists(title string) (bool, *Board, error)
	NewDashboard(board *Board, folderId uint, overwrite bool) (*Board, error)
	DeleteDashboard(uid string) (bool, error)
	GetDashboardDetails(uid string) (*Board, error)
	EnsureFolderExists(folderId int, uid, title string) (int, bool, error)
	//DATA SOURCE
	GetAllDataSources() ([]*DataSource, error)
	GetDashSourceById(id int) (*DataSource, error)
	DeleteDashSource(id int) error
	CreateDashSource(ds *DataSource) error
	//NOTIFICATIONS
	GetAllNotificationChannels() ([]NotificationChannel, error)
	CreateNotificationChannel(nc *NotificationChannel) error
}
