package api

// Represents an App, follows the format of 
// https://git.coopcloud.tech/toolshed/coop-cloud-front
type AbraApp struct {
    Server        string `json:"server"`
    Recipe        string `json:"recipe`
	AppName       string `json:"appName"`
	Domain        string `json:"domain"`
	Status        string `json:"status"`
	Chaos         string `json:"chaos"`
	ChaosVersion  string `json:"chaosVersion"`
	Version       string `json:"version"`
	Upgrade       string `json:"upgrade"`
}

type AbraServer struct {
    Name string `json:"name"`
    Host string `json:"host"`
}

type ServerAppsResponse map[string]ServerApps

type ServerApps struct {
    Apps              []AbraApp `json:"apps"`
    AppCount          int       `json:"appCount"`
    VersionCount      int       `json:"versionCount"`
    UnversionedCount  int       `json:"unversionedCount"`
    LatestCount       int       `json:"latestCount"`
    UpgradeCount      int       `json:"upgradeCount"`
}

type DeployState struct {
	AppName  string   `json:"appName"`
	Streams  []DeployStream `json:"streams"`
	Logs     []string `json:"logs"`
	Failed   bool     `json:"failed"`
	TimedOut bool     `json:"timedOut"`
	Quit     bool     `json:"quit"`
	Count    int      `json:"count"`
}

type DeployStream struct {
    Name string         `json:"Name"`
    id  string          `json:"id"`
    status   string     `json:"status"`
	retries  int        `json:"retries"`
	health   string     `json:"health"`
	rollback bool       `json:"rollback"`
}  
