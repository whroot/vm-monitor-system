@echo off
REM Windows PC - WSL2配置脚本 (16GB/6核心配置)

echo ========================================
echo VM监控系统 - WSL2优化配置
echo ========================================
echo.

REM 检查管理员权限
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [错误] 请以管理员身份运行此脚本
    pause
    exit /b 1
)

echo [步骤 1/5] 创建WSL2配置文件...

REM 创建 .wslconfig 文件
echo [wsl2] > "%USERPROFILE%\.wslconfig"
echo memory=10GB >> "%USERPROFILE%\.wslconfig"
echo processors=5 >> "%USERPROFILE%\.wslconfig"
echo swap=4GB >> "%USERPROFILE%\.wslconfig"
echo localhostForwarding=true >> "%USERPROFILE%\.wslconfig"

echo [步骤 2/5] WSL2配置文件已创建
echo 配置内容:
type "%USERPROFILE%\.wslconfig"
echo.

echo [步骤 3/5] 重启WSL2以应用配置...
wsl --shutdown

echo [步骤 4/5] 等待5秒后启动WSL2...
timeout /t 5 /nobreak

echo [步骤 5/5] 重新启动WSL2...
wsl

echo.
echo ========================================
echo ✅ WSL2配置完成！
echo ========================================
echo.
echo 配置详情:
echo - WSL2内存: 10GB
echo - WSL2处理器: 5核心
echo - 交换空间: 4GB
echo.
echo 下一步:
echo 1. 配置Docker Desktop资源
echo 2. 安装开发环境依赖
echo.
pause