# API_VM_VM管理模块_API规范

## 文档履历

| 版本 | 日期 | 修改人 | 修改内容 | 审核状态 |
|------|------|--------|----------|----------|
| v1.0 | 2026-02-03 | BE工程师 | 初始版本，基于REQ_20260202和UI_20260202生成 | 🔄 待审核 |

---

## 模块概述

### 功能范围
- VM设备基础信息管理（CRUD）
- VM分组管理（集群/部门/自定义分组）
- VMware环境信息同步（vCenter集成）
- VM状态监控（在线/离线/错误）
- 批量操作（批量启动/停止/重启）

### 适用角色
- 系统管理员：全部权限
- 运维工程师：查看、编辑、批量操作
- IT经理：查看、报表

### 技术约束
- 支持1500+台VM管理
- 与vCenter Server 6.5+集成
- 实时同步VMware环境变化

---

## 接口清单

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取VM列表 | GET | /api/v1/vms | 分页查询VM列表，支持筛选 | 需要认证 |
| 获取VM详情 | GET | /api/v1/vms/{id} | 获取单个VM详细信息 | 需要认证 |
| 创建VM记录 | POST | /api/v1/vms | 手动添加VM（不常用） | 需要vm:write权限 |
| 更新VM信息 | PUT | /api/v1/vms/{id} | 更新VM基本信息 | 需要vm:write权限 |
| 删除VM记录 | DELETE | /api/v1/vms/{id} | 删除VM监控记录 | 需要vm:write权限 |
| 同步VMware信息 | POST | /api/v1/vms/sync | 从vCenter同步VM信息 | 需要vm:write权限 |
| 获取分组列表 | GET | /api/v1/vms/groups | 获取所有VM分组 | 需要认证 |
| 创建分组 | POST | /api/v1/vms/groups | 创建新分组 | 需要vm:write权限 |
| 更新分组 | PUT | /api/v1/vms/groups/{id} | 更新分组信息 | 需要vm:write权限 |
| 删除分组 | DELETE | /api/v1/vms/groups/{id} | 删除分组 | 需要vm:write权限 |
| 批量操作VM | POST | /api/v1/vms/batch | 批量启动/停止/重启 | 需要vm:write权限 |
| 获取VM状态统计 | GET | /api/v1/vms/statistics | 获取VM状态分布统计 | 需要认证 |

---

## 数据模型

### VMInfo（VM基本信息）
```typescript
interface VMInfo {
  id: string;                    // 内部ID
  vmwareId: string;              // vCenter VM UUID（唯一标识符）
  name: string;                  // VM名称
  ip: string;                    // IP地址
  os: 'Linux' | 'Windows';       // 操作系统类型
  osVersion: string;             // 操作系统版本
  
  // 资源配置
  cpuCores: number;              // CPU核心数
  memoryGB: number;              // 内存容量(GB)
  diskGB: number;                // 磁盘容量(GB)
  networkAdapters: number;       // 网络适配器数量
  
  // VMware环境信息
  powerState: 'poweredOn' | 'poweredOff' | 'suspended';  // 电源状态
  hostId: string;                // 所在ESXi主机ID
  hostName: string;              // ESXi主机名称
  datacenterId: string;          // 所在数据中心ID
  datacenterName: string;        // 数据中心名称
  clusterId: string;             // 所在集群ID
  clusterName: string;           // 集群名称
  
  // 分组和监控
  groupId?: string;              // 分组ID
  groupName?: string;            // 分组名称
  status: 'online' | 'offline' | 'error' | 'unknown';  // 监控状态
  lastSeen: Date;                // 最后在线时间
  
  // VMware Tools状态
  vmwareToolsStatus: 'installed' | 'notInstalled' | 'running' | 'notRunning';
  vmwareToolsVersion?: string;   // VMware Tools版本
  
  // 元数据
  createdAt: Date;               // 创建时间
  updatedAt: Date;               // 更新时间
  tags?: string[];               // 标签列表
  description?: string;          // 描述
}
```

### VMGroup（VM分组）
```typescript
interface VMGroup {
  id: string;                    // 分组ID
  name: string;                  // 分组名称
  description?: string;          // 分组描述
  type: 'datacenter' | 'cluster' | 'host' | 'custom';  // 分组类型
  parentId?: string;             // 父分组ID（用于层级结构）
  
  // 统计信息
  vmCount: number;               // VM数量
  onlineCount: number;           // 在线VM数量
  offlineCount: number;          // 离线VM数量
  errorCount: number;            // 错误VM数量
  
  // VMware关联
  vmwareObjectId?: string;       // VMware对象ID（如果是自动分组）
  
  // 元数据
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;             // 创建者ID
}
```

