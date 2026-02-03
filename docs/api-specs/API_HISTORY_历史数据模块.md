# API_HISTORY_历史数据模块_API规范

## 文档履历

| 版本 | 日期 | 修改人 | 修改内容 | 审核状态 |
|------|------|--------|----------|----------|
| v1.0 | 2026-02-03 | BE工程师 | 初始版本，基于REQ_20260202和UI_20260202生成 | 🔄 待审核 |

---

## 模块概述

### 功能范围
- 历史监控数据查询（时间范围筛选）
- 多维度数据聚合（小时/天/周/月）
- 数据导出（CSV/Excel格式）
- 异常检测与标记
- 问题排查与容量规划双重视角支持

### 适用角色
- 运维工程师：问题排查、数据分析
- IT经理：容量规划、趋势分析
- 系统管理员：数据导出、审计
- 安全工程师：历史异常追溯

### 技术约束
- 历史数据保留：2年（分层存储策略）
- 查询性能：历史数据查询 < 5秒（P99）
- 数据精度：原始数据保留7天，聚合数据保留2年
- 导出限制：单次导出最多10万条记录

---

## 接口清单

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 查询历史数据 | POST | /api/v1/history/query | 按时间范围查询历史指标 | 需要认证 |
| 获取聚合统计 | POST | /api/v1/history/aggregate | 获取时间段聚合统计数据 | 需要认证 |
| 获取趋势分析 | POST | /api/v1/history/trends | 获取长期趋势分析数据 | 需要认证 |
| 异常检测查询 | POST | /api/v1/history/anomalies | 查询时间段内异常事件 | 需要认证 |
| 导出数据 | POST | /api/v1/history/export | 导出历史数据 | 需要认证 |
| 获取导出任务 | GET | /api/v1/history/export/{id} | 查询导出任务状态 | 需要认证 |
| 下载导出文件 | GET | /api/v1/history/export/{id}/download | 下载导出的文件 | 需要认证 |
| 获取时间线事件 | GET | /api/v1/history/timeline/{vmId} | 获取VM时间线事件 | 需要认证 |

---

## 数据模型

### HistoryQueryRequest（历史数据查询请求）

```typescript
interface HistoryQueryRequest {
  // 查询对象
  vmIds: string[];                  // VM ID列表（支持多选）
  groupId?: string;                 // 按分组筛选（与vmIds互斥）
  
  // 时间范围
  startTime: Date;                  // 开始时间
  endTime: Date;                    // 结束时间
  
  // 指标选择
  metrics: Array<'cpu' | 'memory' | 'disk' | 'network'>;  // 指标类型
  
  // 聚合粒度
  aggregation: 'raw' | '1m' | '5m' | '15m' | '1h' | '1d';  // 聚合间隔
  
  // 聚合函数
  aggregationFunc?: 'avg' | 'max' | 'min' | 'p95' | 'p99';  // 默认avg
  
  // 分页
  page?: number;                    // 页码（默认1）
  pageSize?: number;                // 每页数量（默认100，最大1000）
}
```

### HistoryDataPoint（历史数据点）

```typescript
interface HistoryDataPoint {
  timestamp: Date;                  // 数据点时间点
  vmId: string;                     // VM ID
  
  // CPU指标
  cpu?: {
    usagePercent?: number;          // CPU使用率百分比
    usageMHz?: number;              // CPU使用率(MHz)
    ready?: number;                 // CPU就绪时间
    load1min?: number;              // 1分钟负载
  };
  
  // 内存指标
  memory?: {
    usagePercent?: number;          // 内存使用率
    usedMB?: number;                // 已用内存(MB)
    freeMB?: number;                // 可用内存(MB)
  };
  
  // 磁盘指标
  disk?: {
    usagePercent?: number;          // 磁盘使用率
    readLatency?: number;           // 读取延迟(ms)
    writeLatency?: number;          // 写入延迟(ms)
    readIOPS?: number;              // 读取IOPS
    writeIOPS?: number;             // 写入IOPS
  };
  
  // 网络指标
  network?: {
    inBps?: number;                 // 入流量(bps)
    outBps?: number;                // 出流量(bps)
    inBytes?: number;               // 入流量字节
    outBytes?: number;              // 出流量字节
  };
}
```

