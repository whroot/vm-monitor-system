# API_SYSTEM_系统健康模块_API规范

## 文档履历

| 版本 | 日期 | 修改人 | 修改内容 | 审核状态 |
|------|------|--------|----------|----------|
| v1.0 | 2026-02-03 | BE工程师 | 初始版本，基于REQ_20260202和UI_20260202生成 | 🔄 待审核 |

---

## 模块概述

### 功能范围
- 系统健康评分计算与展示
- 监控系统自监控（采集器/存储/API服务状态）
- 系统性能指标（响应时间/吞吐量/错误率）
- 容量监控与预警
- 系统配置管理
- 日志查询与审计

### 适用角色
- 系统管理员：全部权限
- 运维工程师：查看监控、处理告警
- IT经理：查看概览、容量规划

### 技术约束
- 自监控间隔：30秒
- 健康评分计算周期：5分钟
- 系统日志保留：90天
- 审计日志保留：2年

---

## 接口清单

### 系统概览

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取系统概览 | GET | /api/v1/system/overview | 获取系统整体状态 | 需要system:read权限 |
| 获取健康评分 | GET | /api/v1/system/health-score | 获取系统健康评分详情 | 需要system:read权限 |
| 获取健康趋势 | GET | /api/v1/system/health-trend | 获取健康评分历史趋势 | 需要system:read权限 |

### 服务状态

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取服务状态 | GET | /api/v1/system/services | 获取各服务健康状态 | 需要system:read权限 |
| 获取采集器状态 | GET | /api/v1/system/collectors | 获取数据采集器状态 | 需要system:read权限 |
| 获取存储状态 | GET | /api/v1/system/storage | 获取存储系统状态 | 需要system:read权限 |

### 性能指标

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取性能指标 | GET | /api/v1/system/performance | 获取API性能指标 | 需要system:read权限 |
| 获取容量信息 | GET | /api/v1/system/capacity | 获取系统容量使用情况 | 需要system:read权限 |

### 系统配置

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取系统配置 | GET | /api/v1/system/config | 获取系统配置项 | 需要system:admin权限 |
| 更新系统配置 | PUT | /api/v1/system/config | 更新系统配置 | 需要system:admin权限 |
| 获取配置历史 | GET | /api/v1/system/config/history | 查询配置变更历史 | 需要system:admin权限 |

### 日志审计

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 查询系统日志 | GET | /api/v1/system/logs | 查询系统运行日志 | 需要system:read权限 |
| 查询审计日志 | GET | /api/v1/system/audit-logs | 查询操作审计日志 | 需要system:admin权限 |
| 导出日志 | POST | /api/v1/system/logs/export | 导出日志文件 | 需要system:read权限 |

### 系统维护

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 执行数据清理 | POST | /api/v1/system/maintenance/cleanup | 执行数据清理任务 | 需要system:admin权限 |
| 获取任务列表 | GET | /api/v1/system/maintenance/tasks | 获取系统任务列表 | 需要system:read权限 |
| 获取任务详情 | GET | /api/v1/system/maintenance/tasks/{id} | 获取任务执行详情 | 需要system:read权限 |

---

## 数据模型

### SystemOverview（系统概览）