### VMListRequest（VM列表查询参数）
```typescript
interface VMListRequest {
  page?: number;                 // 页码（默认1）
  pageSize?: number;             // 每页数量（默认20，最大100）
  
  // 筛选条件
  status?: 'online' | 'offline' | 'error' | 'all';  // 状态筛选
  os?: 'Linux' | 'Windows';    // 操作系统筛选
  groupId?: string;              // 分组筛选
  hostId?: string;               // ESXi主机筛选
  clusterId?: string;            // 集群筛选
  datacenterId?: string;         // 数据中心筛选
  
  // 搜索
  keyword?: string;              // 关键字搜索（名称、IP）
  
  // 排序
  sortBy?: 'name' | 'status' | 'lastSeen' | 'createdAt';  // 排序字段
  sortOrder?: 'asc' | 'desc';    // 排序方向
}
```

### VMListResponse（VM列表响应）
```typescript
interface VMListResponse {
  list: VMInfo[];                // VM列表
  pagination: {
    page: number;                // 当前页码
    pageSize: number;            // 每页数量
    total: number;               // 总数量
    totalPages: number;        // 总页数
  };
  
  // 统计摘要
  summary: {
    total: number;               // 总数
    online: number;              // 在线数
    offline: number;             // 离线数
    error: number;               // 错误数
  };
}
```

### VMSyncRequest（VM同步请求）
```typescript
interface VMSyncRequest {
  type: 'full' | 'incremental';  // 同步类型：全量/增量
  datacenterId?: string;         // 指定数据中心（可选）
  clusterId?: string;            // 指定集群（可选）
  hostId?: string;               // 指定主机（可选）
}
```

### VMSyncResponse（VM同步响应）
```typescript
interface VMSyncResponse {
  syncId: string;                // 同步任务ID
  status: 'pending' | 'running' | 'completed' | 'failed';
  
  // 同步结果
  result?: {
    totalVMs: number;            // 总VM数
    added: number;               // 新增VM数
    updated: number;             // 更新VM数
    removed: number;             // 移除VM数
    failed: number;              // 失败数
    errors: Array<{
      vmwareId: string;
      error: string;
    }>;
  };
  
  startedAt: Date;
  completedAt?: Date;
}
```

### VMBatchRequest（VM批量操作请求）
```typescript
interface VMBatchRequest {
  action: 'start' | 'stop' | 'restart' | 'delete';  // 操作类型
  vmIds: string[];               // VM ID列表
  force?: boolean;               // 强制操作（用于停止/重启）
}
```

### VMBatchResponse（VM批量操作响应）
```typescript
interface VMBatchResponse {
  taskId: string;                // 批量任务ID
  status: 'pending' | 'running' | 'completed' | 'partial' | 'failed';
  
  // 操作结果
  results: Array<{
    vmId: string;
    vmName: string;
    success: boolean;
    message?: string;            // 成功/失败信息
  }>;
  
  // 统计
  summary: {
    total: number;               // 总数
    success: number;             // 成功数
    failed: number;              // 失败数
  };
  
  createdAt: Date;
  completedAt?: Date;
}
```

### VMStatistics（VM状态统计）
```typescript
interface VMStatistics {
  // 总体统计
  overview: {
    total: number;
    online: number;
    offline: number;
    error: number;
    unknown: number;
  };
  
  // 按OS分布
  byOS: Array<{
    os: 'Linux' | 'Windows';
    count: number;
    onlineCount: number;
  }>;
  
  // 按分组分布
  byGroup: Array<{
    groupId: string;
    groupName: string;
    count: number;
    onlineCount: number;
  }>;
  
  // 按VMware状态分布
  byPowerState: Array<{
    state: 'poweredOn' | 'poweredOff' | 'suspended';
    count: number;
  }>;
  
  // VMware Tools状态分布
  byToolsStatus: Array<{
    status: string;
    count: number;
  }>;
}
```

---

## 接口详情

