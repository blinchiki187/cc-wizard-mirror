package cli
import (
	"log"
	"os/exec"
	"net/http"
)
func (h *abraHandler) handleRemoveApp(w http.ResponseWriter, r *http.Request, appName string) {
	cmd := exec.Command("abra", "app", "remove", appName, "-n")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error: ", string(output))
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}
