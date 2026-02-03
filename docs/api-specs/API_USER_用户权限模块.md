# API_USER_用户权限模块_API规范

## 文档履历

| 版本 | 日期 | 修改人 | 修改内容 | 审核状态 |
|------|------|--------|----------|----------|
| v1.0 | 2026-02-03 | BE工程师 | 初始版本，基于REQ_20260202和UI_20260202生成 | 🔄 待审核 |

---

## 模块概述

### 功能范围
- 用户账号管理（CRUD）
- 角色层级管理（支持继承）
- 权限矩阵配置
- 权限冲突检测
- 用户权限审计

### 适用角色
- 系统管理员：全部权限
- 安全工程师：权限审计、安全管理
- 其他角色：查看自己的权限信息

### 技术约束
- 用户数量：最多500个账号
- 角色层级：最多5层继承
- 权限实时生效：变更后1分钟内生效
- 权限缓存：15分钟有效期

---

## 接口清单

### 用户管理

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取用户列表 | GET | /api/v1/users | 分页查询用户列表 | 需要user:read权限 |
| 获取用户详情 | GET | /api/v1/users/{id} | 获取单个用户详情 | 需要user:read权限 |
| 创建用户 | POST | /api/v1/users | 创建新用户 | 需要user:write权限 |
| 更新用户 | PUT | /api/v1/users/{id} | 更新用户信息 | 需要user:write权限 |
| 删除用户 | DELETE | /api/v1/users/{id} | 删除用户 | 需要user:write权限 |
| 重置密码 | POST | /api/v1/users/{id}/reset-password | 重置用户密码 | 需要user:write权限 |
| 批量更新状态 | PUT | /api/v1/users/batch/status | 批量启用/禁用 | 需要user:write权限 |
| 获取当前用户权限 | GET | /api/v1/users/me/permissions | 获取当前登录用户权限 | 需要认证 |

### 角色管理

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取角色列表 | GET | /api/v1/roles | 获取角色层级列表 | 需要user:read权限 |
| 获取角色详情 | GET | /api/v1/roles/{id} | 获取角色详情 | 需要user:read权限 |
| 创建角色 | POST | /api/v1/roles | 创建新角色 | 需要user:write权限 |
| 更新角色 | PUT | /api/v1/roles/{id} | 更新角色 | 需要user:write权限 |
| 删除角色 | DELETE | /api/v1/roles/{id} | 删除角色 | 需要user:write权限 |
| 获取角色权限 | GET | /api/v1/roles/{id}/permissions | 获取角色权限详情 | 需要user:read权限 |
| 更新角色权限 | PUT | /api/v1/roles/{id}/permissions | 更新角色权限 | 需要user:write权限 |
| 获取角色用户 | GET | /api/v1/roles/{id}/users | 获取角色下的用户 | 需要user:read权限 |

### 权限矩阵

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取权限矩阵 | GET | /api/v1/permissions/matrix | 获取权限矩阵视图 | 需要user:read权限 |
| 批量设置权限 | PUT | /api/v1/permissions/matrix | 批量设置权限 | 需要user:write权限 |
| 获取用户权限详情 | GET | /api/v1/users/{id}/permissions/detail | 获取用户完整权限 | 需要user:read权限 |
| 检查权限冲突 | POST | /api/v1/permissions/check-conflict | 检查权限配置冲突 | 需要user:read权限 |

### 权限审计

| 接口 | 方法 | 路径 | 描述 | 认证要求 |
|------|------|------|------|----------|
| 获取权限变更历史 | GET | /api/v1/permissions/audit | 查询权限变更日志 | 需要user:read权限 |
| 生成权限报告 | POST | /api/v1/permissions/report | 生成权限汇总报告 | 需要user:read权限 |

---

## 数据模型

### User（用户）

```typescript
interface User {
  id: string;                       // 用户ID
  username: string;                 // 用户名（唯一）
  email: string;                    // 邮箱（唯一）
  name: string;                     // 显示名称
  phone?: string;                   // 电话
  department?: string;              // 部门
  
  // 角色
  roles: string[];                  // 角色ID列表
  roleNames: string[];              // 角色名称列表（计算字段）
  
  // 状态
  status: 'active' | 'inactive' | 'locked' | 'expired' | 'pending';
  
  // 安全设置
  passwordExpiredAt?: Date;         // 密码过期时间
  mustChangePassword: boolean;      // 是否强制修改密码
  mfaEnabled: boolean;              // MFA启用状态（预留）
  lastLoginAt?: Date;               // 最后登录时间
  lastLoginIp?: string;             // 最后登录IP
  loginFailCount: number;         // 连续登录失败次数
  lockedUntil?: Date;               // 锁定截止时间
  
  // 偏好设置
  preferences: {
    language: 'en' | 'zh-CN' | 'ja-JP';
    theme: 'dark' | 'light';
    timezone: string;
    dateFormat: string;
  };
  
  // 元数据
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;
  updatedBy: string;
}
```

