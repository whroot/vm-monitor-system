#!/bin/bash
# VM监控系统 - WSL2环境优化脚本 (16GB/6核心配置)

echo "========================================"
echo "VM监控系统 - WSL2环境优化"
echo "========================================"
echo

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then 
    echo "[错误] 请使用sudo运行此脚本"
    exit 1
fi

echo "[步骤 1/6] 更新系统包..."
apt update && apt upgrade -y

echo "[步骤 2/6] 安装基础工具..."
apt install -y curl wget git vim htop net-tools build-essential software-properties-common

echo "[步骤 3/6] 安装Node.js 18+..."
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt install -y nodejs
npm install -g pnpm

echo "[步骤 4/6] 安装Docker CLI..."
if ! command -v docker &> /dev/null; then
    echo "Docker CLI未安装，正在安装..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    usermod -aG docker $USER
    echo "✅ Docker CLI已安装"
else
    echo "✅ Docker CLI已安装"
fi

echo "[步骤 5/6] 安装开发工具..."
# 安装其他开发工具
apt install -y python3 python3-pip jq

echo "[步骤 6/6] 配置系统优化..."
# 优化系统参数
cat >> /etc/sysctl.conf << EOF
# 网络优化
net.core.default_qdisc=fq
net.ipv4.tcp_fastopen=3

# 内存优化
vm.swappiness=10
vm.vfs_cache_pressure=50
EOF

# 应用系统优化
sysctl -p

echo "========================================"
echo "✅ WSL2环境优化完成！"
echo "========================================"
echo
echo "已安装的组件:"
echo "- Node.js: $(node --version)"
echo "- npm: $(npm --version)"
echo "- pnpm: $(pnpm --version)"
echo "- Docker: $(docker --version)"
echo
echo "下一步:"
echo "1. 配置Docker Desktop资源"
echo "2. 克隆项目并安装依赖"
echo "3. 启动开发环境"
echo
echo "当前资源使用:"
free -h
nproc
