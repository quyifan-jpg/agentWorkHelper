/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package toolx

import (
	"AIWorkHelper/internal/domain"
	"AIWorkHelper/internal/logic/chatinternal/toolx/approval"
	"AIWorkHelper/internal/model"
	"AIWorkHelper/internal/svc"
	"AIWorkHelper/pkg/curl"
	"AIWorkHelper/pkg/langchain/outputparserx"
	"AIWorkHelper/pkg/token"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
)

type ApprovalAdd struct {
	svc          *svc.ServiceContext
	Callback     callbacks.Handler
	outputparser outputparserx.Structured
}

func NewApprovalAdd(svc *svc.ServiceContext) *ApprovalAdd {
	return &ApprovalAdd{
		svc:      svc,
		Callback: svc.Callbacks,
		outputparser: outputparserx.NewStructured([]outputparserx.ResponseSchema{
			{
				Name:        "type",
				Description: "approval type; enum: 2=Leave Approval, 3=Card replacement Approval, 4=Go out Approval",
				Type:        "int",
			}, {
				Name:        "leaveType",
				Description: "For Leave Approval only. Leave type: 1=事假, 2=调休, 3=病假, 4=年假, 5=产假, 6=陪产假, 7=婚假, 8=丧假, 9=哺乳假",
				Type:        "int",
			}, {
				Name:        "startTime",
				Description: "For Leave Approval. Start time Unix timestamp in seconds. Use time_parser tool result.",
				Type:        "int64",
			}, {
				Name:        "endTime",
				Description: "For Leave Approval. End time Unix timestamp in seconds. Use time_parser tool result.",
				Type:        "int64",
			}, {
				Name:        "reason",
				Description: "Reason for approval. Extract from user input.",
			}, {
				Name:        "timeType",
				Description: "For Leave Approval. Duration type: 1=小时(Hours), 2=天(Days). Calculate from time difference.",
				Type:        "int",
			}, {
				Name:        "makeCardDateStr",
				Description: "For Card replacement Approval only. REQUIRED. The date-time string from user input. Extract the exact date-time string (e.g., '2025-10-22 17:00', '2025年10月22日17点', '10月22日17点'). If user only provides relative time (e.g., '昨天下班'), use time_parser tool first and put the result timestamp in makeCardDate field instead.",
				Type:        "string",
			}, {
				Name:        "makeCardDate",
				Description: "For Card replacement Approval only. Unix timestamp for relative time expressions. Only use this if user provides relative time (e.g., '昨天下班') - use time_parser tool. Leave empty if makeCardDateStr is provided.",
				Type:        "int64",
			}, {
				Name:        "workCheckType",
				Description: "For Card replacement Approval only. Check type: 1=上班卡(On-work check), 2=下班卡(Off-work check). Infer from user's description (下班=2, 上班=1).",
				Type:        "int",
			}, {
				Name:        "goOutStartTimeStr",
				Description: "For Go out Approval only. The start time string from user input. Extract the exact date-time string (e.g., '2025-10-24 09:00', '2025年10月24日9点', '10月24日9点'). If user provides relative time (e.g., '明天上午9点'), use time_parser tool first and put the result in goOutStartTime field instead.",
				Type:        "string",
			}, {
				Name:        "goOutEndTimeStr",
				Description: "For Go out Approval only. The end time string from user input. Extract the exact date-time string (e.g., '2025-10-24 11:00', '2025年10月24日11点', '10月24日11点'). If user provides relative time (e.g., '明天上午11点'), use time_parser tool first and put the result in goOutEndTime field instead.",
				Type:        "string",
			}, {
				Name:        "goOutStartTime",
				Description: "For Go out Approval only. Start time Unix timestamp for relative time expressions. Only use this if user provides relative time - use time_parser tool. Leave empty if goOutStartTimeStr is provided.",
				Type:        "int64",
			}, {
				Name:        "goOutEndTime",
				Description: "For Go out Approval only. End time Unix timestamp for relative time expressions. Only use this if user provides relative time - use time_parser tool. Leave empty if goOutEndTimeStr is provided.",
				Type:        "int64",
			}, {
				Name:        "goOutReason",
				Description: "For Go out Approval only. REQUIRED. The reason for going out. Extract from user input (e.g., '拜访客户', '外出办事', '出差').",
				Type:        "string",
			},
		}),
	}
}

