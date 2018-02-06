package circleci

import (
	"net/url"

	circleci "github.com/tylux/go-circleci"
)

// Client is an implementation of the Client interface that actually sends
// API calls to CircleCI.
type Client struct {
	c *circleci.Client

	vcsType string
	owner   string
}

// New creates a Client instance with the provided token and returns it.
func New(token, vcsType, owner string) *Client {
	return &Client{
		c: &circleci.Client{
			Token: token,
			Debug: true,
		},
		vcsType: vcsType,
		owner:   owner,
	}
}

// GetBuild ...
func (c *Client) GetBuild(project string, buildNum int) (*circleci.Build, error) {
	return c.c.GetBuild(c.vcsType, c.owner, project, buildNum)
}

// SetBaseURL ...
func (c *Client) SetBaseURL(baseURL *url.URL) {
	c.c.BaseURL = baseURL
}
