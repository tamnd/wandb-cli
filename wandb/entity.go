package wandb

import (
	"context"
	"encoding/json"
)

// GetEntity fetches a public entity (user or team) profile by name.
func (c *Client) GetEntity(ctx context.Context, name string) (*Entity, error) {
	data, err := c.postGQL(ctx, queryEntity, map[string]any{"name": name})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Entity *rawEntity `json:"entity"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.Entity == nil {
		return nil, ErrNotFound
	}
	e := fromRawEntity(*resp.Entity)
	return &e, nil
}
