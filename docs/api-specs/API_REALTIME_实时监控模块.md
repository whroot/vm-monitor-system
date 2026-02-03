# API_REALTIME_实时监控模块_API规范

## 文档履历

| 版本 | 日期 | 修改人 | 修改内容 | 审核状态 |
|------|------|--------|----------|----------|
| v1.0 | 2026-02-03 | BE工程师 | 初始版本，基于REQ_20260202和UI_20260202生成 | 🔄 待审核 |

---

## 模块概述

### 功能范围
- WebSocket实时数据推送（30-60秒间隔）
- 实时指标查询（当前性能数据）
- 多维度数据聚合（按集群/分组/主机）
- 异常检测标记

### 适用角色
- 系统管理员、运维工程师：全部权限
- IT经理、安全工程师：查看权限

### 技术约束
- 采集间隔：30-60秒（可配置）
- 支持1500台VM实时数据推送
- 数据保留：实时数据7天（热数据）
- 并发连接：500+用户同时访问

---

## 接口清单

### WebSocket接口

| 接口 | 路径 | 描述 | 认证方式 |
|------|------|------|----------|
| 实时数据连接 | /ws/v1/realtime | WebSocket连接，接收实时推送 | Token Query参数 |
| 订阅VM数据 | WebSocket消息 | 订阅指定VM的实时指标 | 已连接后发送 |
| 取消订阅 | WebSocket消息 | 取消订阅 | 已连接后发送 |

### REST接口

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取VM当前指标 | GET | /api/v1/realtime/vms/{id} | 获取VM当前所有指标 | 需要认证 |
| 获取多个VM指标 | POST | /api/v1/realtime/vms/batch | 批量获取VM当前指标 | 需要认证 |
| 获取分组聚合指标 | GET | /api/v1/realtime/groups/{id} | 获取分组聚合数据 | 需要认证 |
| 获取集群聚合指标 | GET | /api/v1/realtime/clusters/{id} | 获取集群聚合数据 | 需要认证 |
| 获取全局概览 | GET | /api/v1/realtime/overview | 获取系统整体实时状态 | 需要认证 |

---

## 数据模型

### RealtimeMetrics（实时指标数据）

实时指标数据合并了vSphere API指标（虚拟化层）和VMware Tools指标（操作系统层）：

