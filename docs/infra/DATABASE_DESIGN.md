# VM监控系统数据库设计文档

## 文档履历

| 版本 | 日期 | 修改人 | 修改内容 | 审核状态 |
|------|------|--------|----------|----------|
| v1.0 | 2026-02-03 | BE工程师 | 初始版本，包含完整数据模型 | 🔄 待审核 |

---

## 1. 数据库选型

- **数据库**: PostgreSQL 14+
- **时序数据库**: TimescaleDB (用于监控指标数据)
- **缓存**: Redis (用于Token、会话、热点数据)
- **搜索**: PostgreSQL全文搜索 (暂不使用Elasticsearch)

---

## 2. 表结构总览

### 核心表

| 表名 | 描述 | 存储引擎 |
|------|------|----------|
| `users` | 用户表 | PostgreSQL |
| `roles` | 角色表 | PostgreSQL |
| `role_permissions` | 角色权限关联表 | PostgreSQL |
| `permissions` | 权限定义表 | PostgreSQL |
| `vms` | 虚拟机表 | PostgreSQL |
| `vm_groups` | VM分组表 | PostgreSQL |
| `vm_group_members` | VM分组关联表 | PostgreSQL |
| `alert_rules` | 告警规则表 | PostgreSQL |
| `alert_conditions` | 告警条件表 | PostgreSQL |
| `alert_records` | 告警记录表 | PostgreSQL |
| `alert_notifications` | 告警通知记录表 | PostgreSQL |
| `metrics_raw` | 原始监控指标数据 | TimescaleDB |
| `metrics_hourly` | 小时聚合指标数据 | TimescaleDB |
| `metrics_daily` | 天聚合指标数据 | TimescaleDB |
| `system_logs` | 系统日志表 | TimescaleDB |
| `audit_logs` | 审计日志表 | PostgreSQL |
| `user_sessions` | 用户会话表 | PostgreSQL |

---

## 3. 详细表结构

### 3.1 用户权限相关

#### users (用户表)

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    department VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'locked', 'expired', 'pending')),
    must_change_password BOOLEAN NOT NULL DEFAULT false,
    mfa_enabled BOOLEAN NOT NULL DEFAULT false,
    mfa_secret VARCHAR(255),
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    login_fail_count INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    password_expired_at TIMESTAMP WITH TIME ZONE,
    preferences JSONB NOT NULL DEFAULT '{
        "language": "zh-CN",
        "theme": "dark",
        "timezone": "Asia/Shanghai",
        "dateFormat": "YYYY-MM-DD"
    }',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id)
);

-- 索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_department ON users(department);
CREATE INDEX idx_users_created_at ON users(created_at);
```

#### user_roles (用户角色关联表)

```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

#### roles (角色表)

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES roles(id),
    level INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1 AND level <= 5),
    path VARCHAR(500) NOT NULL, -- 层级路径，如 /admin/operator
    is_system BOOLEAN NOT NULL DEFAULT false, -- 系统内置角色不可删除
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id)
);

