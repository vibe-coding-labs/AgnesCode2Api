<div align="center">

# AgnesCode2Api

**零成本用上 Claude Code — 不需要 Anthropic 付费账号**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react)](https://react.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue?style=flat)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/vibe-coding-labs/AgnesCode2Api)](https://github.com/vibe-coding-labs/AgnesCode2Api/releases)

</div>

---

## 30 秒上手

复制下面这段提示词，粘贴给你的 Claude Code，它会自己搭好一切：

```text
我需要你帮我启动一个 AgnesCode2Api 代理服务，让 Claude Code 可以通过它调用 AgnesCode 的 AI 模型。

请按照以下步骤操作：

1. 如果本机没有安装这个服务，先访问 https://github.com/vibe-coding-labs/AgnesCode2Api/releases 下载最新版的 Linux 二进制文件，保存到 ~/AgnesCodeProxy/ 目录下
2. 赋予执行权限：chmod +x ~/AgnesCodeProxy/AgnesCodeProxy-linux-amd64
3. 启动服务：~/AgnesCodeProxy/AgnesCodeProxy-linux-amd64 serve --tls=false --skip-validation
4. 等待服务启动完成，确认 http://localhost:34891/health 返回 200
5. 给我设置环境变量 ANTHROPIC_BASE_URL=http://localhost:34891 和 ANTHROPIC_API_KEY=agnescode
6. 确认代理正常工作后告知我

注意：如果 ~/AgnesCodeProxy/ 目录下已经有二进制文件，直接从第 2 步开始。
```

**不需要 Anthropic 付费账号，不需要 OpenAI 订阅。有 AgnesCode 权限就能跑 Claude Code。**

---

## 这是什么

AgnesCode2Api 是一个**协议翻译代理**。它把 AgnesCode 的 API 协议翻译成 Anthropic / OpenAI 兼容格式，让 Claude Code、Cursor、Windsurf 这些工具可以直接调用。

```
你习惯的工具 ──→ AgnesCode2Api ──→ AgnesCode 的模型
(Claude Code / Cursor)   ↓                (GLM / Kimi / MiniMax / Doubao…)
                    Anthropic / OpenAI 协议
```

**你不需要改工具配置以外的任何代码。改两个环境变量，就能用 Claude Code 写出 Agent 能力。**

---

## 为什么用这个

| 场景 | 以前 | 现在 |
|------|------|------|
| 想用 Claude Code 但没 Anthropic 账号 | 付费订阅 Anthropic | 用已有 AgnesCode 权限直接跑 |
| 多个账号需要管理 | 人工切 Key，记不清哪个过期 | Dashboard 统一管理，自动保活 |
| 部署到远程服务器 | 复杂配置，依赖多 | 单文件部署，一条命令启动 |
| 想用 Cursor 接不同模型 | 协议不兼容，接不上 | 自动翻译，即配即用 |

---

## 核心功能

**🔀 协议翻译**
- 同时兼容 Anthropic Messages API 和 OpenAI Chat Completions API
- Claude Code → `/v1/messages`，Cursor → `/v1/chat/completions`
- Tool Use 完整映射，不影响 Agent 工具调用
- SSE 流式输出，打字机效果

**🔑 账号管理**
- **OAuth 授权登录**：从 Dashboard 跳转 AgnesCode 登录页，授权后自动添加
- **一键导入**：本机已登录 AgnesCode IDE 时自动读取凭据
- **多账号**：每个账号独立 API Key，可拖拽排序，默认账号自动路由
- **会话保活**：自动检测并刷新过期凭证，不用手动重新登录

**🛡️ 生产就绪**
- 智能上下文截断：对话过长时自动截断早期消息，不会卡死
- 日志轮转：防止磁盘写满
- 内置 Dashboard（React 19 + Ant Design）：数据可视化、账号管理、系统设置
- 单文件部署：前端嵌入 Go 二进制，一个文件就能跑

---

## 快速开始

### 下载安装

从 [Releases](https://github.com/vibe-coding-labs/AgnesCode2Api/releases) 下载对应平台的二进制文件：

```bash
# macOS
chmod +x AgnesCodeProxy-darwin-arm64
./AgnesCodeProxy-darwin-arm64 serve

# Linux
chmod +x AgnesCodeProxy-linux-amd64
./AgnesCodeProxy-linux-amd64 serve
```

### Docker

```bash
docker run -p 34891:34891 ghcr.io/vibe-coding-labs/agnescode-proxy
```

### 或者让 AI 自动安装

把 [docs/guides/INSTALL.md](docs/guides/INSTALL.md) 的链接发给 Claude Code，它会根据你的系统环境自动选择安装方式。

```bash
cd web && npm install && npm run build && cd ..
go build -o agnescode_proxy_bin ./cmd/AgnesCodeProxy/
./agnescode_proxy_bin serve
```

### Docker

```bash
docker build -t agnescode-proxy .
docker run -p 34891:34891 agnescode-proxy
```

> 国内构建时若 Alpine 源连接失败，加构建参数：
> ```bash
> docker build --build-arg ALPINE_MIRROR=https://mirrors.aliyun.com/alpine -t agnescode-proxy .
> ```

### 连接 Claude Code

```bash
export ANTHROPIC_BASE_URL=http://localhost:34891
export ANTHROPIC_API_KEY=agnescode
claude
```

---

## 添加账号

打开 `http://localhost:34891` 进入 Dashboard。

| 方式 | 操作 |
|------|------|
| **OAuth 授权登录** | 点击按钮 → 跳转 AgnesCode 登录页 → 授权后自动添加 |
| **一键导入** | 本机已安装 AgnesCode IDE 时自动读取凭据 |
| **手动添加** | 已有 jwt_token 和 user_id 时直接填写 |

> 远程部署时 OAuth 授权完成后，浏览器会跳转到 `http://127.0.0.1:34891/auth/callback`（这是正常的，因为回调地址指向本机）。把地址栏 URL 复制到弹窗输入框提交即可。

---

## API 端点

| 路径 | 说明 |
|------|------|
| `POST /v1/messages` | Anthropic Messages API（Claude Code 用） |
| `POST /v1/chat/completions` | OpenAI Chat Completions API（Cursor 用） |
| `POST /v1/web-search` | 网页搜索 |
| `POST /v1/rerank` | 文档重排序 |
| `GET /v1/models` | 模型列表 |
| `GET /health` | 健康检查 |
| `GET /` | Dashboard |

---

## 界面截图

<div align="center">
<img src="data/imgs/dashboard.png" alt="Dashboard" width="720" />
<p><sub>数据概览 — 请求量、Token 消耗、延迟统计、模型分布</sub></p>
</div>

<div align="center">
<img src="data/imgs/accounts.png" alt="账号管理" width="720" />
<p><sub>账号管理 — 多账号、OAuth 授权登录、拖拽排序</sub></p>
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

## 项目结构

```
cmd/AgnesCodeProxy/    入口，HTTP 服务器，CLI
pkg/anthropic/         Anthropic 协议翻译
pkg/openai/            OpenAI 协议翻译
pkg/agnes/             AgnesCode API 客户端
pkg/auth/              凭据读取、JWT、中间件
pkg/store/             SQLite 存储
pkg/dashboard/         Dashboard API + 静态文件
pkg/keepalive/         会话保活
pkg/logrot/            日志轮转
pkg/proxy/             会话管理
web/                   前端（React 19 + Ant Design）
```

---

## 使用限制

- 最多 10 个账号
- 使用前请阅读并遵守 AgnesCode 服务条款

---

## 许可证

Apache 2.0

---

> **免责声明：** 本项目仅供个人学习和技术研究使用。禁止用于商业转售、API 中转服务（中转站属于违法行为）、大规模薅号或任何黑灰产/违法违规活动。因不当使用造成的一切后果由使用者自行承担，与项目作者无关。本项目不是 AgnesCode 官方产品。