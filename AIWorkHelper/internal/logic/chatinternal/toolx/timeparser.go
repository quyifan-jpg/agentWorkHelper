/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package toolx

import (
	"AIWorkHelper/internal/svc"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
)

// TimeParser 时间解析工具，将自然语言时间描述转换为Unix时间戳
type TimeParser struct {
	svc      *svc.ServiceContext // 服务上下文
	callback callbacks.Handler   // 回调处理器，用于记录执行日志
}

// NewTimeParser 创建时间解析工具实例
func NewTimeParser(svc *svc.ServiceContext) *TimeParser {
	return &TimeParser{
		svc:      svc,
		callback: svc.Callbacks,
	}
}

// Name 返回工具名称，用于AI代理识别
func (t *TimeParser) Name() string {
	return "time_parser"
}

// Description 返回工具描述和使用说明
func (t *TimeParser) Description() string {
	return `
	a time parser tool that converts natural language time expressions to Unix timestamps.
	use this tool to parse time expressions before creating todos or approvals.
	input: JSON string with "timeExpression" field containing the natural language time description from user input.
	examples:
	  - {"timeExpression": "明天下午14点"} -> tomorrow at 14:00
	  - {"timeExpression": "后天上午9点"} -> day after tomorrow at 09:00
	  - {"timeExpression": "下周一下午3点"} -> next Monday at 15:00
	  - {"timeExpression": "今天晚上8点"} -> today at 20:00
	  - {"timeExpression": "3天后下午2点"} -> 3 days later at 14:00
	output: JSON with "timestamp" (Unix timestamp in seconds) and "readableTime" (human-readable time string)
	keep Chinese output.
`
}

// Call 执行时间解析操作
func (t *TimeParser) Call(ctx context.Context, input string) (string, error) {
	// 记录工具调用日志
	if t.callback != nil {
		t.callback.HandleText(ctx, "time parser start : "+input)
	}

	// 解析输入参数
	var req struct {
		TimeExpression string `json:"timeExpression"`
	}
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		return "", fmt.Errorf("invalid input format: %w", err)
	}

	if req.TimeExpression == "" {
		return "", fmt.Errorf("timeExpression is required")
	}

	// 解析时间表达式
	timestamp, err := parseTimeExpression(req.TimeExpression)
	if err != nil {
		return "", fmt.Errorf("failed to parse time expression: %w", err)
	}

	// 构建返回结果
	result := struct {
		Timestamp    int64  `json:"timestamp"`
		ReadableTime string `json:"readableTime"`
	}{
		Timestamp:    timestamp,
		ReadableTime: time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"),
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to serialize result: %w", err)
	}

	return Success + "\nparsed time:\n" + string(jsonResult), nil
}