```typescript
interface SystemOverview {
  timestamp: Date;                  // 数据时间戳
  
  // 系统整体状态
  status: 'healthy' | 'degraded' | 'unhealthy' | 'maintenance';
  
  // 健康评分
  healthScore: {
    current: number;                // 当前评分（0-100）
    level: 'excellent' | 'good' | 'warning' | 'critical';
    trend: 'up' | 'down' | 'stable';
    change: number;                 // 变化值
  };
  
  // VM监控状态
  vmMonitoring: {
    totalVMs: number;               // VM总数
    onlineVMs: number;              // 在线VM数
    offlineVMs: number;             // 离线VM数
    errorVMs: number;               // 错误VM数
    collectionRate: number;         // 采集成功率（%）
    avgCollectionTime: number;      // 平均采集时间（秒）
  };
  
  // 告警统计
  alerts: {
    critical: number;
    high: number;
    medium: number;
    low: number;
    total: number;
  };
  
  // 核心服务状态
  services: {
    api: ServiceStatus;           // API服务
    collector: ServiceStatus;     // 采集服务
    database: ServiceStatus;      // 数据库服务
    cache: ServiceStatus;         // 缓存服务
    websocket: ServiceStatus;     // WebSocket服务
  };
  
  // 系统运行时间
  uptime: {
    system: number;                 // 系统运行时间（秒）
    api: number;                    // API服务运行时间
    collector: number;              // 采集服务运行时间
  };
  
  // 版本信息
  version: {
    backend: string;                // 后端版本
    frontend?: string;              // 前端版本（预留）
    database: string;             // 数据库版本
  };
}

interface ServiceStatus {
  status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  lastCheck: Date;
  responseTime?: number;          // 响应时间（ms）
  errorRate?: number;             // 错误率（%）
  message?: string;               // 状态说明
}
```

### HealthScoreDetail（健康评分详情）

```typescript
interface HealthScoreDetail {
  current: number;                  // 当前评分（0-100）
  level: 'excellent' | 'good' | 'warning' | 'critical';
  
  // 评分维度
  dimensions: Array<{
    name: string;                   // 维度名称
    weight: number;                 // 权重（%）
    score: number;                  // 该维度得分
    status: 'healthy' | 'warning' | 'critical';
    details: string;                // 详细说明
  }>;
  
  // 评分计算依据
  factors: {
    vmOnlineRate: {                 // VM在线率
      weight: 30;
      score: number;
      actual: number;               // 实际在线率
      target: number;             // 目标在线率（99%）
    };
    collectionSuccessRate: {      // 采集成功率
      weight: 25;
      score: number;
      actual: number;
      target: number;             // 目标成功率（98%）
    };
    alertResolutionRate: {        // 告警解决率
      weight: 20;
      score: number;
      actual: number;
      target: number;             // 目标解决率（95%）
    };
    apiAvailability: {           // API可用性
      weight: 15;
      score: number;
      actual: number;
      target: number;             // 目标可用性（99.9%）
    };
    storageHealth: {              // 存储健康度
      weight: 10;
      score: number;
      actual: number;
      target: number;
    };
  };
  
  // 最近24小时趋势
  history: Array<{
    timestamp: Date;
    score: number;
  }>;
  
  calculatedAt: Date;
}
```

### ServiceHealth（服务健康详情）

```typescript
interface ServiceHealth {
  id: string;                       // 服务ID
  name: string;                     // 服务名称
  type: 'api' | 'collector' | 'database' | 'cache' | 'websocket' | 'notification';
  
  // 当前状态
  status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown' | 'maintenance';
  statusMessage?: string;
  
  // 运行信息
  version: string;                // 服务版本
  uptime: number;                 // 运行时间（秒）
  startedAt: Date;                // 启动时间
  
  // 性能指标
  performance: {
    requestCount: number;           // 请求总数（最近1小时）
    avgResponseTime: number;      // 平均响应时间（ms）
    p95ResponseTime: number;      // P95响应时间
    p99ResponseTime: number;      // P99响应时间
    errorRate: number;            // 错误率（%）
    throughput: number;           // 吞吐量（QPS）
  };
  
  // 资源使用
  resources?: {
    cpuPercent: number;
    memoryPercent: number;
    memoryUsedMB: number;
    memoryTotalMB: number;
  };
  
  // 依赖服务状态
  dependencies: Array<{
    serviceId: string;
    serviceName: string;
    status: 'healthy' | 'degraded' | 'unhealthy';
    latency: number;
  }>;
  
  lastCheckAt: Date;
}
```

### CollectorStatus（采集器状态）

