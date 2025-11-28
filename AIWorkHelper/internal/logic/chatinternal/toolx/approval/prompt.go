/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package approval

const (
	_defaultCreateApprovalTemplate = `You are an intelligent assistant helping to process leave approval requests.

User input:
{{.input}}

Available leave types:
- type=1: 事假 (Personal leave) - for personal matters, errands, family affairs
- type=2: 调休 (Compensatory leave) - for overtime compensation
- type=3: 病假 (Sick leave) - for illness, medical appointments
- type=4: 年假 (Annual leave) - for vacation, rest
- type=5: 产假 (Maternity leave)
- type=6: 陪产假 (Paternity leave)
- type=7: 婚假 (Marriage leave)
- type=8: 丧假 (Bereavement leave)
- type=9: 哺乳假 (Breastfeeding leave)

Instructions:
1. Extract leave type from user input keywords
2. Use time_parser tool to convert time expressions to timestamps (MUST use the tool, don't guess!)
3. Extract reason directly from user's words
4. Determine timeType: less than 8 hours=1(小时), 8+ hours=2(天)

Extract the information and output in the required format.
`
)
