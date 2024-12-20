// Package oktamock provides a docker resource for the okta mock server.
package oktamock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"schubergphilis/mcvs-integrationtest-services/internal/pkg/constants"
	"schubergphilis/mcvs-integrationtest-services/internal/pkg/dockertestutils"
)

// ErrOktaMockServerNotHealthy okta mock server not healthy.
var ErrOktaMockServerNotHealthy = fmt.Errorf("okta mock server not healthy")

// ErrNotRunning okta mock server container not running yet.
var ErrNotRunning = fmt.Errorf("okta mock server container not running yet")

// Resource the docker resource for the okta mock server.
type Resource struct {
	pool     *dockertest.Pool
	network  *dockertest.Network
	resource *dockertest.Resource

	writer io.Writer
}

// NewResource creates a new okta mock server resource.
func NewResource(pool *dockertest.Pool, network *dockertest.Network) *Resource {
	return &Resource{
		pool:    pool,
		network: network,
	}
}

// WithLogger adds a logger to the resources, to track docker logs.
func (r *Resource) WithLogger(writer io.Writer) *Resource {
	r.writer = writer
	return r
}

// Start starts the resource with given run options.
func (r *Resource) Start(opts *dockertest.RunOptions, _ string, hcOpts ...func(*docker.HostConfig)) error {
	opts.Networks = append(opts.Networks, r.network)
	var err error
	projectRoot, err := constants.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to determine the root of the project: %w", err)
	}
	buildArgs := []docker.BuildArg{
		{
			Name:  "APPLICATION",
			Value: "oktamock",
		},
	}
	r.resource, err = r.pool.BuildAndRunWithBuildOptions(&dockertest.BuildOptions{
		Dockerfile: "./Dockerfile",
		ContextDir: projectRoot,
		BuildArgs:  buildArgs,
	}, opts, hcOpts...)
	if err != nil {
		return fmt.Errorf("unable to build okta mock server container: %w", err)
	}

	err = r.waitUntilContainerIsRunning()
	if err != nil {
		return err
	}
	if r.writer != nil {
		dockertestutils.AttachLoggerToResource(r.pool, r.writer, r.ContainerID())
	}

	return r.startupCheck(opts)
}

func (r *Resource) startupCheck(opts *dockertest.RunOptions) error {
	return r.pool.Retry(func() error {
		oktaMockServerPort := "8080"
		if len(opts.ExposedPorts) > 0 {
			oktaMockServerPort = opts.ExposedPorts[0]
		}
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("http://localhost:%s/token", r.resource.GetPort(fmt.Sprintf("%s/tcp", oktaMockServerPort))), io.NopCloser(bytes.NewBufferString("{\"custom_claims\": {\"allowed_services\": \"['*']\"}}")))
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Default().Println("unable to perform http request okta mock server, try again...", err)
			return err
		}
		defer func() {
			err = resp.Body.Close()
			if err != nil {
				log.Default().Println("unable to close response body", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			return ErrOktaMockServerNotHealthy
		}
		return nil
	})
}

func (r *Resource) waitUntilContainerIsRunning() error {
	return r.pool.Retry(func() error {
		container, err := r.pool.Client.InspectContainer(r.ContainerID())
		if err != nil {
			return err
		}
		if container.State.Running {
			return nil
		}
		return ErrNotRunning
	})
}

// GetPort retrieve the mapped docker port.
func (r *Resource) GetPort(port string) string {
	return r.resource.GetPort(port)
}

// Stop stop the resource.
func (r *Resource) Stop() error {
	return r.resource.Close()
}

// ContainerID retrieves the container ID.
func (r *Resource) ContainerID() string {
	return r.resource.Container.ID
}
