package api
import (
	"fmt"
	"log"
	"encoding/json"
	appPkg "coopcloud.tech/abra/pkg/app"
	"coop-cloud-backend/internal"
	"net/http"
)

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
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
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
			appT, err := appTransposeUndeployed(app)
			if err != nil {
				log.Printf("app transpose failed: %s\n", err)
				http.Error(w, fmt.Sprintf("app transpose failed: %s\n", err), http.StatusInternalServerError)
				return
			}
			serverApps.Apps = append(serverApps.Apps, appT)
			serverApps.AppCount += 1
			// assume these are true rn idk
			serverApps.VersionCount += 1
			serverApps.LatestCount += 1
		}
		abraAppResponse[app.Server] = serverApps
	}
	jsonBytes, err := json.Marshal(abraAppResponse)
	if err != nil {
		log.Printf("JSON conversion failed: %s\n", err)
		http.Error(w, fmt.Sprintf("JSON conversion failed: %s\n", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
func appTransposeUndeployed(app appPkg.App) (AbraApp, error) {
	config, err := app.Recipe.GetComposeConfig(app.Env)
	if err != nil {
		return AbraApp{}, err
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
	}, nil
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
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}
	serverNames, err := ReadServerNames()
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
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
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}