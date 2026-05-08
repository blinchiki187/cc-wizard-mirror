package cli
import (
	"context"
	"log"
	"fmt"
	"net/http"

	appPkg "coopcloud.tech/abra/pkg/app"
	"coop-cloud-backend/cli/status"
	"coopcloud.tech/abra/pkg/client"
	"coopcloud.tech/abra/pkg/upstream/stack"

	"github.com/docker/docker/api/types"
	
)
func (h *abraHandler) handleGetLogs(w http.ResponseWriter, r *http.Request, appName string, serviceName string) {
	app, err := appPkg.Get(appName)
	if err != nil {
		log.Printf("error: %s", err)
		return 
	}
	stackName := app.StackName()

	if err := app.Recipe.EnsureExists(); err != nil {
		log.Fatal(err)
		return
	}
	cl, err := client.New(app.Server)
	if err != nil {
		log.Fatal(err)
		return
	}

	deployMeta, err := stack.IsDeployed(context.Background(), cl, stackName)
	if err != nil {
		log.Fatal(err)
		return
	}

	if !deployMeta.IsDeployed {
		log.Printf(fmt.Sprintf("%s is not deployed?", app.Name))
		return
	}
	serviceNames := []string{serviceName}

	f, err := app.Filters(true, false, serviceNames...)
	if err != nil {
		log.Fatal(err)
	}
	ctx := r.Context()

	services, err := cl.ServiceList(
		context.Background(),
		types.ServiceListOptions{Filters: f},
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

	logCh := make (chan string, 50)
	go status.StreamLogs(ctx, cl, services, logCh)
	for {
		select{
		case <- ctx.Done():
			log.Printf("log streaming done")
			return
		case line, ok := <- logCh:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
		}
	}
}
