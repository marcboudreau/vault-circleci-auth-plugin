package main

import (
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type backend struct {
	*framework.Backend
}

// Backend creates a new backend with the provided BackendConfig.
func Backend(c *logical.BackendConfig) *backend {
	var b backend

	b.Backend = &framework.Backend{
		BackendType: logical.TypeCredential,
		Paths: []*framework.Path{
			pathConfig(&b),
		},
	}

	b.Backend.Setup(c)

	return &b
}