```typescript
interface CollectorStatus {
  id: string;                       // 采集器ID
  name: string;                     // 采集器名称
  host: string;                     // 所在主机
  
  // 状态
  status: 'running' | 'stopped' | 'error' | 'maintenance';
  statusMessage?: string;
  
  // 采集配置
  config: {
    interval: number;               // 采集间隔（秒）
    batchSize: number;              // 批次大小
    timeout: number;                // 超时时间（秒）
  };
  
  // 采集统计（最近1小时）
  statistics: {
    totalTasks: number;             // 总任务数
    successTasks: number;           // 成功任务数
    failedTasks: number;            // 失败任务数
    avgTaskTime: number;            // 平均任务时间（秒）
    lastSuccessAt?: Date;
    lastFailureAt?: Date;
  };
  
  // VM采集分布
  vmDistribution: Array<{
    datacenterId: string;
    datacenterName: string;
    vmCount: number;
    avgCollectionTime: number;
  }>;
  
  // 资源使用
  resources: {
    cpuPercent: number;
    memoryMB: number;
    goroutines: number;
  };
  
  connectedAt: Date;                // 连接时间
  lastHeartbeat: Date;              // 最后心跳
}
```

### StorageStatus（存储状态）

```typescript
interface StorageStatus {
  // 数据库状态
  database: {
    type: 'mysql' | 'postgresql' | 'timescaledb';
    version: string;
    status: 'healthy' | 'degraded' | 'unhealthy';
    
    // 连接池
    connections: {
      active: number;
      idle: number;
      max: number;
    };
    
    // 性能
    performance: {
      qps: number;
      avgQueryTime: number;
      slowQueries: number;          // 慢查询数（最近1小时）
    };
  };
  
  // 磁盘使用
  disk: {
    totalGB: number;
    usedGB: number;
    freeGB: number;
    usagePercent: number;
    
    // 数据文件分布
    dataFiles: Array<{
      name: string;
      sizeGB: number;
      path: string;
    }>;
    
    // 存储分层
    tiers: {
      hot: { usedGB: number; retention: string };
      warm: { usedGB: number; retention: string };
      cold: { usedGB: number; retention: string };
    };
  };
  
  // 缓存状态（Redis等）
  cache?: {
    type: string;
    status: 'healthy' | 'degraded' | 'unhealthy';
    memoryUsedMB: number;
    memoryTotalMB: number;
    hitRate: number;
    connectedClients: number;
  };
}
```

### PerformanceMetrics（性能指标）

```typescript
interface PerformanceMetrics {
  // 时间范围
  timeRange: {
    start: Date;
    end: Date;
  };
  
  // API性能
  api: {
    requestCount: number;
    successCount: number;
    errorCount: number;
    
    responseTime: {
      avg: number;
      min: number;
      max: number;
      p50: number;
      p95: number;
      p99: number;
    };
    
    // 按接口统计
    endpoints: Array<{
      path: string;
      method: string;
      count: number;
      avgResponseTime: number;
      errorRate: number;
    }>;
  };
  
  // WebSocket性能
  websocket: {
    connectionCount: number;
    messageCount: number;
    avgLatency: number;
  };
  
  // 数据库性能
  database: {
    queryCount: number;
    avgQueryTime: number;
    slowQueries: number;
  };
}
```

### CapacityInfo（容量信息）

```typescript
interface CapacityInfo {
  // 存储容量
  storage: {
    totalGB: number;
    usedGB: number;
    freeGB: number;
    usagePercent: number;
    
    // 预测
    forecast: {
      dailyGrowthGB: number;        // 日增长量
      daysUntilFull: number;       // 预计满天数
      warningAt: Date;            // 预计达到警告线日期
    };
    
    // 按数据类型
    byType: Array<{
      type: string;
      sizeGB: number;
      percent: number;
      retention: string;
    }>;
  };
  
  // VM容量
  vmCapacity: {
    current: number;              // 当前监控VM数
    max: number;                  // 最大容量
    usagePercent: number;
    
    // 扩展建议
    recommendation?: {
      canAdd: number;             // 还可添加数量
      suggestion: string;
    };
  };
  
  // 告警规则容量
  alertRuleCapacity: {
    current: number;
    max: number;
    usagePercent: number;
  };
  
  // 用户容量
  userCapacity: {
    current: number;
    max: number;
    usagePercent: number;
  };
}
```

