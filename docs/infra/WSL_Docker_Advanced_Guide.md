# Windows+WSL+Docker 高级配置扩展部分

## 10. 高级配置扩展

### 10.1 WSL2 虚拟机优化

#### 10.1.1 自定义 WSL2 配置
```bash
# 创建高级 WSL2 配置
cat > ~/.wslconfig <<EOF
[wsl2]
# 高级内存配置
memory=8GB
processors=8
swap=4GB

# 磁盘优化
localhostForwarding=true
kernelCommandLine=quiet

# 定时任务配置
automaticCheck=false
automaticRepair=false
EOF
```

#### 10.1.2 WSL2 内核优化
```bash
# 更新 WSL2 内核
wsl.exe --update

# 检查 WSL2 内核版本
uname -r

# 优化 WSL2 内核参数
cat >> /etc/sysctl.conf <<EOF
# WSL2 优化参数
vm.swappiness=10
net.core.somaxconn=65535
fs.file-max=2097152
EOF
```

### 10.2 Docker 高级配置

#### 10.2.1 Docker 服务优化
```bash
# 创建 Docker 服务配置
sudo mkdir -p /etc/systemd/system/docker.service.d
sudo tee /etc/systemd/system/docker.service.d/override.conf <<EOF
[Service]
# 内存限制
MemoryMax=4G
# CPU 限制
CPUQuota=200%
# 启动超时
TimeoutStartSec=300
TimeoutStopSec=300
EOF

# 重启 Docker 服务
sudo systemctl daemon-reload
sudo systemctl restart docker
```

#### 10.2.2 Docker Compose 资源优化
```bash
# 创建 Docker Compose 优化配置
cat > docker-compose.resources.yml <<EOF
version: '3.8'

services:
  app:
    deploy:
      resources:
        limits:
          cpus: '0.8'
          memory: '512M'
        reservations:
          cpus: '0.4'
          memory: '256M'
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  mysql:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: '1G'
        reservations:
          cpus: '0.5'
          memory: '512M'

  redis:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: '256M'
EOF
```

### 10.3 安全配置

#### 10.3.1 Docker 安全加固
```bash
# 创建安全配置脚本
cat > docker-security.sh <<'EOF'
#!/bin/bash

echo "Docker 安全检查开始..."

# 1. 检查 Docker 版本
echo "Docker 版本:"
docker version

# 2. 检查运行中的容器
echo "运行中的容器:"
docker ps

# 3. 检查镜像
echo "本地镜像:"
docker images

# 4. 检查网络
echo "Docker 网络:"
docker network ls

# 5. 检查存储卷
echo "存储卷:"
docker volume ls

# 6. 检查安全配置
echo "安全配置检查:"
docker info | grep -i security

# 7. 检查容器特权模式
echo "特权容器检查:"
docker ps --format '{{.Names}}' | xargs -I {} docker inspect {} --format '{{.Name}} {{.HostConfig.Privileged}}' | grep true

echo "安全检查完成!"
EOF

chmod +x docker-security.sh
```

#### 10.3.2 敏感信息管理
```bash
# 创建环境变量模板
cat > .env.example <<EOF
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_NAME=myapp
DB_USER=appuser
DB_PASSWORD=your_password_here

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_password_here

# JWT 配置
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRES_IN=24h

# API 配置
API_KEY=your_api_key
API_SECRET=your_api_secret

# 邮件配置
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your_email@example.com
SMTP_PASSWORD=your_email_password
EOF

# 设置文件权限
chmod 600 .env.example
```

### 10.4 监控配置

#### 10.4.1 基础监控
```bash
# 创建监控脚本
cat > monitor.sh <<'EOF'
#!/bin/bash

# 容器状态监控
docker stats --no-stream

# 检查容器健康状态
docker ps --format '{{.Names}} {{.Status}}'

# 检查磁盘使用
docker system df

# 检查 Docker 服务状态
sudo systemctl status docker --no-pager

# 检查日志错误
docker logs --tail 100 2>&1 | grep -i error || echo "无错误日志"
EOF

chmod +x monitor.sh
```

