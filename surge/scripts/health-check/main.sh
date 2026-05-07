#!/bin/bash
# 节点健康检查主脚本
# 使用方法: ./main.sh [nodes_dir] [output_dir] [test_file]
#
# 环境变量:
#   TEST_URL      - 测试 URL (默认: http://cp.cloudflare.com/generate_204)
#   TIMEOUT       - 超时秒数 (默认: 10)
#   MAX_RETRIES   - 失败重试次数 (默认: 3)
#   REPORT_MODE   - true: 仅报告不删除 (默认: false)

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
TIMEOUT="${TIMEOUT:-10}"
MAX_RETRIES="${MAX_RETRIES:-3}"
REPORT_MODE="${REPORT_MODE:-false}"

echo "========================================"
echo "节点健康检查"
echo "========================================"
echo "节点目录: $NODES_DIR"
echo "测试文件: $TEST_FILE"
echo "测试URL: $TEST_URL"
echo "超时: ${TIMEOUT}s x ${MAX_RETRIES}次重试"
echo "模式: $([ "$REPORT_MODE" == "true" ] && echo "报告模式(不删除)" || echo "删除模式")"
echo "========================================"

# 统计
total=0
available=0
unavailable=0
skipped=0
parse_error=0

# 输入文件
INI_FILE="${NODES_DIR}/${TEST_FILE}"
if [[ ! -f "$INI_FILE" ]]; then
    echo "❌ 文件不存在: $INI_FILE"
    exit 1
fi

# 备份原文件（万一出问题）
cp "$INI_FILE" "${INI_FILE}.bak"

# 临时文件
TMP_FILE=$(mktemp)
> "$TMP_FILE"  # 清空

# 处理节点
echo ""
echo "开始检查节点..."

while IFS= read -r line || [[ -n "$line" ]]; do
    # 跳过空行
    [[ -z "$line" ]] && continue

    ((total++))

    # 解析节点
    node_json=$(parse_node "$line")
    if [[ -z "$node_json" || "$node_json" == "null" ]]; then
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

    # 跳过解析失败的
    if [[ "$server" == "null" || "$port" == "null" || "$server" == "" || "$port" == "" ]]; then
        echo "$line" >> "$TMP_FILE"
        ((parse_error++))
        echo "⚠️ 解析错误: ${line:0:50}..."
        continue
    fi

    echo -n "测试: $name ($server:$port) ... "

    # 提取密码/额外参数
    extra=$(echo "$node_json" | jq -r '.extra // empty')

    # 测试连通性
    if test_node "$type" "$server" "$port" "$extra"; then
        echo "✅ 可用"
        echo "$line" >> "$TMP_FILE"
        ((available++))
    else
        echo "❌ 不可用"
        ((unavailable++))
    fi

done < "$INI_FILE"

echo ""
echo "========================================"
echo "检查完成"
echo "========================================"
echo "总节点数: $total"
echo "可用: $available"
echo "不可用: $unavailable (将移除)"
echo "跳过: $skipped"
echo "解析错误: $parse_error"
echo "========================================"

# 删除备份
rm -f "${INI_FILE}.bak"

# 更新文件
if [[ "$REPORT_MODE" == "true" ]]; then
    rm -f "$TMP_FILE"
    echo "📋 报告模式，文件未修改"
else
    if [[ "$unavailable" -gt 0 ]]; then
        echo "正在更新文件..."

        # 检查可用节点数
        if [[ "$available" -eq 0 && "$skipped" -eq 0 ]]; then
            # 全部不可用，清空文件但不删除
            echo "⚠️ 所有节点不可用，清空文件"
            > "$INI_FILE"
        else
            # 用临时文件替换
            mv "$TMP_FILE" "$INI_FILE"
        fi

        echo "✅ 已移除 $unavailable 个不可用节点"
    else
        rm -f "$TMP_FILE"
        echo "✅ 所有节点可用，文件未修改"
    fi
fi

# 输出报告
echo ""
echo "========================================"
echo "Markdown 报告"
echo "========================================"
echo "## 节点健康检查报告"
echo ""
echo "| 项目 | 数量 |"
echo "|------|------|"
echo "| 总节点 | $total |"
echo "| 可用 | $available |"
echo "| 不可用 | $unavailable |"
echo "| 跳过 | $skipped |"
echo "| 解析错误 | $parse_error |"
echo ""

if [[ "$unavailable" -gt 0 && "$REPORT_MODE" != "true" ]]; then
    echo "✅ 已自动移除 $unavailable 个不可用节点"
fi

# 设置退出码：有不可用节点且不是报告模式
if [[ "$unavailable" -gt 0 && "$REPORT_MODE" != "true" ]]; then
    exit 0  # 成功，有变化
else
    exit 0  # 成功，无变化
fi