### SystemConfig（系统配置）

```typescript
interface SystemConfig {
  // 采集配置
  collection: {
    interval: number;               // 采集间隔（秒，默认30）
    timeout: number;                // 采集超时（秒，默认10）
    retryCount: number;             // 重试次数（默认3）
    batchSize: number;              // 批次大小（默认50）
  };
  
  // 数据保留策略
  retention: {
    rawData: number;                // 原始数据保留天数（默认7）
    hourAggregation: number;        // 小时聚合保留天数（默认30）
    dayAggregation: number;       // 天聚合保留天数（默认730，2年）
    alertHistory: number;           // 告警历史保留天数（默认730）
    auditLog: number;               // 审计日志保留天数（默认730）
    systemLog: number;              // 系统日志保留天数（默认90）
  };
  
  // 告警配置
  alerting: {
    evaluationInterval: number;     // 告警评估间隔（秒，默认60）
    cooldown: number;               // 默认冷却时间（秒，默认300）
    maxRulesPerVM: number;          // 单VM最大规则数（默认50）
    maxGlobalRules: number;         // 全局最大规则数（默认500）
  };
  
  // 性能配置
  performance: {
    maxQueryRange: number;          // 最大查询时间范围（天，默认365）
    maxExportRecords: number;       // 最大导出记录数（默认100000）
    cacheTTL: number;               // 缓存过期时间（秒，默认900）
  };
  
  // 安全配置
  security: {
    maxLoginAttempts: number;       // 最大登录尝试次数（默认5）
    lockoutDuration: number;      // 锁定时间（分钟，默认30）
    passwordExpiry: number;       // 密码过期天数（默认90）
    sessionTimeout: number;       // 会话超时（分钟，默认60）
    passwordComplexity: {
      minLength: number;
      requireUppercase: boolean;
      requireLowercase: boolean;
      requireNumbers: boolean;
      requireSpecial: boolean;
    };
  };
}
```

### SystemLog（系统日志）

```typescript
interface SystemLog {
  id: string;
  timestamp: Date;
  
  // 日志级别
  level: 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  
  // 日志来源
  source: string;                   // 服务/模块名称
  instance: string;                 // 实例标识
  
  // 日志内容
  message: string;
  details?: object;                 // 详细信息
  
  // 上下文
  traceId?: string;                 // 追踪ID
  requestId?: string;               // 请求ID
  userId?: string;                  // 用户ID（如适用）
  
  // 位置信息
  file?: string;
  line?: number;
  function?: string;
}
```

### MaintenanceTask（维护任务）

```typescript
interface MaintenanceTask {
  id: string;
  name: string;
  type: 'cleanup' | 'optimize' | 'backup' | 'custom';
  
  // 任务状态
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  
  // 任务配置
  config: {
    target?: string;                // 操作目标
    params?: Record<string, any>;   // 参数
  };
  
  // 执行信息
  schedule?: {
    type: 'immediate' | 'once' | 'recurring';
    cron?: string;                  // 定时表达式
    nextRun?: Date;
  };
  
  // 执行结果
  result?: {
    startTime: Date;
    endTime?: Date;
    duration?: number;
    message?: string;
    details?: object;
  };
  
  // 操作者
  createdBy: string;
  startedBy?: string;
  
  // 时间
  createdAt: Date;
  startedAt?: Date;
  completedAt?: Date;
}
```

---

## 接口详情

### 系统概览

