package main

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type backend struct {
	*framework.Backend

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

	allPaths := append(b.ProjectMap.Paths(), pathConfig(&b))
	b.Backend = &framework.Backend{
		BackendType: logical.TypeCredential,
		Paths: allPaths,
	}

	b.Backend.Setup(c)

	return &b
}