package cli
import (
	"log"
	"io"
	"os"
	"net/http"
	appPkg "coopcloud.tech/abra/pkg/app"
)
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
	w.WriteHeader(http.StatusOK)
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
	body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request", http.StatusBadRequest)
        return
    }

    defer r.Body.Close()

    err = os.WriteFile(appFile.Path, body, 0644)
    if err != nil {
        http.Error(w, "Failed to write file", http.StatusInternalServerError)
        return
    }
	w.WriteHeader(http.StatusOK)
}