func (a *ApprovalAdd) Name() string {
	return "approval_add"
}

func (a *ApprovalAdd) Description() string {
	return `
	An approval creation interface.

	CRITICAL TIME HANDLING RULES:
	1. If user provides SPECIFIC date-time (e.g., "2025-10-22 17:00", "2025年10月22日17点", "10月22日下午5点"):
	   - Parse it DIRECTLY to Unix timestamp
	   - DO NOT use time_parser tool
	   - Support formats: "YYYY-MM-DD HH:mm", "YYYY年MM月DD日HH点", "MM月DD日 HH:mm"

	2. If user provides RELATIVE time (e.g., "明天下午2点", "后天9点", "下周一"):
	   - Use time_parser tool to convert to Unix timestamp

	3. For Card replacement Approval - time handling:
	   - If user says "2025-10-22 17:00下班忘记打卡" → makeCardDate = parse("2025-10-22 17:00") to Unix timestamp
	   - If user says "昨天下班忘记打卡" → use time_parser with "昨天下班"
	   - Priority: Extract specific date-time first, then use time_parser for relative expressions

	If any required information is MISSING or UNCLEAR:
	- DO NOT call this tool yet
	- Respond to user in Chinese with:
	  1. What information is missing
	  2. Provide an example in the correct format

	Example response when information is incomplete (for Card replacement):
	"请提供以下信息来创建补卡审批：
	1. 补卡日期时间（格式：YYYY-MM-DD HH:mm，如 2025-10-22 17:00）
	2. 打卡类型（上班卡/下班卡）
	3. 补卡原因

	示例：2025-10-22 17:00 下班忘记打卡了"

	For Card replacement Approval (type=3), required fields:
	- makeCardDate (补卡日期时间):
	  * Format: Unix timestamp in seconds
	  * User should provide specific date-time like "2025-10-22 17:00" or "10月22日17点"
	  * Parse formats like "YYYY-MM-DD HH:mm", "YYYY年MM月DD日HH点", "MM月DD日HH:mm" DIRECTLY to timestamp
	  * Only use time_parser for relative expressions like "昨天下班", "明天上班"

	- workCheckType (打卡类型):
	  * 1=上班卡 (On-work check)
	  * 2=下班卡 (Off-work check)
	  * Infer from keywords: "下班"→2, "上班"→1

	- reason (补卡原因):
	  * Extract from user input
	  * Common valid reasons: "忘记打卡", "卡坏了", "手机没电"
	  * DO NOT ask for additional details if reason is provided

	For Leave Approval (type=2), required fields:
	- startTime, endTime: Unix timestamps (parse specific dates directly, use time_parser for relative dates)
	- Leave type (请假类型): 1=事假, 2=调休, 3=病假, 4=年假, 5=产假, 6=陪产假, 7=婚假, 8=丧假, 9=哺乳假
	- reason (原因)
	- timeType: 1=小时(Hours), 2=天(Days)

	For Go out Approval (type=4), required fields:
	- goOutStartTimeStr / goOutStartTime (开始时间):
	  * If user provides specific date-time (e.g., "2025-10-24 09:00"), extract it as goOutStartTimeStr string
	  * If user provides relative time (e.g., "明天上午9点"), use time_parser tool first, then use goOutStartTime timestamp
	  * Supported formats: "YYYY-MM-DD HH:mm", "YYYY年MM月DD日HH点", "MM月DD日HH:mm"

	- goOutEndTimeStr / goOutEndTime (结束时间):
	  * Same format requirements as start time
	  * Must be AFTER start time

	- goOutReason (外出原因):
	  * REQUIRED. Extract from user input
	  * Common valid reasons: "拜访客户", "外出办事", "出差", "见客户"
	  * DO NOT ask for additional details if reason is provided

	After extracting all required information, call this tool to create the approval.
	Keep Chinese output.
` + a.outputparser.GetFormatInstructions()
}

