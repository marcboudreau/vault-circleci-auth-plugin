package mock

import (
	"net/url"

	circleci "github.com/tylux/go-circleci"
)

// Client is an implementation of the Client interface that mocks out calls to the
// CircleCI API.
type Client struct {
	Build *circleci.Build
	Err   error
}

// GetBuild ...
func (c *Client) GetBuild(project string, buildNum int) (*circleci.Build, error) {
	return c.Build, c.Err
}

// SetBaseURL ...
func (c *Client) SetBaseURL(baseURL *url.URL) {

}
