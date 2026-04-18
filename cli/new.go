package cli
import (
	"log"
	"os/exec"
	"net/http"
	"encoding/json"
	"fmt"
)
func (h *abraHandler) handleNewApp(w http.ResponseWriter, r *http.Request, appName string) {
	args := []string{"app", "new", appName, "-n"}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // catch unwanted fields

	// anonymous struct type: handy for one-time use
	body := struct {
		Domain  	  *string 	`json:"domain"`
		Server  	  *string 	`json:"server"`
		Chaos   	  *bool   	`json:"chaos"` 
		Secrets		  *bool   	`json:"secrets"` 
	}{}
	
	err := d.Decode(&body)
	if err != nil {
		log.Printf("???\n")
		// bad JSON or unrecognized json field
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.Domain == nil {	
		http.Error(w, "missing field 'domain' from JSON object", http.StatusBadRequest)
		return
	}
	if body.Server == nil {	
		http.Error(w, "missing field 'server' from JSON object", http.StatusBadRequest)
		return
	}
	args = append(args, fmt.Sprintf("--domain=%s", *body.Domain))
	args = append(args, fmt.Sprintf("--server=%s", *body.Server))
	if body.Chaos != nil && *body.Chaos == true {
		args = append(args, "-C")
	}
	if body.Secrets != nil && *body.Secrets == true {
		args = append(args, "--secrets")
	}
	log.Printf("%v", args)
	cmd := exec.Command("abra", args...)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error: ", string(output))
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}