func (a *ApprovalAdd) Call(ctx context.Context, input string) (string, error) {
	if a.Callback != nil {
		a.Callback.HandleText(ctx, "approval add start input : "+input)
	}

	out, err := a.outputparser.Parse(input)
	if err != nil {
		return "", err
	}
	data := out.(map[string]any)

	var approvalType float64
	if t, ok := data["type"]; ok {
		approvalType = t.(float64)
	}

	// 如果是请假审批，直接使用AI提取的结构化数据
	if model.ApprovalType(approvalType) == model.LeaveApproval {
		return a.createLeaveApproval(ctx, data)
	}

	// 如果是补卡审批，直接使用AI提取的结构化数据
	if model.ApprovalType(approvalType) == model.MakeCardApproval {
		return a.createMakeCardApproval(ctx, data)
	}

	// 如果是外出审批，直接使用AI提取的结构化数据
	if model.ApprovalType(approvalType) == model.GoOutApproval {
		return a.createGoOutApproval(ctx, data)
	}

	// 其他类型的审批保持原有逻辑
	ap, err := approval.NewApproval(a.svc, model.ApprovalType(approvalType))
	if err != nil {
		return "", err
	}

	// 构造input（对于非请假审批）
	userInput := ""
	if v, ok := data["reason"]; ok {
		userInput = v.(string)
	}

	id, err := ap.Create(ctx, userInput)
	if err != nil {
		return "", err
	}

	return Success + "\ncreated approval id : " + id, nil
}