```typescript
interface RealtimeMetrics {
  vmId: string;                    // VM ID
  timestamp: Date;                 // 数据采集时间戳
  
  // 数据源标记
  dataSources: {
    vsphere: boolean;              // 是否来自vSphere API
    guestOS: boolean;              // 是否来自VMware Tools
  };
  
  // ========== CPU指标 ==========
  cpu: {
    // vSphere层指标
    usageMHz?: number;             // CPU使用率(MHz)
    ready?: number;                // CPU就绪时间百分比
    wait?: number;                 // CPU等待时间百分比
    limit?: number;                // CPU限制(MHz)
    reservation?: number;        // CPU预留(MHz)
    
    // 操作系统层指标
    usagePercent?: number;         // CPU使用率百分比
    load1min?: number;             // 1分钟负载平均值
    load5min?: number;             // 5分钟负载平均值
    load15min?: number;            // 15分钟负载平均值
    contextSwitches?: number;      // 上下文切换次数
  };
  
  // ========== 内存指标 ==========
  memory: {
    // vSphere层指标
    usageMB?: number;              // 内存使用量(MB)
    grantedMB?: number;            // 已分配内存(MB)
    activeMB?: number;             // 活跃内存(MB)
    balloonedMB?: number;          // 气球内存(MB)
    compressedMB?: number;         // 压缩内存(MB)
    swappedMB?: number;            // 交换内存(MB)
    
    // 操作系统层指标
    totalMB?: number;              // 总内存(MB)
    usedMB?: number;               // 已用内存(MB)
    freeMB?: number;               // 可用内存(MB)
    buffersMB?: number;            // 缓冲区内存(MB)
    cachedMB?: number;             // 缓存内存(MB)
    swapUsedMB?: number;           // 交换分区使用(MB)
    usagePercent?: number;         // 内存使用率百分比
    
    // Windows特有
    availableMB?: number;          // Windows可用内存
    committedMB?: number;          // Windows已提交内存
  };
  
  // ========== 磁盘指标 ==========
  disk: {
    // vSphere层指标
    readLatency?: number;        // 磁盘读取延迟(ms)
    writeLatency?: number;         // 磁盘写入延迟(ms)
    readIOPS?: number;             // 磁盘读取IOPS
    writeIOPS?: number;            // 磁盘写入IOPS
    throughputMBps?: number;       // 磁盘吞吐量(MB/s)
    
    // 操作系统层指标
    usagePercent?: number;         // 磁盘使用率百分比
    usedMB?: number;               // 已用空间(MB)
    freeMB?: number;               // 可用空间(MB)
    readBytes?: number;            // 磁盘读取字节数
    writeBytes?: number;           // 磁盘写入字节数
    inodesTotal?: number;          // 总inode数(Linux)
    inodesUsed?: number;           // 已用inode数(Linux)
  };
  
  // ========== 网络指标 ==========
  network: {
    // vSphere层指标
    inBps?: number;                // 网络入流量(bps)
    outBps?: number;               // 网络出流量(bps)
    inPps?: number;                // 入包数(pps)
    outPps?: number;               // 出包数(pps)
    droppedPackets?: number;       // 丢包数
    
    // 操作系统层指标
    inBytes?: number;              // 网络入流量字节数
    outBytes?: number;             // 网络出流量字节数
    inPackets?: number;            // 入包数
    outPackets?: number;           // 出包数
    errors?: number;               // 网络错误包数
    dropped?: number;              // 丢包数
  };
  
  // ========== 系统指标 ==========
  system: {
    uptime?: number;               // 系统运行时间(秒)
    processTotal?: number;         // 总进程数
    processRunning?: number;       // 运行中进程数
    processSleeping?: number;      // 休眠进程数(Linux)
  };
  
  // ========== 告警标记 ==========
  alerts?: Array<{
    metric: 'cpu' | 'memory' | 'disk' | 'network';
    severity: 'low' | 'medium' | 'high' | 'critical';
    threshold: number;
    currentValue: number;
    message: string;
  }>;
}
```

### AggregatedMetrics（聚合指标）

```typescript
interface AggregatedMetrics {
  scope: 'global' | 'datacenter' | 'cluster' | 'group' | 'host';
  scopeId: string;
  scopeName: string;
  timestamp: Date;
  
  // VM统计
  vmCount: {
    total: number;
    online: number;
    offline: number;
    error: number;
  };
  
  // CPU聚合
  cpu: {
    avgUsagePercent: number;       // 平均CPU使用率
    maxUsagePercent: number;       // 最大CPU使用率
    minUsagePercent: number;       // 最小CPU使用率
    totalCores: number;            // 总核心数
    activeCores: number;           // 活跃核心数
  };
  
  // 内存聚合
  memory: {
    avgUsagePercent: number;       // 平均内存使用率
    maxUsagePercent: number;
    minUsagePercent: number;
    totalGB: number;               // 总内存
    usedGB: number;                // 已用内存
  };
  
  // 磁盘聚合
  disk: {
    avgUsagePercent: number;
    maxUsagePercent: number;
    totalReadIOPS: number;
    totalWriteIOPS: number;
    avgReadLatency: number;
    avgWriteLatency: number;
  };
  
  // 网络聚合
  network: {
    totalInBps: number;
    totalOutBps: number;
    avgInBps: number;
    avgOutBps: number;
  };
}
```

### SystemOverview（系统概览）

