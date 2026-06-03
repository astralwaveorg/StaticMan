## Context

当前 MagicHub 以纯静态文件形式部署在个人服务器上，通过 GitHub Actions SSH 推送 + `git pull` 完成部署。Surge 客户端通过 managed URL 拉取 `.conf` 文件，Mihomo 网关通过 cron 拉取 `config.yaml`。

目录结构混乱，文件类型单一（仅代理配置），敏感信息（节点密码、API 密钥）完全暴露在公网可访问的静态文件上，且没有 Web 界面进行浏览和管理。

项目将在 `v2` 分支全新构建，保留核心代理配置文件的兼容性。

## Goals / Non-Goals

**Goals:**

- 构建一个自托管的通用配置文件管理平台，提供优秀的 Web 浏览体验
- 实现简易密码保护机制，控制敏感文件的可见性和内容访问
- 支持任意类型的配置文件托管（vim、git、ssh、shell 等），不限于代理配置
- 保持向后兼容：Surge managed URL 和 Mihomo cron 拉取继续正常工作
- 部署简便：Docker Compose 一键部署

**Non-Goals:**

- 不实现多用户鉴权体系（仅单密码保护）
- 不实现文件在线编辑功能（仅浏览和复制）
- 不实现文件版本历史（依赖 Git 自身版本管理）
- 不实现文件上传 Web 界面（通过 Git push 管理）
- 不替换现有 Python 脚本的节点/规则更新功能

## Decisions

### D1: 技术栈选择 — Go + Vue 3

**选择**: Go (Gin) 后端 + Vue 3 (Vite) 前端

**理由**:
- Go 编译为单一二进制，部署简单，无运行时依赖
- Gin 是成熟的 Go Web 框架，API 开发效率高
- Vue 3 + Vite 构建 SPA，加载快、交互流畅
- 单二进制 + 嵌入式前端资源，可用 Docker 单容器部署
- 替代方案：
  - *Node.js + React*: 运行时较重，内存占用高，不适配低配服务器
  - *Python + Flask*: 异步性能弱，进程模型不适合文件服务场景
  - *纯静态 + Nginx*: 无法实现密码保护等动态功能

### D2: 文件存储 — Git 仓库即数据库

**选择**: 配置文件继续以 Git 仓库形式存储，后端直接读取文件系统

**理由**:
- 配置文件已经通过 Git 管理，保持 Git 工作流不变
- 无需引入数据库，文件即数据源
- Git 历史提供天然版本追踪
- 替代方案：
  - *SQLite*: 引入额外依赖，与 Git 管理方式冲突
  - *对象存储*: 增加部署复杂度，个人服务器场景过度

### D3: 目录结构 — 按类型分类

**选择**: `data/configs/<type>/` 按配置类型组织

```
data/
├── configs/
│   ├── proxy/           # 代理配置 (原 surge/ + clash/)
│   │   ├── surge/       # Surge 配置和节点
│   │   │   ├── nodes/   # 节点文件 (密码保护)
│   │   │   ├── rules/   # 规则文件
│   │   │   ├── macOS.conf
│   │   │   ├── iOS.conf
│   │   │   └── Macmini.conf
│   │   └── mihomo/      # Mihomo 配置
│   │       ├── config.yaml
│   │       └── config-android.yaml
│   ├── vim/             # Vim 配置
│   │   └── vimrc
│   ├── git/             # Git 配置
│   │   └── gitconfig
│   ├── shell/           # Shell 配置
│   │   ├── zshrc
│   │   └── bashrc
│   └── ...              # 可扩展
├── metadata.yaml        # 文件可见性配置
└── password.yaml        # 密码配置
```

**理由**:
- 清晰的类型分区，方便前端分类展示和搜索
- 新增类型只需创建目录，无需代码改动
- 代理配置从 `surge/`、`clash/` 迁移到 `data/configs/proxy/` 子目录
- 替代方案：
  - *扁平目录*: 文件多了之后难以管理，不符合分类需求
  - *标签系统*: 需要额外元数据管理，增加复杂度

### D4: 密码保护机制 — Session Token + 配置驱动

