package main

import (
	"context"
	"fmt"
	"log"

	appPkg "coopcloud.tech/abra/pkg/app"
	containerPkg "coopcloud.tech/abra/pkg/container"

	// "github.com/docker/docker/api/types"
	// "github.com/docker/docker/api/types/filters"

	"coopcloud.tech/abra/pkg/client"
	// dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/api/types/filters"
	// "github.com/spf13/cobra"

	// "coopcloud.tech/abra/pkg/log"
	"coop-cloud-backend/cli"
)

// getEnv reads env variables from docker services.
// func getEnv(cl *dockerclient.Client, stackName string) (envfile.AppEnv, error) {
// 	envMap := make(map[string]string)
// 	filter := filters.NewArgs()
// 	filter.Add("label", fmt.Sprintf("%s=%s", convert.LabelNamespace, stackName))

// 	services, err := cl.ServiceList(context.Background(), types.ServiceListOptions{Filters: filter})
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, service := range services {
// 		envList := service.Spec.TaskTemplate.ContainerSpec.Env
// 		for _, envString := range envList {
// 			splitString := strings.SplitN(envString, "=", 2)
// 			if len(splitString) != 2 {
// 				log.Debug(i18n.G("can't separate key from value: %s (this variable is probably unset)", envString))
// 				continue
// 			}
// 			k := splitString[0]
// 			v := splitString[1]
// 			log.Debugf(i18n.G("for %s read env %s with value: %s from docker service", stackName, k, v))
// 			envMap[k] = v
// 		}
// 	}

//		return envMap, nil
//	}
func printApp(app appPkg.App) error {
	fmt.Printf("Name: %s | Domain: %s | Server: %s | Path: %s",
		app.Name, app.Domain, app.Server, app.Path)
	return nil
}
func connectToContainer(app appPkg.App, serviceName string) {
	cl, err := client.New(app.Server)
	if err != nil {
		log.Fatal(err)
	}

	stackAndServiceName := fmt.Sprintf("^%s_%s", app.StackName(), serviceName)

	filters := filters.NewArgs()
	filters.Add("name", stackAndServiceName)

	fmt.Printf("HI?")

	targetContainer, err := containerPkg.GetContainer(context.Background(), cl, filters, false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Container ID: %s", targetContainer.ID)
}
func main() {
	appNames, err := appPkg.GetAppNames()
	if err != nil {
		log.Fatal(err)
	}
	for _, appName := range appNames {
		val := fmt.Sprintf("app , %v", appName)
		fmt.Println(val)
	}

	cli.StartAPI()
}
