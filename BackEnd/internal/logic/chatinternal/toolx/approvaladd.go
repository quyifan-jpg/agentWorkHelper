package toolx

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/svc"
	"BackEnd/pkg/langchain/outputparserx"
	"context"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/callbacks"
)

type ApprovalAddLogic interface {
	Create(ctx context.Context, req *domain.Approval) (resp *domain.IdResp, err error)
}

type ApprovalAdd struct {
	svc          *svc.ServiceContext
	callback     callbacks.Handler
	outputparser outputparserx.Structured
	logic        ApprovalAddLogic
}

func NewApprovalAdd(svc *svc.ServiceContext, l ApprovalAddLogic) *ApprovalAdd {
	return &ApprovalAdd{
		svc:      svc,
		callback: svc.Callbacks,
		logic:    l,
		outputparser: outputparserx.NewStructured([]outputparserx.ResponseSchema{
			{
				Name:        "type",
				Description: "approval type: 1=Leave, 2=GoOut, 4=MakeCard(ClockIn Fix)",
				Type:        "int",
			},
			{
				Name:        "reason",
				Description: "reason for approval",
				Type:        "string",
			},
			// Mixed fields for different types
			{
				Name:        "startTime",
				Description: "start time (yyyy-MM-dd HH:mm:ss) for Leave/GoOut",
				Type:        "string",
			},
			{
				Name:        "endTime",
				Description: "end time (yyyy-MM-dd HH:mm:ss) for Leave/GoOut",
				Type:        "string",
			},
			{
				Name:        "leaveType",
				Description: "leave type: 1=Personal, 2=Sick, 3=Annual, etc.",
				Type:        "int",
			},
			{
				Name:        "date",
				Description: "date (yyyy-MM-dd) for MakeCard",
				Type:        "string",
			},
		}),
	}
}

func (t *ApprovalAdd) Name() string {
	return "approval_add"
}

func (t *ApprovalAdd) Description() string {
	return "Useful for submitted a new approval application (leave, go out, clock-in fix). Requires distinct parameters based on type." + t.outputparser.GetFormatInstructions()
}

func (t *ApprovalAdd) Call(ctx context.Context, input string) (string, error) {
	if t.callback != nil {
		t.callback.HandleText(ctx, "Creating approval: "+input)
	}

	// 1. Parse Input
	params, err := t.outputparser.Parse(input)
	if err != nil {
		return "", err
	}
	p := params.(map[string]any)

	req := &domain.Approval{}

	// Extract Type
	if v, ok := p["type"].(float64); ok {
		req.Type = int(v)
	} else {
		return "", fmt.Errorf("missing or invalid 'type'")
	}

	// Extract Reason
	if v, ok := p["reason"].(string); ok {
		req.Reason = v
	}

	// Helper for time parsing
	parseTime := func(s string) int64 {
		// Try full time format
		t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
		if err == nil {
			return t.Unix()
		}
		// Try date format
		t, err = time.ParseInLocation("2006-01-02", s, time.Local)
		if err == nil {
			return t.Unix()
		}
		return 0
	}

	// Construct Details based on Type
	switch req.Type {
	case 1: // Leave
		startTimeStr := getString(p, "startTime")
		endTimeStr := getString(p, "endTime")
		leaveType := int(getFloat(p, "leaveType"))

		if startTimeStr == "" || endTimeStr == "" {
			return "", fmt.Errorf("startTime and endTime are required for Leave")
		}

		req.Leave = &domain.Leave{
			Type:      leaveType,
			StartTime: parseTime(startTimeStr),
			EndTime:   parseTime(endTimeStr),
			Reason:    req.Reason,
		}

	case 2: // GoOut
		startTimeStr := getString(p, "startTime")
		endTimeStr := getString(p, "endTime")

		if startTimeStr == "" || endTimeStr == "" {
			return "", fmt.Errorf("startTime and endTime are required for GoOut")
		}

		req.GoOut = &domain.GoOut{
			StartTime: parseTime(startTimeStr),
			EndTime:   parseTime(endTimeStr),
			Reason:    req.Reason,
		}

	case 4: // MakeCard (ClockIn Fix)
		dateStr := getString(p, "date")
		if dateStr == "" {
			return "", fmt.Errorf("date is required for MakeCard")
		}

		req.MakeCard = &domain.MakeCard{
			Date:   parseTime(dateStr),
			Reason: req.Reason,
		}

	default:
		return "", fmt.Errorf("unsupported approval type: %d", req.Type)
	}

	// Set status to Auditing (1) by default
	req.Status = 1

	// 2. Call Logic
	resp, err := t.logic.Create(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create approval: %v", err)
	}

	return fmt.Sprintf("Successfully created approval. ID: %s", resp.Id), nil
}

// Helpers
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]any, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}
