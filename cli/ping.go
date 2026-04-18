package cli
import (
	"os/exec"
	"log"
	"sort"
	"net/http"
	"encoding/json"
)
func (h *abraHandler) handleGetAppServices (w http.ResponseWriter, r *http.Request, appName string) {
	var tmpMsg map[string]map[string]string
	cmd := exec.Command("abra", "app", "ps", appName, "-m")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error: ", err)
		InternalServerErrorHandler(w, r)
		return
	}
	if err = json.Unmarshal(output, &tmpMsg); err != nil {
		log.Printf("Error: ", err)
		InternalServerErrorHandler(w, r)
		return
	}
	var services []AbraAppService
	for _,v := range tmpMsg {
		services = append(services, AbraAppService{
			Service: v["service"],
			Chaos: v["chaos"] == "true",
			Created: v["created"],
			Image: v["image"],
			Ports: v["ports"],
			State: v["state"],
			Status: v["status"],
			Version: v["version"],
		})
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].Service < services[j].Service
	})
	jsonBytes, err := json.Marshal(services)
	if err != nil {
		log.Printf("JSON conversion failed: %s\n", err)
		InternalServerErrorHandler(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}