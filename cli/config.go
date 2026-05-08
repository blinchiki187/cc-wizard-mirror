package cli
import (
	"log"
	"io/ioutil"
	"os"
	"net/http"
	"encoding/json"
	appPkg "coopcloud.tech/abra/pkg/app"
)
type FileResponse struct {
	Content string `json:"content"`
}

func (h *abraHandler) handleGetConfig(w http.ResponseWriter, r *http.Request, appName string) {
	files, err := appPkg.LoadAppFiles("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appFile, exists := files[appName]
	if !exists {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("path: %s | server: %s", appFile.Path, appFile.Server)
	log.Printf("Ending...")
	// TODO: sanitize 
	file, err := ioutil.ReadFile(appFile.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := FileResponse{Content: string(file)}
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *abraHandler) handlePostConfig(w http.ResponseWriter, r *http.Request, appName string) {
	files, err := appPkg.LoadAppFiles("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appFile, exists := files[appName]
	if !exists {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // catch unwanted fields

	var config FileResponse
	err = d.Decode(&config)
    if err != nil {
		log.Printf("err: %s", err);
        http.Error(w, "Failed to read request", http.StatusBadRequest)
        return
    }

    err = os.WriteFile(appFile.Path, []byte(config.Content), 0644)
    if err != nil {
        http.Error(w, "Failed to write file", http.StatusInternalServerError)
        return
    }
	w.WriteHeader(http.StatusOK)
}