CREATE INDEX idx_roles_parent_id ON roles(parent_id);
CREATE INDEX idx_roles_level ON roles(level);
CREATE INDEX idx_roles_path ON roles(path);
```

#### permissions (权限定义表)

```sql
CREATE TABLE permissions (
    id VARCHAR(100) PRIMARY KEY, -- 如：vm:read, vm:write
    name VARCHAR(100) NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL, -- vm, alert, user等
    action VARCHAR(50) NOT NULL, -- read, write, delete, admin
    level VARCHAR(20) NOT NULL DEFAULT 'read' CHECK (level IN ('none', 'read', 'write', 'admin')),
    scope VARCHAR(20) DEFAULT 'global' CHECK (scope IN ('global', 'own', 'department')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_level ON permissions(level);
```

#### role_permissions (角色权限关联表)

```sql
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id VARCHAR(100) NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    is_inherited BOOLEAN NOT NULL DEFAULT false, -- 是否继承自父角色
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
```

#### user_sessions (用户会话表)

```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    access_token_hash VARCHAR(255) NOT NULL, -- Token哈希，用于验证
    refresh_token_hash VARCHAR(255), -- Refresh Token哈希
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    refresh_expires_at TIMESTAMP WITH TIME ZONE,
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token_hash ON user_sessions(access_token_hash);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active) WHERE is_active = true;
```

### 3.2 VM管理相关

#### vms (虚拟机表)

```sql
CREATE TABLE vms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vmware_id VARCHAR(100) UNIQUE, -- vCenter VM UUID
    name VARCHAR(200) NOT NULL,
    ip INET,
    os_type VARCHAR(20) CHECK (os_type IN ('Linux', 'Windows')),
    os_version VARCHAR(100),
    cpu_cores INTEGER,
    memory_gb INTEGER,
    disk_gb INTEGER,
    network_adapters INTEGER,
    power_state VARCHAR(20) CHECK (power_state IN ('poweredOn', 'poweredOff', 'suspended')),
    host_id VARCHAR(100),
    host_name VARCHAR(200),
    datacenter_id VARCHAR(100),
    datacenter_name VARCHAR(200),
    cluster_id VARCHAR(100),
    cluster_name VARCHAR(200),
    group_id UUID REFERENCES vm_groups(id),
    status VARCHAR(20) NOT NULL DEFAULT 'unknown' CHECK (status IN ('online', 'offline', 'error', 'unknown')),
    last_seen TIMESTAMP WITH TIME ZONE,
    vmware_tools_status VARCHAR(20) CHECK (vmware_tools_status IN ('installed', 'notInstalled', 'running', 'notRunning')),
    vmware_tools_version VARCHAR(50),
    tags JSONB DEFAULT '[]',
    description TEXT,
    metadata JSONB DEFAULT '{}', -- 扩展字段
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id)
);

-- 索引
CREATE INDEX idx_vms_vmware_id ON vms(vmware_id);
CREATE INDEX idx_vms_name ON vms(name);
CREATE INDEX idx_vms_ip ON vms(ip);
CREATE INDEX idx_vms_status ON vms(status);
CREATE INDEX idx_vms_os_type ON vms(os_type);
CREATE INDEX idx_vms_group_id ON vms(group_id);
CREATE INDEX idx_vms_power_state ON vms(power_state);
CREATE INDEX idx_vms_host_id ON vms(host_id);
CREATE INDEX idx_vms_cluster_id ON vms(cluster_id);
CREATE INDEX idx_vms_datacenter_id ON vms(datacenter_id);
CREATE INDEX idx_vms_last_seen ON vms(last_seen);
CREATE INDEX idx_vms_is_deleted ON vms(is_deleted) WHERE is_deleted = false;
CREATE INDEX idx_vms_tags ON vms USING GIN(tags);
```

#### vm_groups (VM分组表)

```sql
CREATE TABLE vm_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL DEFAULT 'custom' CHECK (type IN ('datacenter', 'cluster', 'host', 'custom')),
    parent_id UUID REFERENCES vm_groups(id),
    vmware_object_id VARCHAR(100), -- VMware对象ID（自动分组时使用）
    color VARCHAR(7) DEFAULT '#2196F3', -- 分组颜色标识
    sort_order INTEGER DEFAULT 0,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id)
);

CREATE INDEX idx_vm_groups_parent_id ON vm_groups(parent_id);
CREATE INDEX idx_vm_groups_type ON vm_groups(type);
CREATE INDEX idx_vm_groups_vmware_object_id ON vm_groups(vmware_object_id);
```

#### vm_group_members (VM分组关联表)

```sql
CREATE TABLE vm_group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vm_id UUID NOT NULL REFERENCES vms(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES vm_groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(vm_id, group_id)
);

