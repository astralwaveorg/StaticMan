# StaticMan

[![Deploy](https://github.com/astralwaveorg/StaticMan/actions/workflows/deploy.yml/badge.svg)](https://github.com/astralwaveorg/StaticMan/actions/workflows/deploy.yml)

私人配置文件管理平台 — 浏览、搜索、保护你的配置文件。

## 功能

- 🗂️ **分类浏览** — 按类型组织配置（代理 / Vim / Git / Shell…），Web 界面一键查看
- 🔒 **密码保护** — 登录解锁受保护文件，公开文件无需登录
- 🔑 **auth_key 认证** — 受保护文件通过 `?key=<auth_key>` 直接访问，密钥用户自定义
- 🔍 **全文搜索** — 文件名 + 内容搜索，受保护内容自动脱敏
- 📋 **一键复制** — 登录后复制 Raw 链接自动附 `?key=`
- 🎨 **语法高亮** — YAML、INI、JSON、Shell、Vim 等格式自动高亮
- 🔗 **旧 URL 兼容** — Surge managed URL 和 Mihomo cron 拉取继续工作
- ♻️ **热加载** — 配置文件变更 30 秒内自动生效，无需重启

## 快速开始

```bash
git clone https://github.com/astralwaveorg/StaticMan.git && cd StaticMan

# 创建环境变量文件
cat > .env <<EOF
PORT=8080
DATA_DIR=./data
EOF

# 前端构建 + 后端编译
cd web && npm install && npm run build && cd ..
go build -tags withweb -o staticman ./cmd/server
./staticman
```

或开发模式（前后端分离）：

```bash
# 后端
go run ./cmd/server/

# 前端（自动代理 /api → localhost:8080）
cd web && npm install && npm run dev
```

访问 `http://localhost:8080`。

## 保护模型

### 两级策略

| 策略 | 行为 | 适用场景 |
|------|------|---------|
| **hide** | 完全不可见，直接访问返回 404 | `.git`、临时文件、配置文件自身 |
| **protect** | 可见（带🔒），内容需认证 | 私密配置、密钥文件 |

### 认证方式

登录后获得 `auth_key`（来自 `password.yaml`），用于所有后续请求：

- **API 请求**：`Authorization: Bearer <auth_key>`
- **原始文件**：`/raw/<path>?key=<auth_key>`

未登录时，受保护文件的操作按钮（Raw 链接、下载、复制）自动隐藏。

### 规则引擎

支持全局配置（`password.yaml`）和目录级配置（`.encrypt` 文件），规则语法支持 Glob 和正则：

```yaml
# password.yaml
rules:
  hide:
    - ".git/"
    - "*.tmp"
  protect:
    - "*/private/*"
    - "*.key"
    - "regex:.*credential.*"
```

## 路由架构

| 层级 | 路径 | 用途 |
|------|------|------|
| API 层 | `/api/*` | Web UI 专用，JSON 响应，受保护文件未认证返回脱敏内容 |
| 原始文件层 | `/raw/*`, `/<category>/*` | 短 URL，浏览器可直接预览 |
| 兼容层 | `/d/*` | 旧 URL 重写，直接放行（机器客户端） |

## 配置

### `password.yaml`（数据目录）

```yaml
# 登录密码
password: "your-password"

# 认证密钥（用于 API 鉴权和 URL 访问）
auth_key: "your-auth-key"

# 受保护路径
protected:
  - path: "Surge/macOS.conf"

# 规则引擎
rules:
  hide: [".git/", "*.tmp"]
  protect: ["*.key"]
```

### `metadata.yaml`（数据目录）

```yaml
categories:
  Surge:
    name: "Surge"
    icon: "shield"
    description: "Surge 代理配置"
    color: "#6366f1"
```

修改后 30 秒自动生效。

### 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `PORT` | `8080` | 服务端口 |
| `DATA_DIR` | `data` | 数据目录路径 |
| `SITE_TITLE` | `StaticMan` | 站点标题 |
| `SITE_TITLE_CN` | — | 中文品牌名 |
| `SITE_TITLE_EN` | — | 英文品牌名 |
| `SITE_DESCRIPTION` | — | 站点描述 |
| `SITE_LOGO` | `/logo.svg` | Logo URL |

## 项目结构

```
cmd/server/          入口
internal/
  auth/              认证（密码验证）
  config/            配置热加载 + 规则引擎
  handler/           HTTP 路由与业务逻辑
  masker/            正则脱敏引擎
  middleware/        CORS + 请求日志 + 限流
  cache/             内存缓存
  mime/              MIME 类型检测
  web/               前端资源嵌入（-tags withweb）
web/                 Vue 3 + Vite 前端
```

## CI/CD

Push `main` → GitHub Actions：前端构建 → Go 编译（`-tags withweb`）→ SCP 二进制到服务器 → `systemctl restart staticman` → 健康检查。

详见 [docs/DEPLOY.md](docs/DEPLOY.md)。

## 许可

私人项目，仅供个人使用。