#### 10.4.2 健康检查配置
```bash
# 创建健康检查配置
cat > healthcheck.yml <<EOF
version: '3.8'

services:
  app:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  mysql:
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h localhost -u root -p$$MYSQL_ROOT_PASSWORD"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 30s

  redis:
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost/"]
      interval: 30s
      timeout: 10s
      retries: 3
EOF
```

### 10.5 日志管理

#### 10.5.1 日志配置
```bash
# 创建日志配置
cat > logging.yml <<EOF
version: '3.8'

services:
  app:
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"
        labels: "production"

  nginx:
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "3"

  mysql:
    logging:
      driver: "json-file"
      options:
        max-size: "200m"
        max-file: "3"
EOF
```

#### 10.5.2 日志查看工具
```bash
# 创建日志查看脚本
cat > logs.sh <<'EOF'
#!/bin/bash

# 查看所有容器日志
docker compose logs -f

# 查看特定服务日志
if [ -n "$1" ]; then
    docker compose logs -f "$1"
fi

# 查看最近 100 行日志
docker compose logs --tail 100

# 查看实时错误日志
docker compose logs 2>&1 | grep -i error
EOF

chmod +x logs.sh
```

### 10.6 备份恢复

#### 10.6.1 数据库备份
```bash
# 创建备份脚本
cat > backup.sh <<'EOF'
#!/bin/bash

# 备份目录
BACKUP_DIR=./backups
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# MySQL 备份
echo "开始备份 MySQL..."
docker exec mysql-server mysqldump -u root -p${MYSQL_ROOT_PASSWORD} mydb > "$BACKUP_DIR/mydb_$DATE.sql"
echo "MySQL 备份完成: $BACKUP_DIR/mydb_$DATE.sql"

# PostgreSQL 备份
echo "开始备份 PostgreSQL..."
docker exec postgres-server pg_dump -U appuser mydb > "$BACKUP_DIR/mydb_$DATE.sql"
echo "PostgreSQL 备份完成: $BACKUP_DIR/mydb_$DATE.sql"

# Redis 备份
echo "开始备份 Redis..."
docker exec redis-server BGSAVE
docker exec redis-server lastsave
docker cp redis-server:/data/dump.rdb "$BACKUP_DIR/redis_$DATE.rdb"
echo "Redis 备份完成: $BACKUP_DIR/redis_$DATE.rdb"

# 清理旧备份（保留 7 天）
find "$BACKUP_DIR" -name "*.sql" -mtime +7 -delete
find "$BACKUP_DIR" -name "*.rdb" -mtime +7 -delete

echo "备份完成!"
EOF

chmod +x backup.sh
```

#### 10.6.2 数据恢复
```bash
# 创建恢复脚本
cat > restore.sh <<'EOF'
#!/bin/bash

if [ -z "$1" ]; then
    echo "用法: $0 <备份文件路径>"
    exit 1
fi

BACKUP_FILE="$1"

# MySQL 恢复
echo "恢复 MySQL..."
docker exec -i mysql-server mysql -u root -p${MYSQL_ROOT_PASSWORD} mydb < "$BACKUP_FILE"
echo "MySQL 恢复完成!"

# PostgreSQL 恢复
echo "恢复 PostgreSQL..."
docker exec -i postgres-server psql -U appuser -d mydb < "$BACKUP_FILE"
echo "PostgreSQL 恢复完成!"

echo "数据恢复完成!"
EOF

chmod +x restore.sh
```

### 10.7 开发工具

#### 10.7.1 常用别名
```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
cat >> ~/.bashrc <<'EOF'

# Docker 快捷命令
alias d='docker'
alias dc='docker compose'
alias dcu='docker compose up'
alias dcd='docker compose down'
alias dcl='docker compose logs'
alias dclf='docker compose logs -f'
alias dcb='docker compose build'
alias dcr='docker compose restart'
alias dps='docker ps'
alias dpsa='docker ps -a'
alias dim='docker images'
alias dvol='docker volume ls'
alias dnet='docker network ls'

# 清理命令
alias dcleani='docker image prune -a'
alias dcleanc='docker container prune'
alias dcleanv='docker volume prune'
alias dcleanall='docker system prune -a'

# 调试命令
alias dex='docker exec -it'
alias dtop='docker top'
alias dstats='docker stats'
alias dinfo='docker info'

# 进入容器
alias dbash='docker exec -it'
EOF

# 使配置生效
source ~/.bashrc
```