CREATE INDEX idx_vm_group_members_vm_id ON vm_group_members(vm_id);
CREATE INDEX idx_vm_group_members_group_id ON vm_group_members(group_id);
```

### 3.3 告警管理相关

#### alert_rules (告警规则表)

```sql
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    scope VARCHAR(20) NOT NULL CHECK (scope IN ('global', 'vm', 'group', 'cluster')),
    scope_id UUID, -- 可以是vm.id 或 vm_group.id
    scope_name VARCHAR(200), -- 冗余存储，避免联表查询
    condition_logic VARCHAR(10) NOT NULL DEFAULT 'and' CHECK (condition_logic IN ('and', 'or')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    cooldown INTEGER NOT NULL DEFAULT 300, -- 冷却时间（秒）
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    notification_config JSONB NOT NULL DEFAULT '{}', -- 通知配置JSON
    trigger_count INTEGER NOT NULL DEFAULT 0,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id)
);

CREATE INDEX idx_alert_rules_scope ON alert_rules(scope, scope_id);
CREATE INDEX idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX idx_alert_rules_severity ON alert_rules(severity);
CREATE INDEX idx_alert_rules_is_deleted ON alert_rules(is_deleted) WHERE is_deleted = false;
CREATE INDEX idx_alert_rules_created_by ON alert_rules(created_by);
```

#### alert_conditions (告警条件表)

```sql
CREATE TABLE alert_conditions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    metric VARCHAR(50) NOT NULL CHECK (metric IN ('cpu', 'memory', 'disk', 'network', 'vmStatus')),
    metric_type VARCHAR(100) NOT NULL, -- 如：cpu.usagePercent, memory.usagePercent
    operator VARCHAR(10) NOT NULL CHECK (operator IN ('>', '<', '>=', '<=', '==', '!=', 'in', 'not_in')),
    threshold DECIMAL(18, 4) NOT NULL, -- 支持数值型阈值
    threshold_str VARCHAR(255), -- 字符串阈值（如状态值）
    duration INTEGER NOT NULL DEFAULT 60, -- 持续时间（秒）
    aggregation VARCHAR(20) DEFAULT 'last' CHECK (aggregation IN ('avg', 'max', 'min', 'last')),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alert_conditions_rule_id ON alert_conditions(rule_id);
CREATE INDEX idx_alert_conditions_metric ON alert_conditions(metric);
```

#### alert_records (告警记录表)

```sql
CREATE TABLE alert_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES alert_rules(id),
    rule_name VARCHAR(200) NOT NULL, -- 冗余存储
    vm_id UUID REFERENCES vms(id),
    vm_name VARCHAR(200), -- 冗余存储
    group_id UUID REFERENCES vm_groups(id),
    cluster_id VARCHAR(100),
    metric VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    trigger_value DECIMAL(18, 4) NOT NULL,
    threshold DECIMAL(18, 4) NOT NULL,
    condition_str TEXT, -- 触发条件描述
    triggered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    duration INTEGER, -- 告警持续时间（秒）
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'resolved', 'ignored')),
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_by_name VARCHAR(100), -- 冗余存储
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledge_note TEXT,
    resolved_by UUID REFERENCES users(id),
    resolved_by_name VARCHAR(100), -- 冗余存储
    resolution TEXT,
    snapshot JSONB, -- 触发时的指标快照
    notification_status JSONB DEFAULT '[]', -- 通知发送状态
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alert_records_rule_id ON alert_records(rule_id);
CREATE INDEX idx_alert_records_vm_id ON alert_records(vm_id);
CREATE INDEX idx_alert_records_status ON alert_records(status);
CREATE INDEX idx_alert_records_severity ON alert_records(severity);
CREATE INDEX idx_alert_records_triggered_at ON alert_records(triggered_at);
CREATE INDEX idx_alert_records_resolved_at ON alert_records(resolved_at) WHERE resolved_at IS NOT NULL;
```

### 3.4 监控指标数据（时序表）

#### metrics_raw (原始监控指标数据)

```sql
-- 创建 hypertable（TimescaleDB扩展）
CREATE TABLE metrics_raw (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    vm_id UUID NOT NULL REFERENCES vms(id),
    
    -- CPU指标
    cpu_usage_percent DECIMAL(5, 2),
    cpu_usage_mhz INTEGER,
    cpu_ready DECIMAL(5, 2),
    cpu_wait DECIMAL(5, 2),
    cpu_load_1min DECIMAL(6, 2),
    cpu_load_5min DECIMAL(6, 2),
    cpu_load_15min DECIMAL(6, 2),
    
    -- 内存指标
    memory_usage_percent DECIMAL(5, 2),
    memory_usage_mb INTEGER,
    memory_granted_mb INTEGER,
    memory_active_mb INTEGER,
    memory_ballooned_mb INTEGER,
    memory_compressed_mb INTEGER,
    memory_swapped_mb INTEGER,
    memory_free_mb INTEGER,
    memory_buffers_mb INTEGER,
    memory_cached_mb INTEGER,
    
    -- 磁盘指标
    disk_usage_percent DECIMAL(5, 2),
    disk_read_latency DECIMAL(8, 2),
    disk_write_latency DECIMAL(8, 2),
    disk_read_iops INTEGER,
    disk_write_iops INTEGER,
    disk_throughput_mbps DECIMAL(8, 2),
    disk_free_mb INTEGER,
    disk_used_mb INTEGER,
    
    -- 网络指标
    network_in_bps BIGINT,
    network_out_bps BIGINT,
    network_in_pps INTEGER,
    network_out_pps INTEGER,
    network_dropped_packets INTEGER,
    network_in_bytes BIGINT,
    network_out_bytes BIGINT,
    network_errors INTEGER,
    
    -- 数据来源标记
    data_source VARCHAR(20) CHECK (data_source IN ('vSphere', 'GuestOS', 'both')),
    
    -- 元数据
    collector_id VARCHAR(100),
    collection_duration_ms INTEGER
);

