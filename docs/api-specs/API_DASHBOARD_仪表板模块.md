# API_DASHBOARD_仪表板模块

## 文档履历

| 版本 | 日期 | 修改人 | 修改内容 | 审核状态 |
|------|------|--------|----------|----------|
| v1.0 | 2026-02-07 | AI后端工程师 | 初始版本，仪表盘API接口定义 | 🔄 待审核 |

---

## 概述

仪表盘API提供系统整体健康状态、核心指标概览、VM状态分布和最新告警数据。

## 基础信息

- **基础路径**: `/api/v1/dashboard`
- **认证方式**: Bearer Token (JWT)
- **数据格式**: JSON

---

## API接口列表

### 1. 获取仪表盘概览数据

获取系统整体健康状态和核心指标数据。

**请求信息**

| 项目 | 说明 |
|------|------|
| URL | `/api/v1/dashboard/overview` |
| Method | GET |
| 认证 | 必须 |

**响应数据**

```json
{
  "code": 200,
  "data": {
    "healthScore": 95,
    "healthTrend": "up",
    "lastUpdated": "2026-02-07T22:00:00+08:00",
    "systemStatus": "healthy",
    "summary": {
      "totalVMs": 1500,
      "onlineVMs": 1420,
      "offlineVMs": 50,
      "warningVMs": 25,
      "criticalVMs": 5
    },
    "metrics": {
      "cpu": {
        "usagePercent": 65.5,
        "trend": "stable",
        "trendValue": 2.5
      },
      "memory": {
        "usagePercent": 72.3,
        "trend": "up",
        "trendValue": 1.8
      },
      "disk": {
        "usagePercent": 58.2,
        "trend": "stable",
        "trendValue": 0.5
      },
      "network": {
        "inboundMbps": 125.5,
        "outboundMbps": 89.3,
        "trend": "up",
        "trendValue": 5.2
      }
    },
    "topResources": {
      "byCPU": [
        {"vmId": "vm-001", "vmName": "web-server-01", "usagePercent": 95.2},
        {"vmId": "vm-002", "vmName": "db-server-01", "usagePercent": 89.1}
      ],
      "byMemory": [
        {"vmId": "vm-003", "vmName": "app-server-01", "usagePercent": 92.5}
      ]
    }
  }
}
```

**字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| healthScore | number | 系统健康评分 (0-100) |
| healthTrend | string | 趋势: up/stable/down |
| systemStatus | string | 系统状态: healthy/warning/critical |
| summary | object | VM统计摘要 |
| summary.totalVMs | number | 总VM数量 |
| summary.onlineVMs | number | 在线VM数量 |
| summary.offlineVMs | number | 离线VM数量 |
| summary.warningVMs | number | 警告状态VM数量 |
| summary.criticalVMs | number | 严重状态VM数量 |
| metrics | object | 核心指标数据 |
| metrics.cpu | object | CPU指标 |
| metrics.cpu.usagePercent | number | CPU使用率百分比 |
| metrics.cpu.trend | string | 趋势 |
| metrics.cpu.trendValue | number | 变化幅度 |
| metrics.memory | object | 内存指标 |
| metrics.memory.usagePercent | number | 内存使用率百分比 |
| metrics.network | object | 网络指标 |
| metrics.network.inboundMbps | number | 入站带宽 (Mbps) |
| metrics.network.outboundMbps | number | 出站带宽 (Mbps) |

---

### 2. 获取VM状态分布

获取VM状态分布数据（用于饼图展示）。

**请求信息**

| 项目 | 说明 |
|------|------|
| URL | `/api/v1/dashboard/vm-status` |
| Method | GET |
| 认证 | 必须 |

**响应数据**

```json
{
  "code": 200,
  "data": {
    "distribution": [
      {"status": "online", "count": 1420, "percent": 94.67, "color": "#00d4aa"},
      {"status": "offline", "count": 50, "percent": 3.33, "color": "#607d8b"},
      {"status": "warning", "count": 25, "percent": 1.67, "color": "#ff9800"},
      {"status": "critical", "count": 5, "percent": 0.33, "color": "#f44336"}
    ],
    "byGroup": [
      {
        "groupName": "生产环境",
        "count": 800,
        "online": 780,
        "offline": 10,
        "warning": 8,
        "critical": 2
      },
      {
        "groupName": "测试环境",
        "count": 400,
        "online": 350,
        "offline": 30,
        "warning": 15,
        "critical": 5
      },
      {
        "groupName": "开发环境",
        "count": 300,
        "online": 290,
        "offline": 10,
        "warning": 2,
        "critical": 0
      }
    ],
    "byOS": [
      {"os": "Linux", "count": 1200, "percent": 80},
      {"os": "Windows", "count": 300, "percent": 20}
    ]
  }
}
```

---

### 3. 获取最新告警列表

获取最近的告警记录。

**请求信息**

| 项目 | 说明 |
|------|------|
| URL | `/api/v1/dashboard/alerts` |
| Method | GET |
| 认证 | 必须 |