### HistoryQueryResponse（历史数据查询响应）

```typescript
interface HistoryQueryResponse {
  data: HistoryDataPoint[];         // 数据点列表
  
  // 查询元数据
  meta: {
    startTime: Date;
    endTime: Date;
    aggregation: string;
    aggregationFunc: string;
    totalPoints: number;            // 总数据点数
    vmCount: number;                // 查询的VM数量
  };
  
  // 分页信息
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}
```

### AggregateMetricsRequest（聚合统计请求）

```typescript
interface AggregateMetricsRequest {
  vmIds?: string[];                 // VM列表（可选，默认全部）
  groupId?: string;                 // 分组ID（可选）
  clusterId?: string;               // 集群ID（可选）
  
  startTime: Date;
  endTime: Date;
  
  metrics: Array<'cpu' | 'memory' | 'disk' | 'network'>;
  
  // 统计维度
  dimensions: Array<'avg' | 'max' | 'min' | 'p95' | 'p99' | 'std'>;
  
  // 时间分组
  groupBy?: 'hour' | 'day' | 'week' | 'month';  // 按时间分组统计
}
```

### AggregateMetricsResponse（聚合统计响应）

```typescript
interface AggregateMetricsResponse {
  // 总体统计
  overall: {
    cpu?: {
      avg: number;
      max: number;
      min: number;
      p95?: number;
      p99?: number;
      std?: number;
    };
    memory?: {
      avg: number;
      max: number;
      min: number;
      p95?: number;
      p99?: number;
      std?: number;
    };
    disk?: {
      avg: number;
      max: number;
      min: number;
      p95?: number;
      p99?: number;
      std?: number;
    };
    network?: {
      avgInBps: number;
      avgOutBps: number;
      maxInBps: number;
      maxOutBps: number;
      totalInBytes: number;
      totalOutBytes: number;
    };
  };
  
  // 按时间分组统计（当groupBy指定时）
  timeGroups?: Array<{
    time: Date;
    cpu?: { avg: number; max: number };
    memory?: { avg: number; max: number };
    disk?: { avg: number; max: number };
    network?: { avgInBps: number; avgOutBps: number };
  }>;
  
  // 按VM分组统计
  vmGroups?: Array<{
    vmId: string;
    vmName: string;
    cpu?: { avg: number; max: number };
    memory?: { avg: number; max: number };
    disk?: { avg: number; max: number };
  }>;
}
```

### TrendAnalysisRequest（趋势分析请求）

```typescript
interface TrendAnalysisRequest {
  vmIds?: string[];
  groupId?: string;
  clusterId?: string;
  
  startTime: Date;                  // 趋势分析起始时间（建议3个月以上）
  endTime: Date;
  
  metrics: Array<'cpu' | 'memory' | 'disk'>;  // 趋势分析指标
  
  // 预测配置
  forecast?: {
    enabled: boolean;                 // 是否启用预测
    horizon: number;                  // 预测未来天数（默认30天）
    method: 'linear' | 'polynomial';  // 预测方法
  };
  
  // 容量预警
  capacityThreshold?: number;         // 容量预警阈值百分比（默认80%）
}
```

### TrendAnalysisResponse（趋势分析响应）

