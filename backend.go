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

	AttemptedBuilds       *CircleCIBuildList
	AttemptedBuildsBuffer time.Duration
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
	b.AttemptedBuildsBuffer = 5 * time.Hour

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

func (b *backend) periodicFunc(_ *logical.Request) error {
	b.Logger().Trace("periodicFunc called with time", time.Now().Add(-b.AttemptedBuildsBuffer).Format(time.UnixDate))
	b.AttemptedBuilds.Cleanup(time.Now().Add(-b.AttemptedBuildsBuffer), b)
	return nil
}
