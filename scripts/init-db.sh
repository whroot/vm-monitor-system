#!/bin/bash
# VM监控系统 - 数据库初始化脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}开始初始化数据库...${NC}"

# PostgreSQL配置
export PGHOST=${PGHOST:-localhost}
export PGPORT=${PGPORT:-5432}
export PGUSER=${PGUSER:-postgres}
export PGPASSWORD=${PGPASSWORD:-postgres}

# 创建数据库和用户
echo -e "${YELLOW}创建数据库和用户...${NC}"

# 创建用户
psql -h $PGHOST -p $PGPORT -U $PGUSER << EOF
-- 创建应用用户
DO \$\$ 
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'admin') THEN
        CREATE USER admin WITH PASSWORD 'password';
    END IF;
END
\$\$;

-- 创建数据库
SELECT 'CREATE DATABASE vm_monitor OWNER admin' 
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'vm_monitor')\gexec

\q
EOF

echo -e "${GREEN}PostgreSQL初始化完成${NC}"

# Redis配置
echo -e "${YELLOW}配置Redis...${NC}"

redis-cli CONFIG SET maxmemory 128mb
redis-cli CONFIG SET maxmemory-policy allkeys-lru

echo -e "${GREEN}Redis配置完成${NC}"

# 创建配置文件
cat > .env << EOF
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=password
DB_NAME=vm_monitor
DB_SSLMODE=disable

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0
REDIS_PASSWORD=

# InfluxDB配置
INFLUXDB_URL=http://localhost:8086
INFLUXDB_TOKEN=your-influxdb-token
INFLUXDB_ORG=vm-monitor
INFLUXDB_BUCKET=metrics

# RabbitMQ配置
RABBITMQ_URL=amqp://admin:password@localhost:5672/

# 应用配置
API_PORT=8080
NODE_ENV=development
EOF

echo -e "${GREEN}配置文件 .env 已创建${NC}"
echo -e "${GREEN}初始化完成！${NC}"
