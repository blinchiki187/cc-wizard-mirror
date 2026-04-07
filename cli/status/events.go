package status
import (
	"context"
	"os"
	"io" 
	"log"
	"fmt"
	"strings"
	"time"
	"encoding/json"

	dockerClient "github.com/docker/docker/client"
	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/cli/cli/command/service/progress"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/api/types/filters"

	"coopcloud.tech/abra/pkg/ui"
)
type Event interface {
	ServiceID() string
}

type ProgressEvent struct {
	Service ServiceState
	Err error
	Failed bool
}
func (e ProgressEvent) ServiceID() string { return e.Service.Id }

type HealthEvent struct {
	Service ServiceState
	Err error
	Health string
}
func (e HealthEvent) ServiceID() string { return e.Service.Id }

type StatusEvent struct {
	Service ServiceState
	Err error
	JsonMsg jsonmessage.JSONMessage
}
func (e StatusEvent) ServiceID() string { return e.Service.Id }


type ServiceState struct {
	Name 	 string `json:"name"`
	Err		 error  `json:"err"`
	Id		 string	`json:"id"`
	Status	 string	`json:"status"`
	Retries	 int	`json:"retries"`
	Health	 string	`json:"health"`
	Rollback bool	`json:"rollback"`
	Failed 	 bool	`json:"failed"`
}

type DeployState struct {
	count int
	total int
	failed bool
	quit bool
}

type DeployMsg int
const (
	FailMsg DeployMsg = iota
	CompleteMsg
	QuitMsg
)

func (ds DeployState) complete() bool {
	return ds.count == ds.total
}

func progressProducer(ctx context.Context, cl *dockerClient.Client, s ServiceState, w *io.PipeWriter, ch chan<- Event) {
	go func() {
		log.Printf("producing progress...")
		err := progress.ServiceProgress(ctx, cl, s.Id, w)
		log.Printf("got some service progress here")
		if err != nil {
			log.Printf("Error in service progress")
			ch <- ProgressEvent{
				Service: s,
				Failed: true,
				Err:     err,
			}
		} else {
			ch <- ProgressEvent{
				Service: s,
			}
		}
	}()
}

func statusProducer(ctx context.Context, decoder *json.Decoder, s ServiceState, ch chan<- Event) {
	go func() {
		log.Printf("producing status...")
		for {
			var msg jsonmessage.JSONMessage

			if err := decoder.Decode(&msg); err != nil {
				if err == io.EOF {
					return
				}

				ch <- StatusEvent{
					Service: s,
					Err:     err,
				}
				return
			}
			log.Printf("Maybe writing JSON message")
			msg.Display(os.Stdout, false); 			
			ch <- StatusEvent{
				Service:    s,
				JsonMsg: msg,
			}
		}
	}()
}

func healthProducer(ctx context.Context, cl *dockerClient.Client, s ServiceState, ch chan<- Event) {
	go func() {
		log.Printf("producing health...")
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				health := ""

				filters := filters.NewArgs()
				filters.Add("name", fmt.Sprintf("^%s", s.Name))

				containers, err := cl.ContainerList(ctx, containerTypes.ListOptions{Filters: filters})
				if err != nil {
					ch <- HealthEvent{Service: s, Err: err}
					continue
				}
				if len(containers) == 0 {
					ch <- HealthEvent{Service: s}
					continue
				}

				containerState, err := cl.ContainerInspect(ctx, containers[0].ID)
				if err != nil {
					ch <- HealthEvent{Service: s, Err: err}
					continue
				}

				if containerState.State.Health != nil {
					health = containerState.State.Health.Status
				}
				ch <- HealthEvent{
					Service: s,
					Health:  health,
				}
			}
		}
	}()
}

func processEvent(ctx context.Context, events <- chan Event, info chan <- DeployMsg, s ServiceState) error {
	for {
		select {
		case event := <- events:
			if event.ServiceID() != s.Id {
				continue
			}
			switch event := event.(type) {
			case ProgressEvent:
				if event.Err != nil {
					s.Err = event.Err
					log.Printf("Error in progress event: %s", s.Err)
				}
				if event.Failed {
					info <- FailMsg
				}
				// print for debugging purposes
				log.Printf("Service: %s is complete", event.Service.Id)
				// end print
				info <- CompleteMsg

			case StatusEvent:
				if event.Err != nil {
					s.Err = event.Err
					log.Printf("Error in status event: %s", s.Err)
				}

				// print for debugging purposes
				b, err := json.MarshalIndent(event.JsonMsg, "", "  	")
				if err != nil {
					log.Printf("Problem with json message: %s", err)
				}
				log.Printf(string(b))
				// end print
				
				if event.JsonMsg.ID == "rollback" {
					log.Printf("Failed in rollback")
					s.Failed = true
					s.Rollback = true
				}

				if event.JsonMsg.ID != "overall progress" {
					newStatus := strings.ToLower(event.JsonMsg.Status)
					currentStatus := s.Status

					if !strings.Contains(currentStatus, "starting") &&
						strings.Contains(newStatus, "starting") {
						s.Retries += 1
					}

					if s.Rollback {
						if event.JsonMsg.ID == "rollback" {
							s.Status = newStatus
						}
					} else {
						s.Status = newStatus
					}
				}
		
			case HealthEvent:
				if event.Err != nil {
					s.Err = event.Err
					log.Printf("Error in health event: %s", s.Err)
				}
				h := "?"
				if s.Health != "" {
					h = s.Health
				}
				if event.Health != "" {
					h = event.Health
				}
				s.Health = h
			}

		case <- ctx.Done():
			log.Printf("context is done???")
			return ctx.Err()
		}
	}
}
func processService(ctx context.Context, info chan <- DeployMsg, cl *dockerClient.Client, s ServiceState, decoder *json.Decoder, writer *io.PipeWriter) {
	events := make (chan Event, 50)
	progressProducer(ctx, cl, s, writer, events)
	statusProducer(ctx, decoder, s, events)
	healthProducer(ctx, cl, s, events)

	go processEvent(ctx, events, info, s)
}

func WaitOnServices(pctx context.Context, cl *dockerClient.Client, services []ui.ServiceMeta, filters filters.Args) {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	ds := DeployState{
		count: 0,
		total: len(services),
		failed: false,
		quit: false,
	}
	info := make (chan DeployMsg, 50)
	for _, service := range services {
		r, w := io.Pipe()
		d := json.NewDecoder(r)
		s := ServiceState{
			Name: service.Name,
			Id: service.ID,
			Retries: -1,
			Health: "?",
		}
		log.Printf("Processing Service: %s", s.Id)
		processService(ctx, info, cl, s, d, w)
	}
	for {
		select {
		case msg := <-info:
			switch msg {
			case FailMsg:
				ds.failed = true
			case CompleteMsg:
				ds.count += 1
				if ds.complete() {
					log.Printf("deploy completed")
					cancel()
				}
			case QuitMsg:
				ds.quit = true
				cancel()
			}
		case <-ctx.Done():
			log.Printf("Context finished because: %s", ctx.Err())
			return
		}
	}
}