-- 转换为 hypertable
SELECT create_hypertable('metrics_raw', 'time', chunk_time_interval => INTERVAL '1 day');

-- 索引
CREATE INDEX idx_metrics_raw_vm_id_time ON metrics_raw(vm_id, time DESC);
CREATE INDEX idx_metrics_raw_time ON metrics_raw(time DESC);
```

#### metrics_hourly (小时聚合指标数据)

```sql
CREATE TABLE metrics_hourly (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    vm_id UUID NOT NULL REFERENCES vms(id),
    
    -- CPU聚合
    cpu_usage_percent_avg DECIMAL(5, 2),
    cpu_usage_percent_max DECIMAL(5, 2),
    cpu_usage_percent_min DECIMAL(5, 2),
    cpu_usage_percent_p95 DECIMAL(5, 2),
    
    -- 内存聚合
    memory_usage_percent_avg DECIMAL(5, 2),
    memory_usage_percent_max DECIMAL(5, 2),
    memory_usage_percent_min DECIMAL(5, 2),
    
    -- 磁盘聚合
    disk_usage_percent_avg DECIMAL(5, 2),
    disk_usage_percent_max DECIMAL(5, 2),
    disk_read_iops_avg INTEGER,
    disk_write_iops_avg INTEGER,
    
    -- 网络聚合
    network_in_bps_avg BIGINT,
    network_in_bps_max BIGINT,
    network_out_bps_avg BIGINT,
    network_out_bps_max BIGINT,
    
    -- 数据点数
    data_points INTEGER NOT NULL
);

SELECT create_hypertable('metrics_hourly', 'time', chunk_time_interval => INTERVAL '7 days');

CREATE INDEX idx_metrics_hourly_vm_id_time ON metrics_hourly(vm_id, time DESC);
```

#### metrics_daily (天聚合指标数据)

```sql
CREATE TABLE metrics_daily (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    vm_id UUID NOT NULL REFERENCES vms(id),
    
    cpu_usage_percent_avg DECIMAL(5, 2),
    cpu_usage_percent_max DECIMAL(5, 2),
    memory_usage_percent_avg DECIMAL(5, 2),
    memory_usage_percent_max DECIMAL(5, 2),
    disk_usage_percent_avg DECIMAL(5, 2),
    disk_usage_percent_max DECIMAL(5, 2),
    
    -- 可用性统计
    online_minutes INTEGER,
    total_minutes INTEGER,
    availability_percent DECIMAL(5, 2),
    
    data_points INTEGER NOT NULL
);

