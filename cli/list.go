package cli
import (
	"log"
	"net/http"
	"os/exec"
	"encoding/json"
	"strings"
	"sort"
	appPkg "coopcloud.tech/abra/pkg/app"
	"coopcloud.tech/tagcmp"

)

 func (h *abraHandler) handleListApps(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("abra", "app", "ls", "-m")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error: ", err)
		InternalServerErrorHandler(w, r)
		return
	}
	var resp ServerAppsResponse
	// no filtering
	listAppServer := ""
	recipeFilter := ""

	if err := json.Unmarshal(output, &resp); err != nil {
		http.Error(w, "invalid JSON from CLI", http.StatusInternalServerError)
		return
	}

	appFiles, err := appPkg.LoadAppFiles(listAppServer)
	if err != nil {
		log.Fatal(err)
	}

	apps, err := appPkg.GetApps(appFiles, recipeFilter)
	if err != nil {
		log.Fatal(err)
	}
	statuses := make(map[string]map[string]string)
	alreadySeen := make(map[string]bool)

	for _, app := range apps {
		if _, ok := alreadySeen[app.Server]; !ok {
			alreadySeen[app.Server] = true
		}
	}

	statuses, err = appPkg.GetAppStatuses(apps, true)
	if err != nil {
		log.Fatal(err)
	}

	for _, app := range apps {
		var stats ServerApps

		stats = resp[app.Server]
		var appStats *AbraApp
		for _, sapp := range stats.Apps {
			if sapp.AppName == app.Name {
				appStats = &sapp
			}
		}

		status := "unknown"
		version := "unknown"
		chaos := "unknown"
		chaosVersion := "unknown"
		if statusMeta, ok := statuses[app.StackName()]; ok {
			if currentVersion, exists := statusMeta["version"]; exists {
				if currentVersion != "" {
					version = currentVersion
				}
			}
			if chaosDeploy, exists := statusMeta["chaos"]; exists {
				chaos = chaosDeploy
			}
			if chaosDeployVersion, exists := statusMeta["chaosVersion"]; exists {
				chaosVersion = chaosDeployVersion
			}
			if statusMeta["status"] != "" {
				status = statusMeta["status"]
			}
			stats.VersionCount++
		} else {
			stats.UnversionedCount++
		}

		appStats.Status = status
		appStats.Chaos = chaos
		appStats.ChaosVersion = chaosVersion
		appStats.Version = version
		localApp := true

		var newUpdates []string
		if version != "unknown" && chaos == "false" {
			if err := app.Recipe.EnsureExists(); err != nil {
				log.Printf("unable to clone %s: %s", app.Name, err)
				InternalServerErrorHandler(w, r)
				return
			}

			updates, err := app.Recipe.Tags()
			if err != nil {
				localApp = false
			} else {
				parsedVersion, err := tagcmp.Parse(version)
				if err != nil {
					log.Fatal(err)
				}

				for _, update := range updates {
					if ok := tagcmp.IsParsable(update); !ok {
						log.Printf("unable to parse %s, skipping as upgrade option", update)
						continue
					}

					parsedUpdate, err := tagcmp.Parse(update)
					if err != nil {
						log.Fatal(err)
					}

					if update != version && parsedUpdate.IsGreaterThan(parsedVersion) {
						newUpdates = append(newUpdates, update)
					}
				}
			}
		}

		if len(newUpdates) == 0 {
			if version == "unknown" {
				appStats.Upgrade = "unknown"
			}  else if localApp {
				appStats.Upgrade = "latest"
				stats.LatestCount++
			} else {
				appStats.Upgrade = "unknown"
			}
		} else {
			newUpdates = SortVersionsDesc(newUpdates)
			appStats.Upgrade = strings.Join(newUpdates, "\n")
			stats.UpgradeCount++
		}
		
		for i, sapp := range stats.Apps {
			if sapp.AppName == app.Name {
				stats.Apps[i] = *appStats
			}
		}
		resp[app.Server] = stats
	}
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("JSON conversion failed: %s\n", err)
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
func (h *abraHandler) handleListServers (w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("abra", "server", "ls", "-m")
	abraServers, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error: ", err)
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(abraServers)
}

func SortVersionsDesc(versions []string) []string {
	var tags []tagcmp.Tag

	for _, v := range versions {
		parsed, _ := tagcmp.Parse(v) // skips unsupported tags
		tags = append(tags, parsed)
	}

	sort.Sort(tagcmp.ByTagDesc(tags))

	var desc []string
	for _, t := range tags {
		desc = append(desc, t.String())
	}

	return desc
}
