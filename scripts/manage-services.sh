#!/bin/bash
# VM监控系统 - 16GB内存环境服务管理脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_ROOT"

echo "========================================"
echo "VM监控系统 - 服务管理 (16GB优化)"
echo "========================================"
echo

# 显示当前资源使用情况
function show_resources() {
    echo -e "${YELLOW}当前资源使用情况:${NC}"
    echo "内存使用:"
    free -h | grep "Mem:"
    echo
    echo "CPU使用:"
    top -bn1 | grep "Cpu(s)"
    echo
    echo "Docker资源使用:"
    if command -v docker &> /dev/null; then
        docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" || true
    else
        echo "Docker未运行"
    fi
    echo
}

# 启动基础开发环境（最轻量）
function start_basic() {
    echo -e "${GREEN}启动基础开发环境...${NC}"
    docker-compose -f docker-compose.optimized.yml up -d postgresql redis nginx
    echo "✅ 基础开发环境已启动"
    echo "  - PostgreSQL: localhost:5432"
    echo "  - Redis: localhost:6379"
    echo "  - Nginx: http://localhost"
}

# 启动前端开发环境
function start_frontend() {
    echo -e "${GREEN}启动前端开发环境...${NC}"
    docker-compose -f docker-compose.optimized.yml up -d postgresql redis
    echo "✅ 前端开发环境已启动"
    echo "  - 数据库服务已启动"
    echo "  请手动启动前端: cd frontend && npm run dev"
}

# 启动后端开发环境
function start_backend() {
    echo -e "${GREEN}启动后端开发环境...${NC}"
    docker-compose -f docker-compose.optimized.yml up -d postgresql redis rabbitmq influxdb
    echo "✅ 后端开发环境已启动"
    echo "  - PostgreSQL: localhost:5432"
    echo "  - Redis: localhost:6379"
    echo "  - RabbitMQ: localhost:5672 (管理: http://localhost:15672)"
    echo "  - InfluxDB: localhost:8086"
    echo "  请手动启动后端: cd backend && npm run dev"
}

# 启动完整测试环境（中等负载）
function start_full() {
    echo -e "${GREEN}启动完整测试环境...${NC}"
    echo -e "${YELLOW}注意: 完整环境需要较多内存，可能影响系统性能${NC}"
    docker-compose -f docker-compose.optimized.yml --profile full up -d
    echo "✅ 完整测试环境已启动"
    echo "  所有服务已启动，包括:"
    echo "  - 数据库服务"
    echo "  - 中间件服务"
    echo "  - 监控服务 (Prometheus: http://localhost:9090)"
    echo "  - 可视化服务 (Grafana: http://localhost:3000)"
    echo "  - ClickHouse: localhost:8123"
}

# 停止所有服务
function stop_all() {
    echo -e "${YELLOW}停止所有服务...${NC}"
    docker-compose -f docker-compose.optimized.yml down
    echo "✅ 所有服务已停止"
}

# 停止监控服务（节省内存）
function stop_monitoring() {
    echo -e "${YELLOW}停止监控服务...${NC}"
    docker-compose -f docker-compose.optimized.yml stop prometheus grafana
    echo "✅ 监控服务已停止"
}

# 重启服务
function restart_service() {
    if [ -z "$1" ]; then
        echo -e "${RED}请指定要重启的服务${NC}"
        exit 1
    fi
    echo -e "${GREEN}重启服务: $1...${NC}"
    docker-compose -f docker-compose.optimized.yml restart "$1"
    echo "✅ 服务 $1 已重启"
}

# 显示服务状态
function show_status() {
    echo -e "${GREEN}服务状态:${NC}"
    docker-compose -f docker-compose.optimized.yml ps
}

# 查看日志
function show_logs() {
    if [ -z "$1" ]; then
        echo -e "${RED}请指定要查看日志的服务${NC}"
        exit 1
    fi
    docker-compose -f docker-compose.optimized.yml logs -f "$1"
}

# 清理未使用的资源
function cleanup() {
    echo -e "${YELLOW}清理未使用的Docker资源...${NC}"
    docker system prune -f
    echo "✅ 清理完成"
}

# 显示帮助信息
function show_help() {
    cat << EOF
VM监控系统 - 服务管理脚本

用法: $0 [选项]

选项:
    basic              启动基础开发环境 (PostgreSQL + Redis + Nginx)
    frontend           启动前端开发环境
    backend            启动后端开发环境
    full               启动完整测试环境 (所有服务)
    stop               停止所有服务
    stop-monitoring     停止监控服务 (节省内存)
    restart <service>  重启指定服务
    status             显示服务状态
    logs <service>     查看指定服务日志
    resources          显示当前资源使用情况
    cleanup            清理未使用的Docker资源
    help               显示此帮助信息

示例:
    $0 basic          # 启动基础开发环境
    $0 backend        # 启动后端开发环境
    $0 restart redis  # 重启Redis服务
    $0 logs postgres  # 查看PostgreSQL日志
    $0 resources      # 查看资源使用情况

注意:
    - 16GB内存环境建议按需启动服务
    - 基础环境约需3-4GB内存
    - 完整环境约需8-10GB内存
    - 定期使用 resources 命令监控内存使用
EOF
}

# 主逻辑
case "$1" in
    basic)
        show_resources
        start_basic
        ;;
    frontend)
        show_resources
        start_frontend
        ;;
    backend)
        show_resources
        start_backend
        ;;
    full)
        show_resources
        start_full
        ;;
    stop)
        stop_all
        ;;
    stop-monitoring)
        stop_monitoring
        ;;
    restart)
        restart_service "$2"
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "$2"
        ;;
    resources)
        show_resources
        ;;
    cleanup)
        cleanup
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}未知选项: $1${NC}"
        echo
        show_help
        exit 1
        ;;
esac

echo
echo "========================================"
echo "操作完成"
echo "========================================"