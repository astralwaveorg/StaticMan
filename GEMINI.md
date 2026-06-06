# StaticMan 项目指令 (Codebase)

这是 StaticMan 的主代码仓库指令文件。

## 项目定位
- **名称**: StaticMan
- **定位**: 轻量级静态文件管理器与分发服务器。
- **技术栈**: Go + Vue 3 (Rolldown/Vite 8)。

## 关键配置
- **运行方式**: Systemd 服务 (非 Docker)
- **核心路径**:
  - 二进制文件: /opt/magichub/bin/staticman
  - 源码与配置: /opt/magichub/src/
  - 数据目录: /opt/magichub/data/
- **环境变量**: ACCESS_KEY, PORT, DATA_DIR (定义于 /opt/magichub/src/.env)。

## 运维流程
- **域名**: file.magichub.top (Nginx 反向代理至 8080 端口)。
- **部署**: GitHub Actions 自动执行前端打包和 Go 编译，并通过 SCP 传输二进制文件至服务器，最后重启 systemd 服务。
- **构建**: go build -tags withweb (嵌入静态资源)。
- **Secrets**: SERVER_HOST, SSH_PRIVATE_KEY。