#### 1. 获取系统概览

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/system/overview`
- 认证: 需要Access Token
- 权限: `system:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "timestamp": "2026-02-03T14:30:00Z",
    "status": "healthy",
    "healthScore": {
      "current": 87,
      "level": "good",
      "trend": "stable",
      "change": 0
    },
    "vmMonitoring": {
      "totalVMs": 150,
      "onlineVMs": 140,
      "offlineVMs": 5,
      "errorVMs": 5,
      "collectionRate": 98.5,
      "avgCollectionTime": 25.3
    },
    "alerts": {
      "critical": 0,
      "high": 3,
      "medium": 8,
      "low": 15,
      "total": 26
    },
    "services": {
      "api": {
        "status": "healthy",
        "lastCheck": "2026-02-03T14:30:00Z",
        "responseTime": 45,
        "errorRate": 0.01
      },
      "collector": {
        "status": "healthy",
        "lastCheck": "2026-02-03T14:30:00Z"
      },
      "database": {
        "status": "healthy",
        "lastCheck": "2026-02-03T14:30:00Z"
      },
      "cache": {
        "status": "healthy",
        "lastCheck": "2026-02-03T14:30:00Z"
      },
      "websocket": {
        "status": "healthy",
        "lastCheck": "2026-02-03T14:30:00Z"
      }
    },
    "uptime": {
      "system": 2592000,
      "api": 2592000,
      "collector": 2592000
    },
    "version": {
      "backend": "v1.0.0",
      "database": "PostgreSQL 14.5"
    }
  }
}
```

---

#### 2. 获取健康评分详情

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/system/health-score`
- 认证: 需要Access Token
- 权限: `system:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "current": 87,
    "level": "good",
    "dimensions": [
      {
        "name": "VM在线率",
        "weight": 30,
        "score": 93,
        "status": "healthy",
        "details": "当前在线率93.3%，目标99%"
      },
      {
        "name": "采集成功率",
        "weight": 25,
        "score": 98,
        "status": "healthy",
        "details": "采集成功率98.5%，目标98%"
      },
      {
        "name": "告警解决率",
        "weight": 20,
        "score": 85,
        "status": "warning",
        "details": "24小时内告警解决率85%，目标95%"
      },
      {
        "name": "API可用性",
        "weight": 15,
        "score": 100,
        "status": "healthy",
        "details": "API可用性100%，目标99.9%"
      },
      {
        "name": "存储健康度",
        "weight": 10,
        "score": 70,
        "status": "warning",
        "details": "磁盘使用率70%，建议清理"
      }
    ],
    "factors": {
      "vmOnlineRate": {
        "weight": 30,
        "score": 93,
        "actual": 93.3,
        "target": 99
      },
      "collectionSuccessRate": {
        "weight": 25,
        "score": 98,
        "actual": 98.5,
        "target": 98
      }
    },
    "history": [
      {
        "timestamp": "2026-02-02T14:30:00Z",
        "score": 85
      },
      {
        "timestamp": "2026-02-03T14:30:00Z",
        "score": 87
      }
    ],
    "calculatedAt": "2026-02-03T14:30:00Z"
  }
}
```

---

### 服务状态

#### 3. 获取采集器状态

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/system/collectors`
- 认证: 需要Access Token
- 权限: `system:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "collectors": [
      {
        "id": "collector_001",
        "name": "主采集器",
        "host": "monitor-server-01",
        "status": "running",
        "config": {
          "interval": 30,
          "batchSize": 50,
          "timeout": 10
        },
        "statistics": {
          "totalTasks": 5000,
          "successTasks": 4925,
          "failedTasks": 75,
          "avgTaskTime": 23.5,
          "lastSuccessAt": "2026-02-03T14:29:30Z"
        },
        "vmDistribution": [
          {
            "datacenterId": "dc_001",
            "datacenterName": "数据中心A",
            "vmCount": 150,
            "avgCollectionTime": 25.3
          }
        ],
        "resources": {
          "cpuPercent": 25.5,
          "memoryMB": 512,
          "goroutines": 150
        },
        "connectedAt": "2026-01-01T00:00:00Z",
        "lastHeartbeat": "2026-02-03T14:30:00Z"
      }
    ]
  }
}
```

---

#### 4. 获取容量信息

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/system/capacity`
- 认证: 需要Access Token
- 权限: `system:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "storage": {
      "totalGB": 2000,
      "usedGB": 850,
      "freeGB": 1150,
      "usagePercent": 42.5,
      "forecast": {
        "dailyGrowthGB": 2.5,
        "daysUntilFull": 460,
        "warningAt": "2027-05-15T00:00:00Z"
      },
      "byType": [
        {
          "type": "raw_data",
          "sizeGB": 150,
          "percent": 17.6,
          "retention": "7 days"
        },
        {
          "type": "hour_aggregation",
          "sizeGB": 300,
          "percent": 35.3,
          "retention": "30 days"
        },
        {
          "type": "day_aggregation",
          "sizeGB": 400,
          "percent": 47.1,
          "retention": "2 years"
        }
      ],
      "tiers": {
        "hot": {
          "usedGB": 150,
          "retention": "7 days"
        },
        "warm": {
          "usedGB": 300,
          "retention": "30 days"
        },
        "cold": {
          "usedGB": 400,
          "retention": "2 years"
        }
      }
    },
    "vmCapacity": {
      "current": 150,
      "max": 5000,
      "usagePercent": 3,
      "recommendation": {
        "canAdd": 4850,
        "suggestion": "当前容量充足，可继续添加VM"
      }
    },
    "alertRuleCapacity": {
      "current": 50,
      "max": 500,
      "usagePercent": 10
    },
    "userCapacity": {
      "current": 45,
      "max": 500,
      "usagePercent": 9
    }
  }
}
```

