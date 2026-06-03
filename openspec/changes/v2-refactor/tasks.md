## 1. 项目初始化

- [x] 1.1 创建 `v2` 分支并清理不需要的旧文件（保留 surge/、clash/、config/ 等核心数据目录）
- [x] 1.2 初始化 Go 模块项目 (`go mod init github.com/athena/magichub`)，创建 `cmd/server/main.go` 入口
- [x] 1.3 初始化 Vue 3 + Vite 前端项目 (`web/` 目录)，安装 Element Plus 和代码高亮依赖
- [x] 1.4 创建 Docker Compose 配置和 Dockerfile（Go 编译 + 前端嵌入方案）
- [x] 1.5 创建 `data/` 目录结构和配置文件骨架 (`metadata.yaml`, `password.yaml`)

## 2. 数据层与配置

- [x] 2.1 创建 `data/configs/` 目录结构，按类型划分子目录 (`proxy/`, `vim/`, `git/`, `shell/`)
- [x] 2.2 编写迁移脚本 `scripts/migrate-v1.sh`，将核心文件从旧路径复制到新路径
- [x] 2.3 编写 `data/password.yaml` 默认配置，标记敏感文件（节点文件、含密钥配置）为 `hidden` 或 `masked`
- [x] 2.4 编写 `data/metadata.yaml` 默认配置，定义代理、Vim 等分类的元数据和文件可见性
- [x] 2.5 实现配置文件热加载机制（30 秒轮询或 fsnotify 监听 `password.yaml` 和 `metadata.yaml`）

## 3. API Server — 核心路由

- [x] 3.1 实现文件树 API (`GET /api/tree`)：扫描 `data/configs/` 返回层级结构，根据认证状态过滤 hidden/marked 项
- [x] 3.2 实现文件内容 API (`GET /api/file/:path`)：读取文件内容，对 masked 文件进行脱敏处理（替换密码/密钥模式为 `***`）
- [x] 3.3 实现原始文件服务：三层路由架构 — `/proxy/*` 短URL原始文件层 + `/api/file/*` API层(带脱敏) + `/d/*` 兼容层
- [x] 3.4 实现搜索 API (`GET /api/search?q=keyword&type=name|content`)：文件名搜索和内容搜索，脱敏敏感结果
- [x] 3.5 实现配置类型发现接口 (`GET /api/categories`)：自动扫描 `data/configs/` 顶层目录返回类型列表

## 4. API Server — 认证与兼容

- [x] 4.1 实现密码验证 API (`POST /api/auth`)：接收密码、常量时间比较、签发 JWT token（7 天有效期）
- [x] 4.2 实现 JWT 中间件：解析 `Authorization: Bearer` token，注入认证状态到请求上下文
- [x] 4.3 实现内容脱敏引擎：识别并替换 SS 密码、API key/token、`password=` 等模式为 `***`
- [x] 4.4 实现旧 URL 路由映射：`/d/surge/*` → `configs/proxy/surge/*`，`/d/clash/*` → `configs/proxy/*`，三层路由架构（原始文件层短 URL、API 层带认证、兼容层重写）
- [x] 4.5 为旧 URL 路由添加 IP 白名单或 User-Agent 识别逻辑（Surge/Mihomo 客户端绕过密码保护）

## 5. 前端 UI — 框架与布局

- [x] 5.1 搭建 Vue 3 SPA 骨架：路由配置、全局布局（顶栏 + 侧边栏 + 主内容区）
- [x] 5.2 实现首页分类卡片视图：从 `/api/categories` 获取类型列表，展示图标+名称+描述+文件数
- [x] 5.3 实现文件树组件：左侧递归树展示 `data/configs/` 结构，展开/折叠，锁定图标标记 protected 项

## 6. 前端 UI — 核心功能

- [x] 6.1 实现文件预览页：右侧内容展示区，集成语法高亮（基于文件扩展名选择高亮器）
- [x] 6.2 实现密码解锁交互：顶栏锁图标 → 弹窗输入密码 → 调用 `/api/auth` → 存储 JWT 到 localStorage → 刷新树和内容
- [x] 6.3 实现一键复制功能：复制完整 URL（`/api/raw/:path` 绝对路径）和相对路径按钮
- [x] 6.4 实现搜索功能：顶栏搜索框，支持文件名搜索和内容搜索，搜索结果高亮匹配行
- [x] 6.5 实现大文件处理：超过 1MB 的文件仅显示前 1000 行 + 全文下载按钮

## 7. 部署与文档

- [x] 7.1 编写 Dockerfile：多阶段构建（Node 编译前端 → Go 编译后端 → 精简运行镜像）
- [x] 7.2 编写 docker-compose.yml：服务定义、端口映射、数据卷挂载、环境变量配置
- [x] 7.3 编写部署文档 README：Docker Compose 部署步骤、环境变量说明、旧 URL 兼容性配置
- [x] 7.4 编写配置文件文档：`password.yaml` 和 `metadata.yaml` 的完整格式说明和示例
- [x] 7.5 端到端测试：验证 Surge managed URL 和 Mihomo cron 拉取在旧路由下正常工作