```typescript
interface TrendAnalysisResponse {
  // 历史趋势数据（按天）
  historical: Array<{
    date: Date;
    cpu?: number;
    memory?: number;
    disk?: number;
  }>;
  
  // 增长率分析
  growthRates: {
    cpu?: {
      daily: number;                // 日增长率
      weekly: number;               // 周增长率
      monthly: number;              // 月增长率
    };
    memory?: {
      daily: number;
      weekly: number;
      monthly: number;
    };
    disk?: {
      daily: number;
      weekly: number;
      monthly: number;
    };
  };
  
  // 容量预测（当forecast.enabled=true时）
  forecast?: {
    cpu?: {
      predictedValue: number;       // 预测值
      confidence: number;             // 置信度(0-1)
      capacityExceedDate?: Date;    // 预计超过阈值日期
    };
    memory?: {
      predictedValue: number;
      confidence: number;
      capacityExceedDate?: Date;
    };
    disk?: {
      predictedValue: number;
      confidence: number;
      capacityExceedDate?: Date;
    };
  };
  
  // 容量预警
  capacityAlerts?: Array<{
    metric: string;
    currentUsage: number;
    threshold: number;
    predictedExceedDate: Date;
    severity: 'low' | 'medium' | 'high';
  }>;
  
  // 资源优化建议
  recommendations?: Array<{
    type: 'scale_up' | 'scale_down' | 'optimize';
    metric: string;
    description: string;
    potentialSavings?: string;
  }>;
}
```

### AnomalyDetectionRequest（异常检测请求）

```typescript
interface AnomalyDetectionRequest {
  vmIds?: string[];
  groupId?: string;
  
  startTime: Date;
  endTime: Date;
  
  metrics?: Array<'cpu' | 'memory' | 'disk' | 'network'>;  // 空表示全部
  
  // 异常检测配置
  sensitivity?: 'low' | 'medium' | 'high';  // 敏感度（默认medium）
  
  // 异常类型筛选
  anomalyTypes?: Array<'spike' | 'drop' | 'trend_change' | 'pattern_break'>;
}
```

### AnomalyEvent（异常事件）

```typescript
interface AnomalyEvent {
  id: string;                       // 异常事件ID
  vmId: string;                     // VM ID
  vmName: string;                   // VM名称
  
  timestamp: Date;                  // 异常发生时间
  metric: 'cpu' | 'memory' | 'disk' | 'network';
  
  // 异常特征
  type: 'spike' | 'drop' | 'trend_change' | 'pattern_break';
  severity: 'low' | 'medium' | 'high' | 'critical';
  
  // 数值信息
  value: number;                    // 异常值
  baseline: number;                 // 基线值（正常范围）
  deviation: number;                // 偏离程度（百分比或倍数）
  
  // 上下文
  duration?: number;                // 异常持续时间（秒）
  relatedVMs?: string[];          // 相关VM（同时异常）
  
  // 根因分析
  possibleCauses?: string[];        // 可能原因
  suggestedActions?: string[];    // 建议操作
  
  // 状态
  status: 'active' | 'acknowledged' | 'resolved';
  acknowledgedBy?: string;          // 确认人
  acknowledgedAt?: Date;            // 确认时间
  resolvedAt?: Date;                // 解决时间
  
  createdAt: Date;
}
```

### ExportRequest（数据导出请求）

```typescript
interface ExportRequest {
  // 查询条件（同HistoryQueryRequest）
  vmIds: string[];
  groupId?: string;
  startTime: Date;
  endTime: Date;
  metrics: Array<'cpu' | 'memory' | 'disk' | 'network'>;
  aggregation: 'raw' | '1m' | '5m' | '15m' | '1h' | '1d';
  
  // 导出配置
  format: 'csv' | 'excel' | 'json';
  filename?: string;                // 自定义文件名（可选）
  
  // 字段选择（可选，默认全部）
  fields?: string[];
  
  // 高级选项
  options?: {
    includeHeaders: boolean;        // 包含表头（默认true）
    timezone: string;               // 时区（默认UTC）
    dateFormat: string;             // 日期格式
    numberFormat: string;           // 数字格式
  };
}
```

### ExportTask（导出任务）