---

### 系统配置

#### 5. 获取系统配置

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/system/config`
- 认证: 需要Access Token
- 权限: `system:admin`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "collection": {
      "interval": 30,
      "timeout": 10,
      "retryCount": 3,
      "batchSize": 50
    },
    "retention": {
      "rawData": 7,
      "hourAggregation": 30,
      "dayAggregation": 730,
      "alertHistory": 730,
      "auditLog": 730,
      "systemLog": 90
    },
    "alerting": {
      "evaluationInterval": 60,
      "cooldown": 300,
      "maxRulesPerVM": 50,
      "maxGlobalRules": 500
    },
    "performance": {
      "maxQueryRange": 365,
      "maxExportRecords": 100000,
      "cacheTTL": 900
    },
    "security": {
      "maxLoginAttempts": 5,
      "lockoutDuration": 30,
      "passwordExpiry": 90,
      "sessionTimeout": 60,
      "passwordComplexity": {
        "minLength": 8,
        "requireUppercase": true,
        "requireLowercase": true,
        "requireNumbers": true,
        "requireSpecial": true
      }
    }
  }
}
```

---

#### 6. 更新系统配置

**基本信息**
- 方法: `PUT`
- 路径: `/api/v1/system/config`
- 认证: 需要Access Token
- 权限: `system:admin`

**请求参数**
```json
{
  "collection": {
    "interval": 60,
    "timeout": 15,
    "retryCount": 3,
    "batchSize": 50
  },
  "retention": {
    "rawData": 7,
    "hourAggregation": 30,
    "dayAggregation": 730,
    "alertHistory": 730,
    "auditLog": 730,
    "systemLog": 90
  }
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "配置更新成功",
  "data": {
    "updatedAt": "2026-02-03T14:35:00Z",
    "affectedModules": ["collector", "storage"]
  }
}
```

---

### 日志审计

#### 7. 查询系统日志

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/system/logs`
- 认证: 需要Access Token
- 权限: `system:read`

**查询参数**
```
GET /api/v1/system/logs?page=1&pageSize=50&level=error&startTime=2026-02-03T00:00:00Z&source=collector
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": "log_001",
        "timestamp": "2026-02-03T14:30:00Z",
        "level": "error",
        "source": "collector",
        "instance": "collector_001",
        "message": "采集VM vm_005超时",
        "details": {
          "vmId": "vm_005",
          "vmName": "db-server-01",
          "timeout": 10,
          "error": "connection timeout"
        },
        "traceId": "trace_001",
        "file": "collector/vm.go",
        "line": 156,
        "function": "CollectVMMetrics"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 50,
      "total": 1250,
      "totalPages": 25
    }
  }
}
```

---

#### 8. 查询审计日志

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/system/audit-logs`
- 认证: 需要Access Token
- 权限: `system:admin`

