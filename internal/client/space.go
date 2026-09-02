package client

import (
	"context"
	"fmt"
	"net/url"
)

// Space is a Confluence space.
type Space struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description struct {
		Plain struct {
			Value string `json:"value"`
		} `json:"plain"`
	} `json:"description"`
}

// GetSpace fetches a single space by its key.
func (c *Client) GetSpace(ctx context.Context, key string) (*Space, error) {
	path := fmt.Sprintf("/rest/api/space/%s?%s", url.PathEscape(key), url.Values{
		"expand": {"description.plain"},
	}.Encode())

	var space Space
	if err := c.do(ctx, "GET", path, nil, &space); err != nil {
		return nil, err
	}
	return &space, nil
}
