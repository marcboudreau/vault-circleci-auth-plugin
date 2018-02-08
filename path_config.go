package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathConfig(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"circleci_token": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The CircleCI access token that allows this plugin to make CircleCI API calls to verify the authentication information.",
			},
			"base_url": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The base URL used to construct all endpoint URLs for this plugin.",
			},
			"ttl": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Duration of the token's lifetime, unless renewed.",
			},
			"max_ttl": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Maximum duration of the token's lifetime.",
			},
			"vcs_type": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The version control system type where the project is hosted.  Supported values are github and bitbucket.",
			},
			"owner": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The user or organization that owns the project in the VCS.",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathConfigRead,
			logical.UpdateOperation: b.pathConfigWrite,
		},
	}
}

func (b *backend) pathConfigRead(req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	config, err := b.Config(req.Storage)
	if err != nil {
		return nil, err
	}

	config.TTL /= time.Second
	config.MaxTTL /= time.Second

	return &logical.Response{
		Data: map[string]interface{}{
			"circleci_token": config.CircleCIToken,
			"base_url":       config.BaseURL,
			"ttl":            config.TTL,
			"max_ttl":        config.MaxTTL,
			"vcs_type":       config.VCSType,
			"owner":          config.Owner,
		},
	}, nil
}

func (b *backend) pathConfigWrite(req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	circleCIToken := d.Get("circleci_token").(string)
	baseURL := d.Get("base_url").(string)
	if len(baseURL) > 0 {
		// Try parsing the URL to make sure that it's valid.
		if _, err := url.Parse(baseURL); err != nil {
			return logical.ErrorResponse(fmt.Sprintf("Error parsing given base_url: %s", err)), nil
		}
	}

	ttl, err := parseDurationField("ttl", d)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Invalid 'ttl': %s", err)), nil
	}

	maxTTL, err := parseDurationField("max_ttl", d)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Invalid 'max_ttl': %s", err)), nil
	}

	vcsType := d.Get("vcs_type").(string)
	owner := d.Get("owner").(string)

	entry, err := logical.StorageEntryJSON("config", config{
		CircleCIToken: circleCIToken,
		BaseURL:       baseURL,
		TTL:           ttl,
		MaxTTL:        maxTTL,
		VCSType:       vcsType,
		Owner:         owner,
	})
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(entry); err != nil {
		return nil, err
	}

	// Clear out the client so that it gets reconstructed using the new vcs_type and owner values.
	b.client = nil

	return nil, nil
}

func parseDurationField(fieldName string, d *framework.FieldData) (time.Duration, error) {
	var value time.Duration
	var err error

	raw, ok := d.GetOk(fieldName)
	if !ok || len(raw.(string)) == 0 {
		value = 0
	} else {
		value, err = time.ParseDuration(raw.(string))
	}

	return value, err
}

// Config reads the config object out of the provided Storage.
func (b *backend) Config(s logical.Storage) (*config, error) {
	entry, err := s.Get("config")
	if err != nil {
		return nil, err
	}

	var result config
	if entry != nil {
		if err := entry.DecodeJSON(&result); err != nil {
			return nil, fmt.Errorf("error reading configuration: %s", err)
		}
	}

	return &result, nil
}

type config struct {
	CircleCIToken string        `json:"circleci_token"`
	BaseURL       string        `json:"base_url"`
	TTL           time.Duration `json:"ttl"`
	MaxTTL        time.Duration `json:"max_ttl"`
	VCSType       string        `json:"vcs_type"`
	Owner         string        `json:"owner"`
}
