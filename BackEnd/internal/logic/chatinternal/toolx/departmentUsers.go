package toolx

import (
	"BackEnd/internal/domain"
	"BackEnd/internal/svc"
	"BackEnd/pkg/langchain/outputparserx"
	"context"
	"fmt" // Correct way to implement, but currently unused in this partial snippet if we remove URL logic

	"github.com/tmc/langchaingo/callbacks"
)

type DepartmentLogic interface {
	SetDepartmentUsers(ctx context.Context, req *domain.SetDepartmentUser) error
}

type DepartmentUsers struct {
	svc          *svc.ServiceContext      // Service Context
	callback     callbacks.Handler        // Callback Handler
	outputparser outputparserx.Structured // Output Parser
	logic        DepartmentLogic          // Interface instead of concrete package ref
}

// NewDepartmentUsers creates a new instance of DepartmentUsers tool
func NewDepartmentUsers(svc *svc.ServiceContext, l DepartmentLogic) *DepartmentUsers {
	return &DepartmentUsers{
		svc:      svc,
		callback: svc.Callbacks,
		// Logic is injected dependency
		logic: l,
		outputparser: outputparserx.NewStructured([]outputparserx.ResponseSchema{
			{
				Name:        "depId",
				Description: "department id to update users for",
				Type:        "string",
			},
			{
				Name:        "userIds",
				Description: "list of user ids to set for the department",
				Type:        "[]string",
			},
		}),
	}
}

func (t *DepartmentUsers) Name() string {
	return "department_users"
}

func (t *DepartmentUsers) Description() string {
	return "Useful for managing users within a department (specifically setting users). Requires 'depId' and 'userIds'." + t.outputparser.GetFormatInstructions()
}

func (t *DepartmentUsers) Call(ctx context.Context, input string) (string, error) {
	if t.callback != nil {
		t.callback.HandleText(ctx, "department users start : "+input)
	}

	// 1. Parse Input
	data, err := t.outputparser.Parse(input)
	if err != nil {
		return "", err
	}

	params := data.(map[string]any)

	// 2. Extract parameters (robustly)
	depId, ok := params["depId"].(string)
	if !ok || depId == "" {
		return "", fmt.Errorf("missing or invalid 'depId'")
	}

	// Handle userIds which might be []interface{} from JSON parsing
	var userIds []string
	if uids, ok := params["userIds"].([]interface{}); ok {
		for _, uid := range uids {
			if s, ok := uid.(string); ok {
				userIds = append(userIds, s)
			}
		}
	} else if uids, ok := params["userIds"].([]string); ok {
		userIds = uids
	} else {
		// Try to handle single string case if LLM messes up
		if uidStr, ok := params["userIds"].(string); ok {
			userIds = append(userIds, uidStr)
		} else {
			return "", fmt.Errorf("missing or invalid 'userIds'")
		}
	}

	// 3. Call Logic Directly (Internal Call)
	// We map the tool's intent to `SetDepartmentUsers` logic
	req := &domain.SetDepartmentUser{
		DepId:   depId,
		UserIds: userIds,
	}

	if err := t.logic.SetDepartmentUsers(ctx, req); err != nil {
		return "", fmt.Errorf("failed to set department users: %v", err)
	}

	return "Successfully updated department users.", nil
}
