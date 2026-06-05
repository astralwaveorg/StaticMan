# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目定位

MagicHub 是一个**私人网络代理订阅管理平台**，v2 重构后采用 Go + Vue 3 全栈架构，替代原 v1 纯 Python 脚本方案。

生产站点：`https://list.magichub.top`（静态文件服务 + Web UI）| `https://sub.magichub.top`（订阅转换）

## 架构总览

```
外部订阅源 / 规则仓库
       │
       ▼
bin/update_sub_list.py ──→  data/configs/ (v2 数据目录)
bin/update_rules.py     ──→  surge/rules/*.list
       │
       ▼
Go HTTP Server (cmd/server)
  ├─ /api/*          ──→  JSON API（文件树、分类、搜索、鉴权）
  ├─ /raw/*          ──→  统一原始文件路由（受保护文件需 ?key=JWT）
  ├─ /{category}/*   ──→  分类短 URL（运行时扫描 data/ 动态注册，等价于 /raw）
  ├─ /d/*            ──→  v1 兼容路由（无鉴权）
  └─ /               ──→  Vue SPA（Go embed 嵌入）
       │
       ▼
Docker Compose 部署（magichub:v2）
```

## 常用命令

```bash
# 前端开发
cd web && npm run dev          # Vite 开发服务器（代理 /api → localhost:8080）
cd web && npm run build        # 构建 → ../internal/web/dist/

# 后端开发
go run ./cmd/server/           # 运行 Go 服务（默认 :8080）
go build -tags withweb ./cmd/server/  # 构建含嵌入前端的单二进制

# Docker
docker-compose up -d           # 部署（需先构建 magichub:v2 镜像）

# v1 脚本（仍在使用）
python3 bin/update_sub_list.py # 更新代理节点
python3 bin/update_rules.py    # 更新 Surge 规则
python3 tools/clash2surge.py   # Clash YAML → Surge INI
python3 surge/scripts/optimize_surge.py  # 优化 Surge 配置
bash bin/cfstmodule.sh         # Cloudflare 优选 IP（需 Linux）
```

## V2 架构详情

### Go 后端（`internal/`）

| 包 | 职责 |
|---|---|
| `cmd/server` | 入口：加载配置、注册路由、启动 HTTP |
| `config` | 热加载 `data/password.yaml` + `metadata.yaml`（30s 轮询） |
| `auth` | 密码验证（常量时间比较）、JWT 签发/校验（HS256, 7天） |
| `handler` | 三层路由 + **动态注册**：运行时扫描 `data/` 下顶级目录作为分类前缀路由 |
| `masker` | 内容脱敏：4 个正则模式（明文 key:value、JSON 凭据、URI 凭据、YAML 凭据） |
| `middleware` | CORS + 请求日志 |
| `web` | `//go:embed dist/` 嵌入前端 SPA；构建标签 `withweb` |

### Vue 3 前端（`web/`）

| 文件 | 职责 |
|---|---|
| `App.vue` | 根布局：Header（搜索/主题/登录） + 路由视图 + 登录弹窗 |
| `views/HomeView.vue` | 分类卡片网格 |
| `views/BrowseView.vue` | 左侧文件树 + 右侧文件查看器 |
| `views/SearchView.vue` | 文件名/内容搜索 |
| `components/TreeNode.vue` | 递归文件树组件 |
| `components/FileViewer.vue` | highlight.js 代码高亮 + 操作按钮 |
| `api/index.ts` | Axios API 客户端（JWT 自动附加，同时存 localStorage + cookie） |
| `stores/auth.ts` | Pinia 认证状态（login/logout） |
| `stores/ui.ts` | Pinia UI 状态（登录弹窗/命令面板开关） |
| `components/CommandPalette.vue` | 全局搜索命令面板（Ctrl+K 触发，集成 SearchView） |
| `styles/global.css` | 设计系统 CSS 变量（dark/light 主题） |

**路由（仅两条，均在 `web/src/router/index.ts`）：**
- `/` → HomeView（分类卡片）
- `/browse/:pathMatch(.*)*` → BrowseView（文件树 + 查看器）
- 搜索功能通过 `CommandPalette` 组件全局嵌入，无独立路由

### 数据目录（`data/`）

- `data/configs/` — 按分类存放配置文件（proxy/surge/, proxy/mihomo/, vim/, git/, shell/）
- `data/password.yaml` — 访问密码 + 受保护路径列表
- `data/metadata.yaml` — 分类元数据（图标、描述、颜色）

## 关键文件

| 文件 | 用途 |
|---|---|
| `internal/handler/handler.go` | 所有 HTTP 路由定义和业务逻辑 |
| `internal/config/config.go` | 配置热加载，`IsProtected(path)` 检查 |
| `internal/masker/masker.go` | 内容脱敏正则引擎 |
| `internal/web/embed.go` | `//go:embed` 前端资源（需 `-tags withweb`） |
| `web/vite.config.ts` | 开发代理 + 构建输出到 `internal/web/dist/` |
| `Dockerfile` | 三阶段构建：Node 前端 → Go 后端 → Alpine 运行 |
| `docker-compose.yml` | 生产部署配置（端口 8080，挂载 `./data`） |
| `data/password.yaml` | 认证密码和受保护路径配置 |
| `data/metadata.yaml` | 分类展示元数据 |

## 重要约定

- **v1 兼容**：`/d/*` 路由绕过所有鉴权，确保旧客户端不断线
- **双格式维护**：Surge（INI）和 Mihomo（YAML）需同时更新
- **受保护文件**：未认证时内容脱敏（masker）；认证后原文返回
- **构建标签**：生产构建须 `go build -tags withweb`，否则无前端 SPA
- **规则优先级**：局域网 → 广告拦截 → 国内直连 → AI → 社交 → 流媒体 → Google → 开发工具 → 兜底

## 部署流程

**Docker（推荐 v2）**：
```bash
docker-compose up -d    # 需先构建镜像
```

**v1 方式（仍在使用）**：
```bash
git push origin main → GitHub Actions → SSH 到生产服务器 git pull
```

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `JWT_SECRET` | `changeme` | JWT 签名密钥 |
| `PORT` | `8080` | HTTP 监听端口 |
| `DATA_DIR` | `./data` | 数据目录路径 |

## 无测试/lint

v2 前后端均无测试框架和 linter 配置。