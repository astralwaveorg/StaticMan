# 部署指南

## 架构概览

```
User → Nginx (443) → StaticMan (8080) → /opt/magichub/data/
```

- **Nginx**：反向代理 + SSL 终止
- **StaticMan**：Go 二进制（嵌入前端），systemd 管理
- **数据目录**：`/opt/magichub/data`（magicdata 仓库，CI 自动 pull）

## 服务器目录结构

```
/opt/magichub/
├── bin/
│   └── staticman              # 当前运行版本
├── data/                      # magicdata 仓库 (Git)
│   ├── password.yaml          # 认证密钥 + 保护规则
│   ├── metadata.yaml          # 分类元数据
│   ├── Surge/                 # 代理配置文件
│   └── Clash/
└── .env                       # 环境变量
```

## CI/CD 自动部署

Push 到 `main` 分支自动触发 GitHub Actions：

```
前端构建 → Go 编译 (-tags withweb) → SCP 上传 → systemctl restart → 健康检查
```

详见 `.github/workflows/deploy.yml`。

## 环境变量

```env
PORT=8080
DATA_DIR=/opt/magichub/data
SITE_TITLE_CN=魔匣
SITE_TITLE_EN=MagicBox
SITE_DESCRIPTION=私人网络代理配置管理中心
SITE_LOGO=/logo.svg
```

站点相关变量可通过 GitHub Secrets 配置，未设置的不会覆盖服务器已有值。

## 手动回滚

```bash
ssh hkb

# 查看备份
ls -lt /opt/magichub/bin/staticman.bak.*

# 回滚
systemctl stop staticman
cp /opt/magichub/bin/staticman.bak.XXXXXXX /opt/magichub/bin/staticman
systemctl start staticman
```