**Query参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| limit | number | 否 | 返回数量，默认5，最大20 |

**响应数据**

```json
{
  "code": 200,
  "data": {
    "alerts": [
      {
        "id": "alert-001",
        "vmId": "vm-001",
        "vmName": "web-server-01",
        "vmIP": "192.168.1.100",
        "alertType": "cpu",
        "severity": "critical",
        "message": "CPU使用率持续超过95%超过5分钟",
        "value": "97.5%",
        "threshold": "95%",
        "occurredAt": "2026-02-07T21:55:00+08:00",
        "status": "active",
        "acknowledged": false
      },
      {
        "id": "alert-002",
        "vmId": "vm-002",
        "vmName": "db-server-01",
        "vmIP": "192.168.1.101",
        "alertType": "memory",
        "severity": "warning",
        "message": "内存使用率超过80%",
        "value": "85.2%",
        "threshold": "80%",
        "occurredAt": "2026-02-07T21:50:00+08:00",
        "status": "active",
        "acknowledged": true
      }
    ],
    "total": 156,
    "unreadCount": 12
  }
}
```

**告警级别定义**

| 级别 | 值 | 说明 | 颜色 |
|------|-----|------|------|
| low | 1 | 低级别 | #2196f3 |
| medium | 2 | 中级别 | #ff9800 |
| high | 3 | 高级别 | #f44336 |
| critical | 4 | 严重级别 | #b71c1c |

---

### 4. 获取健康度历史趋势

获取系统健康评分的历史趋势数据。

**请求信息**

| 项目 | 说明 |
|------|------|
| URL | `/api/v1/dashboard/health-trend` |
| Method | GET |
| 认证 | 必须 |

**Query参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| period | string | 否 | 时间范围: 24h/7d/30d, 默认7d |

**响应数据**

```json
{
  "code": 200,
  "data": {
    "period": "7d",
    "currentScore": 95,
    "trend": "up",
    "dataPoints": [
      {"timestamp": "2026-02-01T00:00:00+08:00", "score": 92},
      {"timestamp": "2026-02-02T00:00:00+08:00", "score": 93},
      {"timestamp": "2026-02-03T00:00:00+08:00", "score": 88},
      {"timestamp": "2026-02-04T00:00:00+08:00", "score": 91},
      {"timestamp": "2026-02-05T00:00:00+08:00", "score": 94},
      {"timestamp": "2026-02-06T00:00:00+08:00", "score": 96},
      {"timestamp": "2026-02-07T00:00:00+08:00", "score": 95}
    ]
  }
}
```

---

### 5. 获取问题VM列表

获取当前存在问题的VM列表（用于故障模式）。

**请求信息**

| 项目 | 说明 |
|------|------|
| URL | `/api/v1/dashboard/problem-vms` |
| Method | GET |
| 认证 | 必须 |

**Query参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| severity | string | 否 | 过滤级别: warning/critical |
| limit | number | 否 | 返回数量，默认20 |

**响应数据**

```json
{
  "code": 200,
  "data": {
    "total": 30,
    "vms": [
      {
        "vmId": "vm-001",
        "vmName": "web-server-01",
        "vmIP": "192.168.1.100",
        "group": "生产环境",
        "severity": "critical",
        "issues": [
          {"type": "cpu", "message": "CPU使用率97.5%", "value": "97.5%"}
        ],
        "firstDetected": "2026-02-07T21:50:00+08:00",
        "duration": "10分钟"
      },
      {
        "vmId": "vm-002",
        "vmName": "db-server-01",
        "vmIP": "192.168.1.101",
        "group": "生产环境",
        "severity": "warning",
        "issues": [
          {"type": "memory", "message": "内存使用率85.2%", "value": "85.2%"},
          {"type": "disk", "message": "磁盘使用率92%", "value": "92%"}
        ],
        "firstDetected": "2026-02-07T21:30:00+08:00",
        "duration": "35分钟"
      }
    ]
  }
}
```

---

## 错误响应

```json
{
  "code": 401,
  "message": "无效或已过期的Token"
}
```

```json
{
  "code": 403,
  "message": "没有查看仪表盘的权限"
}
```

```json
{
  "code": 500,
  "message": "获取仪表盘数据失败"
}
```

---

## 健康评分计算规则

健康评分基于以下维度计算：

| 维度 | 权重 | 计算方式 |
|------|------|----------|
| VM在线率 | 30% | (在线VM数 / 总VM数) × 100 |
| 性能指标 | 30% | 100 - 平均CPU/内存使用率 |
| 告警数量 | 25% | 基于告警严重程度的扣分 |
| 系统错误 | 15% | 系统级错误数量扣分 |

评分结果：
- 90-100: 健康 (绿色)
- 70-89: 良好 (蓝色)
- 50-69: 警告 (橙色)
- <50: 严重 (红色)

---

**文档创建日期**: 2026-02-07
**最后更新**: 2026-02-07
