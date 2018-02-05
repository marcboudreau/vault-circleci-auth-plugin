package main

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"

	"github.com/marcboudreau/vault-circleci-auth-plugin/circleci"
)

type backend struct {
	*framework.Backend

	client Client
	ProjectMap *framework.PolicyMap
}

// Backend creates a new backend with the provided BackendConfig.
func Backend(c *logical.BackendConfig) *backend {
	var b backend

	b.ProjectMap = &framework.PolicyMap{
		PathMap: framework.PathMap{
			Name: "projects",
		},
		DefaultKey: "default",
	}

	allPaths := append(b.ProjectMap.Paths(), pathConfig(&b), pathLogin(&b))
	b.Backend = &framework.Backend{
		BackendType: logical.TypeCredential,
		Paths: allPaths,
	}

	b.Backend.Setup(c)

	return &b
}

func (b *backend) GetClient(token string) Client {
	if b.client == nil {
		b.client = circleci.New(token)
		
	}

	return b.client
}