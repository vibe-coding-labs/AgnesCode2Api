# AgnesCode2Api

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react)](https://react.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue?style=flat)](./LICENSE)

---

## 为什么有这个项目

**Claude Code、Cursor、Windsurf 这些 AI 编程工具只认 Anthropic 和 OpenAI 的 API 协议。** 如果拿到了其他平台的 API 权限（比如 AgnesCode），想用这些工具调用，协议不兼容，接不上。

**AgnesCode2Api 就是做这个的：一个协议翻译层。**

它在中间把 AgnesCode 的 API 协议实时翻译成 Anthropic Messages API 和 OpenAI Chat Completions API 格式。你不需要改工具，不需要写适配代码，改两个环境变量就能用。

```
你用的工具 (Claude Code / Cursor)
       ↓  Anthropic / OpenAI 协议
AgnesCode2Api (协议翻译层)
       ↓  AgnesCode 协议
AgnesCode API (模型服务)
```

---

## 适合谁用

- 有 AgnesCode 账号，想用 Claude Code 编程的人
- 有多账号，需要统一管理 API Key 和用量的人
- 在 macOS 上已安装 AgnesCode IDE，想一键导入凭据的人
- 在 Docker 或远程服务器上部署，需要通过 OAuth 授权登录的人

---

## 核心功能

**协议翻译**
- Anthropic Messages API ↔ AgnesCode API
- OpenAI Chat Completions API ↔ AgnesCode API
- Tool Use 完整映射，不影响 Claude Code 正常使用
- SSE 流式输出，打字机效果

**账号管理**
- OAuth 授权登录，服务端自动兑换 token（CSRF 防护）
- 一键导入已登录的 AgnesCode IDE 凭据
- 支持多账号，每个账号独立 API Key，拖拽排序
- 默认账号自动路由，请求无需指定密钥

**运维能力**
- 会话保活：自动检测并刷新过期凭证，不用手动重新登录
- 智能上下文截断：对话过长时自动截断早期消息，不会卡死
- 日志轮转，防止磁盘写满
- 内置 Dashboard（React 19 + Ant Design + Recharts），数据可视化
- 单文件部署，前端嵌入 Go 二进制

---

## 快速开始

```bash
# 构建
cd web && npm install && npm run build && cd ..
go build -o agnescode_proxy_bin ./cmd/AgnesCodeProxy/

# 启动
./agnescode_proxy_bin serve

# 终端 2：连接到 Claude Code
export ANTHROPIC_BASE_URL=http://localhost:34891
export ANTHROPIC_API_KEY=agnescode
claude
```

### Docker

```bash
docker build -t agnescode-proxy .
docker run -p 34891:34891 agnescode-proxy
```

> 国内构建时如遇 Alpine 源连接失败，换用国内镜像：
> ```bash
> docker build --build-arg ALPINE_MIRROR=https://mirrors.aliyun.com/alpine -t agnescode-proxy .
> ```

---

## 添加账号

打开 `http://localhost:34891` 进入 Dashboard。

| 方式 | 说明 |
|------|------|
| **OAuth 授权登录** | 点击 → 跳转 AgnesCode 登录页 → 授权后自动添加 |
| **一键导入** | 本机已登录 AgnesCode IDE 时自动读取凭据 |
| **手动添加** | 已有 jwt_token 和 user_id 时直接填写 |

**远程部署的 OAuth 回调：** 授权完成后浏览器会跳转到 `http://127.0.0.1:34891/auth/callback`（这是正常的，因为回调地址指向本机）。把地址栏 URL 复制到弹窗输入框提交即可。

---

## API 端点

| 路径 | 协议 |
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

## 界面截图

<div align="center">
<img src="data/imgs/dashboard.png" alt="Dashboard" width="720" />
<p><sub>数据概览</sub></p>
</div>

<div align="center">
<img src="data/imgs/accounts.png" alt="账号管理" width="720" />
<p><sub>账号管理</sub></p>
</div>

<div align="center">
<img src="data/imgs/account-detail.png" alt="账号详情" width="720" />
<p><sub>账号详情</sub></p>
</div>

<div align="center">
<img src="data/imgs/settings.png" alt="系统设置" width="720" />
<p><sub>系统设置</sub></p>
</div>

---

## 使用限制

- 最多 10 个账号
- 使用前请阅读并遵守 AgnesCode 服务条款

---

## 许可证

Apache 2.0

---

> **免责声明：** 本项目仅供个人学习和技术研究使用。禁止用于商业转售、API 中转服务（中转站属于违法行为）、大规模薅号或任何黑灰产/违法违规活动。因不当使用造成的一切后果由使用者自行承担，与项目作者无关。本项目不是 AgnesCode 官方产品。