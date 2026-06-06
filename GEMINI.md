# StaticMan 项目指令 (Codebase)

这是 StaticMan 的主代码仓库指令文件。

## 项目定位
- **名称**: StaticMan
- **定位**: 轻量级静态文件管理器与分发服务器，支持访问控制与自动热加载。
- **技术栈**: Go (后端) + Vue 3 TypeScript (前端)。

## 关键配置
- **运行方式**: Docker Compose
- **核心变量**: ACCESS_KEY (在服务器 .env 中配置)，用于受保护文件的鉴权。
- **数据映射**: 容器内的 /app/data 映射自宿主机的 /opt/magichub/data。

## 运维流程
- **部署**: 推送至 main 分支触发 GitHub Actions。
- **构建**: 采用多阶段 Dockerfile 编译前端并嵌入 Go 二进制文件。
- **环境**: 生产环境部署在 hkb 服务器的 /opt/magichub/src 路径。
- **Secrets**: 需在 GitHub 配置 SERVER_HOST 和 SSH_PRIVATE_KEY。
