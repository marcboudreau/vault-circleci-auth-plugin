package main

import (
	"context"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"

	"github.com/marcboudreau/vault-circleci-auth-plugin/circleci"

	cache "github.com/patrickmn/go-cache"
)

type backend struct {
	*framework.Backend

	client     Client
	ProjectMap *framework.PolicyMap

	AttemptsCache *cache.Cache
	CacheExpiry   time.Duration
}

// Backend creates a new backend with the provided BackendConfig.
func Backend(ctx context.Context, c *logical.BackendConfig) *backend {
	var b backend

	b.ProjectMap = &framework.PolicyMap{
		PathMap: framework.PathMap{
			Name: "projects",
		},
		DefaultKey: "default",
	}

	b.AttemptsCache = cache.New(5*time.Hour, cache.NoExpiration)
	b.CacheExpiry = 5 * time.Hour

	allPaths := append(b.ProjectMap.Paths(), pathConfig(&b), pathLogin(&b))

	b.Backend = &framework.Backend{
		PeriodicFunc: b.periodicFunc,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{
				"login",
			},
		},
		Paths:       allPaths,
		BackendType: logical.TypeCredential,
	}

	b.Backend.Setup(ctx, c)

	return &b
}

func (b *backend) GetClient(token, vcsType, owner string) Client {
	if b.client == nil {
		b.client = circleci.New(token, vcsType, owner)
	}

	return b.client
}

func (b *backend) periodicFunc(_ context.Context, _ *logical.Request) error {
	b.Logger().Trace("periodicFunc called")
	b.AttemptsCache.DeleteExpired()

	return nil
}
