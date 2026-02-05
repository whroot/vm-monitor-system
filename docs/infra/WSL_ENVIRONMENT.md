# VM监控系统 - WSL环境配置指南

## 已安装服务

| 服务 | 状态 | 连接信息 |
|------|------|----------|
| PostgreSQL 16 | ✅ 运行中 | localhost:5432 |
| Redis 7 | ✅ 运行中 | localhost:6379 |
| InfluxDB 2 | �待安装 | localhost:8086 |
| RabbitMQ | �待安装 | localhost:5672 |

## 已运行服务测试

```bash
# Redis测试
wsl redis-cli ping
# 期望输出: PONG

# PostgreSQL测试
wsl sudo -u postgres psql -c "SELECT version();"
```

## 手动安装InfluxDB和RabbitMQ

### 1. 安装InfluxDB

```bash
# 下载并安装
wsl sudo apt-get update
wsl wget https://dl.influxdata.com/influxdb/releases/influxdb2-2.7.6-amd64.deb
wsl sudo dpkg -i influxdb2-2.7.6-amd64.deb
wsl sudo systemctl enable influxdb
wsl sudo systemctl start influxdb

# 初始化 (首次启动后访问 http://localhost:8086)
```

### 2. 安装RabbitMQ

```bash
# 安装
wsl sudo apt-get install -y rabbitmq-server

# 启用管理插件
wsl sudo rabbitmq-plugins enable rabbitmq_management

# 创建管理员用户
wsl sudo rabbitmqctl add_user admin password
wsl sudo rabbitmqctl set_user_tags admin administrator
wsl sudo rabbitmqctl set_permissions -p / admin ".*" ".*" ".*"

# 启动服务
wsl sudo systemctl enable rabbitmq-server
wsl sudo systemctl start rabbitmq-server
```

## 连接配置

创建 `.env` 文件：

```bash
# 数据库配置
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=vm_monitor

# Redis配置
export REDIS_HOST=localhost
export REDIS_PORT=6379

# InfluxDB配置 (安装后配置)
export INFLUXDB_URL=http://localhost:8086
export INFLUXDB_TOKEN=your-token
export INFLUXDB_ORG=vm-monitor
export INFLUXDB_BUCKET=metrics

# RabbitMQ配置 (安装后配置)
export RABBITMQ_URL=amqp://admin:password@localhost:5672/

# 应用配置
export API_PORT=8080
```

## 启动后端服务

```bash
wsl
cd /mnt/d/work/OpenCode/server
go mod tidy
go run ./cmd
```

## 验证中间件连接

```bash
# 测试PostgreSQL
wsl sudo -u postgres psql -d vm_monitor -c "SELECT 1;"

# 测试Redis
wsl redis-cli ping

# 测试InfluxDB (安装后)
wsl curl http://localhost:8086/health

# 测试RabbitMQ (安装后)
wsl sudo rabbitmqctl status
```
