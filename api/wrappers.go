package api
import (
	appPkg "coopcloud.tech/abra/pkg/app"
	"coopcloud.tech/abra/pkg/config"

	composetypes "github.com/docker/cli/cli/compose/types"	
)
func GetLabel(compose *composetypes.Config, stackName string, label string) string {
	return appPkg.GetLabel(compose, stackName, label)
}

func GetAppStatuses(apps []appPkg.App) (map[string]map[string]string, error) {
	return appPkg.GetAppStatuses(apps, true)
}
func GetApp(appName string) (appPkg.App, error) {
	return appPkg.Get(appName)
}

func GetAppNames() ([]string, error) {
	return appPkg.GetAppNames()
}

func GetAppServiceNames(appName string) ([]string, error) {
	return appPkg.GetAppServiceNames(appName)
}

func GetServers() ([]string, error) {
	return config.GetServers()
}

func ReadServerNames() ([]string, error) {
	return config.ReadServerNames()
}
