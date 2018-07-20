package main

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/marcboudreau/vault-circleci-auth-plugin/mock"
	"github.com/stretchr/testify/assert"
	circleci "github.com/tylux/go-circleci"
)

func TestVerifyBuild(t *testing.T) {
	testcases := []struct {
		Backend     *backend
		Req         *logical.Request
		VCSType     string
		User        string
		Project     string
		BuildNum    int
		VCSRevision string

		ExpectError    bool
		ExpectResponse bool
	}{
		{
			Backend: &backend{},
			Req: &logical.Request{
				Storage: &GetErrorStorage{},
			},
			Project:     "p",
			BuildNum:    1,
			VCSRevision: "r",
			ExpectError: true,
		},
		{
			Backend: &backend{},
			Req: &logical.Request{
				Storage: &GetValidStorage{},
			},
			Project:        "p",
			BuildNum:       1,
			VCSRevision:    "r",
			ExpectResponse: true,
		},
		{
			Backend: &backend{
				client: &mock.Client{
					Build: &circleci.Build{
						Lifecycle:   "running",
						VcsRevision: "r",
					},
					Err: nil,
				},
				ProjectMap: &framework.PolicyMap{},
			},
			Req: &logical.Request{
				Storage: &GetValidStorage{
					Entry: &logical.StorageEntry{
						Key:   "config",
						Value: []byte("{\"circleci_token\": \"fake-token\",\"vcs_type\":\"github\",\"owner\":\"fred\"}"),
					},
				},
			},
			Project:        "p",
			BuildNum:       1,
			VCSRevision:    "r",
			ExpectResponse: false,
		},
		{
			Backend: &backend{
				client: &mock.Client{
					Build: &circleci.Build{
						Lifecycle:   "running",
						VcsRevision: "r",
					},
					Err: nil,
				},
				ProjectMap: &framework.PolicyMap{},
			},
			Req: &logical.Request{
				Storage: &GetValidStorage{
					Entry: &logical.StorageEntry{
						Key:   "config",
						Value: []byte("{\"circleci_token\": \"fake-token\",\"vcs_type\":\"github\"}"),
					},
				},
			},
			Project:        "p",
			BuildNum:       1,
			VCSRevision:    "r",
			ExpectResponse: true,
		},
		{
			Backend: &backend{
				client: &mock.Client{
					Build: &circleci.Build{
						Lifecycle:   "running",
						VcsRevision: "r",
					},
					Err: nil,
				},
				ProjectMap: &framework.PolicyMap{},
			},
			Req: &logical.Request{
				Storage: &GetValidStorage{
					Entry: &logical.StorageEntry{
						Key:   "config",
						Value: []byte("{\"circleci_token\": \"fake-token\",\"owner\":\"fred\"}"),
					},
				},
			},
			Project:        "p",
			BuildNum:       1,
			VCSRevision:    "r",
			ExpectResponse: true,
		},
		{
			Backend: &backend{
				client: &mock.Client{
					Build: &circleci.Build{
						Lifecycle:   "running",
						VcsRevision: "r",
					},
					Err: nil,
				},
				ProjectMap: &framework.PolicyMap{},
			},
			Req: &logical.Request{
				Storage: &GetValidStorage{
					Entry: &logical.StorageEntry{
						Key:   "config",
						Value: []byte("{\"circleci_token\":\"fake-token\",\"base_url\":\"https://circleci.com\",\"vcs_type\":\"github\",\"owner\":\"fred\"}"),
					},
				},
			},
			Project:        "p",
			BuildNum:       1,
			VCSRevision:    "r",
			ExpectResponse: false,
		},
		{
			Backend: &backend{
				client: &mock.Client{
					Build: nil,
					Err:   errors.New("SomeError"),
				},
			},
			Req: &logical.Request{
				Storage: &GetValidStorage{
					Entry: &logical.StorageEntry{
						Key:   "config",
						Value: []byte("{\"circleci_token\":\"fake-token\",\"vcs_type\":\"github\",\"owner\":\"fred\"}"),
					},
				},
			},
			Project:        "p",
			BuildNum:       1,
			VCSRevision:    "r",
			ExpectResponse: true,
		},
		{
			Backend: &backend{
				client: &mock.Client{
					Build: &circleci.Build{
						Lifecycle:   "canceled",
						VcsRevision: "r",
					},
					Err: nil,
				},
				ProjectMap: &framework.PolicyMap{},
			},
			Req: &logical.Request{
				Storage: &GetValidStorage{
					Entry: &logical.StorageEntry{
						Key:   "config",
						Value: []byte("{\"circleci_token\": \"fake-token\",\"vcs_type\":\"github\",\"owner\":\"fred\"}"),
					},
				},
			},
			Project:        "p",
			BuildNum:       1,
			VCSRevision:    "r",
			ExpectResponse: true,
		},
		{
			Backend: &backend{
				client: &mock.Client{
					Build: &circleci.Build{
						Lifecycle:   "running",
						VcsRevision: "R",
					},
					Err: nil,
				},
				ProjectMap: &framework.PolicyMap{},
			},
			Req: &logical.Request{
				Storage: &GetValidStorage{
					Entry: &logical.StorageEntry{
						Key:   "config",
						Value: []byte("{\"circleci_token\": \"fake-token\",\"vcs_type\":\"github\",\"owner\":\"fred\"}"),
					},
				},
			},
			Project:        "p",
			BuildNum:       1,
			VCSRevision:    "r",
			ExpectResponse: true,
		},
	}

	for _, tc := range testcases {
		verifyResponse, resp, err := tc.Backend.verifyBuild(context.TODO(), tc.Req, tc.Project, tc.BuildNum, tc.VCSRevision)
		if tc.ExpectError {
			assert.NotNil(t, err)
			assert.Nil(t, resp)
			assert.Nil(t, verifyResponse)
		} else if tc.ExpectResponse {
			assert.Nil(t, err)
			assert.NotNil(t, resp)
			assert.Nil(t, verifyResponse)
		} else {
			assert.Nil(t, err)
		}
	}
}
