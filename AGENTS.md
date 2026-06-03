# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## 项目定位

MagicHub 是一个**私人网络代理订阅管理仓库**，服务于 Surge（macOS/iOS）和 Mihomo（Linux/Android）两种代理客户端。核心功能：聚合代理节点、管理分流规则、自动生成并部署配置文件。

生产站点：`https://list.magichub.top`（静态文件服务）| `https://sub.magichub.top`（订阅转换）

## 架构总览

```
外部订阅源 / 规则仓库
       │
       ▼
bin/update_sub_list.py  ──→  clash/list.yaml + surge/list.ini
bin/update_rules.py     ──→  surge/rules/*.list
tools/clash2surge.py    ──→  Clash YAML → Surge INI 转换
surge/scripts/optimize_surge.py ──→  重写 Surge [Rule] + [Host] 段
bin/cfstmodule.sh       ──→  surge/modules/cfst.sgmodule（Cloudflare 优选 IP）
       │
       ▼
git push ──→ GitHub Actions (deploy.yml) ──→ SSH 到生产服务器 git pull
       │
       ├─→ Surge 客户端通过 managed URL 拉取 .conf
       └─→ Mihomo NAS 网关通过每日 cron 拉取 config.yaml
```

## 常用命令

```bash
# 安装 Python 依赖
pip install -r requirements.txt

# 更新 Surge 规则文件（从 rules.yaml 中定义的上游 URL 下载并合并）
python3 bin/update_rules.py

# 更新代理节点列表（从订阅 URL 拉取，黑名单过滤后写入 clash/list.yaml 和 surge/list.ini）
python3 bin/update_sub_list.py

# Clash YAML 转 Surge INI
python3 tools/clash2surge.py <input.yaml> [output.ini]

# 优化所有 Surge 配置（重写 Rule 段、修正 Proxy Group、追加 DNS hints）
python3 surge/scripts/optimize_surge.py

# Cloudflare 优选 IP 并生成 Surge 模块（需在 Linux 服务器运行）
bash bin/cfstmodule.sh

# 手动 git 推送（macOS 用当前目录，Linux 用 /data/wwwroot/MagicHub）
bash bin/push.sh
```

## 无编译/测试/lint

此项目没有构建系统、没有测试、没有 linter 配置。`requirements.txt` 仅包含 `requests` 和 `PyYAML`。

## 关键文件与格式

| 文件 | 格式 | 用途 |
|------|------|------|
| `surge/macOS.conf` / `iOS.conf` / `Macmini.conf` | Surge INI | 三设备配置，含 `#!MANAGED-CONFIG` 自动更新 |
| `clash/config.yaml` / `clash/mihomo/config.yaml` | YAML | Mihomo 主配置（97 rule-providers、30+ proxy-groups） |
| `surge/rules/rules.yaml` | YAML | 规则源定义（URL + 自定义规则），供 `update_rules.py` 读取 |
| `surge/rules/**/` | `.list` 文本 | 按分类组织（ai/、apple/、google/、social/ 等）的域名规则 |
| `clash/templates/*.ini` | INI | 订阅转换器模板，配合 `sub.magichub.top` 使用 |
| `clash/mihomo/` | — | NAS 网关部署套件（systemd unit、维护脚本、crontab） |

## 重要约定

- **双格式并行**：每次变更节点或规则，需同时维护 Surge（INI）和 Clash/Mihomo（YAML）两套格式
- **Surge managed URL**：`.conf` 文件通过 `#!MANAGED-CONFIG URL INTERVAL` 指向 `list.magichub.top`，客户端每 6 小时自动拉取
- **规则优先级顺序**：局域网 → 广告拦截 → 国内直连 → AI 分流 → 社交 → 流媒体 → Google → 开发工具 → 兜底。修改 `optimize_surge.py` 中的 `NEW_RULE_SECTION` 可全局重写
- **Mihomo rule-providers 必须带 `proxy: "节点选择"`**：从 GitHub 下载规则文件需要走代理，否则国内无法加载
- **Telegram 告警**：`update_sub_list.py` 内置错误告警，有 60 分钟同错误冷却机制
- **Cloudflare 优选**：`cfstmodule.sh` 在 Linux 上运行 CloudflareSpeedTest，提取 Top 3 IP 写入 Surge Module

## 部署流程

1. 本地修改配置文件
2. `git push origin main`
3. GitHub Actions 自动 SSH 到 `/data/magichub` 执行 `git pull`
4. Mihomo 网关每天 03:20 cron 自动同步远程 `config.yaml` 并重启

## Git 子模块

`external/subs-check` 指向 `https://github.com/beck-8/subs-check.git`，当前未初始化。如需使用需执行 `git submodule update --init`。