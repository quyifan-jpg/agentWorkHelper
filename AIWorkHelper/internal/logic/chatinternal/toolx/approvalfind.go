/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package toolx

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/curl"
	"AIWorkHelper/pkg/httpx"
	"AIWorkHelper/pkg/langchain/outputparserx"
	"AIWorkHelper/pkg/token"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
)

type ApprovalFind struct {
	svc          *svc.ServiceContext
	Callback     callbacks.Handler
	outputparser outputparserx.Structured
}

func NewApprovalFind(svc *svc.ServiceContext) *ApprovalFind {
	return &ApprovalFind{
		svc:      svc,
		Callback: svc.Callbacks,
		outputparser: outputparserx.NewStructured([]outputparserx.ResponseSchema{
			{
				Name:        "type",
				Description: "approval type; enum : 0. General, 1. Leave, 2. Card replacement, 3. Go out, 4. Reimbursement, 5. Payment, 6. Purchase, 7. Collection; number to be completed",
				Type:        "int",
			}, {
				Name:        "id",
				Description: "approval id",
			}, {
				Name:        "status",
				Description: "approval status; enum : 0. No beginning, 1. In progress 2. Done-Passed, 3. Revocation, 4. refused; number to be completed",
			}, {
				Name:        "createId",
				Description: "id of creator",
			},
		}),
	}
}

func (a *ApprovalFind) Name() string {
	return "approval_find"
}

func (a *ApprovalFind) Description() string {
	return `
	a approval find interface.
	use when you need to find, query, search or list approvals.
	use when user asks: "我的审批", "查询审批", "有哪些审批", "审批单", "find my approvals", "审批进度", "审批状态", etc.
	If user doesn't provide specific conditions (id, type, status, createId), query all approvals of current user by leaving those fields empty.
	If the condition is null, return {}
	keep Chinese output.
` + a.outputparser.GetFormatInstructions()
}

func (a *ApprovalFind) Call(ctx context.Context, input string) (string, error) {
	if a.Callback != nil {
		a.Callback.HandleText(ctx, "approval find start input : "+input)
	}

	out, err := a.outputparser.Parse(input)
	if err != nil {
		return "", err
	}

	data := out.(map[string]any)
	if data == nil {
		data = make(map[string]any)
	}
	// AI查询时默认查询当前用户提交的审批（Type=1表示"我提交的"）
	// 设置 type=1 和 userId=当前用户ID
	data["type"] = 1                   // 查询"我提交的"审批
	data["userId"] = token.GetUId(ctx) // 当前用户ID

	// 设置查询数量限制，避免返回过多数据
	if data["count"] == nil {
		data["count"] = 10
	}

	res, err := curl.GetRequest(token.GetTokenStr(ctx), a.svc.Config.Host+"/v1/approval/list", data)
	if err != nil {
		return "", err
	}

	if a.Callback != nil {
		a.Callback.HandleText(ctx, "approval find end data : "+string(res))
	}

	// 解析API响应并格式化输出（对标Java版本的handleFindApproval方法）
	return a.formatApprovalList(res)
}

// formatApprovalList 格式化审批列表输出
// 对标Java版本的handleFindApproval方法（ApprovalAIHandler.java:326-371）
func (a *ApprovalFind) formatApprovalList(res []byte) (string, error) {
	// 解析HTTP响应
	var apiResponse httpx.Response
	if err := json.Unmarshal(res, &apiResponse); err != nil {
		return "", err
	}

	// 检查响应状态码
	if apiResponse.Code != 200 {
		return "", errors.New(apiResponse.Msg)
	}

	// 解析审批列表数据
	var listResp domain.ApprovalListResp
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(dataBytes, &listResp); err != nil {
		return "", err
	}

	// 如果没有审批记录
	if listResp.List == nil || len(listResp.List) == 0 {
		return "您当前没有审批记录。", nil
	}

	// 格式化输出审批列表（对标Java版本第343-365行）
	var result strings.Builder
	result.WriteString(fmt.Sprintf("📋 找到 %d 个审批单:\n\n", len(listResp.List)))

	for i, approval := range listResp.List {
		// 格式化序号和标题
		result.WriteString(fmt.Sprintf("%d. 【%s 提交的】%s\n",
			i+1,
			getCreatorName(approval),
			getApprovalTitle(approval)))

		// 格式化状态
		result.WriteString(fmt.Sprintf("   📌 类型: %s\n", getApprovalTypeName(approval.Type)))
		result.WriteString(fmt.Sprintf("   📊 状态: %s\n", getApprovalStatusName(approval.Status)))

		// 格式化创建时间
		result.WriteString(fmt.Sprintf("   🕐 创建时间: %s\n",
			formatTimestamp(approval.CreateAt)))

		// 根据审批类型显示详细信息
		result.WriteString(fmt.Sprintf("   📝 详情: 【%s】 %s\n",
			approval.Title, approval.Abstract))

		result.WriteString("\n")
	}

	return result.String(), nil
}

// getCreatorName 获取创建者名称
func getCreatorName(approval *domain.ApprovalList) string {
	// 从标题中提取创建者名称（标题格式通常为"【创建者】xxx审批"）
	title := approval.Title
	if strings.Contains(title, "【") && strings.Contains(title, "】") {
		start := strings.Index(title, "【") + len("【")
		end := strings.Index(title, "】")
		if start < end {
			return title[start:end]
		}
	}
	return "未知"
}

// getApprovalTitle 获取审批标题
func getApprovalTitle(approval *domain.ApprovalList) string {
	return approval.Title
}

// getApprovalTypeName 获取审批类型名称（对标Java版本ApprovalType枚举）
func getApprovalTypeName(approvalType int) string {
	switch approvalType {
	case 1:
		return "通用审批"
	case 2:
		return "请假审批"
	case 3:
		return "补卡审批"
	case 4:
		return "外出审批"
	case 5:
		return "报销审批"
	case 6:
		return "付款审批"
	case 7:
		return "采购审批"
	case 8:
		return "收款审批"
	default:
		return "其他"
	}
}

// getApprovalStatusName 获取审批状态名称（对标Java版本ApprovalStatus枚举）
func getApprovalStatusName(status int) string {
	switch status {
	case 0:
		return "未开始"
	case 1:
		return "进行中"
	case 2:
		return "已通过"
	case 3:
		return "已撤销"
	case 4:
		return "已拒绝"
	default:
		return "未知状态"
	}
}

// formatTimestamp 格式化时间戳（对标Java版本第395-404行）
func formatTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "未设置"
	}
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04")
}
