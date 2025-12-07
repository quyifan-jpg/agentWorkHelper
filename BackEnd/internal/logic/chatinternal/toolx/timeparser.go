package toolx

import (
	"BackEnd/internal/svc"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
)

// TimeParser 时间解析工具，将自然语言时间转换为Unix时间戳
type TimeParser struct {
	svc      *svc.ServiceContext // 服务上下文
	callback callbacks.Handler   // 回调处理器
}

// NewTimeParser 创建时间解析工具实例
func NewTimeParser(svc *svc.ServiceContext) *TimeParser {
	return &TimeParser{
		svc:      svc,
		callback: svc.Callbacks,
	}
}

// Name 返回工具名称
func (t *TimeParser) Name() string {
	return "time_parser"
}

// Description 返回工具描述
func (t *TimeParser) Description() string {
	return `
	a time parser interface.
	use when you need to convert natural language time expressions (like "tomorrow at 2pm", "next Friday", "今天下午3点") into a unix timestamp.
	input: a string containing the time expression.
	output: the unix timestamp (int64) representing the time.
`
}

// Call 执行时间解析操作
func (t *TimeParser) Call(ctx context.Context, input string) (string, error) {
	if t.callback != nil {
		t.callback.HandleText(ctx, "time parser start : "+input)
	}

	timestamp, err := parseTimeExpression(input)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", timestamp), nil
}

// parseTimeExpression 解析自然语言时间表达式
func parseTimeExpression(expr string) (int64, error) {
	now := time.Now()
	targetTime := now

	// 简单的规则匹配实现
	// 注意：这是一个简化的实现，实际生产环境可能需要更复杂的NLP处理或调用第三方服务

	hasDate := false
	hasTime := false

	// 处理相对日期
	if strings.Contains(expr, "今天") {
		hasDate = true
	} else if strings.Contains(expr, "明天") {
		targetTime = targetTime.AddDate(0, 0, 1)
		hasDate = true
	} else if strings.Contains(expr, "后天") {
		targetTime = targetTime.AddDate(0, 0, 2)
		hasDate = true
	} else if strings.Contains(expr, "下周") {
		// 简单处理下周：默认下周一，或者根据具体星期几
		daysToAdd := 0
		if strings.Contains(expr, "下周一") {
			// 计算距离下周一的天数
			offset := int(time.Monday - now.Weekday())
			if offset <= 0 {
				offset += 7
			}
			daysToAdd = offset
		} else if strings.Contains(expr, "下周二") {
			offset := int(time.Tuesday - now.Weekday())
			if offset <= 0 {
				offset += 7
			}
			daysToAdd = offset
		} else if strings.Contains(expr, "下周三") {
			offset := int(time.Wednesday - now.Weekday())
			if offset <= 0 {
				offset += 7
			}
			daysToAdd = offset
		} else if strings.Contains(expr, "下周四") {
			offset := int(time.Thursday - now.Weekday())
			if offset <= 0 {
				offset += 7
			}
			daysToAdd = offset
		} else if strings.Contains(expr, "下周五") {
			offset := int(time.Friday - now.Weekday())
			if offset <= 0 {
				offset += 7
			}
			daysToAdd = offset
		} else if strings.Contains(expr, "下周六") {
			offset := int(time.Saturday - now.Weekday())
			if offset <= 0 {
				offset += 7
			}
			daysToAdd = offset
		} else if strings.Contains(expr, "下周日") || strings.Contains(expr, "下周天") {
			offset := int(time.Sunday - now.Weekday())
			if offset <= 0 {
				offset += 7
			}
			daysToAdd = offset
		} else {
			// 默认下周一
			offset := int(time.Monday - now.Weekday())
			if offset <= 0 {
				offset += 7
			}
			daysToAdd = offset
		}
		targetTime = targetTime.AddDate(0, 0, daysToAdd)
		hasDate = true
	} else if strings.Contains(expr, "周") || strings.Contains(expr, "星期") {
		// 处理本周几
		var targetWeekday time.Weekday
		if strings.Contains(expr, "周一") || strings.Contains(expr, "星期一") {
			targetWeekday = time.Monday
		} else if strings.Contains(expr, "周二") || strings.Contains(expr, "星期二") {
			targetWeekday = time.Tuesday
		} else if strings.Contains(expr, "周三") || strings.Contains(expr, "星期三") {
			targetWeekday = time.Wednesday
		} else if strings.Contains(expr, "周四") || strings.Contains(expr, "星期四") {
			targetWeekday = time.Thursday
		} else if strings.Contains(expr, "周五") || strings.Contains(expr, "星期五") {
			targetWeekday = time.Friday
		} else if strings.Contains(expr, "周六") || strings.Contains(expr, "星期六") {
			targetWeekday = time.Saturday
		} else if strings.Contains(expr, "周日") || strings.Contains(expr, "星期日") || strings.Contains(expr, "周天") || strings.Contains(expr, "星期天") {
			targetWeekday = time.Sunday
		}

		// 计算距离目标星期几的天数（假设是本周或下周即将到来的那一天）
		offset := int(targetWeekday - now.Weekday())
		if offset <= 0 {
			offset += 7 // 如果是今天或过去，则算作下周
		}
		targetTime = targetTime.AddDate(0, 0, offset)
		hasDate = true
	}

	// 如果没有找到日期关键词，默认为今天
	if !hasDate {
		targetTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	} else {
		// 如果有日期，先重置时间为0点
		targetTime = time.Date(targetTime.Year(), targetTime.Month(), targetTime.Day(), 0, 0, 0, 0, targetTime.Location())
	}

	// 解析时间部分
	var hour, minute int

	// 匹配具体时间：14点、14:30、14点30分、下午2点、晚上8点等
	// 简单正则匹配
	reTime := regexp.MustCompile(`(\d{1,2})[:点](\d{1,2})?`)
	matches := reTime.FindStringSubmatch(expr)
	if len(matches) > 1 {
		h, _ := strconv.Atoi(matches[1])
		hour = h
		if len(matches) > 2 && matches[2] != "" {
			m, _ := strconv.Atoi(matches[2])
			minute = m
		}
		hasTime = true
	} else {
		// 尝试匹配 "X点"
		reHour := regexp.MustCompile(`(\d{1,2})点`)
		matchesHour := reHour.FindStringSubmatch(expr)
		if len(matchesHour) > 1 {
			h, _ := strconv.Atoi(matchesHour[1])
			hour = h
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
			hour = 18 // 晚上至少从18点开始，如果只说"晚上"没说几点
		}
	} else if strings.Contains(expr, "上午") || strings.Contains(expr, "早上") || strings.Contains(expr, "早晨") {
		if hour >= 12 {
			hour -= 12
		}
	} else if strings.Contains(expr, "中午") {
		if !hasTime {
			hour = 12
			minute = 0
			hasTime = true
		}
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