### Role（角色）

```typescript
interface Role {
  id: string;                       // 角色ID
  name: string;                     // 角色名称（唯一）
  description?: string;             // 角色描述
  
  // 层级关系
  parentId?: string;                // 父角色ID（支持继承）
  level: number;                    // 层级（1-5）
  path: string;                     // 路径（如：/admin/operator）
  
  // 权限
  permissions: Permission[];        // 直接权限
  inheritedPermissions: Permission[]; // 继承权限（计算字段）
  effectivePermissions: Permission[]; // 有效权限（合并后）
  
  // 统计
  userCount: number;                // 关联用户数
  
  // 元数据
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;
  updatedBy: string;
}
```

### Permission（权限）

```typescript
interface Permission {
  id: string;                       // 权限ID（如：vm:read）
  name: string;                     // 权限名称
  description?: string;             // 权限描述
  
  // 资源
  resource: string;                 // 资源类型（vm, alert, user等）
  action: string;                   // 操作（read, write, delete等）
  
  // 级别
  level: 'none' | 'read' | 'write' | 'admin';  // 权限级别
  
  // 范围
  scope?: 'global' | 'own' | 'department';  // 数据范围
}
```

### PermissionMatrix（权限矩阵）

```typescript
interface PermissionMatrix {
  // 角色列表（按层级排序）
  roles: Array<{
    id: string;
    name: string;
    level: number;
    parentId?: string;
    userCount: number;
  }>;
  
  // 功能模块列表
  modules: Array<{
    id: string;
    name: string;
    permissions: string[];        // 模块下的权限ID列表
  }>;
  
  // 权限矩阵数据
  matrix: Array<{
    roleId: string;
    moduleId: string;
    permissionId: string;
    level: 'none' | 'read' | 'write' | 'admin';
    source: 'direct' | 'inherited';  // 权限来源
    inheritedFrom?: string;          // 继承自哪个角色
  }>;
  
  // 冲突检测
  conflicts?: Array<{
    roleId: string;
    permissionId: string;
    conflictType: string;
    message: string;
  }>;
}
```

### UserPermissionDetail（用户权限详情）

```typescript
interface UserPermissionDetail {
  userId: string;
  userName: string;
  
  // 角色信息
  roles: Array<{
    id: string;
    name: string;
    level: number;
  }>;
  
  // 权限清单
  permissions: Array<{
    id: string;
    name: string;
    resource: string;
    action: string;
    level: 'read' | 'write' | 'admin';
    source: Array<{
      roleId: string;
      roleName: string;
      type: 'direct' | 'inherited';
    }>;
  }>;
  
  // 资源访问范围
  resourceScopes: Record<string, {
    scope: 'global' | 'own' | 'department';
    departmentId?: string;
  }>;
  
  // 生成时间
  generatedAt: Date;
}
```

### PermissionAuditLog（权限审计日志）

```typescript
interface PermissionAuditLog {
  id: string;
  
  // 操作信息
  action: 'create' | 'update' | 'delete' | 'grant' | 'revoke';
  resourceType: 'user' | 'role' | 'permission';
  resourceId: string;
  resourceName: string;
  
  // 变更详情
  changes: Array<{
    field: string;
    oldValue: any;
    newValue: any;
  }>;
  
  // 操作者
  operatorId: string;
  operatorName: string;
  operatorIp: string;
  
  // 时间
  createdAt: Date;
  
  // 备注
  note?: string;
}
```

---

## 接口详情

### 用户管理

