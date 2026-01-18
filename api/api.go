package api
import (
	"fmt"
	"log"
	"encoding/json"
	appPkg "coopcloud.tech/abra/pkg/app"
	configPkg "coopcloud.tech/abra/pkg/config"
	composetypes "github.com/docker/cli/cli/compose/types"	
	"net/http"
)
type abraHandler struct{
	mux *http.ServeMux
}
func newAbraHandler() *abraHandler {
	h := &abraHandler{
		mux: http.NewServeMux(),
	}
	h.mux.HandleFunc("/api/abra/apps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return 
		}
		switch r.Method {
		case http.MethodGet: 
			h.handleListApps(w, r)
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/abra/servers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return 
		}
		switch r.Method {
		case http.MethodGet: 
			h.handleListServers(w, r)
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	return h
}
func StartAPI() {
	h := newAbraHandler()
	fmt.Println("Server started on port 3000")
	http.ListenAndServe(":3000", h)
}

func (h *abraHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Pre-processing: logging
	log.Printf("Incoming %s request: %s\n", r.Method, r.URL.Path)

	// Delegate to internal mux
	h.mux.ServeHTTP(w, r)
}
// func (h *appHandler) handleStartApp(w http.ResponseWriter, r *http.Request) {
// 	appName := r.PathValue("appName")
// 	w.Write([]byte("starting app: " + appName))

// }

func (h *abraHandler) handleListApps(w http.ResponseWriter, r *http.Request) {
	appNames, err := GetAppNames()
	if err != nil {
		log.Printf("Error getting app names: %s\n", err)
		InternalServerErrorHandler(w, r)
		return
	}
	// counts the number of apps in a server to initialize slices later
	serverAppCount := make(map[string]int)

	remoteApps := make([]appPkg.App, 0, len(appNames))
	for _, appName := range appNames {
		remoteApp, err := GetApp(appName)
		serverAppCount[remoteApp.Server] += 1
		if err != nil {
			log.Printf("Error getting app %s: %s\n", appName, err)
			InternalServerErrorHandler(w, r)
	 		return
		}
		remoteApps = append(remoteApps, remoteApp)
	}

	appStatuses, err := GetAppStatuses(remoteApps)
	if err != nil {
		log.Printf("GetAppStatuses Falied\n")
		fmt.Println("Error: ", err)
		InternalServerErrorHandler(w, r)
		return
	}
	abraAppResponse := ServerAppsResponse{}
	for _, app := range remoteApps {
		serverApps, ok := abraAppResponse[app.Server]
		if !ok { // create the slice
			// ever other field initializes to 0
			serverApps = ServerApps{
				Apps: make([]AbraApp, 0, serverAppCount[app.Server]),
			}
		}
		appInfo, ok := appStatuses[app.StackName()]
		if ok {
			log.Printf("app %s is deployed\n", app.Name)
			serverApps.Apps = append(serverApps.Apps, appTranspose(app, appInfo))
			serverApps.AppCount += 1
			// assume these are true rn idk
			serverApps.VersionCount += 1
			serverApps.LatestCount += 1
		} else {
			log.Printf("app %s is undeployed\n", app.Name)
			serverApps.Apps = append(serverApps.Apps, appTransposeUndeployed(app))
			serverApps.AppCount += 1
			// assume these are true rn idk
			serverApps.VersionCount += 1
			serverApps.LatestCount += 1
		}
		abraAppResponse[app.Server] = serverApps
	}
	jsonBytes, err := json.Marshal(abraAppResponse)
	if err != nil {
		log.Printf("Error converting to json: %s\n", err)
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
func appTransposeUndeployed(app appPkg.App) AbraApp {
	config, err := app.Recipe.GetComposeConfig(app.Env)
	if err != nil {
		log.Fatal(err)
	}
	version := GetLabel(config, app.StackName(), "version")
	if version == "" {
		version = "unknown"
	}
	return AbraApp{
		AppName: app.Name,
		Server:  app.Server,
		Recipe:  app.Recipe.Name,
		Domain:  app.Domain,
		Chaos: "false",
		Status: "undeployed",
		ChaosVersion: "unknown",
		Version: version,
		Upgrade: "latest",
	}
}
func appTranspose(app appPkg.App, psInfo map[string]string) AbraApp {
	return AbraApp{
		AppName: app.Name,
		Server:  app.Server,
		Recipe:  app.Recipe.Name,
		Domain:  app.Domain,
		Chaos: psInfo["chaos"],
		Status: psInfo["status"],
		ChaosVersion: getOrDefault(psInfo, "chaosVersion", "unknown"),
		Version: psInfo["version"],
		Upgrade: "latest",
	}
}

func getOrDefault(m map[string]string, key, def string) string {
    if v, ok := m[key]; ok {
        return v
    }
    return def
}

func (h *abraHandler) handleListServers (w http.ResponseWriter, r *http.Request) {
	servers, err := GetServers()
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	serverNames, err := ReadServerNames()
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	abraServers := make([]AbraServer, 0, len(servers))
	for i := range len(servers) {
		abraServers = append(abraServers, 
						   	AbraServer{
							   	Name: serverNames[i], 
							   	Host: servers[i],
						   	},
						  )
	}
	jsonBytes, err := json.Marshal(abraServers)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
func GetLabel(compose *composetypes.Config, stackName string, label string) string {
	return appPkg.GetLabel(compose, stackName, label)
}

func GetAppStatuses(apps []appPkg.App) (map[string]map[string]string, error) {
	return appPkg.GetAppStatuses(apps, true)
}
func GetApp(appName string) (appPkg.App, error) {
	return appPkg.Get(appName)
}

func GetAppNames() ([]string, error) {
	return appPkg.GetAppNames()
}

func GetAppServiceNames(appName string) ([]string, error) {
	return appPkg.GetAppServiceNames(appName)
}

func GetServers() ([]string, error) {
	return configPkg.GetServers()
}

func ReadServerNames() ([]string, error) {
	return configPkg.ReadServerNames()
}