```typescript
interface ExportTask {
  id: string;                       // 任务ID
  status: 'pending' | 'processing' | 'completed' | 'failed';
  
  // 查询条件摘要
  query: {
    vmCount: number;
    startTime: Date;
    endTime: Date;
    aggregation: string;
  };
  
  // 导出配置
  format: 'csv' | 'excel' | 'json';
  filename: string;
  
  // 进度信息
  progress?: {
    total: number;                  // 总记录数
    processed: number;              // 已处理记录数
    percentage: number;             // 进度百分比
  };
  
  // 结果（当status=completed时）
  result?: {
    fileUrl: string;                // 下载链接
    fileSize: number;               // 文件大小（字节）
    recordCount: number;            // 记录数
    expiresAt: Date;                // 过期时间（默认7天）
  };
  
  // 错误信息（当status=failed时）
  error?: {
    code: string;
    message: string;
  };
  
  createdAt: Date;
  startedAt?: Date;
  completedAt?: Date;
  createdBy: string;
}
```

### TimelineEvent（时间线事件）

```typescript
interface TimelineEvent {
  id: string;
  vmId: string;
  
  timestamp: Date;
  type: 'metric_alert' | 'power_change' | 'anomaly' | 'manual' | 'maintenance';
  
  // 事件详情
  title: string;
  description?: string;
  
  // 指标数据（当type=metric_alert或anomaly时）
  metricData?: {
    metric: string;
    value: number;
    threshold?: number;
  };
  
  // 状态变更（当type=power_change时）
  stateChange?: {
    from: string;
    to: string;
  };
  
  // 元数据
  severity?: 'low' | 'medium' | 'high' | 'critical';
  acknowledged: boolean;
  createdBy?: string;               // 创建者（手动事件）
  
  createdAt: Date;
}
```

---

## 接口详情

### 1. 查询历史数据

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/history/query`
- 认证: 需要Access Token
- 权限: `vm:read`

**请求参数**
```json
{
  "vmIds": ["vm_001", "vm_002"],
  "startTime": "2026-02-01T00:00:00Z",
  "endTime": "2026-02-03T23:59:59Z",
  "metrics": ["cpu", "memory"],
  "aggregation": "1h",
  "aggregationFunc": "avg",
  "page": 1,
  "pageSize": 100
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "data": [
      {
        "timestamp": "2026-02-01T00:00:00Z",
        "vmId": "vm_001",
        "cpu": {
          "usagePercent": 35.2,
          "usageMHz": 1408
        },
        "memory": {
          "usagePercent": 52.5,
          "usedMB": 4300
        }
      },
      {
        "timestamp": "2026-02-01T01:00:00Z",
        "vmId": "vm_001",
        "cpu": {
          "usagePercent": 38.1,
          "usageMHz": 1524
        },
        "memory": {
          "usagePercent": 54.2,
          "usedMB": 4440
        }
      }
    ],
    "meta": {
      "startTime": "2026-02-01T00:00:00Z",
      "endTime": "2026-02-03T23:59:59Z",
      "aggregation": "1h",
      "aggregationFunc": "avg",
      "totalPoints": 144,
      "vmCount": 2
    },
    "pagination": {
      "page": 1,
      "pageSize": 100,
      "total": 144,
      "totalPages": 2
    }
  }
}
```

**约束说明**
- 时间范围最大跨度：2年
- 原始数据(raw)查询限制：最多7天
- 分页最大pageSize：1000

---

### 2. 获取聚合统计

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/history/aggregate`
- 认证: 需要Access Token
- 权限: `vm:read`

