# Windows+WSL+Docker 环境搭建完整指南

## 目录
1. [环境概述](#1-环境概述)
2. [WSL2 安装与配置](#2-wsl2-安装与配置)
3. [Docker 安装与配置](#3-docker-安装与配置)
4. [中间件安装](#4-中间件安装)
5. [开发环境模板](#5-开发环境模板)
6. [项目部署指南](#6-项目部署指南)
7. [维护与优化](#7-维护与优化)
8. [常见问题](#8-常见问题)

---

## 1. 环境概述

### 1.1 当前环境信息
- **操作系统**: Windows
- **WSL 发行版**: Ubuntu 24.04.3 LTS
- **WSL 版本**: WSL 2
- **Docker 版本**: 28.2.2
- **系统资源**: 5 CPU, 9.7GB 内存

### 1.2 架构说明
```
Windows 主机
    │
    ├── WSL 2 (Ubuntu 24.04)
    │       │
    │       ├── Docker Engine 28.2.2
    │       │       │
    │       │       ├── MySQL 8.0
    │       │       ├── PostgreSQL 15
    │       │       ├── Redis 7
    │       │       ├── RabbitMQ
    │       │       ├── Kafka
    │       │       └── MinIO
    │       │
    │       └── 开发工具
    │               ├── Node.js
    │               ├── Python
    │               └── Go
    │
    └── 开发工具
            ├── VS Code
            ├── Git
            └── Docker Desktop (可选)
```

---

## 2. WSL2 安装与配置

### 2.1 系统要求
- **Windows 10**: 版本 2004+ (Build 19041+)
- **Windows 11**: 所有版本
- **BIOS**: 必须启用虚拟化

### 2.2 安装 WSL2

#### 2.2.1 启用 WSL 功能
```powershell
# 以管理员身份运行 PowerShell

# 启用 WSL 功能
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart

# 启用虚拟机平台
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart

# 重启计算机
Restart-Computer
```

#### 2.2.2 安装 WSL2 更新包
```powershell
# 下载并安装 WSL2 Linux 内核更新包
# 下载地址: https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi

# 或者使用命令安装
Invoke-WebRequest -Uri https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi -OutFile wsl_update_x64.msi
Start-Process -FilePath wsl_update_x64.msi -ArgumentList "/passive" -Wait
```

#### 2.2.3 设置 WSL2 为默认版本
```powershell
# 设置默认版本
wsl --set-default-version 2

# 安装 Ubuntu 发行版
wsl --install -d Ubuntu

# 或者安装其他发行版
wsl --install -d Ubuntu-20.04
wsl --install -d Ubuntu-22.04
wsl --install -d Debian
wsl --install -d Kali-linux
wsl --install -d openSUSE-Leap-15.5
wsl --install -d Fedora-37
```

### 2.3 WSL2 配置优化

#### 2.3.1 创建 WSL 配置文件
在 Windows 中创建 `%USERPROFILE%\.wslconfig` 文件：

```ini
[wsl2]
# 内存限制 (建议为物理内存的 50-80%)
memory=4GB

# CPU 核心数 (建议为物理核心数的 50-75%)
processors=4

# 交换文件大小
swap=4GB

# 是否启用 localhost 转发
localhostForwarding=true

# 磁盘 I/O 优化
# 如果 SSD，建议设置为 true
nestedVirtualization=true

# GUI 应用支持 (Windows 11)
guiApplications=true
```

#### 2.3.2 WSL2 内存管理
```bash
# 在 WSL2 中优化内存使用
# 编辑 /etc/sysctl.conf
sudo tee -a /etc/sysctl.conf <<EOF

# WSL2 内存优化
vm.swappiness=10
vm.vfs_cache_pressure=50

# 网络优化
net.core.somaxconn=65535
net.ipv4.tcp_max_syn_backlog=65535
EOF

# 应用配置
sudo sysctl -p
```

### 2.4 WSL2 常用命令

```bash
# 列出所有 WSL 发行版
wsl -l -v

# 设置默认发行版
wsl -s Ubuntu

# 以特定用户运行
wsl -u username

# 在 WSL 中执行命令
wsl -e command

# 导出 WSL 发行版
wsl --export Ubuntu backup.tar

# 导入 WSL 发行版
wsl --import Ubuntu2 backup.tar --version 2

# 更新 WSL2 内核
wsl --update

# 关闭 WSL2
wsl --shutdown
```

---

## 3. Docker 安装与配置

### 3.1 Docker 安装

#### 3.1.1 安装 Docker Engine (Ubuntu)
```bash
# 更新软件包索引
sudo apt update

# 安装必要依赖
sudo apt install -y ca-certificates curl gnupg lsb-release

# 添加 Docker 官方 GPG 密钥
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# 添加 Docker 仓库
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装 Docker 引擎
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 验证安装
docker --version
docker compose version
```

#### 3.1.2 安装 Docker Compose V2
```bash
# Docker Compose V2 已随 docker-compose-plugin 安装
docker compose version

# 如果需要单独安装
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 3.2 Docker 服务配置

#### 3.2.1 启动 Docker 服务
```bash
# 启动 Docker 服务
sudo systemctl start docker

# 设置开机自启
sudo systemctl enable docker

# 查看服务状态
sudo systemctl status docker

# 重启服务
sudo systemctl restart docker
```

#### 3.2.2 Docker 用户组配置
```bash
# 将当前用户添加到 docker 组
sudo usermod -aG docker $USER

# 验证用户组
groups

# 重新登录或执行以下命令使权限生效
newgrp docker
```

### 3.3 Docker 配置文件优化

#### 3.3.1 创建 Docker 守护进程配置
```bash
# 创建配置目录
sudo mkdir -p /etc/docker

# 创建配置文件
sudo tee /etc/docker/daemon.json <<'EOF'
{
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.size=80%"
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3",
    "labels": "production"
  },
  "default-ulimits": {
    "nofile": {
      "Name": "nofile",
      "Hard": 1048576,
      "Soft": 1048576
    },
    "nproc": {
      "Name": "nproc",
      "Hard": 4096,
      "Soft": 2048
    }
  },
  "live-restore": true,
  "max-concurrent-downloads": 10,
  "max-concurrent-uploads": 5,
  "features": {
    "buildkit": true
  }
}
EOF

# 重启 Docker 服务
sudo systemctl restart docker
```

#### 3.3.2 配置镜像加速器
```bash
# 编辑配置文件
sudo tee -a /etc/docker/daemon.json <<'EOF',
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn/",
    "https://hub-mirror.c.163.com/",
    "https://mirror.baidubce.com",
    "https://registry.docker-cn.com"
  ]
EOF

# 重启 Docker 服务
sudo systemctl restart docker

# 验证配置
docker info | grep "Registry Mirrors"
```

### 3.4 Docker 验证测试

```bash
# 运行测试容器
sudo docker run hello-world

# 查看 Docker 信息
docker info

# 查看 Docker 版本
docker version

# 查看运行中的容器
docker ps

# 查看所有容器
docker ps -a

# 查看本地镜像
docker images
```

---

## 4. 中间件安装

### 4.1 数据库

#### 4.1.1 MySQL 8.0
```bash
# 创建项目目录
mkdir -p ~/docker/mysql
cd ~/docker/mysql

# 创建配置文件
mkdir -p config init

cat > config/my.cnf <<EOF
[mysqld]
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
max_connections = 1000
innodb_buffer_pool_size = 512M

[client]
default-character-set = utf8mb4
EOF

# 创建 docker-compose.yml
cat > docker-compose.mysql.yml <<EOF
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: mysql-server
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root123456
      MYSQL_DATABASE: mydb
      MYSQL_USER: appuser
      MYSQL_PASSWORD: app123456
      MYSQL_CHARSET: utf8mb4
      MYSQL_COLLATION: utf8mb4_unicode_ci
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./config/my.cnf:/etc/mysql/my.cnf
      - ./init:/docker-entrypoint-initdb.d
    networks:
      - mysql_network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mysql_data:

networks:
  mysql_network:
    driver: bridge
EOF

# 启动 MySQL
docker-compose -f docker-compose.mysql.yml up -d

# 验证启动
docker ps | grep mysql

# 连接测试
docker exec -it mysql-server mysql -u root -p
```

#### 4.1.2 PostgreSQL 15
```bash
# 创建目录
mkdir -p ~/docker/postgresql
cd ~/docker/postgresql

# 创建配置
mkdir -p config init

cat > config/postgresql.conf <<EOF
listen_addresses = '*'
max_connections = 1000
shared_buffers = 256MB
work_mem = 4MB
maintenance_work_mem = 128MB
effective_cache_size = 1GB
log_min_duration_statement = 100
log_connections = on
log_disconnections = on
EOF

cat > docker-compose.postgres.yml <<EOF
version: '3.8'

services:
  postgres:
    image: postgres:15
    container_name: postgres-server
    restart: unless-stopped
    environment:
      POSTGRES_DB: mydb
      POSTGRES_USER: appuser
      POSTGRES_PASSWORD: app123456
      PGDATA: /var/lib/postgresql/data/pgdata
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./config/postgresql.conf:/etc/postgresql/postgresql.conf
      - ./init:/docker-entrypoint-initdb.d
    networks:
      - postgres_network
    command: postgres -c config_file=/etc/postgresql/postgresql.conf

volumes:
  postgres_data:

networks:
  postgres_network:
    driver: bridge
EOF

# 启动 PostgreSQL
docker compose -f docker-compose.postgres.yml up -d
```

### 4.2 缓存系统

#### 4.2.1 Redis 7
```bash
# 创建目录
mkdir -p ~/docker/redis
cd ~/docker/redis

# 创建配置文件
cat > redis.conf <<EOF
# 内存限制
maxmemory 512mb
maxmemory-policy allkeys-lru

# 持久化
appendonly yes
appendfsync everysec

# 网络配置
bind 0.0.0.0
port 6379

# 日志
loglevel notice
logfile ""

# 安全
# requirepass your_password_here
EOF

cat > docker-compose.redis.yml <<EOF
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    container_name: redis-server
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      - ./redis.conf:/usr/local/etc/redis/redis.conf
    networks:
      - redis_network
    command: redis-server /usr/local/etc/redis/redis.conf
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  redis_data:

networks:
  redis_network:
    driver: bridge
EOF

# 启动 Redis
docker compose -f docker-compose.redis.yml up -d

# 测试连接
docker exec -it redis-server redis-cli ping
```

#### 4.2.2 Memcached
```bash
# 创建目录
mkdir -p ~/docker/memcached
cd ~/docker/memcached

cat > docker-compose.memcached.yml <<EOF
version: '3.8'

services:
  memcached:
    image: memcached:1.6-alpine
    container_name: memcached-server
    restart: unless-stopped
    ports:
      - "11211:11211"
    environment:
      - MEMCACHED_MAX_CONN=1024
      - MEMCACHED_CACHE_SIZE=128
    networks:
      - memcached_network
    command: memcached -m 128 -p 11211 -u root

networks:
  memcached_network:
    driver: bridge
EOF

# 启动 Memcached
docker compose -f docker-compose.memcached.yml up -d
```

### 4.3 消息队列

#### 4.3.1 RabbitMQ
```bash
# 创建目录
mkdir -p ~/docker/rabbitmq
cd ~/docker/rabbitmq

# 创建配置文件
cat > rabbitmq.conf <<EOF
loopback_users.guest = false
listeners.tcp.default = 5672
management.tcp.port = 15672
EOF

cat > docker-compose.rabbitmq.yml <<EOF
version: '3.8'

services:
  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: rabbitmq-server
    restart: unless-stopped
    environment:
      RABBITMQ_DEFAULT_USER: admin
      RABBITMQ_DEFAULT_PASS: admin123456
      RABBITMQ_DEFAULT_VHOST: /
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
      - ./rabbitmq.conf:/etc/rabbitmq/rabbitmq.conf
    networks:
      - rabbitmq_network
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "check_port_connectivity"]
      interval: 30s
      timeout: 10s
      retries: 10

volumes:
  rabbitmq_data:

networks:
  rabbitmq_network:
    driver: bridge
EOF

# 启动 RabbitMQ
docker compose -f docker-compose.rabbitmq.yml up -d

# 访问管理界面
# http://localhost:15672
# 用户名: admin
# 密码: admin123456
```

#### 4.3.2 Apache Kafka
```bash
# 创建目录
mkdir -p ~/docker/kafka
cd ~/docker/kafka

cat > docker-compose.kafka.yml <<EOF
version: '3.8'

services:
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    container_name: zookeeper-server
    restart: unless-stopped
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"
    networks:
      - kafka_network

  kafka:
    image: confluentinc/cp-kafka:latest
    container_name: kafka-server
    restart: unless-stopped
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
    networks:
      - kafka_network

networks:
  kafka_network:
    driver: bridge
EOF

# 启动 Kafka
docker compose -f docker-compose.kafka.yml up -d

# 测试 Kafka
docker exec -it kafka-server kafka-topics --create --topic test --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
docker exec -it kafka-server kafka-topics --list --bootstrap-server localhost:9092
```

### 4.4 文件存储

#### 4.4.1 MinIO (S3 兼容)
```bash
# 创建目录
mkdir -p ~/docker/minio
cd ~/docker/minio

cat > docker-compose.minio.yml <<EOF
version: '3.8'

services:
  minio:
    image: minio/minio:latest
    container_name: minio-server
    restart: unless-stopped
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin123
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"
    networks:
      - minio_network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

volumes:
  minio_data:

networks:
  minio_network:
    driver: bridge
EOF

# 启动 MinIO
docker compose -f docker-compose.minio.yml up -d

# 访问控制台
# http://localhost:9001
# Access Key: minioadmin
# Secret Key: minioadmin123
```

### 4.5 管理工具

#### 4.5.1 Portainer (Docker GUI)
```bash
# 创建目录
mkdir -p ~/docker/portainer
cd ~/docker/portainer

cat > docker-compose.portainer.yml <<EOF
version: '3.8'

services:
  portainer:
    image: portainer/portainer-ce:latest
    container_name: portainer
    restart: unless-stopped
    ports:
      - "9443:9443"
      - "8000:8000"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - portainer_data:/data
    networks:
      - portainer_network

volumes:
  portainer_data:

networks:
  portainer_network:
    driver: bridge
EOF

# 启动 Portainer
docker compose -f docker-compose.portainer.yml up -d

# 访问管理界面
# https://localhost:9443
```

---

## 5. 开发环境模板

### 5.1 Node.js 项目模板

#### 5.1.1 创建项目
```bash
# 创建项目目录
mkdir -p ~/projects/nodejs-api
cd ~/projects/nodejs-api

# 创建目录结构
mkdir -p src/{routes,controllers,models,middleware,config} tests

# 创建 package.json
cat > package.json <<EOF
{
  "name": "nodejs-api",
  "version": "1.0.0",
  "description": "Node.js API 项目",
  "main": "src/index.js",
  "scripts": {
    "start": "node src/index.js",
    "dev": "nodemon src/index.js",
    "test": "jest"
  },
  "dependencies": {
    "express": "^4.18.2",
    "mysql2": "^3.6.0",
    "redis": "^4.6.0",
    "dotenv": "^16.3.1",
    "cors": "^2.8.5",
    "helmet": "^7.0.0",
    "morgan": "^1.10.0"
  },
  "devDependencies": {
    "nodemon": "^3.0.1",
    "jest": "^29.7.0"
  }
}
EOF
```

#### 5.1.2 创建 Dockerfile
```dockerfile
cat > Dockerfile <<EOF
FROM node:18-alpine

WORKDIR /app

# 复制依赖文件
COPY package*.json ./

# 安装依赖
RUN npm ci --only=production

# 复制源代码
COPY . .

# 暴露端口
EXPOSE 3000

# 启动命令
CMD ["node", "src/index.js"]
EOF
```

#### 5.1.3 创建 docker-compose.yml
```dockerfile
cat > docker-compose.yml <<EOF
version: '3.8'

services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=development
      - PORT=3000
      - DB_HOST=mysql-server
      - DB_PORT=3306
      - DB_NAME=mydb
      - DB_USER=appuser
      - DB_PASSWORD=app123456
      - REDIS_HOST=redis-server
      - REDIS_PORT=6379
    depends_on:
      mysql-server:
        condition: service_healthy
      redis-server:
        condition: service_healthy
    volumes:
      - .:/app
      - /app/node_modules
    networks:
      - app_network
    restart: unless-stopped

  mysql-server:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root123456
      MYSQL_DATABASE: mydb
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - app_network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis-server:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - app_network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mysql_data:
  redis_data:

networks:
  app_network:
    driver: bridge
EOF
```

#### 5.1.4 创建应用代码
```javascript
// src/index.js
const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const morgan = require('morgan');

const app = express();
const PORT = process.env.PORT || 3000;

// 中间件
app.use(helmet());
app.use(cors());
app.use(morgan('dev'));
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// 路由
app.get('/', (req, res) => {
  res.json({ 
    message: '欢迎使用 Node.js API',
    version: '1.0.0',
    status: '运行中'
  });
});

app.get('/health', (req, res) => {
  res.json({ status: '健康' });
});

// 错误处理
app.use((err, req, res, next) => {
  console.error(err.stack);
  res.status(500).json({ error: '服务器内部错误' });
});

// 启动服务器
app.listen(PORT, () => {
  console.log(`服务器运行在端口 ${PORT}`);
});
```

### 5.2 Python Flask 项目模板

```bash
# 创建项目
mkdir -p ~/projects/python-api
cd ~/projects/python-api

# 创建文件
cat > requirements.txt <<EOF
Flask==3.0.0
Flask-SQLAlchemy==3.1.1
Flask-Redis==0.4.0
psycopg2-binary==2.9.9
python-dotenv==1.0.0
gunicorn==21.2.0
EOF

cat > Dockerfile <<EOF
FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .

EXPOSE 5000

CMD ["gunicorn", "--bind", "0.0.0.0:5000", "app:app"]
EOF

cat > docker-compose.yml <<EOF
version: '3.8'

services:
  app:
    build: .
    ports:
      - "5000:5000"
    environment:
      - FLASK_ENV=development
      - FLASK_APP=app.py
      - DATABASE_URL=postgresql://appuser:app123456@postgres-server:5432/mydb
      - REDIS_URL=redis://redis-server:6379/0
    volumes:
      - .:/app
    depends_on:
      postgres-server:
        condition: service_healthy
      redis-server:
        condition: service_healthy
    networks:
      - app_network

  postgres-server:
    image: postgres:15
    environment:
      POSTGRES_DB: mydb
      POSTGRES_USER: appuser
      POSTGRES_PASSWORD: app123456
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - app_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U appuser -d mydb"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis-server:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    networks:
      - app_network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  redis_data:

networks:
  app_network:
    driver: bridge
EOF

cat > app.py <<EOF
from flask import Flask, jsonify
from flask_sqlalchemy import SQLAlchemy
from flask_redis import FlaskRedis

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'user:app123postgresql://app456@postgres-server:5432/mydb'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

db = SQLAlchemy(app)
redis = FlaskRedis(app)

@app.route('/')
def index():
    return jsonify({
        'message': '欢迎使用 Flask API',
        'version': '1.0.0',
        'status': '运行中'
    })

@app.route('/health')
def health():
    return jsonify({'status': '健康'})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
EOF
```

---

## 6. 项目部署指南

### 6.1 生产环境配置

#### 6.1.1 生产环境 docker-compose.yml
```bash
cat > docker-compose.prod.yml <<EOF
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.prod
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - PORT=3000
      - DB_HOST=mysql-prod
      - DB_PORT=3306
      - DB_NAME=mydb
      - DB_USER=appuser
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_HOST=redis-prod
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    networks:
      - prod_network
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: '512M'
        reservations:
          cpus: '0.25'
          memory: '256M'
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    networks:
      - prod_network
    restart: unless-stopped

  mysql-prod:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: mydb
    volumes:
      - mysql_prod_data:/var/lib/mysql
    networks:
      - prod_network
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: '1G'

  redis-prod:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_prod_data:/data
    networks:
      - prod_network
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: '256M'

volumes:
  mysql_prod_data:
  redis_prod_data:

networks:
  prod_network:
    driver: bridge
EOF
```

### 6.2 部署步骤

```bash
# 1. 构建镜像
docker compose -f docker-compose.prod.yml build

# 2. 设置环境变量
cp .env.example .env
# 编辑 .env 文件，填入实际密码

# 3. 启动服务
docker compose -f docker-compose.prod.yml up -d

# 4. 检查状态
docker compose -f docker-compose.prod.yml ps

# 5. 查看日志
docker compose -f docker-compose.prod.yml logs -f

# 6. 健康检查
curl http://localhost/health
```

---

## 7. 维护与优化

### 7.1 定期维护任务

```bash
# 创建维护脚本
cat > maintenance.sh <<'EOF'
#!/bin/bash

echo "开始系统维护..."

# 1. 检查 Docker 状态
echo "检查 Docker 服务..."
sudo systemctl status docker

# 2. 检查容器状态
echo "检查容器状态..."
docker ps -a

# 3. 检查磁盘使用
echo "检查磁盘使用..."
docker system df

# 4. 清理未使用的资源
echo "清理未使用的镜像..."
docker image prune -af

echo "清理未使用的容器..."
docker container prune -f

echo "清理未使用的网络..."
docker network prune -f

echo "清理构建缓存..."
docker builder prune -f

# 5. 检查运行中的服务
echo "检查运行中的服务..."
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo "维护完成!"
EOF

chmod +x maintenance.sh
```

### 7.2 备份策略

```bash
# 创建备份脚本
cat > backup.sh <<'EOF'
#!/bin/bash

BACKUP_DIR=~/backups
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

echo "开始备份..."

# 备份 MySQL
echo "备份 MySQL..."
docker exec mysql-server mysqldump -u root -proot123456 mydb > $BACKUP_DIR/mydb_$DATE.sql

# 备份 PostgreSQL
echo "备份 PostgreSQL..."
docker exec postgres-server pg_dump -U appuser mydb > $BACKUP_DIR/mydb_pg_$DATE.sql

# 备份 Redis
echo "备份 Redis..."
docker exec redis-server BGSAVE
docker cp redis-server:/data/dump.rdb $BACKUP_DIR/redis_$DATE.rdb

# 清理 7 天前的备份
echo "清理旧备份..."
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.rdb" -mtime +7 -delete

echo "备份完成! 文件保存在: $BACKUP_DIR"
EOF

chmod +x backup.sh
```

### 7.3 性能监控

```bash
# 创建监控脚本
cat > monitor.sh <<'EOF'
#!/bin/bash

echo "=== Docker 容器监控 ==="

# 容器资源使用
echo "容器资源使用:"
docker stats --no-stream

echo ""
echo "=== 容器状态 ==="
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "=== 磁盘使用 ==="
docker system df

echo ""
echo "=== Docker 信息 ==="
docker info | grep -E "Server Version|Storage Driver|Logging Driver|Cgroup Driver"
EOF

chmod +x monitor.sh
```

---

## 8. 常见问题

### 8.1 Docker 无法启动

```bash
# 检查错误日志
sudo journalctl -u docker -n 100

# 检查 Docker 守护进程状态
sudo systemctl status docker

# 重启 Docker
sudo systemctl restart docker

# 检查端口占用
sudo netstat -tlnp | grep docker
```

### 8.2 容器连接问题

```bash
# 检查网络
docker network ls
docker network inspect bridge

# 测试容器连通性
docker exec container_name ping google.com

# 检查 DNS
docker exec container_name cat /etc/resolv.conf
```

### 8.3 磁盘空间不足

```bash
# 检查磁盘使用
docker system df

# 清理未使用资源
docker system prune -af

# 清理构建缓存
docker builder prune -af

# 检查实际磁盘空间
df -h
```

### 8.4 权限问题

```bash
# 添加用户到 docker 组
sudo usermod -aG docker $USER

# 重新登录
newgrp docker

# 或者使用 sudo 运行
sudo docker ps
```

---

## 附录

### A. 常用命令速查

```bash
# 容器管理
docker compose up -d      # 启动服务
docker compose down        # 停止服务
docker compose restart    # 重启服务
docker compose logs -f    # 查看日志
docker compose ps         # 查看状态

# 镜像管理
docker images             # 列出镜像
docker build -t name .    # 构建镜像
docker rmi image_name     # 删除镜像

# 调试
docker exec -it container_name bash  # 进入容器
docker logs container_name          # 查看日志
docker inspect container_name       # 检查容器
docker top container_name           # 查看进程

# 清理
docker system prune -af  # 清理所有未使用资源
docker volume prune -f   # 清理未使用卷
```

### B. 默认端口映射

| 服务 | 端口 |
|------|------|
| MySQL | 3306 |
| PostgreSQL | 5432 |
| Redis | 6379 |
| Memcached | 11211 |
| RabbitMQ | 5672, 15672 |
| Kafka | 9092 |
| MinIO | 9000, 9001 |
| Portainer | 9443 |
| Node.js API | 3000 |
| Python API | 5000 |

### C. 默认凭据

| 服务 | 用户名 | 密码 |
|------|--------|------|
| MySQL | root | root123456 |
| MySQL | appuser | app123456 |
| PostgreSQL | appuser | app123456 |
| Redis | 无 | 无 (可配置) |
| RabbitMQ | admin | admin123456 |
| MinIO | minioadmin | minioadmin123 |

---

**文档版本**: 1.0  
**最后更新**: 2026-02-04  
**适用系统**: Windows 10/11 + WSL2 + Ubuntu + Docker

建议配合高级配置扩展文档一起使用，获得更完整的配置指导。