SELECT create_hypertable('metrics_daily', 'time', chunk_time_interval => INTERVAL '1 month');

CREATE INDEX idx_metrics_daily_vm_id_time ON metrics_daily(vm_id, time DESC);
```

### 3.5 日志相关

#### audit_logs (审计日志表)

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action VARCHAR(50) NOT NULL CHECK (action IN ('create', 'update', 'delete', 'grant', 'revoke', 'login', 'logout', 'export', 'import')),
    resource_type VARCHAR(50) NOT NULL, -- user, role, vm, alert_rule等
    resource_id VARCHAR(100) NOT NULL,
    resource_name VARCHAR(200),
    changes JSONB, -- 变更详情
    operator_id UUID REFERENCES users(id),
    operator_name VARCHAR(100), -- 冗余存储
    operator_ip INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_operator ON audit_logs(operator_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

#### system_logs (系统日志表 - 时序)

```sql
CREATE TABLE system_logs (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    level VARCHAR(20) NOT NULL CHECK (level IN ('debug', 'info', 'warn', 'error', 'fatal')),
    source VARCHAR(100) NOT NULL, -- 服务/模块名称
    instance VARCHAR(100), -- 实例标识
    message TEXT NOT NULL,
    details JSONB,
    trace_id VARCHAR(100),
    request_id VARCHAR(100),
    user_id UUID REFERENCES users(id),
    file VARCHAR(255),
    line INTEGER,
    function VARCHAR(255)
);

SELECT create_hypertable('system_logs', 'time', chunk_time_interval => INTERVAL '7 days');

CREATE INDEX idx_system_logs_level ON system_logs(level);
CREATE INDEX idx_system_logs_source ON system_logs(source);
CREATE INDEX idx_system_logs_trace_id ON system_logs(trace_id);
```

---

## 4. 数据保留策略

```sql
-- 原始数据保留7天（TimescaleDB自动清理）
SELECT add_retention_policy('metrics_raw', INTERVAL '7 days');

-- 小时聚合保留30天
SELECT add_retention_policy('metrics_hourly', INTERVAL '30 days');

-- 天聚合保留2年
SELECT add_retention_policy('metrics_daily', INTERVAL '730 days');

-- 系统日志保留90天
SELECT add_retention_policy('system_logs', INTERVAL '90 days');
```

---

## 5. 初始数据

### 5.1 内置权限数据

```sql
INSERT INTO permissions (id, name, description, resource, action, level) VALUES
-- VM管理权限
('vm:read', '查看VM', '查看虚拟机信息', 'vm', 'read', 'read'),
('vm:write', '编辑VM', '编辑虚拟机信息', 'vm', 'write', 'write'),
('vm:admin', '管理VM', '管理虚拟机（包括删除）', 'vm', 'admin', 'admin'),

-- 告警管理权限
('alert:read', '查看告警', '查看告警规则和记录', 'alert', 'read', 'read'),
('alert:write', '编辑告警', '编辑告警规则', 'alert', 'write', 'write'),
('alert:admin', '管理告警', '管理告警（包括删除）', 'alert', 'admin', 'admin'),

-- 历史数据权限
('history:read', '查看历史数据', '查询历史监控数据', 'history', 'read', 'read'),
('history:export', '导出数据', '导出历史数据', 'history', 'write', 'write'),

-- 用户管理权限
('user:read', '查看用户', '查看用户信息', 'user', 'read', 'read'),
('user:write', '编辑用户', '编辑用户信息', 'user', 'write', 'write'),
('user:admin', '管理用户', '管理用户（包括删除）', 'user', 'admin', 'admin'),

