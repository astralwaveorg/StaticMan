## Why

当前 MagicHub 是一个纯静态文件托管仓库，没有 Web 界面、没有权限控制、文件类型扩展困难。查看配置需要直接访问裸文件 URL，敏感信息（节点密钥、代理密码）完全暴露，且无法方便地浏览、搜索、复制路径或管理文件可见性。需要将其重构为一个功能完整的**自托管配置文件管理平台**，在 v2 分支上全新构建。

## What Changes

- **新增 Web UI**：基于 Vue 3 + Vite 构建单页应用，提供文件树浏览、语法高亮预览、一键复制路径、文件搜索、分类标签等功能
- **新增后端服务**：基于 Go + Gin 构建 REST API 服务，提供文件浏览、内容读取、密码验证等接口，替代当前的纯静态部署方式
- **新增简易权限控制**：单密码机制保护敏感文件，无需完整鉴权体系。通过配置文件标记哪些文件/目录需要密码才能访问，未认证时返回脱敏或隐藏内容
- **新增文件可见性管理**：通过 YAML 配置文件标记每个文件/目录的可见性等级（公开/需密码/隐藏），灵活控制展示
- **扩展配置类型支持**：重构目录结构，从仅支持代理配置扩展为通用配置托管平台，支持 vim、git、ssh、shell 等任意类型的 dotfiles 和配置文件
- **保留核心代理配置**：Surge 节点文件、Surge/Mihomo 配置文件、规则文件等核心代理功能完整保留并迁移至新目录结构
- **保留部署流程兼容**：Surge managed URL 需要继续工作，Mihomo cron 拉取需继续工作，通过 API 路由兼容旧 URL

## Capabilities

### New Capabilities

- `web-ui`: 前端单页应用，提供文件树浏览、语法高亮、搜索、复制路径、密码输入等交互功能
- `api-server`: Go 后端服务，提供文件浏览 API、内容读取 API、密码验证 API、配置管理 API，兼容旧静态文件 URL
- `access-control`: 基于单密码的文件访问控制，通过配置文件定义敏感文件标记，未认证请求返回脱敏内容或隐藏条目
- `file-visibility`: 文件可见性管理系统，通过 YAML 配置声明每个文件/目录的可见性等级和元数据（标签、描述、图标）
- `config-types`: 通用配置类型支持框架，将配置文件按类型组织（proxy/vim/git/shell 等），每种类型可注册语法高亮规则和预览模板
- `migration`: 从当前目录结构到 v2 目录结构的迁移方案，保留核心代理配置文件，确保 Surge managed URL 和 Mihomo cron 兼容性

### Modified Capabilities

（无现有 spec，全部为新增）

## Impact

- **目录结构重构**：当前 `surge/`、`clash/`、`config/` 等目录将重新组织到 `data/configs/` 下按类型分类
- **部署方式变更**：从纯静态文件 + GitHub Actions SSH 部署，变更为 Docker Compose 部署 Go 服务 + 前端静态资源
- **URL 兼容性**：需确保 `https://list.magichub.top/d/surge/Macmini.conf` 等旧 URL 继续可用（API 路由层做兼容）
- **依赖新增**：Go 运行时、Node.js 构建、Docker 部署环境
- **核心文件保护**：`surge/nodes/*.ini`、`surge/*.conf`、`clash/config.yaml` 等含密钥文件需标记为密码保护