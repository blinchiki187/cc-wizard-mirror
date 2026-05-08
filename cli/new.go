package cli
import (
	"log"
	"os/exec"
	"net/http"
	"encoding/json"
	"fmt"

	appPkg "coopcloud.tech/abra/pkg/app"
	"coopcloud.tech/abra/pkg/secret"
	"coopcloud.tech/abra/pkg/recipe"
	"coopcloud.tech/abra/pkg/client"
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
	log.Printf("%v", args)
	cmd := exec.Command("abra", args...)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error: ", string(output))
		InternalServerErrorHandler(w, r)
		return
	}
	var fs []AppSecret
	if body.Secrets != nil && *body.Secrets == true {
		appSecrets, err := createSecrets(appName, *body.Domain, *body.Server)
		if err != nil {
			log.Printf("Error creating secrets: %s", err)
			InternalServerErrorHandler(w, r)
			return
		}
		for k, v := range appSecrets {
			fs = append(fs, AppSecret{
				Name: k,
				Value: v,
				Version: "v1",
			})
		}
	}
	jsonBytes, err := json.Marshal(fs)
	if err != nil {
		log.Printf("JSON conversion failed: %s\n", err)
		InternalServerErrorHandler(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func createSecrets(appName string, domain string, server string) (map[string]string, error) {
	cl, err := client.New(server)
	if err != nil {
		return nil, err
	}
	recipe := recipe.Get(appName)


	sampleEnv, err := recipe.SampleEnv()
	if err != nil {
		return nil, err
	}

	composeFiles, err := recipe.GetComposeFiles(sampleEnv)
	if err != nil {
		return nil, err
	}

	//sanitisedAppName := appPkg.SanitiseAppName(domain)
	secretsConfig, err := secret.ReadSecretsConfig(
		recipe.SampleEnvPath,
		composeFiles,
		appPkg.StackName(domain),
	)
	if err != nil {
		return nil, err
	}

	secrets, err := secret.GenerateSecrets(cl, secretsConfig, server)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}