**请求参数**
```json
{
  "groupId": "grp_001",
  "startTime": "2026-01-01T00:00:00Z",
  "endTime": "2026-02-03T23:59:59Z",
  "metrics": ["cpu", "memory", "disk"],
  "dimensions": ["avg", "max", "p95"],
  "groupBy": "day"
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "overall": {
      "cpu": {
        "avg": 42.5,
        "max": 89.3,
        "p95": 78.2
      },
      "memory": {
        "avg": 58.3,
        "max": 92.1,
        "p95": 85.4
      },
      "disk": {
        "avg": 62.1,
        "max": 95.7,
        "p95": 89.1
      }
    },
    "timeGroups": [
      {
        "time": "2026-02-01T00:00:00Z",
        "cpu": { "avg": 40.2, "max": 85.1 },
        "memory": { "avg": 56.3, "max": 88.2 },
        "disk": { "avg": 60.1, "max": 92.3 }
      },
      {
        "time": "2026-02-02T00:00:00Z",
        "cpu": { "avg": 44.8, "max": 89.3 },
        "memory": { "avg": 60.2, "max": 92.1 },
        "disk": { "avg": 64.1, "max": 95.7 }
      }
    ],
    "vmGroups": [
      {
        "vmId": "vm_001",
        "vmName": "web-server-01",
        "cpu": { "avg": 38.2, "max": 82.1 },
        "memory": { "avg": 52.5, "max": 85.3 }
      }
    ]
  }
}
```

---

### 3. 获取趋势分析

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/history/trends`
- 认证: 需要Access Token
- 权限: `vm:read`

**请求参数**
```json
{
  "groupId": "grp_001",
  "startTime": "2025-11-01T00:00:00Z",
  "endTime": "2026-02-03T23:59:59Z",
  "metrics": ["cpu", "memory", "disk"],
  "forecast": {
    "enabled": true,
    "horizon": 30,
    "method": "linear"
  },
  "capacityThreshold": 80
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "分析成功",
  "data": {
    "historical": [
      { "date": "2025-11-01T00:00:00Z", "cpu": 38.5, "memory": 55.2, "disk": 58.3 },
      { "date": "2025-12-01T00:00:00Z", "cpu": 40.2, "memory": 56.8, "disk": 60.1 },
      { "date": "2026-01-01T00:00:00Z", "cpu": 42.1, "memory": 58.3, "disk": 62.5 },
      { "date": "2026-02-01T00:00:00Z", "cpu": 43.8, "memory": 60.1, "disk": 64.2 }
    ],
    "growthRates": {
      "cpu": { "daily": 0.05, "weekly": 0.35, "monthly": 1.5 },
      "memory": { "daily": 0.04, "weekly": 0.28, "monthly": 1.2 },
      "disk": { "daily": 0.06, "weekly": 0.42, "monthly": 1.8 }
    },
    "forecast": {
      "cpu": {
        "predictedValue": 48.2,
        "confidence": 0.85
      },
      "memory": {
        "predictedValue": 64.5,
        "confidence": 0.82,
        "capacityExceedDate": "2026-06-15T00:00:00Z"
      },
      "disk": {
        "predictedValue": 72.8,
        "confidence": 0.88,
        "capacityExceedDate": "2026-05-20T00:00:00Z"
      }
    },
    "capacityAlerts": [
      {
        "metric": "disk",
        "currentUsage": 64.2,
        "threshold": 80,
        "predictedExceedDate": "2026-05-20T00:00:00Z",
        "severity": "medium"
      },
      {
        "metric": "memory",
        "currentUsage": 60.1,
        "threshold": 80,
        "predictedExceedDate": "2026-06-15T00:00:00Z",
        "severity": "low"
      }
    ],
    "recommendations": [
      {
        "type": "scale_up",
        "metric": "disk",
        "description": "建议在2026-05-01前扩容磁盘容量",
        "potentialSavings": "避免服务中断"
      }
    ]
  }
}
```

---

### 4. 异常检测查询

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/history/anomalies`
- 认证: 需要Access Token
- 权限: `vm:read`

