package chatinternal

const (
	BASE_PROMPAT_TEMPLATE = `Current conversation:
{{.history}}
Human: {{.input}}
AI:`

	_defaultMrklPrefix = `Today is {{.today}}.
Answer the following questions as best you can. You have access to the following tools:

{{.tool_descriptions}}` + BASE_PROMPAT_TEMPLATE

	OUT_PROMPT_TEMPLATE = `<< instructions >>
- Your response should follow the JSON format.
- Your response should have the following structure: {"chatType": {{.chatType}}, "data": {{.data}} }
- "chatType" this is a fixed output`

	// ChatWithToolsTemplate 带有工具描述的通用聊天提示词
	ChatWithToolsTemplate = `The following is a friendly conversation between a human and an AI.
The AI is an office assistant for an application called AgentWorkHelper ("Agent Work Helper").

The AI is aware of the following specialized tools available in the system, which are used for specific tasks:
{{.tool_descriptions}}

If the human asks about "what tools you have" or "what can you do", the AI should summarize the above tools.
Otherwise, the AI should just engage in conversation normally.
Note: You are NOT executing these tools right now. You are just chat handling.

Current conversation:
{{.history}}
Human: {{.input}}
AI:`

	// _defaultChatLogPrompts AI群消息总结的提示词模板，用于指导LLM进行聊天记录分析和总结
	_defaultChatLogPrompts = `Please summarize based on the following chat conversations
- Role: You are an office assistant who helps employees summarize communication matters at work. Matters mainly include things such as "tasks to be done, leave for leave"
- work
1. You must first read the chat content to understand the relationship between employees and distinguish between superiors and subordinates
2. You need to distinguish what matters are based on the chat content, such as [tasks to be done, leave for leave]
4. You need to first obtain the overall context of the matter, and then summarize it based on the context, such as: the person initiated the leave, whether there are any follow-up, etc.
5. Personnel information needs to be fully output
6. keep Chinese output

- chatlog
{{.input}}

- require
The output content needs to be output in the following format and can be parsed by json
[
    {
       "type": int,      // task type; enum : 1. task to be done, 2 approval
       "title": string,  // title
       "content": string // content
    }, {
       "type": int,      // task typ; enum : 1. task to be done, 2 approval
       "title": string,  // title
       "content": string // content
    }
]
`
)
