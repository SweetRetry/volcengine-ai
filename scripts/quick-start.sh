#!/bin/bash

# 🚀 Volcengine AI Server 快速启动脚本
# 使用方法：./scripts/quick-start.sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 显示欢迎信息
echo -e "${BLUE}"
echo "🚀 Volcengine AI Server 快速启动"
echo "=================================="
echo -e "${NC}"

# 检查Docker和Docker Compose
check_dependencies() {
    echo -e "${YELLOW}🔍 检查依赖...${NC}"
    
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker未安装，请先安装Docker${NC}"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose未安装，请先安装Docker Compose${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 依赖检查通过${NC}"
}

# 检查环境变量文件
check_env_file() {
    echo -e "${YELLOW}🔧 检查环境配置...${NC}"
    
    if [ ! -f ".env" ]; then
        echo -e "${YELLOW}⚠️  .env文件不存在，正在创建示例配置...${NC}"
        
        cat > .env << EOF
# 服务器配置
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info

# 数据库配置
MONGO_URL=mongodb://mongodb:27017/volcengine_db

# Redis配置
REDIS_URL=redis://redis:6379

# 火山方舟API密钥（请填写您的密钥）
ARK_API_KEY=your_ark_api_key_here

# AI服务超时配置
AI_TIMEOUT=30s
EOF
        
        echo -e "${YELLOW}📝 请编辑 .env 文件，填写您的火山方舟API密钥${NC}"
        echo -e "${YELLOW}   获取密钥: https://console.volcengine.com/ark${NC}"
        echo -e "${YELLOW}   配置ARK_API_KEY变量${NC}"
        
        read -p "是否现在编辑 .env 文件? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            ${EDITOR:-nano} .env
        fi
    else
        echo -e "${GREEN}✅ 环境配置文件存在${NC}"
    fi
}

# 构建和启动服务
start_services() {
    echo -e "${BLUE}🏗️  构建和启动服务...${NC}"
    
    # 停止现有服务
    docker-compose down 2>/dev/null || true
    
    # 构建镜像
    echo -e "${YELLOW}📦 构建Docker镜像...${NC}"
    docker-compose build
    
    # 启动核心服务
    echo -e "${YELLOW}🚀 启动核心服务...${NC}"
    docker-compose up -d mongodb redis
    
    # 等待数据库启动
    echo -e "${YELLOW}⏳ 等待数据库启动...${NC}"
    sleep 10
    
    # 启动应用服务
    echo -e "${YELLOW}🎯 启动应用服务...${NC}"
    docker-compose up -d api-server queue-worker
    
    echo -e "${GREEN}✅ 服务启动完成！${NC}"
}

# 显示服务状态
show_status() {
    echo -e "${BLUE}📊 服务状态:${NC}"
    docker-compose ps
    
    echo -e "\n${BLUE}🔗 服务地址:${NC}"
    echo -e "  🌐 API服务器: ${CYAN}http://localhost:8080${NC}"
    echo -e "  📊 Redis管理: ${CYAN}http://localhost:8081${NC} (可选)"
    echo -e "  📈 监控面板: ${CYAN}http://localhost:3000${NC} (可选)"
}

# 运行健康检查
health_check() {
    echo -e "${YELLOW}🏥 运行健康检查...${NC}"
    
    # 等待服务完全启动
    sleep 15
    
    # 检查API服务器
    if curl -s http://localhost:8080/health > /dev/null; then
        echo -e "${GREEN}✅ API服务器健康${NC}"
    else
        echo -e "${RED}❌ API服务器不健康${NC}"
    fi
    
    # 检查Redis
    if docker-compose exec -T redis redis-cli ping | grep -q PONG; then
        echo -e "${GREEN}✅ Redis服务健康${NC}"
    else
        echo -e "${RED}❌ Redis服务不健康${NC}"
    fi
    
    # 检查MongoDB
    if docker-compose exec -T mongodb mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ MongoDB服务健康${NC}"
    else
        echo -e "${RED}❌ MongoDB服务不健康${NC}"
    fi
}

# 显示使用示例
show_examples() {
    echo -e "\n${PURPLE}🎯 使用示例:${NC}"
    
    echo -e "\n${YELLOW}1. 创建图像生成任务:${NC}"
    cat << 'EOF'
curl -X POST http://localhost:8080/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪在花园里玩耍",
    "user_id": "user123",
    "size": "1024x1024"
  }'
EOF

    echo -e "\n${YELLOW}2. 查询任务状态:${NC}"
    echo "curl http://localhost:8080/ai/image/result/{task_id}"
    
    echo -e "\n${YELLOW}3. 查看日志:${NC}"
    echo "docker-compose logs -f api-server"
    echo "docker-compose logs -f queue-worker"
    
    echo -e "\n${YELLOW}4. 停止服务:${NC}"
    echo "docker-compose down"
    
    echo -e "\n${YELLOW}5. 启用监控:${NC}"
    echo "docker-compose --profile monitoring up -d"
}

# 主函数
main() {
    check_dependencies
    check_env_file
    start_services
    show_status
    health_check
    show_examples
    
    echo -e "\n${GREEN}🎉 Volcengine AI Server 启动完成！${NC}"
    echo -e "${CYAN}📖 查看完整文档: README.md${NC}"
}

# 运行主函数
main "$@" 