**选择**: 单密码验证 + JWT session token + YAML 配置标记敏感文件

**机制**:
1. `data/password.yaml` 定义密码和敏感文件/目录规则
2. 未认证用户浏览时，敏感文件在文件树中隐藏或显示为锁定状态
3. 请求敏感文件内容时，返回脱敏内容（密钥替换为 `***`）
4. 通过 `/api/auth` 接口提交密码，验证通过后获取 JWT token
5. 带 token 的请求可查看完整内容
6. Token 7 天有效期，存在浏览器 localStorage 中

`password.yaml` 示例:
```yaml
password: "passward"  # 访问密码
protected:
  - path: "configs/proxy/surge/nodes"  # 整个目录密码保护
    mode: "hidden"                       # hidden=隐藏 | masked=脱敏
  - path: "configs/proxy/surge/macOS.conf"
    mode: "masked"
  - path: "configs/proxy/mihomo/config.yaml"
    mode: "masked"
```

**理由**:
- 单密码机制极简，个人服务器场景足够
- YAML 配置驱动，新增保护文件只需编辑配置
- hidden 模式完全不在文件树显示，安全性更高
- masked 模式保留文件可见但替换密钥，平衡安全与可浏览性
- JWT token 无服务端状态，部署简单

### D5: 向后兼容 — API 路由映射

**选择**: 后端路由直接映射旧 URL 到新路径

```
旧 URL                                          → 新路径
/d/surge/Macmini.conf                           → /api/raw/configs/proxy/surge/Macmini.conf
/d/surge/iOS.conf                               → /api/raw/configs/proxy/surge/iOS.conf
/d/surge/nodes/dawang.ini                       → /api/raw/configs/proxy/surge/nodes/dawang.ini (需密码)
/d/clash/config.yaml                            → /api/raw/configs/proxy/mihomo/config.yaml
```

**理由**: Surge managed URL 和 Mihomo cron 已经硬编码了旧路径，必须兼容

### D6: 前端架构 — 功能模块

核心页面和功能:
1. **文件浏览器** — 左侧文件树 + 右侧内容预览
2. **搜索** — 全文搜索 + 文件名搜索
3. **密码解锁** — 顶栏锁图标，点击弹出密码输入框
4. **分类视图** — 按配置类型（proxy/vim/git/shell）分类展示
5. **复制路径** — 文件操作栏一键复制完整访问 URL

### D7: 部署方式 — Docker Compose

**选择**: Docker Compose 单容器部署

```yaml
services:
  magichub:
    image: ghcr.io/user/magichub:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - JWT_SECRET=xxx
```

**理由**:
- 单容器包含前端静态资源 + Go 后端
- 数据卷挂载 `data/` 目录，Git 仓库数据持久化
- 环境变量配置 JWT 密钥，不硬编码

## Risks / Trade-offs

- **[旧 URL 兼容性断裂]** → 通过 API 路由映射层保证，并在部署后用 curl 验证所有 managed URL
- **[密码泄露后无法撤销已发出的 token]** → 设置 7 天 token 有效期，改密码后旧 token 自然过期；也可提供管理端强制失效功能作为后续增强
- **[大型配置文件渲染性能]** → 前端对超过 1MB 的文件仅显示前 1000 行 + 全文下载链接
- **[Go 二进制不含 Git 操作]** → 后端仅读取文件系统，不改写文件，Git 操作保持手动或 CI 触发
- **[Docker 部署增加运维复杂度]** → 提供一键部署脚本和 systemd service 文件作为备选

## Migration Plan

1. 创建 `v2` 分支
2. 搭建 Go 项目骨架 + Vue 前端项目
3. 将核心代理配置文件迁移到 `data/configs/proxy/` 目录结构
4. 配置 `metadata.yaml` 和 `password.yaml`
5. 实现 API Server 和前端 UI
6. 配置 API 路由兼容旧 URL
7. 编写 Docker Compose 和部署文档
8. 在服务器上部署测试，确认 Surge managed URL 和 Mihomo cron 正常工作
9. 切换域名指向新服务

**回滚策略**: 旧静态文件服务部署不变，v2 在不同端口运行，确认无问题后切换 Nginx 指向