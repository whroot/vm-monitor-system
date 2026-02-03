# API_ALERT_告警管理模块_API规范

## 文档履历

| 版本 | 日期 | 修改人 | 修改内容 | 审核状态 |
|------|------|--------|----------|----------|
| v1.0 | 2026-02-03 | BE工程师 | 初始版本，基于REQ_20260202和UI_20260202生成 | 🔄 待审核 |

---

## 模块概述

### 功能范围
- 告警规则配置（CRUD操作）
- 告警触发记录管理
- 告警通知配置（邮件/短信/站内信）
- 告警确认与处理流程
- 告警统计与趋势分析

### 适用角色
- 系统管理员：全部权限
- 运维工程师：创建/编辑规则、确认告警
- IT经理：查看告警统计、导出报告
- 安全工程师：安全告警监控

### 技术约束
- 告警规则数量：单VM最多50条，全局最多500条
- 告警触发频率：最小间隔5分钟
- 通知频率：同一告警最小间隔15分钟
- 告警历史保留：2年

---

## 接口清单

### 告警规则管理

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取告警规则列表 | GET | /api/v1/alerts/rules | 分页查询告警规则 | 需要认证 |
| 获取告警规则详情 | GET | /api/v1/alerts/rules/{id} | 获取单个规则详情 | 需要认证 |
| 创建告警规则 | POST | /api/v1/alerts/rules | 创建新告警规则 | 需要alert:write权限 |
| 更新告警规则 | PUT | /api/v1/alerts/rules/{id} | 更新告警规则 | 需要alert:write权限 |
| 删除告警规则 | DELETE | /api/v1/alerts/rules/{id} | 删除告警规则 | 需要alert:write权限 |
| 批量启用/禁用规则 | PUT | /api/v1/alerts/rules/batch/status | 批量修改规则状态 | 需要alert:write权限 |
| 导入规则 | POST | /api/v1/alerts/rules/import | 导入JSON规则配置 | 需要alert:write权限 |
| 导出规则 | POST | /api/v1/alerts/rules/export | 导出规则为JSON | 需要alert:read权限 |

### 告警记录管理

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取告警记录列表 | GET | /api/v1/alerts/records | 分页查询告警记录 | 需要认证 |
| 获取告警记录详情 | GET | /api/v1/alerts/records/{id} | 获取单个告警详情 | 需要认证 |
| 确认告警 | PUT | /api/v1/alerts/records/{id}/acknowledge | 确认告警 | 需要alert:write权限 |
| 批量确认告警 | PUT | /api/v1/alerts/records/batch/acknowledge | 批量确认 | 需要alert:write权限 |
| 解决告警 | PUT | /api/v1/alerts/records/{id}/resolve | 标记告警已解决 | 需要alert:write权限 |
| 忽略告警 | PUT | /api/v1/alerts/records/{id}/ignore | 忽略告警 | 需要alert:write权限 |
| 获取告警统计 | GET | /api/v1/alerts/statistics | 告警统计信息 | 需要认证 |
| 获取告警趋势 | GET | /api/v1/alerts/trends | 告警趋势数据 | 需要认证 |

### 通知配置

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取通知配置 | GET | /api/v1/alerts/notifications/config | 获取通知方式配置 | 需要认证 |
| 更新通知配置 | PUT | /api/v1/alerts/notifications/config | 更新通知配置 | 需要alert:write权限 |
| 测试通知 | POST | /api/v1/alerts/notifications/test | 发送测试通知 | 需要alert:write权限 |
| 获取通知记录 | GET | /api/v1/alerts/notifications/history | 查询通知发送记录 | 需要认证 |

---

## 数据模型

### AlertRule（告警规则）

```typescript
interface AlertRule {
  id: string;                       // 规则ID
  name: string;                     // 规则名称
  description?: string;             // 规则描述
  
  // 作用范围
  scope: 'global' | 'vm' | 'group' | 'cluster';  // 规则范围
  scopeId?: string;                 // 范围对象ID（vmId/groupId/clusterId）
  scopeName?: string;               // 范围对象名称
  
  // 触发条件
  conditions: AlertCondition[];     // 触发条件（支持多条件组合）
  conditionLogic: 'and' | 'or';    // 多条件逻辑关系
  
  // 触发控制
  enabled: boolean;                 // 是否启用
  cooldown: number;                 // 冷却时间（秒，默认300）
  
  // 严重级别
  severity: 'low' | 'medium' | 'high' | 'critical';
  
  // 通知配置
  notifications: AlertNotificationConfig;
  
  // 元数据
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;              // 创建者ID
  updatedBy: string;                // 更新者ID
  
  // 统计
  triggerCount: number;             // 触发次数
  lastTriggeredAt?: Date;           // 最后触发时间
}
```

