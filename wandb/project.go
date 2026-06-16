package wandb

import (
	"context"
	"encoding/json"
)

// GetProject fetches a public project by entity and project name.
func (c *Client) GetProject(ctx context.Context, entityName, projectName string) (*Project, error) {
	data, err := c.postGQL(ctx, queryProject, map[string]any{
		"entityName": entityName,
		"name":       projectName,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Project *rawProject `json:"project"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.Project == nil {
		return nil, ErrNotFound
	}
	p := fromRawProject(*resp.Project)
	return &p, nil
}
