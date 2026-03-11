package api
import (
	"fmt"
	"errors"
	"log"
	"coopcloud.tech/abra/pkg/upstream/stack"

	appPkg "coopcloud.tech/abra/pkg/app"
	formatter "coopcloud.tech/abra/pkg/formatter"
	secret "coopcloud.tech/abra/pkg/secret"

	dockerClient "github.com/docker/docker/client"

	"coop-cloud-backend/internal"
)
func getDeployVersion(deployMeta stack.DeployMeta, app appPkg.App) (string, error) {
	if internal.Chaos {
		v, err := app.Recipe.ChaosVersion()
		if err != nil {
			return "", err
		}
		log.Printf("version: taking chaos version: %s", v)
		return v, nil
	}

	v, err := getLatestVersionOrCommit(app)
	log.Printf("version: taking new recipe version: %s", v)
	if err != nil {
		return "", err
	}
	return v, nil
}


func getLatestVersionOrCommit(app appPkg.App) (string, error) {
	recipeVersions, warnings, err := app.Recipe.GetRecipeVersions()
	if err != nil {
		return "", err
	}

	for _, warning := range warnings {
		log.Printf(warning)
	}

	if len(recipeVersions) > 0 && !internal.Chaos {
		latest := recipeVersions[len(recipeVersions)-1]
		for tag := range latest {
			log.Printf("selected latest recipe version: %s (from %d available versions)", tag, len(recipeVersions))
			return tag, nil
		}
	}

	head, err := app.Recipe.Head()
	if err != nil {
		return "", err
	}

	return formatter.SmallSHA(head.String()), nil
}

func validateSecrets(cl *dockerClient.Client, app appPkg.App) error {
	composeFiles, err := app.Recipe.GetComposeFiles(app.Env)
	if err != nil {
		return err
	}

	secretsConfig, err := secret.ReadSecretsConfig(app.Path, composeFiles, app.StackName())
	if err != nil {
		return err
	}

	secStats, err := secret.PollSecretsStatus(cl, app)
	if err != nil {
		return err
	}

	for _, secStat := range secStats {
		if !secStat.CreatedOnRemote {
			secretConfig := secretsConfig[secStat.LocalName]
			if secretConfig.SkipGenerate {
				return errors.New(fmt.Sprintf("secret not inserted (#generate=false): %s", secStat.LocalName))
			}
			return errors.New(fmt.Sprintf("secret not generated: %s", secStat.LocalName))
		}
	}

	return nil
}
