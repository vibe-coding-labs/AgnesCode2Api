# AgnesCode2Api — 安装引导

> 本文档描述如何自动完成 AgnesCode2Api 的安装和配置。适用于手动安装，也可以将链接发给 AI 编程助手让它帮你搭。

## 任务概述

用户需要你帮忙部署 AgnesCode2Api 代理服务。这是一个协议翻译层，把 AgnesCode 的 API 翻译成 Anthropic/OpenAI 兼容格式，让 Claude Code 等工具可以直接使用。

你需要做的是：**检测操作系统 → 选择安装方式 → 启动服务 → 配置环境变量 → 验证服务正常。**

## 第一步：检测系统环境

```bash
# 检测操作系统类型
uname -s
# → Linux 或 Darwin

# 检测架构
uname -m
# → x86_64 或 arm64

# 检测是否已安装 Docker
which docker 2>/dev/null && echo "DOCKER_AVAILABLE" || echo "NO_DOCKER"

# 检测是否已安装 Node.js
which node 2>/dev/null && node --version || echo "NO_NODE"

# 检测是否已安装 Go
which go 2>/dev/null && go version || echo "NO_GO"
```

## 第二步：选择安装方式（按优先级）

### 方式 A：Docker 安装（推荐，最省事）

**前提：** 系统有 Docker。

```bash
docker run -d \
  --name agnescode-proxy \
  --restart unless-stopped \
  -p 34891:34891 \
  ghcr.io/vibe-coding-labs/agnescode-proxy:latest

# 等待启动
echo "Waiting for service..."
for i in $(seq 1 15); do
  if curl -sf http://localhost:34891/health > /dev/null 2>&1; then
    echo "Service ready"
    break
  fi
  sleep 1
done
```

### 方式 B：下载二进制文件

**前提：** 系统没有 Docker，但有 curl。

```bash
# 确定下载地址
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

if [ "$OS" = "darwin" ] && [ "$ARCH" = "arm64" ]; then
  BINARY="AgnesCodeProxy-darwin-arm64"
elif [ "$OS" = "linux" ] && [ "$ARCH" = "x86_64" ]; then
  BINARY="AgnesCodeProxy-linux-amd64"
else
  echo "Unsupported platform: $OS $ARCH"
  echo "Falling back to source build..."
  # 走方式 C
fi

# 如果二进制名已确定，下载并安装
if [ -n "$BINARY" ]; then
  DOWNLOAD_DIR="$HOME/AgnesCodeProxy"
  mkdir -p "$DOWNLOAD_DIR"
  
  # 获取最新 Release 的下载链接
  RELEASE_URL=$(curl -s https://api.github.com/repos/vibe-coding-labs/AgnesCode2Api/releases/latest \
    | grep "browser_download_url.*$BINARY" \
    | cut -d '"' -f 4)
  
  echo "Downloading from: $RELEASE_URL"
  curl -L -o "$DOWNLOAD_DIR/$BINARY" "$RELEASE_URL"
  chmod +x "$DOWNLOAD_DIR/$BINARY"
  
  # 启动服务
  "$DOWNLOAD_DIR/$BINARY" serve --tls=false --skip-validation &
  SERVER_PID=$!
  
  echo "Waiting for service..."
  for i in $(seq 1 15); do
    if curl -sf http://localhost:34891/health > /dev/null 2>&1; then
      echo "Service ready (PID: $SERVER_PID)"
      break
    fi
    sleep 1
  done
fi
```

### 方式 C：源码编译

**前提：** 系统没有 Docker，且平台不提供预编译二进制（如 linux-arm64）。

```bash
# 需要 Go 1.22+ 和 Node.js 18+
# 先编译前端
cd /tmp
git clone https://github.com/vibe-coding-labs/AgnesCode2Api.git
cd AgnesCode2Api/web
npm install
npm run build
cd ..

# 编译后端
go build -o /tmp/agnescode_proxy_bin ./cmd/AgnesCodeProxy/

# 启动服务
/tmp/agnescode_proxy_bin serve --tls=false --skip-validation &
SERVER_PID=$!

echo "Waiting for service..."
for i in $(seq 1 15); do
  if curl -sf http://localhost:34891/health > /dev/null 2>&1; then
    echo "Service ready (PID: $SERVER_PID)"
    break
  fi
  sleep 1
done
```

## 第三步：验证服务

```bash
# 健康检查
curl -s http://localhost:34891/health
# 期望返回: {"endpoints":["/v1/chat/completions","/v1/models","/v1/web-search","/v1/rerank"],"service":"agnescode-openai-proxy","status":"ok"}

# 检查可用模型
curl -s http://localhost:34891/v1/models
# 应该返回模型列表（可能报 jwt 过期，正常，因为还没有账号）
```

## 第四步：配置环境变量

服务启动后，告诉用户设置以下环境变量即可使用：

```bash
export ANTHROPIC_BASE_URL=http://localhost:34891
export ANTHROPIC_API_KEY=agnescode
```

然后执行 `claude` 即可连接。

## 第五步：Dashboard 添加账号

如果用户需要添加账号，引导他们打开浏览器访问 `http://localhost:34891`，通过 OAuth 授权登录添加账号。

详情见 [README 的账号管理章节](https://github.com/vibe-coding-labs/AgnesCode2Api/#添加账号)。

## 常见问题

### 端口被占用

```bash
# 检查端口占用
lsof -ti:34891 | xargs -r kill
```

### 国内网络问题

Docker 拉取慢时，可以配置镜像加速器，或直接从 GitHub Releases 下载二进制文件。

### 服务挂了

```bash
# 查看日志
cat ~/.agnescode-proxy/logs/*.log 2>/dev/null
# 重启即可
```