### 1. 获取VM列表

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/vms`
- 认证: 需要Access Token
- 权限: `vm:read`

**查询参数**
```
GET /api/v1/vms?page=1&pageSize=20&status=online&groupId=grp_001&keyword=web
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": "vm_001",
        "vmwareId": "421b8e68-3b1f-c3bf-7eb5-5d8d80e8c6d1",
        "name": "web-server-01",
        "ip": "192.168.1.101",
        "os": "Linux",
        "osVersion": "CentOS 7.9",
        "cpuCores": 4,
        "memoryGB": 8,
        "diskGB": 100,
        "networkAdapters": 1,
        "powerState": "poweredOn",
        "hostId": "host_001",
        "hostName": "esxi-01",
        "datacenterId": "dc_001",
        "datacenterName": "数据中心A",
        "clusterId": "cluster_001",
        "clusterName": "集群1",
        "groupId": "grp_001",
        "groupName": "Web服务器组",
        "status": "online",
        "lastSeen": "2026-02-03T12:00:00Z",
        "vmwareToolsStatus": "running",
        "vmwareToolsVersion": "11.3.0",
        "createdAt": "2026-01-01T00:00:00Z",
        "updatedAt": "2026-02-03T12:00:00Z",
        "tags": ["production", "web"],
        "description": "Web服务器"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 150,
      "totalPages": 8
    },
    "summary": {
      "total": 150,
      "online": 140,
      "offline": 5,
      "error": 5
    }
  }
}
```

---

### 2. 获取VM详情

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/vms/{id}`
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
    "id": "vm_001",
    "vmwareId": "421b8e68-3b1f-c3bf-7eb5-5d8d80e8c6d1",
    "name": "web-server-01",
    "ip": "192.168.1.101",
    "os": "Linux",
    "osVersion": "CentOS 7.9",
    "cpuCores": 4,
    "memoryGB": 8,
    "diskGB": 100,
    "networkAdapters": 1,
    "powerState": "poweredOn",
    "hostId": "host_001",
    "hostName": "esxi-01",
    "datacenterId": "dc_001",
    "datacenterName": "数据中心A",
    "clusterId": "cluster_001",
    "clusterName": "集群1",
    "groupId": "grp_001",
    "groupName": "Web服务器组",
    "status": "online",
    "lastSeen": "2026-02-03T12:00:00Z",
    "vmwareToolsStatus": "running",
    "vmwareToolsVersion": "11.3.0",
    "createdAt": "2026-01-01T00:00:00Z",
    "updatedAt": "2026-02-03T12:00:00Z",
    "tags": ["production", "web"],
    "description": "Web服务器"
  }
}
```

**错误响应**
- `404` - VM不存在

---

### 3. 创建VM记录

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/vms`
- 认证: 需要Access Token
- 权限: `vm:write`

**请求参数**
```json
{
  "name": "db-server-01",
  "ip": "192.168.1.201",
  "os": "Linux",
  "osVersion": "Ubuntu 20.04",
  "cpuCores": 8,
  "memoryGB": 16,
  "diskGB": 500,
  "groupId": "grp_002",
  "tags": ["production", "database"],
  "description": "数据库服务器"
}
```

**成功响应 (201)**
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": "vm_new_001",
    "name": "db-server-01",
    "status": "unknown",
    "createdAt": "2026-02-03T12:30:00Z"
  }
}
```

---

### 4. 更新VM信息

**基本信息**
- 方法: `PUT`
- 路径: `/api/v1/vms/{id}`
- 认证: 需要Access Token
- 权限: `vm:write`

**请求参数**
```json
{
  "name": "web-server-01-updated",
  "groupId": "grp_003",
  "tags": ["production", "web", "frontend"],
  "description": "更新后的描述"
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": "vm_001",
    "updatedAt": "2026-02-03T12:35:00Z"
  }
}
```

---

### 5. 删除VM记录

**基本信息**
- 方法: `DELETE`
- 路径: `/api/v1/vms/{id}`
- 认证: 需要Access Token
- 权限: `vm:write`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "删除成功"
}
```

**注意**: 此操作仅从监控系统中删除记录，**不会**在vCenter中删除VM

---

### 6. 同步VMware信息

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/vms/sync`
- 认证: 需要Access Token
- 权限: `vm:write`

**请求参数**
```json
{
  "type": "full",
  "datacenterId": "dc_001"
}
```

**成功响应 (202)**
```json
{
  "code": 202,
  "message": "同步任务已创建",
  "data": {
    "syncId": "sync_20260203_001",
    "status": "pending",
    "startedAt": "2026-02-03T12:40:00Z"
  }
}
```

**异步查询同步进度**
```
GET /api/v1/vms/sync/{syncId}
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "syncId": "sync_20260203_001",
    "status": "completed",
    "result": {
      "totalVMs": 150,
      "added": 5,
      "updated": 145,
      "removed": 3,
      "failed": 0,
      "errors": []
    },
    "startedAt": "2026-02-03T12:40:00Z",
    "completedAt": "2026-02-03T12:42:30Z"
  }
}
```

---

### 7. 获取分组列表

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/vms/groups`
- 认证: 需要Access Token
- 权限: `vm:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": "grp_001",
        "name": "Web服务器组",
        "description": "生产环境Web服务器",
        "type": "custom",
        "vmCount": 20,
        "onlineCount": 19,
        "offlineCount": 1,
        "errorCount": 0,
        "createdAt": "2026-01-01T00:00:00Z",
        "updatedAt": "2026-02-03T10:00:00Z",
        "createdBy": "usr_001"
      },
      {
        "id": "dc_001",
        "name": "数据中心A",
        "description": "主数据中心",
        "type": "datacenter",
        "vmwareObjectId": "datacenter-1",
        "vmCount": 150,
        "onlineCount": 140,
        "offlineCount": 5,
        "errorCount": 5,
        "createdAt": "2026-01-01T00:00:00Z",
        "updatedAt": "2026-02-03T10:00:00Z",
        "createdBy": "system"
      }
    ]
  }
}
```

