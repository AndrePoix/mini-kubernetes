package worker

import (
	"context"
	"log"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Client struct {
	cli *client.Client
	ctx context.Context
}

func (c *Client) initClient(parentCtx context.Context) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	c.ctx = context.TODO()
	c.cli = cli

	go func() {
		defer cli.Close()
		<-c.ctx.Done()
	}()
}

func (c *Client) startContainer(pod *Pod) {

	pod.Phase = Running
	log.Printf("Starting container for pod %s\n", pod.Name)
	containerConfig := &container.Config{
		Image: pod.Image,
		Tty:   true,
	}

	hostConfig := &container.HostConfig{}
	if pod.CPURequest > 0 {
		hostConfig.Resources.NanoCPUs = int64(pod.CPURequest) * 1_000_000 // milliCPU to nanoCPU
	}
	if pod.MemRequest > 0 {
		hostConfig.Resources.Memory = int64(pod.MemRequest) // in bytes
	}
	if pod.ExposePort != "" && pod.HostPort != "" {
		hostConfig.PortBindings = nat.PortMap{
			nat.Port(pod.ExposePort + "/tcp"): []nat.PortBinding{
				{
					HostIP:   "127.0.0.1",
					HostPort: pod.HostPort,
				},
			},
		}

		containerConfig.ExposedPorts = nat.PortSet{
			nat.Port(pod.ExposePort + "/tcp"): struct{}{},
		}
	}

	resp, err := c.cli.ContainerCreate(c.ctx, containerConfig, hostConfig, nil, nil, pod.Name)
	//assign ContainerID
	pod.ContainerID = resp.ID
	if err != nil {
		pod.Phase = Failed
		log.Println("Error creating container:", err)
	}

	if err := c.cli.ContainerStart(c.ctx, resp.ID, container.StartOptions{}); err != nil {
		pod.Phase = Failed
		log.Println("Error starting container:", err)
	}
}

func (c *Client) deleteContainer(pod *Pod) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	timeoutSecs := 5
	log.Printf("Stopping container %s", pod.ContainerID)
	err := c.cli.ContainerStop(ctx, pod.ContainerID, container.StopOptions{
		Timeout: &timeoutSecs,
	})

	if err != nil {
		log.Printf("Failed to stop container %s: %v. Will try force remove", pod.ContainerID, err)

		// Try force remove
		removeErr := c.cli.ContainerRemove(ctx, pod.ContainerID, container.RemoveOptions{
			Force: true,
		})
		if removeErr != nil {
			log.Printf("Force remove failed for %s: %v", pod.ContainerID, removeErr)
			return
		} else {
			log.Printf("Force removed container %s", pod.ContainerID)
		}
	}

	err = c.cli.ContainerRemove(ctx, pod.ContainerID, container.RemoveOptions{})
	pod.Phase = Stopped
}