### AlertCondition（告警条件）

```typescript
interface AlertCondition {
  id: string;                       // 条件ID（在规则内唯一）
  metric: 'cpu' | 'memory' | 'disk' | 'network' | 'vmStatus';
  
  // 指标子类型
  metricType?: string;              // 如：cpu.usagePercent, memory.usagePercent
  
  // 操作符
  operator: '>' | '<' | '>=' | '<=' | '==' | '!=' | 'in' | 'not_in';
  
  // 阈值
  threshold: number | number[] | string;  // 单值/范围/枚举值
  
  // 持续时间（持续满足条件才触发）
  duration: number;                 // 持续时间（秒，默认60）
  
  // 聚合方式（可选）
  aggregation?: 'avg' | 'max' | 'min' | 'last';  // 默认last
}
```

### AlertNotificationConfig（告警通知配置）

```typescript
interface AlertNotificationConfig {
  // 通知方式
  methods: Array<'email' | 'sms' | 'webhook' | 'inApp'>;
  
  // 邮件配置
  email?: {
    enabled: boolean;
    recipients: string[];           // 收件人邮箱列表
    cc?: string[];                  // 抄送
    template?: string;              // 邮件模板ID
  };
  
  // 短信配置
  sms?: {
    enabled: boolean;
    phoneNumbers: string[];         // 手机号列表
    template?: string;
  };
  
  // Webhook配置
  webhook?: {
    enabled: boolean;
    url: string;                    // Webhook URL
    method: 'POST' | 'PUT';
    headers?: Record<string, string>;  // 自定义Header
    secret?: string;                // 签名密钥
  };
  
  // 站内信
  inApp?: {
    enabled: boolean;
    users?: string[];               // 指定用户（空表示全部管理员）
  };
  
  // 升级策略
  escalation?: {
    enabled: boolean;
    levels: Array<{
      delay: number;                // 延迟时间（分钟）
      methods: string[];            // 升级后的通知方式
      recipients: string[];         // 升级后的接收人
    }>;
  };
}
```

### AlertRecord（告警记录）

```typescript
interface AlertRecord {
  id: string;                       // 告警记录ID
  
  // 关联规则
  ruleId: string;                   // 触发规则ID
  ruleName: string;                 // 规则名称
  
  // 作用对象
  vmId?: string;                    // VM ID（VM级别告警）
  vmName?: string;                  // VM名称
  groupId?: string;                 // 分组ID
  clusterId?: string;               // 集群ID
  
  // 告警内容
  metric: string;                   // 触发指标
  severity: 'low' | 'medium' | 'high' | 'critical';
  
  // 触发详情
  triggerValue: number;             // 触发时的值
  threshold: number;                // 阈值
  condition: string;                // 触发条件描述
  
  // 时间
  triggeredAt: Date;                // 触发时间
  resolvedAt?: Date;                // 解决时间
  duration?: number;                // 持续时长（秒）
  
  // 状态
  status: 'active' | 'acknowledged' | 'resolved' | 'ignored';
  
  // 确认信息
  acknowledgedBy?: string;          // 确认人ID
  acknowledgedByName?: string;        // 确认人姓名
  acknowledgedAt?: Date;              // 确认时间
  acknowledgeNote?: string;         // 确认备注
  
  // 解决信息
  resolvedBy?: string;                // 解决人ID
  resolvedByName?: string;          // 解决人姓名
  resolution?: string;                // 解决方案
  
  // 通知状态
  notifications: Array<{
    method: string;
    status: 'sent' | 'failed' | 'pending';
    sentAt?: Date;
    error?: string;
  }>;
  
  // 快照数据（触发时的指标快照）
  snapshot?: {
    cpu?: object;
    memory?: object;
    disk?: object;
    network?: object;
  };
  
  createdAt: Date;
  updatedAt: Date;
}
```

### AlertStatistics（告警统计）

```typescript
interface AlertStatistics {
  // 总体统计
  overview: {
    totalRules: number;             // 总规则数
    activeRules: number;            // 启用规则数
    totalAlerts: number;            // 总告警数
    activeAlerts: number;           // 活跃告警数
    acknowledgedAlerts: number;     // 已确认告警数
    resolvedAlerts: number;         // 已解决告警数
  };
  
  // 按严重级别分布
  bySeverity: {
    critical: { total: number; active: number };
    high: { total: number; active: number };
    medium: { total: number; active: number };
    low: { total: number; active: number };
  };
  
  // 按指标分布
  byMetric: Array<{
    metric: string;
    count: number;
    activeCount: number;
  }>;
  
  // 按VM分布（Top 10）
  byVM: Array<{
    vmId: string;
    vmName: string;
    count: number;
    activeCount: number;
  }>;
  
  // 按规则分布
  byRule: Array<{
    ruleId: string;
    ruleName: string;
    triggerCount: number;
  }>;
  
  // MTTR（平均修复时间）
  mttr?: {
    avg: number;                    // 平均修复时间（分钟）
    bySeverity: Record<string, number>;
  };
  
  // 时间段统计
  timeRange: {
    start: Date;
    end: Date;
  };
}
```

