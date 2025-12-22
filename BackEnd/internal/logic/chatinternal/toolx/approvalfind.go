package toolx

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/svc"
	"BackEnd/pkg/langchain/outputparserx"
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
)

// Reuse the interface from where we would have defined it, or define it here
type ApprovalFindLogic interface {
	List(ctx context.Context, req *domain.ApprovalListReq) (resp *domain.ApprovalListResp, err error)
}

type ApprovalFind struct {
	svc          *svc.ServiceContext
	callback     callbacks.Handler
	outputparser outputparserx.Structured
	logic        ApprovalFindLogic
}

func NewApprovalFind(svc *svc.ServiceContext, l ApprovalFindLogic) *ApprovalFind {
	return &ApprovalFind{
		svc:      svc,
		callback: svc.Callbacks,
		logic:    l,
		outputparser: outputparserx.NewStructured([]outputparserx.ResponseSchema{
			{
				Name:        "userId",
				Description: "The ID of the user to find approvals for.",
				Type:        "string",
			},
			{
				Name:        "type",
				Description: "Approval type (optional). 0=all, 1=leave, 2=go_out, 3=overtime, 4=make_card",
				Type:        "int",
			},
			{
				Name:        "count",
				Description: "Number of records to return (default 5)",
				Type:        "int",
			},
		}),
	}
}

func (t *ApprovalFind) Name() string {
	return "approval_find"
}

func (t *ApprovalFind) Description() string {
	return "Useful for finding approval records for a specific user. Requires 'userId'. If you only have a name, use 'user_list' first to get the ID." + t.outputparser.GetFormatInstructions()
}

func (t *ApprovalFind) Call(ctx context.Context, input string) (string, error) {
	if t.callback != nil {
		t.callback.HandleText(ctx, "Finding approvals: "+input)
	}

	// 1. Parse Input
	params, err := t.outputparser.Parse(input)
	if err != nil {
		return "", err
	}
	p := params.(map[string]any)

	req := &domain.ApprovalListReq{}
	// Defaults
	req.Page = 1
	req.Count = 5

	if v, ok := p["userId"].(string); ok {
		req.UserId = v
	} else {
		// If userId is missing, maybe return error or list mine?
		// Description says "Requires userId".
		return "", fmt.Errorf("missing 'userId'")
	}

	if v, ok := p["type"].(float64); ok {
		req.Type = int(v)
	}
	if v, ok := p["count"].(float64); ok {
		req.Count = int(v)
	}

	// 2. Call Logic
	resp, err := t.logic.List(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to find approvals: %v", err)
	}

	if resp.Count == 0 {
		return "No approval records found for this user.", nil
	}

	// 3. Format Output
	var result []map[string]any
	for _, item := range resp.List {
		statusStr := "Pending"
		if item.Status == 2 {
			statusStr = "Passed"
		} else if item.Status == 3 {
			statusStr = "Rejected"
		}

		typeStr := "Unknown"
		switch item.Type {
		case 1:
			typeStr = "Leave"
		case 2:
			typeStr = "GoOut"
		case 3:
			typeStr = "Overtime"
		case 4:
			typeStr = "MakeCard"
		}

		result = append(result, map[string]any{
			"id":       item.Id,
			"title":    item.Title,
			"type":     typeStr,
			"status":   statusStr,
			"abstract": item.Abstract,
		})
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}