**请求参数**
```json
{
  "vmIds": ["vm_001", "vm_002"],
  "startTime": "2026-02-01T00:00:00Z",
  "endTime": "2026-02-03T23:59:59Z",
  "metrics": ["cpu", "memory"],
  "sensitivity": "medium",
  "anomalyTypes": ["spike", "trend_change"]
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "anomalies": [
      {
        "id": "anomaly_001",
        "vmId": "vm_001",
        "vmName": "web-server-01",
        "timestamp": "2026-02-02T14:30:00Z",
        "metric": "cpu",
        "type": "spike",
        "severity": "high",
        "value": 95.2,
        "baseline": 35.0,
        "deviation": 172,
        "duration": 1800,
        "possibleCauses": ["突发流量", "定时任务执行"],
        "suggestedActions": ["检查应用日志", "评估是否需要扩容"],
        "status": "acknowledged",
        "acknowledgedBy": "usr_001",
        "acknowledgedAt": "2026-02-02T15:00:00Z",
        "createdAt": "2026-02-02T14:30:00Z"
      }
    ],
    "total": 15,
    "bySeverity": {
      "critical": 0,
      "high": 3,
      "medium": 8,
      "low": 4
    }
  }
}
```

---

### 5. 导出数据

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/history/export`
- 认证: 需要Access Token
- 权限: `vm:read`

**请求参数**
```json
{
  "vmIds": ["vm_001", "vm_002"],
  "startTime": "2026-02-01T00:00:00Z",
  "endTime": "2026-02-03T23:59:59Z",
  "metrics": ["cpu", "memory", "disk", "network"],
  "aggregation": "1h",
  "format": "excel",
  "filename": "vm_monitoring_data_feb2026",
  "options": {
    "includeHeaders": true,
    "timezone": "Asia/Shanghai",
    "dateFormat": "YYYY-MM-DD HH:mm:ss",
    "numberFormat": "0.00"
  }
}
```

**成功响应 (202)**
```json
{
  "code": 202,
  "message": "导出任务已创建",
  "data": {
    "id": "export_20260203_001",
    "status": "pending",
    "query": {
      "vmCount": 2,
      "startTime": "2026-02-01T00:00:00Z",
      "endTime": "2026-02-03T23:59:59Z",
      "aggregation": "1h"
    },
    "format": "excel",
    "filename": "vm_monitoring_data_feb2026.xlsx",
    "createdAt": "2026-02-03T13:00:00Z"
  }
}
```

**导出限制**
- 单次导出最多10万条记录
- 文件保留7天
- 支持CSV、Excel、JSON格式

---

### 6. 获取导出任务状态

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/history/export/{id}`
- 认证: 需要Access Token
- 权限: `vm:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "id": "export_20260203_001",
    "status": "completed",
    "query": {
      "vmCount": 2,
      "startTime": "2026-02-01T00:00:00Z",
      "endTime": "2026-02-03T23:59:59Z",
      "aggregation": "1h"
    },
    "format": "excel",
    "filename": "vm_monitoring_data_feb2026.xlsx",
    "progress": {
      "total": 144,
      "processed": 144,
      "percentage": 100
    },
    "result": {
      "fileUrl": "/api/v1/history/export/export_20260203_001/download",
      "fileSize": 24576,
      "recordCount": 144,
      "expiresAt": "2026-02-10T13:00:00Z"
    },
    "createdAt": "2026-02-03T13:00:00Z",
    "startedAt": "2026-02-03T13:00:05Z",
    "completedAt": "2026-02-03T13:00:30Z"
  }
}
```

---

