package testhelpers

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const mongoContainerName = "ccsappvp2-rental-management-dev-mongo-1"

func IsMongoDbContainerRunning() (bool, error) {
	return isContainerRunning(mongoContainerName)
}

func isContainerRunning(containerName string) (bool, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return false, err
	}

	for _, container := range containers {
		// Our docker compose stack sets the container name matching the following condition.
		if container.Names[0] == "/"+containerName && container.State == "running" {
			return true, nil
		}
	}

	return false, nil
}