---

### 8. 创建分组

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/vms/groups`
- 认证: 需要Access Token
- 权限: `vm:write`

**请求参数**
```json
{
  "name": "数据库服务器组",
  "description": "生产环境数据库服务器",
  "type": "custom",
  "parentId": "dc_001"
}
```

**成功响应 (201)**
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": "grp_new_001",
    "name": "数据库服务器组",
    "type": "custom",
    "vmCount": 0,
    "onlineCount": 0,
    "offlineCount": 0,
    "errorCount": 0,
    "createdAt": "2026-02-03T12:45:00Z",
    "createdBy": "usr_001"
  }
}
```

---

### 9. 批量操作VM

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/vms/batch`
- 认证: 需要Access Token
- 权限: `vm:write`

**请求参数**
```json
{
  "action": "restart",
  "vmIds": ["vm_001", "vm_002", "vm_003"],
  "force": false
}
```

**成功响应 (202)**
```json
{
  "code": 202,
  "message": "批量任务已创建",
  "data": {
    "taskId": "batch_20260203_001",
    "status": "running",
    "results": [],
    "summary": {
      "total": 3,
      "success": 0,
      "failed": 0
    },
    "createdAt": "2026-02-03T12:50:00Z"
  }
}
```

**异步查询批量任务进度**
```
GET /api/v1/vms/batch/{taskId}
```

---

### 10. 获取VM状态统计

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/vms/statistics`
- 认证: 需要Access Token
- 权限: `vm:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "overview": {
      "total": 150,
      "online": 140,
      "offline": 5,
      "error": 5,
      "unknown": 0
    },
    "byOS": [
      {
        "os": "Linux",
        "count": 120,
        "onlineCount": 115
      },
      {
        "os": "Windows",
        "count": 30,
        "onlineCount": 25
      }
    ],
    "byGroup": [
      {
        "groupId": "grp_001",
        "groupName": "Web服务器组",
        "count": 20,
        "onlineCount": 19
      }
    ],
    "byPowerState": [
      {
        "state": "poweredOn",
        "count": 140
      },
      {
        "state": "poweredOff",
        "count": 8
      },
      {
        "state": "suspended",
        "count": 2
      }
    ],
    "byToolsStatus": [
      {
        "status": "running",
        "count": 135
      },
      {
        "status": "notRunning",
        "count": 5
      },
      {
        "status": "notInstalled",
        "count": 10
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
| 403 | Forbidden | 权限不足 | アクセス権限がありません | 无权限执行操作 |
| 404 | Not Found | VM不存在 | VMが見つかりません | VM ID不存在 |
| 404-GROUP | Group Not Found | 分组不存在 | グループが見つかりません | 分组ID不存在 |
| 409 | Conflict | VM名称已存在 | VM名が既に存在します | 名称重复 |
| 409-GROUP | Group Conflict | 分组名称已存在 | グループ名が既に存在します | 分组名称重复 |
| 422 | VMware Error | VMware操作失败 | VMware操作に失敗しました | vCenter API调用失败 |
| 429 | Rate Limit | 请求过于频繁 | リクエストが多すぎます | 频率限制 |
| 500 | Server Error | 服务器内部错误 | サーバーエラー | 服务器错误 |

---

## 变更记录

### 版本 v1.0 (2026-02-03)
**修改人**: BE工程师  
**修改原因**: 基于REQ_20260202_VM监控系统需求文档初始创建  
**具体修改**:
- [x] 新增VM CRUD接口
- [x] 新增分组管理接口
- [x] 新增VMware同步接口（异步任务）
- [x] 新增批量操作接口
- [x] 新增VM状态统计接口
- [x] 定义数据模型（VMInfo, VMGroup等）
- [x] 定义分页和筛选规范

**影响范围**:
- 前端界面: 是（VM列表、VM详情、分组管理页面）
- 后端API: 是（VM服务、分组服务、vCenter集成）
- 数据库结构: 是（vms, vm_groups表）
- 部署配置: 是（vCenter连接配置）

**相关文档**:
- REQ_20260202_VM监控系统.md（VMware技术架构、数据规格定义章节）
- UI_20260202_VM监控系统_视觉设计指南.md（主仪表板、VM详细监控页面）
- API_AUTH_认证授权模块.md（权限校验）

---

**文档管理说明**:
1. 此文档为BE/FE契约文件，任何变更需同步更新
2. VMware字段（vmwareId, hostId等）由系统自动同步，不建议手动修改
3. 分组type为`datacenter`/`cluster`/`host`时，由系统自动创建和管理
4. 批量操作通过异步任务执行，需轮询查询进度
5. 字段变更需记录在`api-changes.md`
