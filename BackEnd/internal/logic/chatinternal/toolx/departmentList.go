package toolx

import (
	"BackEnd/internal/svc"
	"BackEnd/pkg/curl"
	"BackEnd/pkg/token"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
)

type DepartmentList struct {
	svc      *svc.ServiceContext
	callback callbacks.Handler
}

// NewDepartmentList creates a new instance of DepartmentList tool
func NewDepartmentList(svc *svc.ServiceContext) *DepartmentList {
	return &DepartmentList{
		svc:      svc,
		callback: svc.Callbacks,
	}
}

func (t *DepartmentList) Name() string {
	return "department_list"
}

func (t *DepartmentList) Description() string {
	return "Useful for listing all departments and organization structure. No parameters required. The output is a tree structure of departments."
}

func (t *DepartmentList) Call(ctx context.Context, input string) (string, error) {
	if t.callback != nil {
		t.callback.HandleText(ctx, "Listing departments...")
	}

	// 1. Build URL
	host := t.svc.Config.Host
	if host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}
	if !strings.Contains(host, ":") || (strings.Count(host, ":") == 1 && strings.HasPrefix(host, "http")) {
		host = fmt.Sprintf("%s:%d", host, t.svc.Config.Port)
	}

	url := host + "/v1/dep/soa"

	// 2. Request API
	// Since the Frontend calls it without params to get the full tree, we do the same.
	res, err := curl.GetRequest(token.GetTokenStr(ctx), url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch departments: %v", err)
	}

	// 3. Format Output
	return t.formatDepartmentList(res)
}

func (t *DepartmentList) formatDepartmentList(res []byte) (string, error) {
	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Child []DepartmentData `json:"child"`
		} `json:"data"`
	}

	if err := json.Unmarshal(res, &response); err != nil {
		// Fallback: try to print raw response if parsing fails (useful for debug)
		return fmt.Sprintf("Error parsing API response: %v. Raw: %s", err, string(res)), nil
	}

	if response.Code != 200 {
		return "", fmt.Errorf("api error: %s", response.Msg)
	}

	if len(response.Data.Child) == 0 {
		return "No departments found.", nil
	}

	var sb strings.Builder
	sb.WriteString("Department Organization Structure:\n")
	t.buildTreeString(&sb, response.Data.Child, 0)

	return sb.String(), nil
}

// buildTreeString recursively builds the tree string
func (t *DepartmentList) buildTreeString(sb *strings.Builder, nodes []DepartmentData, level int) {
	indent := strings.Repeat("  ", level)
	for _, node := range nodes {
		// Format: - [ID] Name (Leader: xxx)
		leaderInfo := ""
		if node.Leader != "" {
			leaderInfo = fmt.Sprintf(" (Leader: %s)", node.Leader)
		}
		sb.WriteString(fmt.Sprintf("%s- [ID: %s] %s%s\n", indent, node.ID, node.Name, leaderInfo))

		// Recursively process children
		if len(node.Child) > 0 {
			t.buildTreeString(sb, node.Child, level+1)
		}
	}
}

// DepartmentData represents the department node structure
type DepartmentData struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Leader string           `json:"leader"`
	Child  []DepartmentData `json:"child"` // Recursive definition
}
