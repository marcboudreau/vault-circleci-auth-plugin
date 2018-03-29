package main

import (
	"context"
	"net/url"
	"os"

	"log"

	"github.com/hashicorp/vault/helper/pluginutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/plugin"
	circleci "github.com/tylux/go-circleci"
)

func main() {
	apiClientMeta := &pluginutil.APIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := pluginutil.VaultPluginTLSProvider(tlsConfig)

	if err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: Factory,
		TLSProviderFunc:    tlsProviderFunc,
	}); err != nil {
		log.Fatal(err)
	}
}

// Factory constructs the plugin instance with the provided BackendConfig.
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(ctx, c)

	return b.Backend, nil
}

// Client is the interface for clients used to talk to the CircleCI API.
type Client interface {
	GetBuild(project string, buildNum int) (*circleci.Build, error)
	SetBaseURL(baseURL *url.URL)
}
