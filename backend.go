package main

import (
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"

	"github.com/marcboudreau/vault-circleci-auth-plugin/circleci"
)

type backend struct {
	*framework.Backend

	client     Client
	ProjectMap *framework.PolicyMap

	AttemptedBuilds *CircleCIBuildList
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

	b.AttemptedBuilds = New()

	allPaths := append(b.ProjectMap.Paths(), pathConfig(&b), pathLogin(&b))
	b.Backend = &framework.Backend{
		BackendType: logical.TypeCredential,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{
				"login",
			},
		},
		Paths:        allPaths,
		PeriodicFunc: cleanupFunc(&b),
	}

	b.Backend.Setup(c)

	return &b
}

func (b *backend) GetClient(token, vcsType, owner string) Client {
	if b.client == nil {
		b.client = circleci.New(token, vcsType, owner)
	}

	return b.client
}

func (b *backend) RecordAttempt(project string, buildNum int) bool {
	return b.AttemptedBuilds.Add(project, buildNum)
}

func cleanupFunc(b *backend) func(*logical.Request) error {
	return func(*logical.Request) error {
		b.AttemptedBuilds.Cleanup(time.Now().Add(-5 * time.Hour))
		return nil
	}
}