#### 10.7.2 开发环境启动
```bash
# 创建开发环境启动脚本
cat > dev-start.sh <<'EOF'
#!/bin/bash

echo "启动开发环境..."

# 启动基础服务
docker compose -f docker-compose.mysql.yml up -d
docker compose -f docker-compose.redis.yml up -d

# 启动应用
docker compose up -d

echo "开发环境已启动!"
echo "MySQL: localhost:3306"
echo "Redis: localhost:6379"
echo "应用: localhost:3000"
EOF

chmod +x dev-start.sh
```

### 10.8 故障排除

#### 10.8.1 常见问题解决
```bash
# Docker 服务无法启动
sudo systemctl status docker
sudo journalctl -u docker -n 100

# 容器无法连接
docker network inspect app_network
docker network connect app_network container_name

# 磁盘空间不足
docker system df
docker image prune -a
docker builder prune -a

# 内存不足
docker stats
# 增加 WSL 内存限制
```

#### 10.8.2 网络问题排查
```bash
# 检查网络
docker network ls
docker network inspect bridge

# 测试容器连通性
docker exec container_name ping google.com

# 检查 DNS
docker exec container_name cat /etc/resolv.conf

# 端口映射检查
docker port container_name
netstat -tlnp | grep docker
```

---

## 11. 官方文档与资源

### 11.1 官方文档链接
- **Docker**: https://docs.docker.com/
- **Docker Compose**: https://docs.docker.com/compose/
- **WSL**: https://learn.microsoft.com/windows/wsl/
- **Prometheus**: https://prometheus.io/docs/
- **Grafana**: https://grafana.com/docs/

### 11.2 推荐工具
- **Portainer**: Docker Web 管理界面
- **lazydocker**: TUI Docker 管理工具
- **ctop**: 容器资源监控
- **Dozzle**: 容器日志查看器
- **DockStation**: Docker 桌面应用

### 11.3 最佳实践资源
- Docker 官方最佳实践: https://docs.docker.com/develop/develop-images/dockerfile_best-practices/
- Docker 安全: https://docs.docker.com/engine/security/
- Docker 性能: https://docs.docker.com/config/containers/resource_constraints/

---

## 12. 维护与更新

### 12.1 定期维护任务
```bash
# 每周任务
1. docker system df          # 检查磁盘使用
2. docker image prune -a     # 清理未使用的镜像
3. docker container prune    # 清理停止的容器
4. docker volume prune       # 清理未使用的卷

# 每月任务
1. docker builder prune -a   # 清理构建缓存
2. docker system prune -a    # 完全清理
3. 检查更新                 # docker version
```

### 12.2 更新流程
```bash
# 更新 Docker
sudo apt update
sudo apt upgrade docker.io

# 更新 docker-compose
sudo apt upgrade docker-compose

# 或手动更新
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

---

## 13. 快速参考

### 13.1 常用命令速查
```bash
# 启动服务
docker compose up -d

# 停止服务
docker compose down

# 查看日志
docker compose logs -f

# 重启服务
docker compose restart

# 进入容器
docker exec -it container_name bash

# 查看资源使用
docker stats

# 查看容器 IP
docker inspect container_name | grep IPAddress
```

### 13.2 端口速查
```bash
# 默认端口
MySQL:      3306
PostgreSQL: 5432
Redis:      6379
MongoDB:    27017
RabbitMQ:   5672, 15672
Kafka:      9092
Elasticsearch: 9200, 9300
Prometheus: 9090
Grafana:    3000
Kibana:     5601
MinIO:      9000, 9001
```

---

**文档版本**: 1.0  
**最后更新**: 2026-02-04  
**适用系统**: Windows 10/11 + WSL2 + Docker

如有问题，请参考官方文档或联系技术支持。