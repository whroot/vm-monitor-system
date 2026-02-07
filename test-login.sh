#!/bin/bash
echo "======================================"
echo "测试登录流程"
echo "======================================"

echo -e "\n1. 测试登录API..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser4","password":"test123"}')

echo "$LOGIN_RESPONSE" | jq .

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .data.accessToken)

if [ "$TOKEN" == "null" ]; then
  echo "❌ 登录失败"
  exit 1
fi

echo -e "\n2. 使用Token获取用户信息..."
curl -s http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n3. 获取用户权限..."
curl -s http://localhost:8080/api/v1/users/3a2e28e4-759f-49b0-b4fe-f90d2769416f/permissions \
  -H "Authorization: Bearer $TOKEN" | jq '.data | length'

echo -e "\n======================================"
echo "后端API测试完成！"
echo "======================================"
echo ""
echo "如果后端API正常，问题可能在浏览器端。"
echo "请在浏览器中按 F12 打开控制台，查看具体错误信息。"