### NotificationConfig（全局通知配置）

```typescript
interface NotificationConfig {
  // 邮件服务器配置
  email?: {
    enabled: boolean;
    smtp: {
      host: string;
      port: number;
      secure: boolean;
      auth: {
        user: string;
        pass: string;
      };
    };
    from: string;                   // 发件人
    fromName: string;               // 发件人名称
  };
  
  // 短信服务配置（预留）
  sms?: {
    enabled: boolean;
    provider: string;               // 服务商
    apiKey?: string;
    apiSecret?: string;
    signature?: string;
  };
  
  // 默认通知模板
  defaultTemplates: {
    email: string;
    sms: string;
    webhook: string;
  };
  
  // 全局通知策略
  globalPolicy: {
    maxRetry: number;               // 最大重试次数
    retryInterval: number;          // 重试间隔（秒）
    quietHours?: {                  // 静默时段
      enabled: boolean;
      start: string;                // HH:mm 格式
      end: string;
    };
  };
}
```

### NotificationRecord（通知记录）

```typescript
interface NotificationRecord {
  id: string;
  alertId: string;                  // 关联告警ID
  ruleId: string;                   // 关联规则ID
  
  method: 'email' | 'sms' | 'webhook' | 'inApp';
  recipient: string;                // 接收人
  
  // 内容
  subject?: string;
  content: string;
  
  // 状态
  status: 'pending' | 'sent' | 'failed' | 'delivered';
  
  // 时间
  createdAt: Date;
  sentAt?: Date;
  deliveredAt?: Date;
  
  // 错误信息
  error?: string;
  retryCount: number;
}
```

---

## 接口详情

### 告警规则管理

#### 1. 获取告警规则列表

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/alerts/rules`
- 认证: 需要Access Token
- 权限: `alert:read`

**查询参数**
```
GET /api/v1/alerts/rules?page=1&pageSize=20&scope=vm&enabled=true
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": "rule_001",
        "name": "CPU高使用率告警",
        "description": "当CPU使用率超过80%时触发",
        "scope": "vm",
        "scopeId": "vm_001",
        "scopeName": "web-server-01",
        "conditions": [
          {
            "id": "cond_001",
            "metric": "cpu",
            "metricType": "cpu.usagePercent",
            "operator": ">=",
            "threshold": 80,
            "duration": 300,
            "aggregation": "avg"
          }
        ],
        "conditionLogic": "and",
        "enabled": true,
        "cooldown": 600,
        "severity": "high",
        "notifications": {
          "methods": ["email", "inApp"],
          "email": {
            "enabled": true,
            "recipients": ["admin@company.com"]
          },
          "inApp": {
            "enabled": true
          }
        },
        "createdAt": "2026-01-01T00:00:00Z",
        "updatedAt": "2026-02-03T10:00:00Z",
        "createdBy": "usr_001",
        "triggerCount": 15,
        "lastTriggeredAt": "2026-02-03T08:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 50,
      "totalPages": 3
    }
  }
}
```

---

#### 2. 创建告警规则

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/alerts/rules`
- 认证: 需要Access Token
- 权限: `alert:write`

**请求参数**
```json
{
  "name": "内存使用率告警",
  "description": "监控内存使用率",
  "scope": "group",
  "scopeId": "grp_001",
  "conditions": [
    {
      "metric": "memory",
      "metricType": "memory.usagePercent",
      "operator": ">=",
      "threshold": 85,
      "duration": 180
    }
  ],
  "conditionLogic": "and",
  "enabled": true,
  "cooldown": 300,
  "severity": "medium",
  "notifications": {
    "methods": ["email", "webhook"],
    "email": {
      "enabled": true,
      "recipients": ["ops@company.com"]
    },
    "webhook": {
      "enabled": true,
      "url": "https://hooks.slack.com/services/xxx",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      }
    }
  }
}
```

