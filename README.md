# StaticMan

[![Deploy StaticMan](https://github.com/astralwaveorg/StaticMan/actions/workflows/deploy.yml/badge.svg)](https://github.com/astralwaveorg/StaticMan/actions/workflows/deploy.yml)

私人配置文件管理平台 — 浏览、搜索、保护你的配置文件。

> **Clone 注意**：`StaticMan` 项目名大小写敏感，请使用 `git clone https://github.com/astralwaveorg/StaticMan.git`（大写 S 和 M）。若 Go module 路径解析失败，请检查本地目录是否为 `StaticMan` 而非 `staticman`。

## 功能

- 🗂️ **分类浏览** — 按类型组织配置（代理 / Vim / Git / Shell…），Web 界面一键查看
- 🔒 **密码保护** — 登录解锁受保护文件，公开文件无需登录
- 🔗 **短 URL + Key 认证** — 受保护文件通过 `?key=<访问密钥>` 直接访问，可浏览器预览
- 🔍 **全文搜索** — 文件名 + 内容搜索，受保护内容自动脱敏
- 📋 **一键复制** — 公开文件复制裸 URL，受保护文件自动附 key
- 🎨 **语法高亮** — YAML、INI、JSON、Shell、Vim 等格式自动高亮
- 🔗 **旧 URL 兼容** — Surge managed URL 和 Mihomo cron 拉取继续工作

## 保护模型

系统支持基于**规则引擎**的多级访问控制：

| 策略 | 行为 | 适用场景 |
|------|------|---------|
| **hide**（隐藏） | 完全不出现在列表、树、搜索中，直接访问返回 404 | `.git`、临时文件、规则文件自身 |
| **protect**（保护） | 列表中可见（带🔒），但内容需认证才能访问 | 私密配置、密钥文件、敏感文档 |

### 规则优先级

1. **hide > protect** — 隐藏策略优先。
2. **目录级 > 全局级** — `.encrypt` 中的规则优先。
3. **深层目录 > 父目录** — 遵循“最具体规则”原则。

### 规则语法

支持 Glob 和正则表达式：
- `*.key`：匹配所有 `.key` 后缀文件。
- `**/private/*`：匹配任意层级下 `private` 目录的内容。
- `Surge/`：以 `/` 结尾表示匹配该目录及其下所有内容。
- `regex:.*password.*`：使用正则表达式匹配。

**设计要点：**
- 未登录时仍可在文件树中看到 `protect` 文件（标注🔒），但内容脱敏。
- `hide` 文件完全不可见且短 URL 返回 404，不暴露存在性。
- 受保护文件短 URL 不带 key 返回 403，带 `?key=<访问密钥>` 返回完整内容。
- 兼容层 `/d/*` 直接放行（Surge/Mihomo 等机器客户端）。
- 登录后通过 `Authorization: Bearer` header 或 `?key=` 参数认证。

## 快速开始

```bash
# 注意项目名为 StaticMan（大小写敏感）
git clone https://github.com/astralwaveorg/StaticMan.git && cd StaticMan

# 创建环境变量文件
cat > .env <<EOF
ACCESS_KEY=changeme
PORT=8080
DATA_DIR=./data
SITE_TITLE_CN=魔匣
SITE_TITLE_EN=MagicBox
SITE_DESCRIPTION=私人网络代理配置管理中心
SITE_LOGO=/logo.svg
EOF

# 启动（需要本地安装 Go 1.22+ 和 Node 20+）
cd web && npm install && npm run build && cd ..
cp -r web/dist internal/web/dist
go build -tags withweb -o staticman ./cmd/server
./staticman
```

或使用 Docker：

```bash
docker compose up -d
```

访问 `http://localhost:8080`。

## 路由架构

### 原始文件层 — 短 URL，浏览器可预览

```
/vim/vimrc                              → 公开文件，直接访问
/proxy/surge/macOS.conf                → 受保护文件，需 ?key=<访问密钥>
/proxy/surge/macOS.conf?key=GEM91816   → 受保护文件，完整内容
/proxy/surge/rules/reject.list          → 公开文件，直接访问
/proxy/surge/assets/icons/github.png    → 图片资源，直接访问
```

### API 层 — Web UI 专用

```
/api/auth          POST  密码认证
/api/tree          GET   文件树（受保护文件标🔒，未登录看脱敏）
/api/categories    GET   分类列表
/api/file/*path    GET   文件内容 JSON（protected 文件未登录看脱敏）
/api/search?q=&type= GET 搜索
```

### 兼容层 — 旧 URL 重写

```
/d/surge/Macmini.conf     → proxy/surge/Macmini.conf（直接放行）
/d/clash/config.yaml      → proxy/mihomo/config.yaml（直接放行）
```

## 配置

### `data/password.yaml`

```yaml
# 访问密码（登录用）
password: "GEM91816"

# 固定访问 key（16 位以上），用于直接访问受保护文件
# 例：http://localhost:8080/raw/Clash/config.yaml?key=GEM91816
static_key: "GEM91816"

# 新增：规则引擎
rules:
  hide:
    - ".git/"
    - "*.tmp"
  protect:
    - "*/private/*"
    - "*.key"
    - "regex:.*credential.*"

# 向后兼容：精确路径保护
protected:
  - path: "proxy/surge/nodes"
  - path: "proxy/surge/macOS.conf"

---

### 2. 目录级配置 (`.encrypt` 文件)

在数据目录任意位置放置 `.encrypt` 文件，定义该目录及其子目录的规则：

```bash
# data/proxy/.encrypt
hide test/
hide *.bak
*.conf    # 默认策略为 protect
```
```

### `data/metadata.yaml`

```yaml
categories:
  proxy:
    name: "代理配置"
    icon: "monitor"
    description: "Surge / Mihomo 代理规则与节点"
    color: "#409EFF"

files:
  "proxy/surge/nodes":
    visibility: "protected"
    description: "代理节点文件（含服务器密码）"
  "proxy/surge/macOS.conf":
    visibility: "protected"
    description: "macOS Surge 配置"
  "vim/vimrc":
    visibility: "public"
    description: "Vim 编辑器配置"
```

修改后 30 秒自动生效。

## 目录结构

```
data/
├── configs/               # 配置文件（按类型组织）
│   ├── proxy/             # 代理配置
│   │   ├── surge/         #   Surge 配置、节点、规则、模块、图标
│   │   └── mihomo/        #   Mihomo 配置
│   ├── vim/               # Vim 配置
│   ├── git/               # Git 配置
│   ├── shell/             # Shell 配置
│   └── ...                # 可扩展
├── password.yaml          # 密码和保护规则
└── metadata.yaml          # 分类元数据和文件可见性
```

## 开发

```bash
# 后端（开发模式从 internal/web/dist 读取，不强制嵌入）
go run ./cmd/server/

# 前端
cd web && npm install && npm run dev

# 前端构建 + 嵌入后端（必须将构建产物复制到 embed 目录）
cd web && npm run build && cd ..
rm -rf internal/web/dist && cp -r web/dist internal/web/dist
go build -tags withweb ./cmd/server/
```

## 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `ACCESS_KEY` | `changeme` | 访问密钥（**生产环境必须修改**） |
| `PORT` | `8080` | 服务端口 |
| `DATA_DIR` | `data` | 数据目录路径 |
| `SITE_TITLE_CN` | `StaticMan` | 中文品牌名（浏览器标签、Logo 主标题） |
| `SITE_TITLE_EN` | `StaticMan` | 英文品牌名/实例名（Logo 副标题） |
| `SITE_DESCRIPTION` | — | 站点描述，与英文名组合显示在副标题 |
| `SITE_LOGO` | `/logo.svg` | 站点 Logo URL |
| `SITE_TITLE` | — | 向后兼容：完整标题（未设置 CN/EN 时回退） |

## 许可

私人项目，仅供个人使用。