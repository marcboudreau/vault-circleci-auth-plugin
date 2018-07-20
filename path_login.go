package main

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathLogin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "login",
		Fields: map[string]*framework.FieldSchema{
			"user": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The owner of the build's repository.",
			},
			"project": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The name of the build's repository.",
			},
			"build_num": &framework.FieldSchema{
				Type:        framework.TypeInt,
				Description: "The number of the current build.",
			},
			"vcs_revision": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The hash of the current build's source control revision.",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathLogin,
		},
	}
}

func (b *backend) pathLogin(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	project := d.Get("project").(string)
	buildNum := d.Get("build_num").(int)
	vcsRevision := d.Get("vcs_revision").(string)

	s := fmt.Sprintf("%#v", req)
	b.Logger().Debug("pathLogin - Request: %s", s)

	if resp := b.lockBuild(project, buildNum); resp != nil {
		return resp, nil
	}

	var verifyResp *verifyBuildResponse
	if verifyResponse, resp, err := b.verifyBuild(ctx, req, project, buildNum, vcsRevision); err != nil {
		return nil, err
	} else if resp != nil {
		return resp, nil
	} else {
		verifyResp = verifyResponse
	}

	config, err := b.Config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	ttl, _, err := b.SanitizeTTLStr(config.TTL.String(), config.MaxTTL.String())
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("error sanitizing TTLs: %s", err)), nil
	}

	resp := &logical.Response{
		Auth: &logical.Auth{
			InternalData: map[string]interface{}{
				"project":      project,
				"build_num":    buildNum,
				"vcs_revision": vcsRevision,
			},
			Policies:    verifyResp.Policies,
			DisplayName: fmt.Sprintf("%s-%d", project, buildNum),
			LeaseOptions: logical.LeaseOptions{
				TTL:       ttl,
				Renewable: true,
			},
		},
	}

	return resp, nil
}

func (b *backend) lockBuild(project string, buildNum int) *logical.Response {
	if err := b.AttemptsCache.Add(fmt.Sprintf("%s/%d", project, buildNum), struct{}{}, b.CacheExpiry); err != nil {
		b.Logger().Debug("Build already in Cache (%s %d)", project, buildNum)
		return logical.ErrorResponse(
			"an attempt to authenticate as this build has already been made")
	}
	return nil
}

func (b *backend) verifyBuild(ctx context.Context, req *logical.Request, project string, buildNum int, vcsRevision string) (*verifyBuildResponse, *logical.Response, error) {
	config, err := b.Config(ctx, req.Storage)
	if err != nil {
		return nil, nil, err
	}

	if config.CircleCIToken == "" || config.VCSType == "" || config.Owner == "" {
		return nil, logical.ErrorResponse(
			"configure the circleci credential backend first"), nil
	}

	client := b.GetClient(config.CircleCIToken, config.VCSType, config.Owner)

	if config.BaseURL != "" {
		parsedURL, err := url.Parse(config.BaseURL)
		if err != nil {
			return nil, nil, fmt.Errorf("Successfully parsed base_url when set but failing to parse now: %s", err)
		}
		client.SetBaseURL(parsedURL)
		log.Printf("BaseURL set to: %s", parsedURL.String())
	}

	build, err := client.GetBuild(project, buildNum)
	if err != nil {
		return nil, logical.ErrorResponse(err.Error()), nil
	}

	// Make sure the build is still running
	if build.Lifecycle != "running" {
		return nil, logical.ErrorResponse("circleci build is not currently running"), nil
	}

	// Make sure the hashes match
	if build.VcsRevision != vcsRevision {
		return nil, logical.ErrorResponse("provided VCS revision does not match the revision reported by circleci"), nil
	}

	projectPolicyList, err := b.ProjectMap.Policies(ctx, req.Storage, build.Reponame)

	return &verifyBuildResponse{
		Policies: projectPolicyList,
	}, nil, nil
}

type verifyBuildResponse struct {
	Policies []string
}