**查询参数**
```
GET /api/v1/system/audit-logs?page=1&pageSize=20&action=update&resourceType=user&startTime=2026-02-01T00:00:00Z
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": "audit_001",
        "timestamp": "2026-02-03T14:30:00Z",
        "action": "update",
        "resourceType": "user",
        "resourceId": "usr_002",
        "resourceName": "运维工程师01",
        "operatorId": "usr_001",
        "operatorName": "系统管理员",
        "operatorIp": "192.168.1.100",
        "changes": [
          {
            "field": "roles",
            "oldValue": ["role_viewer"],
            "newValue": ["role_operator"]
          }
        ],
        "note": "晋升为运维工程师"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 500,
      "totalPages": 25
    }
  }
}
```

---

### 系统维护

#### 9. 执行数据清理

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/system/maintenance/cleanup`
- 认证: 需要Access Token
- 权限: `system:admin`

**请求参数**
```json
{
  "type": "cleanup",
  "target": "expired_data",
  "params": {
    "olderThan": "2025-02-03T00:00:00Z",
    "dryRun": false
  }
}
```

**成功响应 (202)**
```json
{
  "code": 202,
  "message": "清理任务已创建",
  "data": {
    "taskId": "task_cleanup_001",
    "status": "pending",
    "createdAt": "2026-02-03T14:40:00Z"
  }
}
```

---

## 错误码定义

| 错误码 | 英文消息 | 中文消息 | 日文消息 | 说明 |
|--------|---------|---------|---------|------|
| 400 | Bad Request | 请求参数错误 | リクエストパラメータエラー | 参数缺失或格式错误 |
| 401 | Unauthorized | 未授权 | 未認証 | Token无效或过期 |
| 403 | Forbidden | 权限不足 | アクセス権限がありません | 无系统管理权限 |
| 404 | Not Found | 配置项不存在 | 設定項目が見つかりません | 配置项不存在 |
| 404-TASK | Task Not Found | 维护任务不存在 | メンテナンスタスクが見つかりません | 任务ID不存在 |
| 409 | Conflict | 配置冲突 | 設定が競合しています | 配置项冲突 |
| 422 | Invalid Config | 配置值无效 | 設定値が無効です | 配置值超出范围 |
| 500 | Server Error | 服务器内部错误 | サーバーエラー | 服务器错误 |
| 503 | Service Unavailable | 系统服务不可用 | システムサービスが利用できません | 核心服务异常 |

---

## 变更记录

### 版本 v1.0 (2026-02-03)
**修改人**: BE工程师  
**修改原因**: 基于REQ_20260202_VM监控系统需求文档初始创建  
**具体修改**:
- [x] 新增系统概览接口（健康评分、服务状态）
- [x] 新增服务健康详情接口（API/采集器/数据库/缓存）
- [x] 新增性能指标查询接口
- [x] 新增容量监控接口（存储预测、容量预警）
- [x] 新增系统配置管理接口
- [x] 新增系统日志和审计日志查询接口
- [x] 新增系统维护任务接口（数据清理）
- [x] 定义健康评分计算模型和维度
- [x] 定义容量预测和预警模型

**影响范围**:
- 前端界面: 是（系统健康页面、设置页面、日志查询页面）
- 后端API: 是（系统监控服务、配置服务、维护任务服务）
- 数据库结构: 是（system_logs, audit_logs, maintenance_tasks表）
- 部署配置: 是（系统配置中心、监控Agent配置）

**相关文档**:
- REQ_20260202_VM监控系统.md（系统健康状态总览、可靠性要求、监控自监控）
- UI_20260202_VM监控系统_视觉设计指南.md（系统健康页面、健康度评分组件）

---

**文档管理说明**:
1. 健康评分每5分钟计算一次，实时性要求不高可缓存
2. 系统配置变更实时生效，关键配置变更需二次确认
3. 容量预测基于线性回归算法，需至少30天历史数据
4. 系统日志和审计日志分离存储，保留策略不同
5. 维护任务为异步执行，需轮询查询进度
6. 字段变更需记录在`api-changes.md`