```typescript
interface SystemOverview {
  timestamp: Date;
  
  // 健康度评分（0-100）
  healthScore: {
    value: number;
    level: 'excellent' | 'good' | 'warning' | 'critical';
    trend: 'up' | 'down' | 'stable';  // 趋势
  };
  
  // VM状态分布
  vmStatus: {
    total: number;
    online: number;
    offline: number;
    error: number;
    warning: number;
  };
  
  // 核心指标概览
  coreMetrics: {
    cpu: {
      avgUsagePercent: number;
      alertCount: number;
    };
    memory: {
      avgUsagePercent: number;
      alertCount: number;
    };
    disk: {
      avgUsagePercent: number;
      highUsageCount: number;      // 使用率>80%的VM数量
    };
    network: {
      totalInBps: number;
      totalOutBps: number;
    };
  };
  
  // 告警统计
  alerts: {
    critical: number;
    high: number;
    medium: number;
    low: number;
    total: number;
  };
  
  // 最新告警列表（最近5条）
  recentAlerts: Array<{
    id: string;
    vmId: string;
    vmName: string;
    metric: string;
    severity: string;
    message: string;
    timestamp: Date;
  }>;
}
```

### WebSocketMessage（WebSocket消息格式）

```typescript
// 订阅消息（客户端发送）
interface SubscribeMessage {
  type: 'subscribe';
  data: {
    vmIds: string[];               // 要订阅的VM ID列表
    metrics?: string[];            // 指定指标类型（可选，默认全部）
  };
}

// 取消订阅消息（客户端发送）
interface UnsubscribeMessage {
  type: 'unsubscribe';
  data: {
    vmIds?: string[];              // 取消订阅的VM列表（空表示取消全部）
  };
}

// 心跳消息
interface PingMessage {
  type: 'ping';
  timestamp: number;
}

interface PongMessage {
  type: 'pong';
  timestamp: number;
}

// 数据推送消息（服务端发送）
interface MetricsMessage {
  type: 'metrics';
  data: {
    vmId: string;
    metrics: RealtimeMetrics;
  };
}

// 告警推送消息（服务端发送）
interface AlertMessage {
  type: 'alert';
  data: {
    vmId: string;
    vmName: string;
    alert: {
      id: string;
      metric: string;
      severity: string;
      threshold: number;
      currentValue: number;
      message: string;
      timestamp: Date;
    };
  };
}

// 连接确认消息（服务端发送）
interface ConnectedMessage {
  type: 'connected';
  data: {
    clientId: string;
    serverTime: Date;
    subscribedVMs: string[];
  };
}

// 错误消息（服务端发送）
interface ErrorMessage {
  type: 'error';
  data: {
    code: string;
    message: string;
  };
}
```

---

## 接口详情

### WebSocket接口

#### 1. 建立实时数据连接

