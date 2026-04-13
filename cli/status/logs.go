package status
import (
	"bufio"
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	containerTypes "github.com/docker/docker/api/types/container"
)
func StreamLogs(
	ctx context.Context,
	cl *dockerClient.Client,
	services []swarm.Service,
	logCh chan <- string,
) error {
	waitCh := make(chan struct{})
	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		go func(serviceID string) {
			defer wg.Done()
			tail := "50"

			logs, err := cl.ServiceLogs(ctx, serviceID, containerTypes.LogsOptions{
				ShowStderr: true,
				ShowStdout: true,
				Until:      "",
				Timestamps: true,
				Follow:     true,
				Tail:       tail,
				Details:    false,
			})

			if err == nil {
				buf := bufio.NewScanner(logs)
				for buf.Scan() {
					select {
					case <- ctx.Done():
						return
					default:
					line := fmt.Sprintf("%s: %s", service.Spec.Name, buf.Text())
					logCh <- line
					}
				}
				logs.Close()
				return
			}
		}(service.ID)
	}

	wg.Wait()
	close(waitCh)
	close(logCh)
	return nil
}