#### 1. 获取用户列表

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/users`
- 认证: 需要Access Token
- 权限: `user:read`

**查询参数**
```
GET /api/v1/users?page=1&pageSize=20&status=active&roleId=role_admin&keyword=admin
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": "usr_001",
        "username": "admin",
        "email": "admin@company.com",
        "name": "系统管理员",
        "department": "IT部",
        "roles": ["role_admin"],
        "roleNames": ["系统管理员"],
        "status": "active",
        "mustChangePassword": false,
        "mfaEnabled": false,
        "lastLoginAt": "2026-02-03T10:00:00Z",
        "preferences": {
          "language": "zh-CN",
          "theme": "dark",
          "timezone": "Asia/Shanghai",
          "dateFormat": "YYYY-MM-DD"
        },
        "createdAt": "2026-01-01T00:00:00Z",
        "updatedAt": "2026-02-03T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 45,
      "totalPages": 3
    }
  }
}
```

---

#### 2. 创建用户

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/users`
- 认证: 需要Access Token
- 权限: `user:write`

**请求参数**
```json
{
  "username": "operator01",
  "email": "operator01@company.com",
  "name": "运维工程师01",
  "phone": "13800138001",
  "department": "运维部",
  "roles": ["role_operator"],
  "status": "active",
  "initialPassword": "TempPass123!",
  "mustChangePassword": true,
  "preferences": {
    "language": "zh-CN"
  }
}
```

**成功响应 (201)**
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": "usr_new_001",
    "username": "operator01",
    "name": "运维工程师01",
    "status": "active",
    "mustChangePassword": true,
    "createdAt": "2026-02-03T14:00:00Z"
  }
}
```

**约束**
- 用户名：3-50字符，字母数字下划线
- 邮箱：必须唯一，有效邮箱格式
- 初始密码：8-32字符，必须包含大小写字母、数字、特殊字符

---

#### 3. 更新用户

**基本信息**
- 方法: `PUT`
- 路径: `/api/v1/users/{id}`
- 认证: 需要Access Token
- 权限: `user:write`

**请求参数**
```json
{
  "name": "运维工程师01（改名）",
  "department": "运维部",
  "roles": ["role_operator", "role_viewer"],
  "status": "active",
  "mustChangePassword": false,
  "preferences": {
    "language": "zh-CN",
    "theme": "dark"
  }
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": "usr_002",
    "updatedAt": "2026-02-03T14:05:00Z"
  }
}
```

---

#### 4. 重置密码

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/users/{id}/reset-password`
- 认证: 需要Access Token
- 权限: `user:write`

**请求参数**
```json
{
  "newPassword": "NewPass123!",
  "mustChangePassword": true,
  "notifyUser": true
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "密码重置成功",
  "data": {
    "id": "usr_002",
    "passwordChangedAt": "2026-02-03T14:10:00Z",
    "mustChangePassword": true,
    "notificationSent": true
  }
}
```

---

### 角色管理

