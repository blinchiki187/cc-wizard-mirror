package cli
import (
	"log"
	"os/exec"
	"coop-cloud-backend/internal"
	"coop-cloud-backend/cli/status"
	"net/http"
	//"encoding/json"
	"fmt"
	"context"

	"coopcloud.tech/abra/pkg/client"
	//"coopcloud.tech/abra/pkg/upstream/stack"
	"coopcloud.tech/abra/pkg/upstream/convert"
	"coopcloud.tech/abra/pkg/ui"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	dclient "github.com/docker/docker/client"
	"github.com/docker/docker/api/types"


	appPkg "coopcloud.tech/abra/pkg/app"

	//"github.com/gorilla/websocket"
)
func (h *abraHandler) handleDeployApp(w http.ResponseWriter, r *http.Request, appName string) {
	log.Printf("Handling App Deploy!")
	args := []string{"app", "deploy", appName, "-n", "-c"}
	if internal.Chaos {
		args = append(args, "-C")
	}
	cmd := exec.Command("abra", args...)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error: ", string(output))
		InternalServerErrorHandler(w, r)
		return
	}
	app, err :=  appPkg.Get(appName)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}
	// Implement my own WaitOnServices in order to wrap ui.Model in order to emit JSON updates
	// over Websocket connection
	serviceNames, err := appPkg.GetAppServiceNames(app.Name)
	if err != nil {
		log.Fatal(err)
	}

	f, err := app.Filters(true, false, serviceNames...)
	if err != nil {
		log.Fatal(err)
	}

	// STEP 1: collect information about app being deployed
	cl, err := client.New(app.Server)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	var serviceIDs []ui.ServiceMeta
	stackName := app.StackName()
	namespace := convert.NewNamespace(stackName)
	log.Printf("stack name: %s | namespace: %s", stackName, namespace.Name())
	existingServices, err := GetStackServices(context.Background(), cl, namespace.Name())
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	for _, service := range existingServices {
		log.Printf("existingServices contains: %s", service.Spec.Name)
		serviceIDs = append(serviceIDs, ui.ServiceMeta{
			Name: service.Spec.Name,
			ID: service.ID,
		})
	}
	log.Printf("Waiting on service...")
	status.WaitOnServices(context.Background(), cl, serviceIDs, f)

	w.WriteHeader(http.StatusOK)
}


func getStackFilter(namespace string) filters.Args {
	filter := filters.NewArgs()
	filter.Add("label", convert.LabelNamespace+"="+namespace)
	return filter
}
func getStackServiceFilter(namespace string) filters.Args {
	return getStackFilter(namespace)
}

func GetStackServices(ctx context.Context, dockerclient dclient.APIClient, namespace string) ([]swarm.Service, error) {
	return dockerclient.ServiceList(ctx, types.ServiceListOptions{Filters: getStackServiceFilter(namespace)})
}