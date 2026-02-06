const express = require('express');
const cors = require('cors');

const app = express();
const PORT = 8080;

// 中间件
app.use(cors());
app.use(express.json());

// Mock数据
const mockUsers = [
  {
    id: 1,
    username: 'admin',
    email: 'admin@example.com',
    role: 'admin',
    avatar: null,
    createdAt: new Date().toISOString()
  }
];

const mockVMs = [
  {
    id: 'vm-001',
    name: 'web-server-01',
    status: 'running',
    cpuUsage: 45.2,
    memoryUsage: 62.8,
    diskUsage: 78.1,
    uptime: '15 days',
    ip: '192.168.1.100',
    os: 'Ubuntu 20.04'
  },
  {
    id: 'vm-002',
    name: 'database-01',
    status: 'running',
    cpuUsage: 78.5,
    memoryUsage: 89.2,
    diskUsage: 45.6,
    uptime: '30 days',
    ip: '192.168.1.101',
    os: 'CentOS 8'
  }
];

// 认证接口
app.post('/api/v1/auth/login', (req, res) => {
  const { username, password } = req.body;
  
  if (username === 'admin' && password === 'admin') {
    res.json({
      code: 200,
      message: '登录成功',
      data: {
        user: mockUsers[0],
        accessToken: 'mock-access-token',
        refreshToken: 'mock-refresh-token'
      }
    });
  } else {
    res.status(401).json({
      code: 401,
      message: '用户名或密码错误'
    });
  }
});

app.post('/api/v1/auth/refresh', (req, res) => {
  res.json({
    code: 200,
    message: 'Token刷新成功',
    data: {
      accessToken: 'new-mock-access-token',
      refreshToken: 'new-mock-refresh-token'
    }
  });
});

// 用户管理接口
app.get('/api/v1/users', (req, res) => {
  res.json({
    code: 200,
    message: '获取成功',
    data: mockUsers
  });
});

// VM管理接口
app.get('/api/v1/vms', (req, res) => {
  res.json({
    code: 200,
    message: '获取成功',
    data: {
      vms: mockVMs,
      total: mockVMs.length,
      page: 1,
      pageSize: 10
    }
  });
});

app.get('/api/v1/vms/stats', (req, res) => {
  res.json({
    code: 200,
    message: '获取成功',
    data: {
      total: 2,
      running: 2,
      stopped: 0,
      warning: 1
    }
  });
});

// 实时监控接口
app.get('/api/v1/realtime/metrics', (req, res) => {
  res.json({
    code: 200,
    message: '获取成功',
    data: {
      timestamp: new Date().toISOString(),
      cpu: Math.random() * 100,
      memory: Math.random() * 100,
      disk: Math.random() * 100,
      network: Math.random() * 1000
    }
  });
});

// 告警接口
app.get('/api/v1/alerts', (req, res) => {
  res.json({
    code: 200,
    message: '获取成功',
    data: [
      {
        id: 1,
        level: 'warning',
        message: 'database-01 CPU使用率过高',
        vmId: 'vm-002',
        createdAt: new Date().toISOString()
      }
    ]
  });
});

// 系统设置接口
app.get('/api/v1/system/settings', (req, res) => {
  res.json({
    code: 200,
    message: '获取成功',
    data: {
      monitoringInterval: 30,
      alertThresholds: {
        cpu: 80,
        memory: 85,
        disk: 90
      }
    }
  });
});

app.listen(PORT, () => {
  console.log(`🚀 Mock API服务器已启动: http://localhost:${PORT}`);
  console.log('📱 前端现在可以正常访问了！');
  console.log('');
  console.log('测试账号:');
  console.log('用户名: admin');
  console.log('密码: admin');
  console.log('');
  console.log('可用接口:');
  console.log('- POST /api/v1/auth/login (登录)');
  console.log('- GET  /api/v1/vms (虚拟机列表)');
  console.log('- GET  /api/v1/alerts (告警列表)');
  console.log('- GET  /api/v1/realtime/metrics (实时监控)');
  console.log('- 更多接口...');
});