#### 5. 获取角色列表

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/roles`
- 认证: 需要Access Token
- 权限: `user:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "id": "role_admin",
        "name": "系统管理员",
        "description": "拥有所有权限",
        "parentId": null,
        "level": 1,
        "path": "/admin",
        "permissions": ["*"],
        "userCount": 2,
        "createdAt": "2026-01-01T00:00:00Z"
      },
      {
        "id": "role_operator",
        "name": "运维工程师",
        "description": "日常运维操作权限",
        "parentId": null,
        "level": 1,
        "path": "/operator",
        "permissions": ["vm:read", "vm:write", "alert:read", "alert:write"],
        "userCount": 8,
        "createdAt": "2026-01-01T00:00:00Z"
      },
      {
        "id": "role_viewer",
        "name": "只读用户",
        "description": "仅查看权限",
        "parentId": "role_operator",
        "level": 2,
        "path": "/operator/viewer",
        "permissions": ["vm:read", "alert:read"],
        "inheritedPermissions": ["vm:read", "vm:write", "alert:read", "alert:write"],
        "effectivePermissions": ["vm:read", "vm:write", "alert:read", "alert:write"],
        "userCount": 15,
        "createdAt": "2026-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

#### 6. 创建角色

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/roles`
- 认证: 需要Access Token
- 权限: `user:write`

**请求参数**
```json
{
  "name": "高级运维工程师",
  "description": "拥有更多运维权限",
  "parentId": "role_operator",
  "permissions": [
    {
      "id": "vm:admin",
      "level": "admin"
    },
    {
      "id": "alert:admin",
      "level": "admin"
    },
    {
      "id": "history:export",
      "level": "write"
    }
  ]
}
```

**成功响应 (201)**
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": "role_new_001",
    "name": "高级运维工程师",
    "level": 2,
    "path": "/operator/senior",
    "userCount": 0,
    "createdAt": "2026-02-03T14:15:00Z"
  }
}
```

---

#### 7. 更新角色权限

**基本信息**
- 方法: `PUT`
- 路径: `/api/v1/roles/{id}/permissions`
- 认证: 需要Access Token
- 权限: `user:write`

**请求参数**
```json
{
  "permissions": [
    {
      "id": "vm:read",
      "level": "read"
    },
    {
      "id": "vm:write",
      "level": "write"
    },
    {
      "id": "alert:read",
      "level": "read"
    },
    {
      "id": "alert:write",
      "level": "write"
    }
  ]
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "权限更新成功",
  "data": {
    "id": "role_operator",
    "effectivePermissions": ["vm:read", "vm:write", "alert:read", "alert:write"],
    "updatedAt": "2026-02-03T14:20:00Z"
  }
}
```

**约束**
- 实时权限冲突检测
- 子角色权限不能超过父角色
- 变更后1分钟内生效

---

### 权限矩阵

#### 8. 获取权限矩阵

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/permissions/matrix`
- 认证: 需要Access Token
- 权限: `user:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "roles": [
      {
        "id": "role_admin",
        "name": "系统管理员",
        "level": 1,
        "userCount": 2
      },
      {
        "id": "role_operator",
        "name": "运维工程师",
        "level": 1,
        "userCount": 8
      },
      {
        "id": "role_viewer",
        "name": "只读用户",
        "level": 2,
        "parentId": "role_operator",
        "userCount": 15
      }
    ],
    "modules": [
      {
        "id": "vm",
        "name": "VM管理",
        "permissions": ["vm:read", "vm:write", "vm:admin"]
      },
      {
        "id": "alert",
        "name": "告警管理",
        "permissions": ["alert:read", "alert:write", "alert:admin"]
      },
      {
        "id": "history",
        "name": "历史数据",
        "permissions": ["history:read", "history:export"]
      },
      {
        "id": "user",
        "name": "用户管理",
        "permissions": ["user:read", "user:write"]
      }
    ],
    "matrix": [
      {
        "roleId": "role_admin",
        "moduleId": "vm",
        "permissionId": "vm:read",
        "level": "admin",
        "source": "direct"
      },
      {
        "roleId": "role_operator",
        "moduleId": "vm",
        "permissionId": "vm:read",
        "level": "write",
        "source": "direct"
      },
      {
        "roleId": "role_viewer",
        "moduleId": "vm",
        "permissionId": "vm:read",
        "level": "write",
        "source": "inherited",
        "inheritedFrom": "role_operator"
      }
    ],
    "conflicts": []
  }
}
```

---

#### 9. 获取用户权限详情

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/users/{id}/permissions/detail`
- 认证: 需要Access Token
- 权限: `user:read`

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "userId": "usr_002",
    "userName": "运维工程师01",
    "roles": [
      {
        "id": "role_operator",
        "name": "运维工程师",
        "level": 1
      },
      {
        "id": "role_viewer",
        "name": "只读用户",
        "level": 2
      }
    ],
    "permissions": [
      {
        "id": "vm:read",
        "name": "查看VM信息",
        "resource": "vm",
        "action": "read",
        "level": "read",
        "source": [
          {
            "roleId": "role_operator",
            "roleName": "运维工程师",
            "type": "direct"
          }
        ]
      },
      {
        "id": "vm:write",
        "name": "编辑VM信息",
        "resource": "vm",
        "action": "write",
        "level": "write",
        "source": [
          {
            "roleId": "role_operator",
            "roleName": "运维工程师",
            "type": "direct"
          },
          {
            "roleId": "role_viewer",
            "roleName": "只读用户",
            "type": "inherited"
          }
        ]
      }
    ],
    "resourceScopes": {
      "vm": {
        "scope": "global"
      },
      "alert": {
        "scope": "global"
      }
    },
    "generatedAt": "2026-02-03T14:25:00Z"
  }
}
```

---

#### 10. 检查权限冲突

**基本信息**
- 方法: `POST`
- 路径: `/api/v1/permissions/check-conflict`
- 认证: 需要Access Token
- 权限: `user:read`

**请求参数**
```json
{
  "roleId": "role_new_001",
  "parentId": "role_operator",
  "permissions": [
    {
      "id": "user:write",
      "level": "write"
    }
  ]
}
```

**成功响应 (200)**
```json
{
  "code": 200,
  "message": "检查完成",
  "data": {
    "hasConflict": true,
    "conflicts": [
      {
        "permissionId": "user:write",
        "conflictType": "parent_restriction",
        "message": "该权限超出父角色权限范围，父角色无user:write权限"
      }
    ],
    "suggestions": [
      "请将user:write权限授予父角色",
      "或选择其他不超过父角色权限的权限"
    ]
  }
}
```

---

### 权限审计

#### 11. 获取权限变更历史

**基本信息**
- 方法: `GET`
- 路径: `/api/v1/permissions/audit`
- 认证: 需要Access Token
- 权限: `user:read`

**查询参数**
```
GET /api/v1/permissions/audit?page=1&pageSize=20&resourceType=user&startTime=2026-02-01T00:00:00Z
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
        "action": "grant",
        "resourceType": "user",
        "resourceId": "usr_003",
        "resourceName": "运维工程师03",
        "changes": [
          {
            "field": "roles",
            "oldValue": ["role_viewer"],
            "newValue": ["role_operator"]
          }
        ],
        "operatorId": "usr_001",
        "operatorName": "系统管理员",
        "operatorIp": "192.168.1.100",
        "createdAt": "2026-02-03T10:30:00Z",
        "note": "晋升为运维工程师"
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

