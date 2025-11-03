//go:build integration

package oktamock

import (
	"fmt"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/schubergphilis/mcvs-integrationtest-services/internal/pkg/dockertesthelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	oktaMockServerName = "okta-mock-server"
)

func TestCanRunOkta(t *testing.T) {
	pool, err := dockertest.NewPool("")
	assert.NoError(t, err)

	network, err := dockertesthelpers.GetOrCreateNetwork(pool, "integration-test-okta")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, network.Close())
	}()

	oktaResource := NewResource(pool, network)

	defer func() {
		assert.NoError(t, oktaResource.Stop())
	}()

	err = oktaResource.Start(&dockertest.RunOptions{
		Name: oktaMockServerName,
		Tag:  "",
		Env: []string{
			fmt.Sprintf("ISSUER=http://%s:8080", oktaMockServerName),
		},
		ExposedPorts: []string{"8080"},
	}, "..")
	assert.NoError(t, err)
}
