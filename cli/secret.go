package cli
import (
	"log"
	"os/exec"
	"encoding/json"
	"net/http"
)
func (h *abraHandler) handleGetAppSecrets(w http.ResponseWriter, r *http.Request, appName string) {
	cmd := exec.Command("abra", "app", "secret", "list", appName, "-m")
	secrets, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error: ", err)
		InternalServerErrorHandler(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(secrets)
}

func (h *abraHandler) handleInsertAppSecret(w http.ResponseWriter, r *http.Request, appName string) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // catch unwanted fields

	// anonymous struct type: handy for one-time use
	body := struct {
		Name 	*string 	`json:"name"`
		Version *string		`json:"version"`
		Value 	*string		`json:"value"`
	}{}

	err := d.Decode(&body)
	if err != nil {
		log.Printf("???\n")
		// bad JSON or unrecognized json field
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}


	if body.Name == nil {	
		http.Error(w, "missing field 'name' from JSON object", http.StatusBadRequest)
		return
	}
	if body.Version == nil {	
		http.Error(w, "missing field 'version' from JSON object", http.StatusBadRequest)
		return
	}
	if body.Value == nil {	
		http.Error(w, "missing field 'value' from JSON object", http.StatusBadRequest)
		return
	}
	log.Printf("%s %s %s", *body.Name, *body.Version, *body.Value)
	cmd := exec.Command("abra", "app", "secret", "insert", appName, *body.Name, *body.Version, *body.Value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error: ", string(output))
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}