## 错误码定义

| 错误码 | 英文消息 | 中文消息 | 日文消息 | 说明 |
|--------|---------|---------|---------|------|
| 400 | Bad Request | 请求参数错误 | リクエストパラメータエラー | 参数缺失或格式错误 |
| 400-USERNAME | Invalid Username | 用户名格式错误 | ユーザー名の形式が間違っています | 不符合用户名规范 |
| 400-EMAIL | Invalid Email | 邮箱格式错误或已存在 | メールアドレスの形式が間違っているか既に存在します | 邮箱验证失败 |
| 401 | Unauthorized | 未授权 | 未認証 | Token无效或过期 |
| 403 | Forbidden | 权限不足 | アクセス権限がありません | 无权限管理用户 |
| 403-SELF | Cannot Modify Self | 不能修改自己的关键信息 | 自分の重要な情報を変更できません | 安全限制 |
| 404 | Not Found | 用户不存在 | ユーザーが見つかりません | 用户ID不存在 |
| 404-ROLE | Role Not Found | 角色不存在 | ロールが見つかりません | 角色ID不存在 |
| 409 | Conflict | 用户名或邮箱已存在 | ユーザー名またはメールアドレスが既に存在します | 重复数据 |
| 409-PARENT | Invalid Parent Role | 父角色无效或层级超限 | 親ロールが無効または階層制限を超えています | 继承层级超过5层 |
| 409-CONFLICT | Permission Conflict | 权限配置冲突 | 権限設定が競合しています | 权限冲突 |
| 422 | Invalid Permission | 无效权限 | 無効な権限 | 权限ID不存在 |
| 500 | Server Error | 服务器内部错误 | サーバーエラー | 服务器错误 |

---

## 变更记录

### 版本 v1.0 (2026-02-03)
**修改人**: BE工程师  
**修改原因**: 基于REQ_20260202_VM监控系统需求文档初始创建  
**具体修改**:
- [x] 新增用户CRUD接口
- [x] 新增角色层级管理接口（支持继承）
- [x] 新增权限矩阵查询和批量设置接口
- [x] 新增权限冲突检测接口
- [x] 新增用户权限详情查询接口
- [x] 新增权限审计日志接口
- [x] 定义用户、角色、权限数据模型
- [x] 定义权限矩阵和审计日志模型

**影响范围**:
- 前端界面: 是（用户管理页面、角色管理页面、权限矩阵页面）
- 后端API: 是（用户服务、角色服务、权限服务、审计服务）
- 数据库结构: 是（users, roles, permissions, audit_logs表）
- 部署配置: 是（权限缓存配置、RBAC策略配置）

**相关文档**:
- REQ_20260202_VM监控系统.md（用户角色、RBAC、安全性要求）
- UI_20260202_VM监控系统_视觉设计指南.md（用户管理页面、权限矩阵页面、权限管理组件库）
- API_AUTH_认证授权模块.md（用户登录认证）

---

**文档管理说明**:
1. 权限变更实时生效（1分钟内），无需重新登录
2. 角色层级最多5层，防止循环继承
3. 子角色权限不能超过父角色，实时冲突检测
4. 权限缓存15分钟，强制刷新可立即生效
5. 审计日志保留2年，支持合规要求
6. 字段变更需记录在`api-changes.md`
