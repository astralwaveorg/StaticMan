#!/bin/bash

# ================= 环境变量配置 =================

systemctl stop mihomo

# 1. 基础路径配置
# 测速工具所在目录 (存放 cfst, ip.txt 的目录)
CFST_WORK_DIR="/home/archon/cfst"
# Git 仓库根目录
REPO_DIR="/home/archon/cfst/MagicHub"
# 生成的 Surge 模块文件绝对路径
MODULE_FILE="$REPO_DIR/surge/modules/cfst.sgmodule"
# 临时测速结果文件
CSV_FILE="$CFST_WORK_DIR/result_auto.csv"

# 2. 需要加速的域名列表
DOMAINS=("panel.eoysky.com" "vw.eoysky.com" "list.eoysky.com" "dash.cloudflare.com")

# 3. CloudflareSpeedTest 测速参数
# -tp 443: HTTPS端口 (必须)
# -tl 250: 延迟上限 250ms
# -dn 5: 下载测速数量取前5个 (节省时间)
# -dt 10: 下载测速时间10秒
# -n 200: 并发数
# -httping: 使用 HTTP 测速 (更准确)
CMD_ARGS="-tp 443 -tl 250 -dn 5 -dt 10 -n 200 -httping -o $CSV_FILE"

# ================= 脚本逻辑开始 =================

echo "[$(date '+%Y-%m-%d %H:%M:%S')] 🚀 开始执行 Cloudflare 优选流程..."

# 1. 切换到测速工具目录 (确保能读取 ip.txt)
cd "$CFST_WORK_DIR" || { echo "❌ 无法进入目录 $CFST_WORK_DIR"; exit 1; }

# 2. 执行测速
echo "👉 正在运行测速 (请稍候)..."
# 确保程序有执行权限
chmod +x ./cfst
./cfst $CMD_ARGS > /dev/null 2>&1

# 3. 提取 Top 3 最快 IP
# 逻辑：跳过 CSV 标题行，提取第1列，取前3行，用逗号空格拼接
BEST_IPS=$(awk -F, 'NR>1 {print $1}' "$CSV_FILE" | head -n 3 | paste -sd ", " -)

# 检查是否获取到 IP
if [[ -z "$BEST_IPS" ]]; then
    echo "❌ 错误：未能提取到有效 IP，请检查测速结果。"
    exit 1
fi

echo "✅ 测速完成，Top 3 IP: $BEST_IPS"

# 4. 生成 Surge Module 文件
echo "👉 正在更新 Surge 模块..."

cat > "$MODULE_FILE" <<EOF
#!name=Cloudflare 优选 (Auto)
#!desc=自动优选脚本更新。更新时间: $(date "+%Y-%m-%d %H:%M:%S")
#!category=Auto Generated

[Host]
EOF

# 循环写入域名配置
for domain in "${DOMAINS[@]}"; do
    echo "$domain = $BEST_IPS" >> "$MODULE_FILE"
done

echo "✅ 模块文件已写入: $MODULE_FILE"

# 5. Git 提交并推送
echo "👉 正在同步到 GitHub..."

cd "$REPO_DIR" || { echo "❌ 无法进入 Git 目录 $REPO_DIR"; exit 1; }

# 检查文件是否有变化
if [[ -n $(git status -s "$MODULE_FILE") ]]; then
    git add "$MODULE_FILE"
    git commit -m "🚀 Auto update CF IPs to: $BEST_IPS [$(date +%F_%T)]"
    
    # 推送到 main 分支
    git push origin main
    
    if [ $? -eq 0 ]; then
        echo "🎉 推送成功！Surge 端稍后将自动更新。"
    else
        echo "❌ 推送失败，请检查网络或 SSH 密钥。"
    fi
else
    echo "⚠️ IP 列表未发生变化，跳过推送。"
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ 流程结束。"



