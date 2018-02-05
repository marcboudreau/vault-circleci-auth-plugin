package circleci

import (
	"net/url"

	"github.com/jszwedko/go-circleci"
)

// Client is an implementation of the Client interface that actually sends
// API calls to CircleCI.
type Client struct {
	c *circleci.Client
}

// New creates a Client instance with the provided token and returns it.
func New(token string) *Client {
	return &Client{
		c: &circleci.Client{
			Token: token,
		},
	}
}

// GetBuild ...
func (c *Client) GetBuild(user, project string, buildNum int) (*circleci.Build, error) {
	return c.c.GetBuild(user, project, buildNum)
}

// SetBaseURL ...
func (c *Client) SetBaseURL(baseURL *url.URL) {
	c.c.BaseURL = baseURL
}
