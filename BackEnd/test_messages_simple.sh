#!/bin/bash
BASE_URL="http://localhost:8889"

echo "=== 快速测试历史消息功能 ==="
echo ""

# 登录
echo "1. 登录..."
LOGIN=$(curl -s -X POST "$BASE_URL/v1/user/login" -H "Content-Type: application/json" -d '{"name": "root", "password": "123456"}')
TOKEN=$(echo "$LOGIN" | jq -r '.data.token // empty')
if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
  echo "❌ 登录失败，请确保服务已启动"
  echo "响应: $LOGIN"
  exit 1
fi
echo "✅ Token: ${TOKEN:0:30}..."
echo ""

# 测试查询群聊消息
echo "2. 测试查询群聊历史消息 (conversationId=all)..."
RESP=$(curl -s -X GET "$BASE_URL/v1/chat/messages?conversationId=all&page=1&count=10" -H "Authorization: Bearer $TOKEN")
echo "$RESP" | jq '.' 2>/dev/null || echo "$RESP"
echo ""

CODE=$(echo "$RESP" | jq -r '.code // empty')
if [ "$CODE" == "200" ]; then
  TOTAL=$(echo "$RESP" | jq -r '.data.total // 0')
  COUNT=$(echo "$RESP" | jq -r '.data.list | length // 0')
  echo "✅ 成功！总消息数: $TOTAL, 返回: $COUNT 条"
else
  MSG=$(echo "$RESP" | jq -r '.msg // .message // "未知错误"')
  echo "❌ 失败: $MSG"
fi
