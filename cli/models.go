package cli

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

type AppSecret struct {
	Name string 		`json:"name"`
	Version string		`json:"version"`
	Value string		`json:"value"`
}

type AbraAppService struct {
	Service string	`json:"service"`
	Chaos bool		`json:"chaos"`
	Created string	`json:"created"`
	Image string	`json:"image"`
	Ports string	`json:"ports"`
	State string	`json:"state"`
	Status string	`json:"status"`
	Version string	`json:"version"`
}

// type RecipeMeta struct {
// 	Category      string         `json:"category"`
// 	DefaultBranch string         `json:"default_branch"`
// 	Description   string         `json:"description"`
// 	Features      Features       `json:"features"`
// 	Icon          string         `json:"icon"`
// 	Name          string         `json:"name"`
// 	Repository    string         `json:"repository"`
// 	SSHURL        string         `json:"ssh_url"`
// 	Versions      RecipeVersions `json:"versions"`
// 	Website       string         `json:"website"`
// }

// type RecipeCatalogue map[Name]RecipeMeta
