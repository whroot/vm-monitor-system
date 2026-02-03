// API通用响应类型
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  traceId?: string;
  timestamp: string;
}

// 分页请求
export interface PaginationRequest {
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

// 分页响应
export interface PaginationResponse<T> {
  list: T[];
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}

// 登录请求
export interface LoginRequest {
  username: string;
  password: string;
  rememberMe?: boolean;
  language?: string;
}

// 登录响应
export interface LoginResponse {
  user: User;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  permissions: string[];
}

// 用户信息
export interface User {
  id: string;
  username: string;
  email: string;
  name: string;
  phone?: string;
  department?: string;
  roles: Role[];
  status: 'active' | 'inactive' | 'locked' | 'expired' | 'pending';
  mustChangePassword: boolean;
  mfaEnabled: boolean;
  lastLoginAt?: string;
  preferences: {
    language: string;
    theme: string;
    timezone: string;
    dateFormat: string;
  };
  createdAt: string;
  updatedAt: string;
}

// 角色
export interface Role {
  id: string;
  name: string;
  description?: string;
  parentId?: string;
  level: number;
  path: string;
  isSystem: boolean;
  createdAt: string;
  updatedAt: string;
}

// VM信息
export interface VM {
  id: string;
  vmwareId?: string;
  name: string;
  ip?: string;
  os?: 'Linux' | 'Windows';
  osVersion?: string;
  cpuCores?: number;
  memoryGB?: number;
  diskGB?: number;
  networkAdapters?: number;
  powerState?: 'poweredOn' | 'poweredOff' | 'suspended';
  hostId?: string;
  hostName?: string;
  datacenterId?: string;
  datacenterName?: string;
  clusterId?: string;
  clusterName?: string;
  groupId?: string;
  group?: VMGroup;
  status: 'online' | 'offline' | 'error' | 'unknown';
  lastSeen?: string;
  vmwareToolsStatus?: 'installed' | 'notInstalled' | 'running' | 'notRunning';
  vmwareToolsVersion?: string;
  tags?: string[];
  description?: string;
  createdAt: string;
  updatedAt: string;
}

// VM分组
export interface VMGroup {
  id: string;
  name: string;
  description?: string;
  type: 'datacenter' | 'cluster' | 'host' | 'custom';
  parentId?: string;
  vmwareObjectId?: string;
  color?: string;
  isSystem: boolean;
  vmCount: number;
  onlineCount: number;
  offlineCount: number;
  errorCount: number;
  createdAt: string;
  updatedAt: string;
}

// VM列表请求
export interface VMListRequest extends PaginationRequest {
  status?: 'online' | 'offline' | 'error' | 'all';
  os?: 'Linux' | 'Windows';
  groupId?: string;
  hostId?: string;
  clusterId?: string;
  datacenterId?: string;
  keyword?: string;
}

// VM列表响应
export interface VMListResponse {
  list: VM[];
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
  summary: {
    total: number;
    online: number;
    offline: number;
    error: number;
  };
}

// 实时监控指标
export interface RealtimeMetrics {
  vmId: string;
  timestamp: string;
  dataSources: {
    vsphere: boolean;
    guestOS: boolean;
  };
  cpu?: {
    usageMHz?: number;
    ready?: number;
    wait?: number;
    usagePercent?: number;
    load1min?: number;
    load5min?: number;
    load15min?: number;
  };
  memory?: {
    usagePercent?: number;
    usedMB?: number;
    freeMB?: number;
    totalMB?: number;
    buffersMB?: number;
    cachedMB?: number;
  };
  disk?: {
    usagePercent?: number;
    readLatency?: number;
    writeLatency?: number;
    readIOPS?: number;
    writeIOPS?: number;
  };
  network?: {
    inBps?: number;
    outBps?: number;
    inBytes?: number;
    outBytes?: number;
  };
}

// 系统概览
export interface SystemOverview {
  timestamp: string;
  status: 'healthy' | 'degraded' | 'unhealthy' | 'maintenance';
  healthScore: {
    value: number;
    level: 'excellent' | 'good' | 'warning' | 'critical';
    trend: 'up' | 'down' | 'stable';
  };
  vmMonitoring: {
    totalVMs: number;
    onlineVMs: number;
    offlineVMs: number;
    errorVMs: number;
    collectionRate: number;
    avgCollectionTime: number;
  };
  alerts: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
}

// 告警记录
export interface AlertRecord {
  id: string;
  ruleId: string;
  ruleName: string;
  vmId?: string;
  vmName?: string;
  metric: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  triggerValue: number;
  threshold: number;
  triggeredAt: string;
  status: 'active' | 'acknowledged' | 'resolved' | 'ignored';
  acknowledgedByName?: string;
  acknowledgedAt?: string;
  acknowledgeNote?: string;
  resolvedByName?: string;
  resolvedAt?: string;
  resolution?: string;
}

// 历史数据查询请求
export interface HistoryQueryRequest {
  vmIds: string[];
  startTime: string;
  endTime: string;
  metrics?: ('cpu' | 'memory' | 'disk' | 'network')[];
  aggregation?: 'raw' | '1m' | '5m' | '15m' | '1h' | '1d';
  aggregationFunc?: 'avg' | 'max' | 'min' | 'p95' | 'p99';
  page?: number;
  pageSize?: number;
}

// 历史数据点
export interface HistoryDataPoint {
  timestamp: string;
  vmId: string;
  cpu?: {
    usagePercent?: number;
    usageMHz?: number;
    load1min?: number;
  };
  memory?: {
    usagePercent?: number;
    usedMB?: number;
    freeMB?: number;
  };
  disk?: {
    usagePercent?: number;
    readLatency?: number;
    writeLatency?: number;
  };
  network?: {
    inBps?: number;
    outBps?: number;
  };
}

// 导出任务
export interface ExportTask {
  id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  format: 'csv' | 'excel' | 'json';
  filename: string;
  progress?: {
    total: number;
    processed: number;
    percentage: number;
  };
  result?: {
    fileUrl: string;
    fileSize: number;
    recordCount: number;
    expiresAt: string;
  };
  createdAt: string;
}