**基本信息**
- 路径: `/ws/v1/realtime`
- 协议: WebSocket (ws:// 或 wss://)
- 认证: Token通过Query参数传递 `?token={access_token}`

**连接流程**
```
Client -> Server: WebSocket Handshake (with token)
Server -> Client: { type: 'connected', data: {...} }
Client -> Server: { type: 'subscribe', data: { vmIds: ['vm_001', 'vm_002'] } }
Server -> Client: { type: 'metrics', data: {...} } (每30-60秒推送)
Client -> Server: { type: 'ping', timestamp: 1234567890 }
Server -> Client: { type: 'pong', timestamp: 1234567890 }
```

**连接确认消息示例**
```json
{
  "type": "connected",
  "data": {
    "clientId": "ws_client_001",
    "serverTime": "2026-02-03T13:00:00Z",
    "subscribedVMs": []
  }
}
```

**订阅消息示例**
```json
{
  "type": "subscribe",
  "data": {
    "vmIds": ["vm_001", "vm_002", "vm_003"],
    "metrics": ["cpu", "memory", "disk"]
  }
}
```

**数据推送消息示例**
```json
{
  "type": "metrics",
  "data": {
    "vmId": "vm_001",
    "metrics": {
      "vmId": "vm_001",
      "timestamp": "2026-02-03T13:00:00Z",
      "dataSources": {
        "vsphere": true,
        "guestOS": true
      },
      "cpu": {
        "usageMHz": 1200,
        "ready": 0.5,
        "usagePercent": 30,
        "load1min": 0.8
      },
      "memory": {
        "usageMB": 4096,
        "grantedMB": 8192,
        "usagePercent": 50,
        "freeMB": 4096
      },
      "disk": {
        "readLatency": 5,
        "writeLatency": 3,
        "usagePercent": 65,
        "freeMB": 35000
      },
      "network": {
        "inBps": 1000000,
        "outBps": 500000,
        "inBytes": 125000,
        "outBytes": 62500
      },
      "alerts": []
    }
  }
}
```

**告警推送消息示例**
```json
{
  "type": "alert",
  "data": {
    "vmId": "vm_001",
    "vmName": "web-server-01",
    "alert": {
      "id": "alert_001",
      "metric": "cpu",
      "severity": "high",
      "threshold": 80,
      "currentValue": 85,
      "message": "CPU使用率超过阈值80%",
      "timestamp": "2026-02-03T13:00:00Z"
    }
  }
}
```

**心跳机制**
- 客户端每30秒发送ping消息
- 服务端收到后返回pong消息
- 超过90秒未收到pong，客户端应重连

---

### REST接口

#### 2. 获取VM当前指标

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/realtime/vms/{id}`
- 认证: 需要Access Token
- 权限: `vm:read`

**路径参数**
- `id` - VM ID

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "vmId": "vm_001",
    "timestamp": "2026-02-03T13:00:00Z",
    "dataSources": {
      "vsphere": true,
      "guestOS": true
    },
    "cpu": {
      "usageMHz": 1200,
      "ready": 0.5,
      "usagePercent": 30,
      "load1min": 0.8
    },
    "memory": {
      "usageMB": 4096,
      "grantedMB": 8192,
      "usagePercent": 50,
      "freeMB": 4096
    },
    "disk": {
      "readLatency": 5,
      "writeLatency": 3,
      "usagePercent": 65,
      "freeMB": 35000
    },
    "network": {
      "inBps": 1000000,
      "outBps": 500000
    },
    "alerts": []
  }
}
```

---

#### 3. 批量获取VM当前指标

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/realtime/vms/batch`
- 认证: 需要Access Token
- 权限: `vm:read`

**请求参数**
```json
{
  "vmIds": ["vm_001", "vm_002", "vm_003"],
  "metrics": ["cpu", "memory"]
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "metrics": [
      {
        "vmId": "vm_001",
        "timestamp": "2026-02-03T13:00:00Z",
        "cpu": {
          "usagePercent": 30
        },
        "memory": {
          "usagePercent": 50
        }
      },
      {
        "vmId": "vm_002",
        "timestamp": "2026-02-03T13:00:00Z",
        "cpu": {
          "usagePercent": 45
        },
        "memory": {
          "usagePercent": 60
        }
      }
    ],
    "notFound": ["vm_003"]
  }
}
```

---

#### 4. 获取分组聚合指标

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/realtime/groups/{id}`
- 认证: 需要Access Token
- 权限: `vm:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "scope": "group",
    "scopeId": "grp_001",
    "scopeName": "Web服务器组",
    "timestamp": "2026-02-03T13:00:00Z",
    "vmCount": {
      "total": 20,
      "online": 19,
      "offline": 1,
      "error": 0
    },
    "cpu": {
      "avgUsagePercent": 35.5,
      "maxUsagePercent": 78.2,
      "minUsagePercent": 12.3,
      "totalCores": 80,
      "activeCores": 76
    },
    "memory": {
      "avgUsagePercent": 52.1,
      "maxUsagePercent": 85.4,
      "minUsagePercent": 30.2,
      "totalGB": 160,
      "usedGB": 83.4
    },
    "disk": {
      "avgUsagePercent": 55.3,
      "maxUsagePercent": 89.1,
      "totalReadIOPS": 1500,
      "totalWriteIOPS": 800,
      "avgReadLatency": 4.5,
      "avgWriteLatency": 2.8
    },
    "network": {
      "totalInBps": 50000000,
      "totalOutBps": 30000000,
      "avgInBps": 2500000,
      "avgOutBps": 1500000
    }
  }
}
```

---

