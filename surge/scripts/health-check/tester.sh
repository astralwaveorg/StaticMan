#!/bin/bash
# 节点连通性测试脚本
# 支持协议: ss, hysteria2, vmess

TEST_URL="${TEST_URL:-http://cp.cloudflare.com/generate_204}"
TIMEOUT="${TIMEOUT:-5}"
XRAY_BIN="${XRAY_BIN:-/usr/local/bin/xray}"

# 测试 SS 节点（通过 SOCKS5 代理）
test_ss() {
    local server="$1"
    local port="$2"

    local code
    code=$(curl -x "socks5://$server:$port" \
                -o /dev/null -s -w "%{http_code}" \
                --connect-timeout "$TIMEOUT" \
                --max-time "$TIMEOUT" \
                "$TEST_URL" 2>/dev/null)

    [[ "$code" == "200" || "$code" == "204" ]]
}

# 测试 hysteria2 节点
test_hysteria2() {
    local server="$1"
    local port="$2"
    local password="$3"

    local result
    result=$("$XRAY_BIN" test "hysteria2://$password@$server:$port?insecure=1" \
                -t "$TEST_URL" 2>/dev/null)

    [[ "$result" =~ available|Available ]]
}

# 测试 vmess 节点（简化版：只测 TCP 连接）
test_vmess() {
    local server="$1"
    local port="$2"

    # vmess 较复杂，这里简化为 TCP 端口检测
    timeout "$TIMEOUT" bash -c "echo >/dev/tcp/$server/$port" 2>/dev/null
}

# 主测试函数
test_node() {
    local type="$1"
    local server="$2"
    local port="$3"
    shift 3

    case "$type" in
        ss)
            test_ss "$server" "$port"
            ;;
        hysteria2|hysteria)
            test_hysteria2 "$server" "$port" "$1"
            ;;
        vmess)
            test_vmess "$server" "$port"
            ;;
        *)
            # 未知类型，默认跳过（保留节点）
            return 0
            ;;
    esac
}