**成功响应 (201)**
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": "rule_new_001",
    "name": "内存使用率告警",
    "enabled": true,
    "createdAt": "2026-02-03T13:30:00Z"
  }
}
```

---

#### 3. 批量启用/禁用规则

**基本信息**
- 方法: `PUT`
- 路径: `/api/v1/alerts/rules/batch/status`
- 认证: 需要Access Token
- 权限: `alert:write`

**请求参数**
```json
{
  "ruleIds": ["rule_001", "rule_002", "rule_003"],
  "enabled": false
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "批量更新成功",
  "data": {
    "updated": 3,
    "failed": 0
  }
}
```

---

#### 4. 导出规则

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/alerts/rules/export`
- 认证: 需要Access Token
- 权限: `alert:read`

**请求参数**
```json
{
  "ruleIds": ["rule_001", "rule_002"],
  "format": "json"
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "导出成功",
  "data": {
    "content": "{\"rules\":[{\"id\":\"rule_001\",...}]}",
    "filename": "alert_rules_20260203.json",
    "downloadUrl": "/api/v1/alerts/rules/export/download?token=xxx"
  }
}
```

---

### 告警记录管理

#### 5. 获取告警记录列表

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/alerts/records`
- 认证: 需要Access Token
- 权限: `alert:read`

**查询参数**
```
GET /api/v1/alerts/records?page=1&pageSize=20&status=active&severity=high&vmId=vm_001
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": "alert_001",
        "ruleId": "rule_001",
        "ruleName": "CPU高使用率告警",
        "vmId": "vm_001",
        "vmName": "web-server-01",
        "metric": "cpu",
        "severity": "high",
        "triggerValue": 85.5,
        "threshold": 80,
        "condition": "CPU使用率 >= 80%",
        "triggeredAt": "2026-02-03T12:30:00Z",
        "status": "active",
        "notifications": [
          {
            "method": "email",
            "status": "sent",
            "sentAt": "2026-02-03T12:30:05Z"
          }
        ],
        "snapshot": {
          "cpu": {
            "usagePercent": 85.5,
            "usageMHz": 3420
          },
          "memory": {
            "usagePercent": 60.2
          }
        },
        "createdAt": "2026-02-03T12:30:00Z",
        "updatedAt": "2026-02-03T12:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 156,
      "totalPages": 8
    }
  }
}
```

---

#### 6. 确认告警

**基本信息**
- 方法: `PUT`
- 路径: `/api/v1/alerts/records/{id}/acknowledge`
- 认证: 需要Access Token
- 权限: `alert:write`

**请求参数**
```json
{
  "note": "已检查，为正常业务高峰"
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "确认成功",
  "data": {
    "id": "alert_001",
    "status": "acknowledged",
    "acknowledgedBy": "usr_002",
    "acknowledgedByName": "运维工程师",
    "acknowledgedAt": "2026-02-03T13:00:00Z",
    "acknowledgeNote": "已检查，为正常业务高峰"
  }
}
```

---

#### 7. 批量确认告警

**基本信息**
- 方法: `PUT`
- 路径: `/api/v1/alerts/records/batch/acknowledge`
- 认证: 需要Access Token
- 权限: `alert:write`

**请求参数**
```json
{
  "alertIds": ["alert_001", "alert_002", "alert_003"],
  "note": "批量确认"
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "批量确认成功",
  "data": {
    "acknowledged": 3,
    "failed": 0
  }
}
```

---

#### 8. 解决告警

**基本信息**
- 方法: `PUT`
- 路径: `/api/v1/alerts/records/{id}/resolve`
- 认证: 需要Access Token
- 权限: `alert:write`

**请求参数**
```json
{
  "resolution": "重启应用服务后恢复正常"
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "解决成功",
  "data": {
    "id": "alert_001",
    "status": "resolved",
    "resolvedBy": "usr_002",
    "resolvedByName": "运维工程师",
    "resolvedAt": "2026-02-03T14:00:00Z",
    "resolution": "重启应用服务后恢复正常",
    "duration": 5400
  }
}
```

---

#### 9. 获取告警统计

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/alerts/statistics`
- 认证: 需要Access Token
- 权限: `alert:read`

