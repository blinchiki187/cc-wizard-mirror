package api
import (
	"context"
	"fmt"
	"log"
	"coopcloud.tech/abra/pkg/client"
	"coopcloud.tech/abra/pkg/upstream/stack"

	"net/http"
)
func (h *abraHandler) handleUndeployApp(w http.ResponseWriter, r *http.Request, appName string) {
	log.Printf("App Id: %s", appName)
	app, err := GetApp(appName)
	if err != nil {
		log.Printf("Error getting app %s: %s\n", appName, err)
		InternalServerErrorHandler(w, r)
		return
	}
	// think this just checks to make sure we have the recipe for this app
	if err := app.Recipe.EnsureExists(); err != nil {
		log.Fatal(err)
		InternalServerErrorHandler(w, r)
		return
	}

	cl, err := client.New(app.Server)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	deployMeta, err := stack.IsDeployed(context.Background(), cl, app.StackName())
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}
	if !deployMeta.IsDeployed {
		log.Fatal("%s is not deployed?", app.Name)
		InternalServerErrorHandler(w, r)
		return
	}

	version := deployMeta.Version
	if deployMeta.IsChaos {
		version = deployMeta.ChaosVersion
	}

	rmOpts := stack.Remove{
		Namespaces: []string{app.StackName()},
		Detach:     false,
	}
	if err := stack.RunRemove(context.Background(), cl, rmOpts); err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}
	
	if err := app.WriteRecipeVersion(version, false); err != nil {
		log.Printf("writing recipe version failed: %s\n", err)
		http.Error(w, fmt.Sprintf("writing recipe version failed: %s\n", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
