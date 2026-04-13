package discovery

import (
	"encoding/json"
	"fmt"
)

type ProjectServices struct {
	services map[string]bool
}

func (ps *ProjectServices) Has(name string) bool {
	return ps.services[name]
}

func ParseServices(data []byte) (*ProjectServices, error) {
	var parsed struct {
		Raw struct {
			Services map[string]struct {
				Status string `json:"status"`
			} `json:"services"`
		} `json:"raw"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse ddev describe output: %w", err)
	}
	services := make(map[string]bool)
	for name, svc := range parsed.Raw.Services {
		if svc.Status == "running" {
			services[name] = true
		}
	}
	return &ProjectServices{services: services}, nil
}