// parseTimeExpression 解析时间表达式为Unix时间戳
func parseTimeExpression(expr string) (int64, error) {
	now := time.Now()
	expr = strings.TrimSpace(expr)

	// 默认时间（如果没有指定具体时间）
	var targetTime time.Time
	var hasDate bool
	var hasTime bool

	// 解析日期部分
	if strings.Contains(expr, "今天") || strings.Contains(expr, "今日") {
		targetTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		hasDate = true
	} else if strings.Contains(expr, "明天") || strings.Contains(expr, "明日") {
		targetTime = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		hasDate = true
	} else if strings.Contains(expr, "后天") {
		targetTime = time.Date(now.Year(), now.Month(), now.Day()+2, 0, 0, 0, 0, now.Location())
		hasDate = true
	} else if strings.Contains(expr, "大后天") {
		targetTime = time.Date(now.Year(), now.Month(), now.Day()+3, 0, 0, 0, 0, now.Location())
		hasDate = true
	} else if matched, _ := regexp.MatchString(`(\d+)天后`, expr); matched {
		// 匹配 "3天后"、"5天后" 等
		re := regexp.MustCompile(`(\d+)天后`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 1 {
			days, _ := strconv.Atoi(matches[1])
			targetTime = time.Date(now.Year(), now.Month(), now.Day()+days, 0, 0, 0, 0, now.Location())
			hasDate = true
		}
	} else if strings.Contains(expr, "上周") {
		// 上周的某一天
		daysToSubtract := 7
		if strings.Contains(expr, "上周一") || strings.Contains(expr, "上周1") {
			daysToSubtract = int(now.Weekday()) + 6 // 到上周一的天数
			if now.Weekday() == time.Sunday {
				daysToSubtract = 6
			}
		} else if strings.Contains(expr, "上周二") || strings.Contains(expr, "上周2") {
			daysToSubtract = int(now.Weekday()) + 5
			if now.Weekday() == time.Sunday {
				daysToSubtract = 12
			}
		} else if strings.Contains(expr, "上周三") || strings.Contains(expr, "上周3") {
			daysToSubtract = int(now.Weekday()) + 4
			if now.Weekday() == time.Sunday {
				daysToSubtract = 11
			}
		} else if strings.Contains(expr, "上周四") || strings.Contains(expr, "上周4") {
			daysToSubtract = int(now.Weekday()) + 3
			if now.Weekday() == time.Sunday {
				daysToSubtract = 10
			}
		} else if strings.Contains(expr, "上周五") || strings.Contains(expr, "上周5") {
			daysToSubtract = int(now.Weekday()) + 2
			if now.Weekday() == time.Sunday {
				daysToSubtract = 9
			}
		} else if strings.Contains(expr, "上周六") || strings.Contains(expr, "上周6") {
			daysToSubtract = int(now.Weekday()) + 1
			if now.Weekday() == time.Sunday {
				daysToSubtract = 8
			}
		} else if strings.Contains(expr, "上周日") || strings.Contains(expr, "上周天") || strings.Contains(expr, "上周7") {
			daysToSubtract = int(now.Weekday())
			if now.Weekday() == time.Sunday {
				daysToSubtract = 7
			}
		}
		targetTime = time.Date(now.Year(), now.Month(), now.Day()-daysToSubtract, 0, 0, 0, 0, now.Location())
		hasDate = true
	} else if strings.Contains(expr, "下周") {
		// 下周的某一天
		daysToAdd := 7
		if strings.Contains(expr, "下周一") || strings.Contains(expr, "下周1") {
			daysToAdd = 7 - int(now.Weekday()) + 1
			if now.Weekday() == time.Sunday {
				daysToAdd = 1
			}
		} else if strings.Contains(expr, "下周二") || strings.Contains(expr, "下周2") {
			daysToAdd = 7 - int(now.Weekday()) + 2
			if now.Weekday() == time.Sunday {
				daysToAdd = 2
			}
		} else if strings.Contains(expr, "下周三") || strings.Contains(expr, "下周3") {
			daysToAdd = 7 - int(now.Weekday()) + 3
			if now.Weekday() == time.Sunday {
				daysToAdd = 3
			}
		} else if strings.Contains(expr, "下周四") || strings.Contains(expr, "下周4") {
			daysToAdd = 7 - int(now.Weekday()) + 4
			if now.Weekday() == time.Sunday {
				daysToAdd = 4
			}
		} else if strings.Contains(expr, "下周五") || strings.Contains(expr, "下周5") {
			daysToAdd = 7 - int(now.Weekday()) + 5
			if now.Weekday() == time.Sunday {
				daysToAdd = 5
			}
		} else if strings.Contains(expr, "下周六") || strings.Contains(expr, "下周6") {
			daysToAdd = 7 - int(now.Weekday()) + 6
			if now.Weekday() == time.Sunday {
				daysToAdd = 6
			}
		} else if strings.Contains(expr, "下周日") || strings.Contains(expr, "下周天") || strings.Contains(expr, "下周7") {
			daysToAdd = 7
		}
		targetTime = time.Date(now.Year(), now.Month(), now.Day()+daysToAdd, 0, 0, 0, 0, now.Location())
		hasDate = true
	}

	// 如果没有找到日期关键词，默认为今天
	if !hasDate {
		targetTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	// 解析时间部分
	var hour, minute int

	// 匹配具体时间：14点、14:30、14点30分、下午2点、晚上8点等
	if matched, _ := regexp.MatchString(`(\d{1,2})[点:](\d{1,2})`, expr); matched {
		re := regexp.MustCompile(`(\d{1,2})[点:](\d{1,2})`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 2 {
			hour, _ = strconv.Atoi(matches[1])
			minute, _ = strconv.Atoi(matches[2])
			hasTime = true
		}
	} else if matched, _ := regexp.MatchString(`(\d{1,2})点`, expr); matched {
		re := regexp.MustCompile(`(\d{1,2})点`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 1 {
			hour, _ = strconv.Atoi(matches[1])
			minute = 0
			hasTime = true
		}
	}

	// 处理上午/下午/晚上
	if strings.Contains(expr, "下午") || strings.Contains(expr, "午后") {
		if hour < 12 && hour > 0 {
			hour += 12
		}
	} else if strings.Contains(expr, "晚上") || strings.Contains(expr, "晚间") {
		if hour < 12 {
			hour += 12
		}
		if hour < 18 {
			hour = 18 // 晚上至少从18点开始
		}
	} else if strings.Contains(expr, "上午") || strings.Contains(expr, "早上") || strings.Contains(expr, "早晨") {
		if hour >= 12 {
			hour -= 12
		}
	} else if strings.Contains(expr, "中午") {
		hour = 12
		minute = 0
		hasTime = true
	} else if strings.Contains(expr, "凌晨") {
		if hour >= 12 {
			hour -= 12
		}
	}

	// 如果没有指定具体时间，根据时间段设置默认时间
	if !hasTime {
		// 优先处理上班/下班（更具体的场景）
		if strings.Contains(expr, "下班") {
			hour = 18 // 默认下班时间18:00
			minute = 0
		} else if strings.Contains(expr, "上班") {
			hour = 9 // 默认上班时间9:00
			minute = 0
		} else if strings.Contains(expr, "中午") {
			hour = 12
			minute = 0
		} else if strings.Contains(expr, "上午") || strings.Contains(expr, "早上") {
			hour = 9
			minute = 0
		} else if strings.Contains(expr, "下午") {
			hour = 14
			minute = 0
		} else if strings.Contains(expr, "晚上") {
			hour = 19
			minute = 0
		} else {
			// 默认为当前时间的下一个整点
			hour = now.Hour() + 1
			minute = 0
		}
	}

	// 设置最终时间
	targetTime = time.Date(targetTime.Year(), targetTime.Month(), targetTime.Day(), hour, minute, 0, 0, targetTime.Location())

	return targetTime.Unix(), nil
}
