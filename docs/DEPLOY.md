# 部署指南

## 架构概览

```
User → Nginx (80/443) → StaticMan (8080)
```

- **Nginx**: 反向代理 + SSL 终止 + 静态资源缓存
- **StaticMan**: Go 二进制，嵌入前端资源，监听 8080
- **数据目录**: `/opt/magichub/data` (Git 仓库，自动同步)

## 域名配置

| 域名 | 状态 | 说明 |
|------|------|------|
| `files.magichub.top` | ✅ 已生效 | 主域名 |
| `magichub.top` | ⏳ 待 DNS | 配置完成，DNS 解析后自动生效 |
| `www.magichub.top` | ⏳ 待 DNS | 同上 |

### DNS 配置步骤

1. 在 DNS 服务商添加 A 记录：
   - `magichub.top` → `38.147.173.222`
   - `www.magichub.top` → `38.147.173.222`

2. 等待 DNS 生效（通常 5-30 分钟）

3. 在服务器上申请 SSL 证书：
   ```bash
   ssh root@38.147.173.222
   /opt/magichub/setup-ssl.sh
   ```

4. 验证：
   ```bash
   curl -I https://magichub.top
   ```

## 自动部署

### 一键部署（推荐）

在 Mac 开发机上：

```bash
cd /Users/athena/MagicHub/staticman
./scripts/deploy.sh [服务器地址]
```

默认服务器: `root@38.147.173.222`

### 部署流程

```
1. npm run build          (构建前端)
2. go build -tags withweb  (编译 Linux 二进制，嵌入前端)
3. 备份旧版本
4. 停止服务
5. 上传新二进制
6. 启动服务
7. 健康检查 (HTTP 200)
8. 验证前端版本一致性
```

### 关键注意事项

**⚠️ 必须使用 `-tags withweb` 编译**

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags withweb -o staticman-linux ./cmd/server
```

原因：
- `withweb` tag 启用 `embed.go`，将 `internal/web/dist` 嵌入二进制
- 不使用此 tag 会启用 `dev.go`，从文件系统读取，导致部署不同步

**⚠️ 必须先停止服务再替换二进制**

systemd 运行时文件被锁定，直接 scp 会失败。

**⚠️ 清理旧的开发模式 dist**

服务器上的 `/opt/magichub/src/internal/web/dist` 是旧文件，使用嵌入模式后不再需要。

## 回滚

```bash
ssh root@38.147.173.222

# 查看最新备份
ls -lt /opt/magichub/bin/staticman.bak.* | head -5

# 回滚到上一个版本
systemctl stop staticman
cp /opt/magichub/bin/staticman.bak.XXXXXXX /opt/magichub/bin/staticman
systemctl start staticman

# 验证
systemctl status staticman
```

## 服务器目录结构

```
/opt/magichub/
├── bin/
│   ├── staticman              # 当前运行版本
│   ├── staticman.bak.*        # 历史备份 (保留5个)
│   └── staticman.bak.XXXXXXX  # 上一个版本
├── data/                      # 数据仓库 (Git)
│   ├── .git/
│   ├── Clash/
│   ├── Surge/
│   └── ...
├── .env                       # 环境变量配置
├── deploy.sh                  # 服务器端部署脚本
├── setup-ssl.sh               # SSL 证书申请脚本
├── check-dns.sh               # DNS 检查脚本
└── src/                       # 源代码备份 (可选)
```

## 配置文件

### 环境变量 (/opt/magichub/.env)

```env
ACCESS_KEY=GEM91816              # JWT 签名密钥 / 下载密钥
PORT=8080                        # 服务端口
DATA_DIR=/opt/magichub/data      # 数据目录
SITE_TITLE_CN=魔匣               # 中文标题
SITE_TITLE_EN=MagicBox           # 英文标题
SITE_DESCRIPTION=私人网络代理配置管理中心
SITE_LOGO=/logo.svg
```

### Nginx (/etc/nginx/conf.d/magichub.top.conf)

- 双域名并行：files.magichub.top + magichub.top
- HTTP → HTTPS 自动跳转
- ACME 挑战路径支持
- 静态资源长期缓存 (1年)
- HTML/API 禁止缓存
- OCSP Stapling + HSTS

### systemd (/etc/systemd/system/staticman.service)

- 自动重启 (Restart=always)
- 环境变量从 /opt/magichub/.env 加载
- 日志输出到 journal

## 常见问题

### Q: 部署后前端还是旧版本？

A: 检查是否使用了 `-tags withweb` 编译：
```bash
strings staticman-linux | grep 'dist/assets/index'
# 应该有输出，否则未嵌入
```

### Q: scp 上传失败 "Failure"？

A: 服务正在运行，文件被锁定。部署脚本会自动先 `systemctl stop`，
但如果是手动操作，请先停止服务：
```bash
systemctl stop staticman
# 然后 scp
systemctl start staticman
```

### Q: 证书续期？

A: certbot 已配置 systemd timer，自动续期。可手动测试：
```bash
certbot renew --dry-run
```

### Q: 添加新的域名？

A: 修改 `/etc/nginx/conf.d/magichub.top.conf`，添加 server_name，
然后 `nginx -t && nginx -s reload`，最后申请证书。