**查询参数**
```
GET /api/v1/alerts/statistics?startTime=2026-02-01T00:00:00Z&endTime=2026-02-03T23:59:59Z
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "overview": {
      "totalRules": 50,
      "activeRules": 45,
      "totalAlerts": 156,
      "activeAlerts": 8,
      "acknowledgedAlerts": 12,
      "resolvedAlerts": 136
    },
    "bySeverity": {
      "critical": { "total": 3, "active": 0 },
      "high": { "total": 28, "active": 3 },
      "medium": { "total": 85, "active": 4 },
      "low": { "total": 40, "active": 1 }
    },
    "byMetric": [
      { "metric": "cpu", "count": 45, "activeCount": 3 },
      { "metric": "memory", "count": 38, "activeCount": 2 },
      { "metric": "disk", "count": 42, "activeCount": 2 },
      { "metric": "network", "count": 31, "activeCount": 1 }
    ],
    "byVM": [
      { "vmId": "vm_005", "vmName": "db-server-01", "count": 15, "activeCount": 2 },
      { "vmId": "vm_001", "vmName": "web-server-01", "count": 12, "activeCount": 1 }
    ],
    "mttr": {
      "avg": 45.5,
      "bySeverity": {
        "critical": 12.3,
        "high": 38.5,
        "medium": 52.1,
        "low": 78.6
      }
    },
    "timeRange": {
      "start": "2026-02-01T00:00:00Z",
      "end": "2026-02-03T23:59:59Z"
    }
  }
}
```

---

### 通知配置

#### 10. 获取通知配置

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/alerts/notifications/config`
- 认证: 需要Access Token
- 权限: `alert:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "email": {
      "enabled": true,
      "smtp": {
        "host": "smtp.company.com",
        "port": 587,
        "secure": true
      },
      "from": "alerts@company.com",
      "fromName": "VM监控系统"
    },
    "sms": {
      "enabled": false
    },
    "defaultTemplates": {
      "email": "alert_email_template",
      "sms": "alert_sms_template",
      "webhook": "alert_webhook_template"
    },
    "globalPolicy": {
      "maxRetry": 3,
      "retryInterval": 60,
      "quietHours": {
        "enabled": true,
        "start": "23:00",
        "end": "07:00"
      }
    }
  }
}
```

---

#### 11. 测试通知

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/alerts/notifications/test`
- 认证: 需要Access Token
- 权限: `alert:write`

**请求参数**
```json
{
  "method": "email",
  "recipient": "test@company.com"
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "测试通知已发送",
  "data": {
    "status": "sent",
    "sentAt": "2026-02-03T13:30:00Z"
  }
}
```

---

## 错误码定义

| 错误码 | 英文消息 | 中文消息 | 日文消息 | 说明 |
|--------|---------|---------|---------|------|
| 400 | Bad Request | 请求参数错误 | リクエストパラメータエラー | 参数缺失或格式错误 |
| 400-LIMIT | Rule Limit Exceeded | 规则数量超过限制 | ルール数が制限を超えています | 单VM超50条或全局超500条 |
| 401 | Unauthorized | 未授权 | 未認証 | Token无效或过期 |
| 403 | Forbidden | 权限不足 | アクセス権限がありません | 无权限管理告警 |
| 404 | Not Found | 告警规则不存在 | アラートルールが見つかりません | 规则ID不存在 |
| 404-ALERT | Alert Not Found | 告警记录不存在 | アラート記録が見つかりません | 告警ID不存在 |
| 409 | Conflict | 规则名称已存在 | ルール名が既に存在します | 规则名称重复 |
| 422 | Invalid Condition | 告警条件无效 | アラート条件が無効です | 条件配置错误 |
| 500 | Server Error | 服务器内部错误 | サーバーエラー | 服务器错误 |

---

## 变更记录

### 版本 v1.0 (2026-02-03)
**修改人**: BE工程师  
**修改原因**: 基于REQ_20260202_VM监控系统需求文档初始创建  
**具体修改**:
- [x] 新增告警规则CRUD接口
- [x] 新增告警记录管理接口（确认/解决/忽略）
- [x] 新增告警统计与趋势接口
- [x] 新增通知配置与测试接口
- [x] 新增规则导入导出功能
- [x] 定义告警规则、条件、通知配置模型
- [x] 定义告警记录、统计模型
- [x] 支持多条件组合和复杂通知策略

**影响范围**:
- 前端界面: 是（告警管理页面、告警规则配置弹窗、告警列表）
- 后端API: 是（告警引擎、规则服务、通知服务）
- 数据库结构: 是（alert_rules, alert_records, notifications表）
- 部署配置: 是（邮件服务器配置、告警引擎配置）

**相关文档**:
- REQ_20260202_VM监控系统.md（告警规则定义、基础告警系统）
- UI_20260202_VM监控系统_视觉设计指南.md（告警管理页面）
- API_REALTIME_实时监控模块.md（告警推送WebSocket）

---

**文档管理说明**:
1. 告警规则变更实时生效，无需重启服务
2. 告警触发条件支持多条件组合（AND/OR逻辑）
3. 通知升级策略支持多级延迟通知
4. 告警历史保留2年，支持审计追溯
5. 字段变更需记录在`api-changes.md`
