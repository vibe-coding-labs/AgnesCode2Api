# AgnesCodeProxy

> 把 AgnesCode 的 API 协议翻译成 Anthropic / OpenAI 兼容格式。
> 让 Claude Code、Cursor 这些工具直接接上 AgnesCode 的模型。

```
Claude Code / Cursor / Windsurf  →  AgnesCodeProxy  →  AgnesCode API
                                    ↓
                          Anthropic / OpenAI 协议
```

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react)](https://react.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue?style=flat)](./LICENSE)

---

## 它能做什么

**一句话：协议翻译器。**

AgnesCode 的模型能力不错，但它的 API 协议跟 Anthropic 和 OpenAI 不兼容，Claude Code、Cursor 这些工具接不上。AgnesCodeProxy 在中间做了一层翻译，把它们串起来。

**支持模型：** JoyAI-Code · GLM-5.1 · GLM-5 · GLM-4.7 · Kimi-K2.6 · Kimi-K2.5 · MiniMax-M2.7 · Doubao-Seed-2.0-pro

**核心能力：**

- **双协议翻译** — 同时兼容 Anthropic Messages API 和 OpenAI Chat Completions API，Claude Code 和 Cursor 各走各的通道
- **Tool Use 映射** — 工具调用（读写文件、执行命令等）完整翻译，不影响 Claude Code 的正常使用
- **SSE 流式输出** — 实时流式返回，打字机效果，和原生体验一样
- **多账号管理** — 通过 OAuth 授权登录或一键导入添加多个账号，每个账号有独立的 API Key，支持拖拽排序
- **OAuth 授权登录** — 从 Dashboard 跳转 AgnesCode 登录页完成授权，服务端自动兑换 token，CSRF 防护
- **会话保活** — 自动检测并刷新过期凭证，不需要手动重新登录
- **智能上下文截断** — 对话过长时自动截断早期消息，不会卡死，`/compact` 正常工作
- **单文件部署** — 前端打包进 Go 二进制，丢一个文件就能跑

---

## 快速开始

### 构建

```bash
# 前端构建
cd web && npm install && npm run build && cd ..

# 后端构建（前端自动嵌入）
go build -o agnescode_proxy_bin ./cmd/AgnesCodeProxy/
```

### Docker

```bash
docker build -t agnescode-proxy .
docker run -p 34891:34891 agnescode-proxy
```

> 国内构建 Docker 时如遇 Alpine 源连接失败，用国内镜像源：
> ```bash
> docker build --build-arg ALPINE_MIRROR=https://mirrors.aliyun.com/alpine -t agnescode-proxy .
> ```

### 启动

```bash
./agnescode_proxy_bin serve
```

默认监听 `0.0.0.0:34891`。macOS 首次启动会自动读取本地 AgnesCode 客户端凭据。

### 连接到 Claude Code

```bash
export ANTHROPIC_BASE_URL=http://localhost:34891
export ANTHROPIC_API_KEY=agnescode
claude
```

---

## 账号管理

打开 `http://localhost:34891` 进入 Dashboard。

### 三种登录方式

| 方式 | 适用场景 | 说明 |
|------|---------|------|
| **OAuth 授权登录** | 所有场景（推荐） | 点「OAuth授权登录」→ 浏览器跳转 AgnesCode 登录页 → 完成授权后自动添加账号 |
| **一键导入** | macOS 本机已安装 AgnesCode IDE | 自动读取本地登录凭据，无需手动操作 |
| **手动添加** | 已有 jwt_token 和 user_id | 在弹窗中直接填写 |

### 远程 / Docker 部署时的 OAuth 授权

点击「OAuth授权登录」后，浏览器会跳转到 AgnesCode 登录页。完成授权后，浏览器会尝试跳转到 `http://127.0.0.1:34891/auth/callback`——在远程部署时这个地址不可达，这是正常的。**把地址栏中的完整 URL 复制下来，粘贴到弹窗的输入框，点击「提交授权」即可。**

Docker 部署时如需使用「一键导入」，需挂载 AgnesCode IDE 的状态文件：

```bash
docker run -p 34891:34891 \
  -e AGNESCODE_STATE_DB=/data/state.vscdb \
  -v /path/to/AgnesCode/state.vscdb:/data/state.vscdb:ro \
  agnescode-proxy --skip-validation serve
```

---

## 界面截图

<div align="center">
<img src="data/imgs/dashboard.png" alt="Dashboard" width="720" />
<p><sub>数据概览 — 请求量、Token 消耗、延迟统计、模型分布</sub></p>
</div>

<div align="center">
<img src="data/imgs/accounts.png" alt="账号管理" width="720" />
<p><sub>账号管理 — 多账号管理、OAuth 授权登录、拖拽排序</sub></p>
</div>

<div align="center">
<img src="data/imgs/account-detail.png" alt="账号详情" width="720" />
<p><sub>账号详情 — 用量、模型分布、请求记录</sub></p>
</div>

<div align="center">
<img src="data/imgs/settings.png" alt="系统设置" width="720" />
<p><sub>系统设置 — 默认模型、超时时间、日志保留</sub></p>
</div>

---

## API 端点

| 路径 | 说明 |
|------|------|
| `POST /v1/messages` | Anthropic Messages API |
| `POST /v1/chat/completions` | OpenAI Chat Completions API |
| `POST /v1/web-search` | 网页搜索 |
| `POST /v1/rerank` | 文档重排序 |
| `GET /v1/models` | 模型列表 |
| `GET /health` | 健康检查 |
| `GET /` | Dashboard |

---

## 项目结构

```
cmd/AgnesCodeProxy/    入口，HTTP 服务器，CLI 命令
pkg/anthropic/         Anthropic 协议翻译
pkg/openai/            OpenAI 协议翻译
pkg/agnes/             AgnesCode API 客户端
pkg/auth/              凭据读取、JWT 认证、中间件
pkg/store/             SQLite 存储层
pkg/dashboard/         Dashboard API + 静态文件服务
pkg/keepalive/         会话保活、凭证刷新
pkg/logrot/            日志轮转
pkg/proxy/             会话管理
web/                   前端（React 19 + Ant Design + Recharts）
```

---

## 使用限制

- 每个用户最多配置 **10 个账号**
- 使用前请确保你已了解并遵守 AgnesCode 的服务条款

---

## 许可证

Apache 2.0

---

> **免责声明：** 本项目仅供个人学习和技术研究使用。禁止用于商业转售、API 中转服务（中转站属于违法行为）、大规模薅号或任何黑灰产/违法违规活动。因不当使用造成的一切后果由使用者自行承担，与项目作者无关。本项目不是 AgnesCode 官方产品。