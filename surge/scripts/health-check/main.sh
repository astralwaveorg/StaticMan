#!/bin/bash
# 节点健康检查主脚本
# 使用方法: ./main.sh [nodes_dir] [output_dir]

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NODES_DIR="${1:-${SCRIPT_DIR}/../../nodes}"
OUTPUT_DIR="${2:-${SCRIPT_DIR}/../../nodes}"
TEST_FILE="${3:-wugui-test.ini}"

# 加载依赖
source "${SCRIPT_DIR}/parser.sh"
source "${SCRIPT_DIR}/tester.sh"

# 配置
TEST_URL="${TEST_URL:-http://cp.cloudflare.com/generate_204}"
TIMEOUT="${TIMEOUT:-5}"
REPORT_MODE="${REPORT_MODE:-false}"  # true: 只报告不删除

echo "========================================"
echo "节点健康检查"
echo "========================================"
echo "节点目录: $NODES_DIR"
echo "测试文件: $TEST_FILE"
echo "测试URL: $TEST_URL"
echo "超时: ${TIMEOUT}s"
echo "模式: $([ "$REPORT_MODE" == "true" ] && echo "报告模式(不删除)" || echo "删除模式")"
echo "========================================"

# 确保 xray 可用
if ! command -v xray &>/dev/null; then
    echo "⚠️ xray 未安装，跳过 hysteria2/vmess 测试"
    XRAY_AVAILABLE=false
else
    XRAY_AVAILABLE=true
fi

# 统计
total=0
available=0
unavailable=0
skipped=0

# 输入文件
INI_FILE="${NODES_DIR}/${TEST_FILE}"
if [[ ! -f "$INI_FILE" ]]; then
    echo "❌ 文件不存在: $INI_FILE"
    exit 1
fi

# 临时文件
TMP_FILE=$(mktemp)

# 处理节点
echo "开始检查节点..."

while IFS= read -r line || [[ -n "$line" ]]; do
    ((total++))

    # 解析节点
    node_json=$(parse_node "$line")
    if [[ -z "$node_json" ]]; then
        # 非节点行，直接保留
        echo "$line" >> "$TMP_FILE"
        ((skipped++))
        continue
    fi

    # 提取信息
    name=$(echo "$node_json" | jq -r '.name')
    type=$(echo "$node_json" | jq -r '.type')
    server=$(echo "$node_json" | jq -r '.server')
    port=$(echo "$node_json" | jq -r '.port')
    raw=$(echo "$node_json" | jq -r '.raw')

    echo -n "测试: $name ($server:$port) ... "

    # 跳过不支持测试的类型
    if [[ "$type" == "ss+shadow-tls" || "$type" == "trojan" || "$type" == "vless" ]]; then
        echo "⏭️ 跳过(不支持测试)"
        echo "$line" >> "$TMP_FILE"
        ((skipped++))
        continue
    fi

    # 测试连通性
    if test_node "$type" "$server" "$port"; then
        echo "✅ 可用"
        echo "$line" >> "$TMP_FILE"
        ((available++))
    else
        echo "❌ 不可用"
        ((unavailable++))
    fi

done < "$INI_FILE"

echo "========================================"
echo "检查完成"
echo "========================================"
echo "总节点数: $total"
echo "可用: $available"
echo "不可用: $unavailable"
echo "跳过: $skipped"
echo "========================================"

if [[ "$unavailable" -gt 0 && "$REPORT_MODE" != "true" ]]; then
    echo "正在更新文件..."
    mv "$TMP_FILE" "$INI_FILE"
    echo "✅ 已移除 $unavailable 个不可用节点"
else
    rm -f "$TMP_FILE"
    echo "文件未修改"
fi

# 输出报告（供 GitHub Actions 使用）
echo ""
echo "--- Markdown 报告 ---"
echo "## 节点健康检查报告"
echo ""
echo "| 状态 | 数量 |"
echo "|------|------|"
echo "| 总节点 | $total |"
echo "| 可用 | $available |"
echo "| 不可用 | $unavailable |"
echo "| 跳过 | $skipped |"
echo ""
if [[ "$unavailable" -gt 0 ]]; then
    echo "⚠️ 检测到 $unavailable 个不可用节点"
fi
