# StaticMan Engine (Codebase)

StaticMan 是驱动 MagicHub 的核心引擎，基于 Go + Vue 3 的轻量级文件管理系统。

## ⚙️ 技术栈与构建
- **Backend**: Go 1.22+ (标准库 http + fsnotify 热加载)。
- **Frontend**: Vue 3 + Vite 8 + Rolldown。
- **构建标签**: 必须带 `-tags withweb` 才能将前端 UI 嵌入 Go 二进制文件。

## 🚀 生产环境配置 (Server: hkb)
- **服务类型**: Systemd Service (`staticman.service`)。
- **关键路径**:
  - 执行文件: `/opt/magichub/bin/staticman`
  - 环境配置: `/opt/magichub/src/.env`
- **环境变量**:
  - `ACCESS_KEY`: 用于 JWT 签名的核心密钥（由服务器端定义）。
  - `DATA_DIR`: 指向 `/opt/magichub/data`。

## 🌐 网络与安全
- **入口**: `https://files.magichub.top`
- **转发**: Nginx (443) -> localhost (8080)。
- **保护**: 默认忽略 `.git`, `.github`, `.DS_Store`。系统文件 (`password.yaml`, `metadata.yaml`) 自动在导航中隐藏。

## 🔄 CI/CD 工作流
- **触发**: Push to `main` 分支。
- **过程**: 构建前端 -> 编译 Go -> SCP 传输 -> `systemctl restart staticman`。
