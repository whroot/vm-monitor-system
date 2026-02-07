#!/bin/bash

echo "======================================"
echo "VM监控系统前后端集成测试"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go backend is running
echo "1. 检查Go后端服务..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✓${NC} Go后端服务已启动"
    curl -s http://localhost:8080/health | jq '.'
else
    echo -e "${RED}✗${NC} Go后端服务未启动"
    exit 1
fi
echo ""

# Check if Frontend is running
echo "2. 检查前端服务..."
if curl -s http://localhost:5173 > /dev/null; then
    echo -e "${GREEN}✓${NC} 前端服务已启动"
else
    echo -e "${RED}✗${NC} 前端服务未启动"
    exit 1
fi
echo ""

# Test User Registration
echo "3. 测试用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser5","email":"test5@example.com","password":"test123","name":"Test User 5"}')
if echo "$REGISTER_RESPONSE" | jq -e '.code == 201' > /dev/null; then
    echo -e "${GREEN}✓${NC} 用户注册成功"
    USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.data.userId')
    echo "  用户ID: $USER_ID"
else
    echo -e "${YELLOW}⚠${NC} 用户可能已存在 (这可能是预期的)"
fi
echo ""

# Test User Login
echo "4. 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser4","password":"test123"}')
if echo "$LOGIN_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓${NC} 用户登录成功"
    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.accessToken')
    echo "  Token: ${ACCESS_TOKEN:0:50}..."
else
    echo -e "${RED}✗${NC} 用户登录失败"
    exit 1
fi
echo ""

# Test VM List
echo "5. 测试获取VM列表..."
VM_RESPONSE=$(curl -s http://localhost:8080/api/v1/vms \
  -H "Authorization: Bearer $ACCESS_TOKEN")
if echo "$VM_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓${NC} 获取VM列表成功"
    VM_COUNT=$(echo "$VM_RESPONSE" | jq -r '.data.vms | length')
    echo "  VM数量: $VM_COUNT"
else
    echo -e "${RED}✗${NC} 获取VM列表失败"
fi
echo ""

# Test VM Stats
echo "6. 测试获取VM统计..."
STATS_RESPONSE=$(curl -s http://localhost:8080/api/v1/vms/stats \
  -H "Authorization: Bearer $ACCESS_TOKEN")
if echo "$STATS_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓${NC} 获取VM统计成功"
    echo "$STATS_RESPONSE" | jq '.data'
else
    echo -e "${RED}✗${NC} 获取VM统计失败"
fi
echo ""

# Test Alert Rules
echo "7. 测试告警规则..."
ALERT_RULES_RESPONSE=$(curl -s http://localhost:8080/api/v1/alerts/rules \
  -H "Authorization: Bearer $ACCESS_TOKEN")
if echo "$ALERT_RULES_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓${NC} 获取告警规则列表成功"
    RULE_COUNT=$(echo "$ALERT_RULES_RESPONSE" | jq -r '.data.rules | length')
    echo "  规则数量: $RULE_COUNT"
else
    echo -e "${RED}✗${NC} 获取告警规则列表失败"
fi
echo ""

# Test Create Alert Rule
echo "8. 测试创建告警规则..."
CREATE_RULE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/alerts/rules \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"磁盘使用率告警","scope":"vm","severity":"warning"}')
if echo "$CREATE_RULE_RESPONSE" | jq -e '.code == 201' > /dev/null; then
    echo -e "${GREEN}✓${NC} 创建告警规则成功"
    RULE_ID=$(echo "$CREATE_RULE_RESPONSE" | jq -r '.data.ruleId')
    echo "  规则ID: $RULE_ID"
else
    echo -e "${RED}✗${NC} 创建告警规则失败"
fi
echo ""

# Test Alert Stats
echo "9. 测试告警统计..."
ALERT_STATS_RESPONSE=$(curl -s http://localhost:8080/api/v1/alerts/stats \
  -H "Authorization: Bearer $ACCESS_TOKEN")
if echo "$ALERT_STATS_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓${NC} 获取告警统计成功"
    echo "$ALERT_STATS_RESPONSE" | jq '.data'
else
    echo -e "${RED}✗${NC} 获取告警统计失败"
fi
echo ""

# Test Create Test Alert
echo "10. 测试创建测试告警..."
CREATE_ALERT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/alerts/test \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"vmName":"prod-web-01","severity":"critical","metric":"memory","value":92.5}')
if echo "$CREATE_ALERT_RESPONSE" | jq -e '.code == 201' > /dev/null; then
    echo -e "${GREEN}✓${NC} 创建测试告警成功"
    ALERT_ID=$(echo "$CREATE_ALERT_RESPONSE" | jq -r '.data.alertId')
    echo "  告警ID: $ALERT_ID"

    # Test Acknowledge Alert
    echo "11. 测试确认告警..."
    ACKNOWLEDGE_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/alerts/${ALERT_ID}/acknowledge" \
      -H "Authorization: Bearer $ACCESS_TOKEN")
    if echo "$ACKNOWLEDGE_RESPONSE" | jq -e '.code == 200' > /dev/null; then
        echo -e "${GREEN}✓${NC} 确认告警成功"
    else
        echo -e "${RED}✗${NC} 确认告警失败"
    fi

    # Test Resolve Alert
    echo "12. 测试解决告警..."
    RESOLVE_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/alerts/${ALERT_ID}/resolve" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"resolution":"已重启服务"}')
    if echo "$RESOLVE_RESPONSE" | jq -e '.code == 200' > /dev/null; then
        echo -e "${GREEN}✓${NC} 解决告警成功"
    else
        echo -e "${RED}✗${NC} 解决告警失败"
    fi
else
    echo -e "${RED}✗${NC} 创建测试告警失败"
fi
echo ""

# Test Real-time Metrics
echo "13. 测试获取实时监控指标..."
METRICS_RESPONSE=$(curl -s http://localhost:8080/api/v1/vms/73c547bc-19b9-431b-b71f-7bfaf92be02c/metrics \
  -H "Authorization: Bearer $ACCESS_TOKEN")
if echo "$METRICS_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    echo -e "${GREEN}✓${NC} 获取实时监控指标成功"
    CPU_USAGE=$(echo "$METRICS_RESPONSE" | jq -r '.data.cpu.usage')
    MEMORY_USAGE=$(echo "$METRICS_RESPONSE" | jq -r '.data.memory.usage')
    echo "  CPU使用率: $CPU_USAGE%"
    echo "  内存使用率: $MEMORY_USAGE%"
else
    echo -e "${RED}✗${NC} 获取实时监控指标失败"
fi
echo ""

echo "======================================"
echo "集成测试完成!"
echo "======================================"
echo ""
echo "访问地址:"
echo "  - 前端: http://localhost:5173"
echo "  - 后端: http://localhost:8080"
echo ""
echo "测试用户:"
echo "  - 用户名: testuser4"
echo "  - 密码: test123"
echo ""
