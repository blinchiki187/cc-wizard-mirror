package api
import (
	"context"
	"fmt"
	"log"
	"encoding/json"
	"coopcloud.tech/abra/pkg/client"
	"coopcloud.tech/abra/pkg/upstream/stack"
	"coopcloud.tech/abra/pkg/upstream/convert"

	"coopcloud.tech/abra/pkg/ui"

	appPkg "coopcloud.tech/abra/pkg/app"
	"coopcloud.tech/abra/pkg/deploy"

	"net/http"	
	"github.com/gorilla/websocket"
	tea "github.com/charmbracelet/bubbletea"

	"coop-cloud-backend/internal"
)
func (h *abraHandler) handleDeployApp(w http.ResponseWriter, r *http.Request, appName string) {
	log.Printf("App Id: %s", appName)
	app, err := GetApp(appName)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}
	cl, err := client.New(app.Server)
	if err != nil {
		log.Printf("Error Connecting to Docker client: %s\n", err)
		http.Error(w, fmt.Sprintf("Error Connecting to Docker client: %s\n", err), http.StatusInternalServerError)
		return
	}
	deployMeta, err := stack.IsDeployed(context.Background(), cl, app.StackName())
	if err != nil {
		log.Printf("Error checking deploy status: %s\n", err)
		http.Error(w, fmt.Sprintf("Error checking deploy status: %s\n", err), http.StatusInternalServerError)
		return
	}
	if deployMeta.IsDeployed {
		log.Printf("App already deployed\n")
		http.Error(w, "App already deployed\n", http.StatusInternalServerError)
		return
	}

	// logic differs from CLI, we only want to take either
	// 1. the chaos version
	// 2. TODO: the version in the .env file
	// 3. the latest version
	// we never take: specific CLI verison (maybe will support this) or the deployed version
	internal.Chaos = false
	toDeployVersion, err := getDeployVersion(deployMeta, app)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	if err := validateSecrets(cl, app); err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	if err := deploy.MergeAbraShEnv(app.Recipe, app.Env); err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	composeFiles, err := app.Recipe.GetComposeFiles(app.Env)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	stackName := app.StackName()
	deployOpts := stack.Deploy{
		Composefiles: composeFiles,
		Namespace:    stackName,
		Prune:        false,
		ResolveImage: stack.ResolveImageAlways,
		Detach:       false,
	}
	compose, err := appPkg.GetAppComposeConfig(app.Name, deployOpts, app.Env)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	appPkg.SetRecipeLabel(compose, stackName, app.Recipe.Name)
	appPkg.SetChaosLabel(compose, stackName, internal.Chaos)
	if internal.Chaos {
		appPkg.SetChaosVersionLabel(compose, stackName, toDeployVersion)
	}

	versionLabel := toDeployVersion
	if internal.Chaos {
		for _, service := range compose.Services {
			if service.Name == "app" {
				labelKey := fmt.Sprintf("coop-cloud.%s.version", stackName)
				// NOTE(d1): keep non-chaos version labbeling when doing chaos ops
				versionLabel = service.Deploy.Labels[labelKey]
			}
		}
	}
	appPkg.SetVersionLabel(compose, stackName, versionLabel)

	newRecipeWithDeployVersion := fmt.Sprintf("%s:%s", app.Recipe.Name, toDeployVersion)
	appPkg.ExposeAllEnv(stackName, compose, app.Env, newRecipeWithDeployVersion)

	envVars, err := appPkg.CheckEnv(app)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}
	// doesn't really get used at all right now
	deployWarnMessages := []string{}
	for _, envVar := range envVars {
		if !envVar.Present {
			deployWarnMessages = append(deployWarnMessages,
				fmt.Sprintf("%s missing from %s.env", envVar.Name, app.Domain),
			)
		}
	}

	//skipping domain checks like crazy
	//commented code is to show deploy overview before deploy

	/*
	deployedVersion := config.MISSING_DEFAULT
	if deployMeta.IsDeployed {
		deployedVersion = deployMeta.Version
		if deployMeta.IsChaos {
			deployedVersion = deployMeta.ChaosVersion
		}
	}
	ShowUnchanged := false
	secretInfo, err := deploy.GatherSecretsForDeploy(cl, app, ShowUnchanged)
	if err != nil {
		log.Fatal(err)
		InternalServerErrorHandler(w, r)
		return
	}

	// Gather configs
	configInfo, err := deploy.GatherConfigsForDeploy(cl, app, compose, app.Env, ShowUnchanged)
	if err != nil {
		log.Fatal(err)
		InternalServerErrorHandler(w, r)
		return
	}

	// Gather images
	imageInfo, err := deploy.GatherImagesForDeploy(cl, app, compose, ShowUnchanged)
	if err != nil {
		log.Fatal(err)
		InternalServerErrorHandler(w, r)
		return
	}*/

	stack.WaitTimeout, err = appPkg.GetTimeoutFromLabel(compose, stackName)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	serviceNames, err := appPkg.GetAppServiceNames(app.Name)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	f, err := app.Filters(true, false, serviceNames...)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	// in future allow user input?
	DontWaitConverge := true
	NoInput := true
	if err := stack.RunDeploy(
		cl,
		deployOpts,
		compose,
		app.Name,
		app.Server,
		DontWaitConverge,
		NoInput,
		f,
	); err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}
	// Implement my own WaitOnServices in order to wrap ui.Model in order to emit JSON updates
	// over Websocket connection
	var serviceIDs []ui.ServiceMeta
	namespace := convert.NewNamespace(deployOpts.Namespace)

	existingServices, err := stack.GetStackServices(context.Background(), cl, namespace.Name())
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}

	for _, service := range existingServices {
		serviceIDs = append(serviceIDs, ui.ServiceMeta{
			Name: service.Spec.Name,
			ID: service.ID,
		})
	}


	waitOpts := WaitOpts{
		Services:   serviceIDs,
		AppName:    app.Name,
		ServerName: app.Server,
		NoInput:    NoInput,
		Filters:    f,
	}	

	w.WriteHeader(http.StatusOK)
}

type StateEmitter interface {
	Emit(DeployState)
}

type WebsocketEmitter struct {
    conn *websocket.Conn
}

func (w *WebsocketEmitter) Emit(state DeployState) {
    b, _ := json.Marshal(state)
    w.conn.WriteMessage(websocket.TextMessage, b)
}

type WrappedModel struct {
	inner	tea.Model
	emitter	StateEmitter
}

func (m WrappedModel) Init() tea.Cmd {
	return m.inner.Init()
}

func (m WrappedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	innerState, cmds := m.inner.Update(msg)
	
	m.inner = innerState

	m.emitter.Emit(s.State())

	return m, cmds
}

func (m WrappedModel) View() string {
	return m.inner.View()
}

func (m *ui.Model) State() DeployState {
	var streams []DeployStream
	if m.Streams != nil {
		for _, s := range *m.Streams {
			streams = append(streams, DeployStream{
				Name: s.Name,
				id: s.id,
				status: s.status,
				retries: s.retries,
				health: s.health,
				rollback: s.rollback,
			})
		}
	}
	var logs []string
	if m.Logs != nil {
		logs = *m.Logs
	}
	return DeployState{
		AppName: m.appName,
		Streams: streams,
		Logs: logs,
		Failed: m.Failed,
		TimedOut: m.TimedOut,
		Quit: m.Quit,
		Count: m.count,
	}
}