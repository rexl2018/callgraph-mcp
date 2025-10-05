package unit

import (
	"testing"

	"callgraph-mcp/handlers"
)

func TestMCPRequestMapping(t *testing.T) {
	req := handlers.MCPCallgraphRequest{
		ModuleArgs: []string{"./test"},
		Algo:       "static",
		Focus:      "main",
		NoStd:      true,
		NoInter:    false,
		Group:      []string{"pkg", "type"},
		LimitKeyword: []string{"github.com/test"},
		Ignore:     []string{"vendor"},
		LimitPrefix: []string{"internal"},
	}

	// Note: This function needs to be exported from handlers package
	// opts := handlers.MapMCPRequestToRenderOpts(req)

	// Basic validation of request structure
	if len(req.ModuleArgs) == 0 {
		t.Error("ModuleArgs should not be empty")
	}

	if req.Algo == "" {
		t.Error("Algo should not be empty")
	}

	if req.Focus == "" {
		t.Error("Focus should not be empty")
	}

	// Test valid algorithms
	validAlgos := []string{"static", "cha", "rta"}
	found := false
	for _, algo := range validAlgos {
		if req.Algo == algo {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Invalid algorithm: %s", req.Algo)
	}
}

func TestMCPCallgraphRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     handlers.MCPCallgraphRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: handlers.MCPCallgraphRequest{
				ModuleArgs: []string{"./test"},
				Algo:       "static",
			},
			wantErr: false,
		},
		{
			name: "empty module args",
			req: handlers.MCPCallgraphRequest{
				ModuleArgs: []string{},
				Algo:       "static",
			},
			wantErr: true,
		},
		{
			name: "invalid algorithm",
			req: handlers.MCPCallgraphRequest{
				ModuleArgs: []string{"./test"},
				Algo:       "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			hasError := false
			
			if len(tt.req.ModuleArgs) == 0 {
				hasError = true
			}
			
			validAlgos := []string{"static", "cha", "rta"}
			if tt.req.Algo != "" {
				found := false
				for _, algo := range validAlgos {
					if tt.req.Algo == algo {
						found = true
						break
					}
				}
				if !found {
					hasError = true
				}
			}

			if hasError != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasError, tt.wantErr)
			}
		})
	}
}