### 7. 获取时间线事件

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/history/timeline/{vmId}`
- 认证: 需要Access Token
- 权限: `vm:read`

**查询参数**
```
GET /api/v1/history/timeline/vm_001?startTime=2026-02-01T00:00:00Z&endTime=2026-02-03T23:59:59Z&types=metric_alert,anomaly
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "events": [
      {
        "id": "event_001",
        "vmId": "vm_001",
        "timestamp": "2026-02-02T14:30:00Z",
        "type": "anomaly",
        "title": "CPU使用率异常峰值",
        "description": "CPU使用率从35%飙升至95%，持续30分钟",
        "metricData": {
          "metric": "cpu",
          "value": 95.2,
          "threshold": 80
        },
        "severity": "high",
        "acknowledged": true,
        "createdAt": "2026-02-02T14:30:00Z"
      },
      {
        "id": "event_002",
        "vmId": "vm_001",
        "timestamp": "2026-02-01T08:00:00Z",
        "type": "power_change",
        "title": "VM电源状态变更",
        "description": "VM从关机状态启动",
        "stateChange": {
          "from": "poweredOff",
          "to": "poweredOn"
        },
        "acknowledged": false,
        "createdAt": "2026-02-01T08:00:00Z"
      }
    ],
    "total": 25
  }
}
```

---

## 错误码定义

| 错误码 | 英文消息 | 中文消息 | 日文消息 | 说明 |
|--------|---------|---------|---------|------|
| 400 | Bad Request | 请求参数错误 | リクエストパラメータエラー | 参数缺失或格式错误 |
| 400-TIME | Invalid Time Range | 时间范围无效 | 時間範囲が無効です | 时间格式错误或范围过大 |
| 401 | Unauthorized | 未授权 | 未認証 | Token无效或过期 |
| 403 | Forbidden | 权限不足 | アクセス権限がありません | 无权限查看历史数据 |
| 404 | Not Found | VM不存在 | VMが見つかりません | VM ID不存在 |
| 404-EXPORT | Export Task Not Found | 导出任务不存在 | エクスポートタスクが見つかりません | 导出任务ID不存在 |
| 400-LIMIT | Record Limit Exceeded | 导出记录数超过限制 | エクスポート記録数が制限を超えています | 超过10万条限制 |
| 429 | Rate Limit | 请求过于频繁 | リクエストが多すぎます | 频率限制 |
| 500 | Server Error | 服务器内部错误 | サーバーエラー | 服务器错误 |
| 503 | Storage Unavailable | 历史数据存储不可用 | 履歴データストレージが利用できません | 存储服务异常 |

---

## 存储策略

### 分层存储
| 层级 | 数据类型 | 保留时间 | 聚合粒度 | 存储介质 |
|------|---------|---------|---------|----------|
| 热数据 | 原始数据 | 7天 | 30-60秒 | SSD |
| 温数据 | 小时聚合 | 30天 | 1小时 | SSD |
| 冷数据 | 天/周/月聚合 | 2年 | 1天/1周/1月 | HDD |

### 数据压缩
- 原始数据：Snappy压缩
- 聚合数据：GZIP压缩
- 预计压缩率：60-70%

---

## 变更记录

### 版本 v1.0 (2026-02-03)
**修改人**: BE工程师  
**修改原因**: 基于REQ_20260202_VM监控系统需求文档初始创建  
**具体修改**:
- [x] 新增历史数据查询接口（支持多维度筛选和聚合）
- [x] 新增聚合统计接口（支持P95/P99等高级统计）
- [x] 新增趋势分析接口（支持容量预测）
- [x] 新增异常检测接口（基于ML算法）
- [x] 新增数据导出接口（异步任务）
- [x] 新增时间线事件接口
- [x] 定义历史数据模型和导出任务模型
- [x] 定义分层存储策略

**影响范围**:
- 前端界面: 是（历史数据查询页面、问题排查/容量规划双模式）
- 后端API: 是（历史数据查询服务、分析服务、导出服务）
- 数据库结构: 是（timeseries_metrics表、导出任务表）
- 部署配置: 是（时序数据库配置、对象存储配置）

**相关文档**:
- REQ_20260202_VM监控系统.md（历史数据查询、数据持久化存储、分层存储策略）
- UI_20260202_VM监控系统_视觉设计指南.md（历史数据双重视角、异常检测时间轴）
- API_REALTIME_实时监控模块.md（实时数据写入）

---

**文档管理说明**:
1. 原始数据(raw)查询限制为7天内，超过需使用聚合粒度
2. 导出任务为异步执行，需轮询查询进度
3. 趋势分析的预测功能需要至少3个月历史数据
4. 异常检测算法基于统计学方法（3-sigma/z-score）
5. 字段变更需记录在`api-changes.md`
