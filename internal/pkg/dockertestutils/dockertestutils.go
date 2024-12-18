package dockertestutils

import (
	"context"
	"io"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
)

// GetOrCreateNetwork checks if a network for a given name exists in the pool, if so
// it returns the network. Otherwise it returns a new network with the given name.
func GetOrCreateNetwork(pool *dockertest.Pool, name string) (*dockertest.Network, error) {
	networks, err := pool.NetworksByName(name)
	if err != nil {
		return nil, err
	}
	if len(networks) == 0 {
		return pool.CreateNetwork(name)
	}
	return &networks[0], nil
}

// AttachLoggerToResource attaches an io Writer to a container for a given pool.
func AttachLoggerToResource(pool *dockertest.Pool, outputStream io.Writer, containerID string) {
	go func() {
		ctx := context.Background()
		opts := docker.LogsOptions{
			Context: ctx,

			Stderr:      true,
			Stdout:      true,
			Follow:      true,
			Timestamps:  true,
			RawTerminal: true,

			Container: containerID,

			OutputStream: outputStream,
		}
		err := pool.Client.Logs(opts)
		if err != nil {
			log.Errorf("unable to attach logger to resource: %s", err)
		}
	}()
}