-- 系统权限
('system:read', '查看系统信息', '查看系统健康状态', 'system', 'read', 'read'),
('system:admin', '系统管理', '系统配置和管理', 'system', 'admin', 'admin');
```

### 5.2 内置角色数据

```sql
-- 系统管理员角色
INSERT INTO roles (id, name, description, level, path, is_system) VALUES
('role_admin', '系统管理员', '拥有所有权限', 1, '/admin', true);

-- 运维工程师角色
INSERT INTO roles (id, name, description, level, path, is_system) VALUES
('role_operator', '运维工程师', '日常运维操作权限', 1, '/operator', true);

-- 只读用户角色
INSERT INTO roles (id, name, description, parent_id, level, path, is_system) VALUES
('role_viewer', '只读用户', '仅查看权限', 'role_operator', 2, '/operator/viewer', true);

-- IT经理角色
INSERT INTO roles (id, name, description, level, path, is_system) VALUES
('role_manager', 'IT经理', '查看和报表权限', 1, '/manager', true);

-- 安全工程师角色
INSERT INTO roles (id, name, description, level, path, is_system) VALUES
('role_security', '安全工程师', '安全监控和审计权限', 1, '/security', true);
```

### 5.3 角色权限关联

```sql
-- 系统管理员拥有所有权限
INSERT INTO role_permissions (role_id, permission_id) 
SELECT 'role_admin', id FROM permissions;

-- 运维工程师权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
('role_operator', 'vm:read'),
('role_operator', 'vm:write'),
('role_operator', 'alert:read'),
('role_operator', 'alert:write'),
('role_operator', 'history:read'),
('role_operator', 'history:export'),
('role_operator', 'user:read'),
('role_operator', 'system:read');

-- 只读用户权限（继承运维工程师的读权限）
-- 无需单独插入，通过层级继承

-- IT经理权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
('role_manager', 'vm:read'),
('role_manager', 'alert:read'),
('role_manager', 'history:read'),
('role_manager', 'history:export'),
('role_manager', 'system:read');

-- 安全工程师权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
('role_security', 'vm:read'),
('role_security', 'alert:read'),
('role_security', 'alert:admin'),
('role_security', 'history:read'),
('role_security', 'history:export'),
('role_security', 'user:read'),
('role_security', 'system:read'),
('role_security', 'system:admin');
```

---

## 6. 数据库迁移管理

使用 **golang-migrate** 或 **gormigrate** 管理数据库迁移。

### 迁移文件命名规范
```
{序号}_{描述}.sql

示例：
001_create_users_table.sql
002_create_roles_table.sql
003_create_vms_table.sql
```

---

## 7. 性能优化建议

### 7.1 查询优化
- 所有外键字段建立索引
- 时间范围查询字段建立索引
- JSONB字段使用GIN索引

### 7.2 分区策略
- 时序表使用TimescaleDB自动分区
- 按时间维度分区（chunk_time_interval）

### 7.3 缓存策略
- Token验证使用Redis缓存
- 权限数据使用Redis缓存（15分钟过期）
- 热点数据使用Redis缓存

---

## 变更记录

### 版本 v1.0 (2026-02-03)
**修改人**: BE工程师  
**修改原因**: 基于API规范文档设计数据库结构  
**具体修改**:
- [x] 设计用户权限相关表（users, roles, permissions等）
- [x] 设计VM管理相关表（vms, vm_groups等）
- [x] 设计告警管理相关表（alert_rules, alert_records等）
- [x] 设计监控指标时序表（metrics_raw, metrics_hourly, metrics_daily）
- [x] 设计日志审计相关表（audit_logs, system_logs）
- [x] 定义索引策略
- [x] 定义数据保留策略
- [x] 定义初始权限和角色数据

**影响范围**:
- 后端开发: 是（数据库ORM模型）
- 部署配置: 是（需要安装TimescaleDB扩展）

---

**文档管理说明**:
1. 此文档为数据库DDL脚本的基础
2. 实际SQL脚本需要根据具体数据库版本调整
3. 建议配合迁移工具（migrate/gormigrate）使用
4. 时序表必须使用TimescaleDB扩展