#### 5. 获取全局概览

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/realtime/overview`
- 认证: 需要Access Token
- 权限: `vm:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "timestamp": "2026-02-03T13:00:00Z",
    "healthScore": {
      "value": 87,
      "level": "good",
      "trend": "stable"
    },
    "vmStatus": {
      "total": 150,
      "online": 140,
      "offline": 5,
      "error": 5,
      "warning": 10
    },
    "coreMetrics": {
      "cpu": {
        "avgUsagePercent": 42.5,
        "alertCount": 3
      },
      "memory": {
        "avgUsagePercent": 58.2,
        "alertCount": 5
      },
      "disk": {
        "avgUsagePercent": 62.1,
        "highUsageCount": 12
      },
      "network": {
        "totalInBps": 125000000,
        "totalOutBps": 85000000
      }
    },
    "alerts": {
      "critical": 0,
      "high": 3,
      "medium": 8,
      "low": 15,
      "total": 26
    },
    "recentAlerts": [
      {
        "id": "alert_001",
        "vmId": "vm_005",
        "vmName": "db-server-02",
        "metric": "memory",
        "severity": "high",
        "message": "内存使用率超过阈值85%",
        "timestamp": "2026-02-03T12:55:00Z"
      }
    ]
  }
}
```

---

## 错误码定义

| 错误码 | 英文消息 | 中文消息 | 日文消息 | 说明 |
|--------|---------|---------|---------|------|
| 400 | Bad Request | 请求参数错误 | リクエストパラメータエラー | 参数缺失或格式错误 |
| 401 | Unauthorized | 未授权 | 未認証 | Token无效或过期 |
| 403 | Forbidden | 权限不足 | アクセス権限がありません | 无权限查看监控数据 |
| 404 | Not Found | VM不存在 | VMが見つかりません | VM ID不存在 |
| 404-GROUP | Group Not Found | 分组不存在 | グループが見つかりません | 分组ID不存在 |
| 429 | Rate Limit | 请求过于频繁 | リクエストが多すぎます | 频率限制 |
| 503 | Service Unavailable | 实时数据服务不可用 | リアルタイムデータサービスが利用できません | 数据采集服务异常 |

**WebSocket错误码**
| 代码 | 说明 |
|------|------|
| 1008 | Token无效或过期 |
| 1009 | 订阅VM数量超过限制（最大100） |
| 1011 | 服务端内部错误 |

---

## 性能规范

### 数据采集
- **采集间隔**: 30-60秒（可配置，默认30秒）
- **采集超时**: 10秒（单个VM）
- **批量大小**: 每批次最多50个VM

### 数据推送
- **推送频率**: 与采集频率一致（30-60秒）
- **最大订阅数**: 单个WebSocket连接最多订阅100个VM
- **并发连接**: 支持500+并发WebSocket连接

### 查询性能
- **实时数据查询**: < 500ms（P99）
- **聚合数据查询**: < 1s（P99）
- **批量查询**: 单次最多100个VM

---

## 变更记录

### 版本 v1.0 (2026-02-03)
**修改人**: BE工程师  
**修改原因**: 基于REQ_20260202_VM监控系统需求文档初始创建  
**具体修改**:
- [x] 新增WebSocket实时数据推送接口
- [x] 新增VM当前指标查询接口
- [x] 新增批量指标查询接口
- [x] 新增分组/集群聚合指标接口
- [x] 新增系统全局概览接口
- [x] 定义实时指标数据模型（合并vSphere + GuestOS）
- [x] 定义聚合指标和健康评分模型
- [x] 定义WebSocket消息协议

**影响范围**:
- 前端界面: 是（主仪表板、VM详细监控页面、实时图表）
- 后端API: 是（WebSocket服务、数据采集服务）
- 数据库结构: 是（timeseries_metrics表）
- 部署配置: 是（WebSocket端口、采集器配置）

**相关文档**:
- REQ_20260202_VM监控系统.md（监控指标数据定义、性能要求）
- UI_20260202_VM监控系统_视觉设计指南.md（主仪表板双模式、VM详细监控页面）
- API_VM_VM管理模块.md（VM基本信息查询）

---

**文档管理说明**:
1. WebSocket Token通过Query参数传递，避免Header问题
2. 指标字段可能为空（取决于数据源可用性）
3. 告警通过独立WebSocket消息推送，确保及时性
4. 心跳机制防止连接被代理服务器断开
5. 字段变更需记录在`api-changes.md`