// createLeaveApproval 直接创建请假审批，绕过LLMChain
func (a *ApprovalAdd) createLeaveApproval(ctx context.Context, data map[string]any) (string, error) {
	// 提取AI已经解析好的数据
	var leaveData domain.Leave

	fmt.Printf("[DEBUG] 收到的数据: %+v\n", data)

	if v, ok := data["leaveType"]; ok {
		if val, ok := v.(float64); ok {
			leaveData.Type = int(val)
			fmt.Printf("[DEBUG] 请假类型: %d\n", leaveData.Type)
		}
	}
	if v, ok := data["startTime"]; ok {
		if val, ok := v.(float64); ok {
			leaveData.StartTime = int64(val)
			fmt.Printf("[DEBUG] 开始时间: %d\n", leaveData.StartTime)
		}
	}
	if v, ok := data["endTime"]; ok {
		if val, ok := v.(float64); ok {
			leaveData.EndTime = int64(val)
			fmt.Printf("[DEBUG] 结束时间: %d\n", leaveData.EndTime)
		}
	}
	if v, ok := data["reason"]; ok {
		if val, ok := v.(string); ok {
			leaveData.Reason = val
			fmt.Printf("[DEBUG] 请假原因: %s\n", leaveData.Reason)
		}
	}
	if v, ok := data["timeType"]; ok {
		if val, ok := v.(float64); ok {
			leaveData.TimeType = int(val)
			fmt.Printf("[DEBUG] 时长类型: %d\n", leaveData.TimeType)
		}
	}

	// 构造审批请求
	req := domain.Approval{
		Type:  int(model.LeaveApproval),
		Leave: &leaveData,
	}

	fmt.Printf("[DEBUG] 完整审批数据: %+v\n", req)
	fmt.Printf("[DEBUG] Leave详情: Type=%d, StartTime=%d, EndTime=%d, Reason=%s, TimeType=%d\n",
		leaveData.Type, leaveData.StartTime, leaveData.EndTime, leaveData.Reason, leaveData.TimeType)

	// 直接调用API创建审批
	addRes, err := curl.PostRequest(token.GetTokenStr(ctx), a.svc.Config.Host+"/v1/approval", req)
	if err != nil {
		return "", fmt.Errorf("failed to create approval: %w", err)
	}

	fmt.Printf("[DEBUG] API响应: %s\n", string(addRes))

	var idResp domain.IdRespInfo
	if err := json.Unmarshal(addRes, &idResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return Success + "\ncreated approval id : " + idResp.Data.Id, nil
}

// parseDateTimeString 解析各种格式的日期时间字符串为 Unix 时间戳
func parseDateTimeString(dateStr string) (int64, error) {
	dateStr = strings.TrimSpace(dateStr)
	now := time.Now()
	loc := now.Location()

	// 支持的时间格式列表
	formats := []struct {
		pattern string
		layout  string
	}{
		// 标准格式: 2025-10-22 17:00
		{`^\d{4}-\d{2}-\d{2}\s+\d{1,2}:\d{2}$`, "2006-01-02 15:04"},
		// 中文格式: 2025年10月22日17点 或 2025年10月22日17点30分
		{`^\d{4}年\d{1,2}月\d{1,2}日\d{1,2}点`, ""},
		// 简短格式: 10月22日17点 或 10-22 17:00
		{`^\d{1,2}月\d{1,2}日\d{1,2}点`, ""},
		{`^\d{1,2}-\d{1,2}\s+\d{1,2}:\d{2}$`, ""},
	}

	// 尝试标准格式: 2025-10-22 17:00
	if matched, _ := regexp.MatchString(formats[0].pattern, dateStr); matched {
		t, err := time.ParseInLocation(formats[0].layout, dateStr, loc)
		if err == nil {
			return t.Unix(), nil
		}
	}

	// 尝试解析中文格式: 2025年10月22日17点30分 或 2025年10月22日17点
	if matched, _ := regexp.MatchString(`^\d{4}年\d{1,2}月\d{1,2}日`, dateStr); matched {
		re := regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})日(\d{1,2})点(?:(\d{1,2})分)?`)
		matches := re.FindStringSubmatch(dateStr)
		if len(matches) >= 5 {
			year, _ := strconv.Atoi(matches[1])
			month, _ := strconv.Atoi(matches[2])
			day, _ := strconv.Atoi(matches[3])
			hour, _ := strconv.Atoi(matches[4])
			minute := 0
			if len(matches) > 5 && matches[5] != "" {
				minute, _ = strconv.Atoi(matches[5])
			}
			t := time.Date(year, time.Month(month), day, hour, minute, 0, 0, loc)
			return t.Unix(), nil
		}
	}

	// 尝试解析简短中文格式: 10月22日17点 或 10月22日17点30分
	if matched, _ := regexp.MatchString(`^\d{1,2}月\d{1,2}日`, dateStr); matched {
		re := regexp.MustCompile(`(\d{1,2})月(\d{1,2})日(\d{1,2})点(?:(\d{1,2})分)?`)
		matches := re.FindStringSubmatch(dateStr)
		if len(matches) >= 4 {
			month, _ := strconv.Atoi(matches[1])
			day, _ := strconv.Atoi(matches[2])
			hour, _ := strconv.Atoi(matches[3])
			minute := 0
			if len(matches) > 4 && matches[4] != "" {
				minute, _ = strconv.Atoi(matches[4])
			}
			// 使用当前年份
			t := time.Date(now.Year(), time.Month(month), day, hour, minute, 0, 0, loc)
			return t.Unix(), nil
		}
	}

	// 尝试解析简短格式: 10-22 17:00
	if matched, _ := regexp.MatchString(`^\d{1,2}-\d{1,2}\s+\d{1,2}:\d{2}$`, dateStr); matched {
		re := regexp.MustCompile(`(\d{1,2})-(\d{1,2})\s+(\d{1,2}):(\d{2})`)
		matches := re.FindStringSubmatch(dateStr)
		if len(matches) >= 5 {
			month, _ := strconv.Atoi(matches[1])
			day, _ := strconv.Atoi(matches[2])
			hour, _ := strconv.Atoi(matches[3])
			minute, _ := strconv.Atoi(matches[4])
			// 使用当前年份
			t := time.Date(now.Year(), time.Month(month), day, hour, minute, 0, 0, loc)
			return t.Unix(), nil
		}
	}

	return 0, fmt.Errorf("unsupported date format: %s", dateStr)
}

// createMakeCardApproval 直接创建补卡审批，绕过LLMChain
func (a *ApprovalAdd) createMakeCardApproval(ctx context.Context, data map[string]any) (string, error) {
	// 提取AI已经解析好的数据
	var makeCardData domain.MakeCard

	fmt.Printf("[DEBUG] 补卡审批收到的数据: %+v\n", data)

	// 参数验证
	var missingFields []string

	// 检查补卡时间
	var hasValidDate bool
	// 优先使用 makeCardDateStr (用户提供的具体时间字符串)
	if v, ok := data["makeCardDateStr"]; ok && v != nil {
		if dateStr, ok := v.(string); ok && dateStr != "" {
			timestamp, err := parseDateTimeString(dateStr)
			if err != nil {
				return "", fmt.Errorf("时间格式解析失败: %s, 请使用格式如: 2025-10-22 17:00", dateStr)
			}
			makeCardData.Date = timestamp
			hasValidDate = true
			fmt.Printf("[DEBUG] 从字符串解析补卡时间: %s -> %d\n", dateStr, timestamp)
		}
	}

	// 如果没有 makeCardDateStr，尝试使用 makeCardDate (time_parser 返回的时间戳)
	if !hasValidDate {
		if v, ok := data["makeCardDate"]; ok && v != nil {
			if val, ok := v.(float64); ok && val > 0 {
				makeCardData.Date = int64(val)
				hasValidDate = true
				fmt.Printf("[DEBUG] 使用时间戳补卡时间: %d\n", makeCardData.Date)
			}
		}
	}

	if !hasValidDate {
		missingFields = append(missingFields, "补卡日期时间")
	}

	// 检查打卡类型
	if v, ok := data["workCheckType"]; ok && v != nil {
		if val, ok := v.(float64); ok && val > 0 {
			makeCardData.CheckType = int(val)
			fmt.Printf("[DEBUG] 打卡类型: %d\n", makeCardData.CheckType)
		} else {
			missingFields = append(missingFields, "打卡类型(上班卡/下班卡)")
		}
	} else {
		missingFields = append(missingFields, "打卡类型(上班卡/下班卡)")
	}

	// 检查补卡原因
	if v, ok := data["reason"]; ok && v != nil {
		if val, ok := v.(string); ok && val != "" {
			makeCardData.Reason = val
			fmt.Printf("[DEBUG] 补卡原因: %s\n", makeCardData.Reason)
		} else {
			missingFields = append(missingFields, "补卡原因")
		}
	} else {
		missingFields = append(missingFields, "补卡原因")
	}

	// 如果有缺失字段，返回提示信息
	if len(missingFields) > 0 {
		errorMsg := fmt.Sprintf("创建补卡审批缺少必要信息：%s\n\n", strings.Join(missingFields, "、"))
		errorMsg += "请提供以下信息：\n"
		errorMsg += "1. 补卡日期时间（格式：YYYY-MM-DD HH:mm，如 2025-10-22 17:00）\n"
		errorMsg += "2. 打卡类型（上班卡/下班卡）\n"
		errorMsg += "3. 补卡原因\n\n"
		errorMsg += "示例：2025-10-22 17:00 下班忘记打卡了"
		return "", fmt.Errorf(errorMsg)
	}

	// 构造审批请求
	req := domain.Approval{
		Type:     int(model.MakeCardApproval),
		MakeCard: &makeCardData,
	}

	fmt.Printf("[DEBUG] 完整补卡审批数据: %+v\n", req)
	fmt.Printf("[DEBUG] MakeCard详情: Date=%d, CheckType=%d, Reason=%s\n",
		makeCardData.Date, makeCardData.CheckType, makeCardData.Reason)

	// 直接调用API创建审批
	addRes, err := curl.PostRequest(token.GetTokenStr(ctx), a.svc.Config.Host+"/v1/approval", req)
	if err != nil {
		return "", fmt.Errorf("failed to create make card approval: %w", err)
	}

	fmt.Printf("[DEBUG] 补卡审批API响应: %s\n", string(addRes))

	var idResp domain.IdRespInfo
	if err := json.Unmarshal(addRes, &idResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return Success + "\ncreated approval id : " + idResp.Data.Id, nil
}

// createGoOutApproval 直接创建外出审批，绕过LLMChain
func (a *ApprovalAdd) createGoOutApproval(ctx context.Context, data map[string]any) (string, error) {
	var goOutData domain.GoOut
	fmt.Printf("[DEBUG] 外出审批收到的数据: %+v\n", data)

	var missingFields []string
	var hasValidStartTime bool
	if v, ok := data["goOutStartTimeStr"]; ok && v != nil {
		if dateStr, ok := v.(string); ok && dateStr != "" {
			timestamp, err := parseDateTimeString(dateStr)
			if err != nil {
				return "", fmt.Errorf("开始时间格式解析失败: %s, 请使用格式如: 2025-10-24 09:00", dateStr)
			}
			goOutData.StartTime = timestamp
			hasValidStartTime = true
			fmt.Printf("[DEBUG] 从字符串解析开始时间: %s -> %d\n", dateStr, timestamp)
		}
	}
	if !hasValidStartTime {
		if v, ok := data["goOutStartTime"]; ok && v != nil {
			if val, ok := v.(float64); ok && val > 0 {
				goOutData.StartTime = int64(val)
				hasValidStartTime = true
				fmt.Printf("[DEBUG] 使用时间戳开始时间: %d\n", goOutData.StartTime)
			}
		}
	}
	if !hasValidStartTime {
		missingFields = append(missingFields, "开始时间")
	}

	var hasValidEndTime bool
	if v, ok := data["goOutEndTimeStr"]; ok && v != nil {
		if dateStr, ok := v.(string); ok && dateStr != "" {
			timestamp, err := parseDateTimeString(dateStr)
			if err != nil {
				return "", fmt.Errorf("结束时间格式解析失败: %s, 请使用格式如: 2025-10-24 11:00", dateStr)
			}
			goOutData.EndTime = timestamp
			hasValidEndTime = true
			fmt.Printf("[DEBUG] 从字符串解析结束时间: %s -> %d\n", dateStr, timestamp)
		}
	}
	if !hasValidEndTime {
		if v, ok := data["goOutEndTime"]; ok && v != nil {
			if val, ok := v.(float64); ok && val > 0 {
				goOutData.EndTime = int64(val)
				hasValidEndTime = true
				fmt.Printf("[DEBUG] 使用时间戳结束时间: %d\n", goOutData.EndTime)
			}
		}
	}
	if !hasValidEndTime {
		missingFields = append(missingFields, "结束时间")
	}

	if hasValidStartTime && hasValidEndTime && goOutData.EndTime <= goOutData.StartTime {
		return "", fmt.Errorf("结束时间必须晚于开始时间")
	}

	if v, ok := data["goOutReason"]; ok && v != nil {
		if val, ok := v.(string); ok && val != "" {
			goOutData.Reason = val
			fmt.Printf("[DEBUG] 外出原因: %s\n", goOutData.Reason)
		} else {
			missingFields = append(missingFields, "外出原因")
		}
	} else {
		missingFields = append(missingFields, "外出原因")
	}

	if len(missingFields) > 0 {
		errorMsg := fmt.Sprintf("创建外出审批缺少必要信息：%s\n\n", strings.Join(missingFields, "、"))
		errorMsg += "请提供以下信息：\n"
		errorMsg += "1. 开始时间（格式：YYYY-MM-DD HH:mm，如 2025-10-24 09:00）\n"
		errorMsg += "2. 结束时间（格式：YYYY-MM-DD HH:mm，如 2025-10-24 11:00）\n"
		errorMsg += "3. 外出原因\n\n"
		errorMsg += "示例：2025-10-24 09:00 到 2025-10-24 11:00 我要拜访客户"
		return "", fmt.Errorf(errorMsg)
	}

	req := domain.Approval{
		Type:  int(model.GoOutApproval),
		GoOut: &goOutData,
	}

	fmt.Printf("[DEBUG] 完整外出审批数据: %+v\n", req)
	fmt.Printf("[DEBUG] GoOut详情: StartTime=%d, EndTime=%d, Reason=%s\n",
		goOutData.StartTime, goOutData.EndTime, goOutData.Reason)

	addRes, err := curl.PostRequest(token.GetTokenStr(ctx), a.svc.Config.Host+"/v1/approval", req)
	if err != nil {
		return "", fmt.Errorf("failed to create go out approval: %w", err)
	}

	fmt.Printf("[DEBUG] 外出审批API响应: %s\n", string(addRes))

	var idResp domain.IdRespInfo
	if err := json.Unmarshal(addRes, &idResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return Success + "\ncreated approval id : " + idResp.Data.Id, nil
}
