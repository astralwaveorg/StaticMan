#!/bin/bash
# 节点连通性测试脚本
# 支持协议: ss, hysteria2, vmess, trojan, vless

TEST_URL="${TEST_URL:-http://cp.cloudflare.com/generate_204}"
TIMEOUT="${TIMEOUT:-5}"

# 测试 SS 节点（通过 SOCKS5 代理）
test_ss() {
    local server="$1"
    local port="$2"

    local code
    code=$(curl -x "socks5://$server:$port" \
                -o /dev/null -s -w "%{http_code}" \
                --connect-timeout "$TIMEOUT" \
                --max-time "$((TIMEOUT + 2))" \
                "$TEST_URL" 2>/dev/null)

    [[ "$code" == "200" || "$code" == "204" ]]
}

# 测试 hysteria2 节点
test_hysteria2() {
    local server="$1"
    local port="$2"
    local password="$3"

    # 使用 hysteria2 客户端测试
    if command -v hysteria2 &>/dev/null; then
        hysteria2 ping -l "$server:$port" -p "$password" --insecure -t "$TIMEOUT" &>/dev/null
        return $?
    elif command -v hy2 &>/dev/null; then
        hy2 ping -l "$server:$port" -p "$password" --insecure -t "$TIMEOUT" &>/dev/null
        return $?
    fi

    # 备用：使用 socat 测试 TCP 端口
    timeout "$TIMEOUT" bash -c "echo >/dev/tcp/$server/$port" 2>/dev/null
}

# 测试 vmess 节点（使用 xray）
test_vmess() {
    local server="$1"
    local port="$2"

    if ! command -v xray &>/dev/null; then
        # 备用：TCP 端口检测
        timeout "$TIMEOUT" bash -c "echo >/dev/tcp/$server/$port" 2>/dev/null
        return $?
    fi

    # xray 作为 SOCKS5 代理测试
    local xray_port=10888
    local config_file=$(mktemp)

    # 创建简单的 xray 配置
    cat > "$config_file" <<EOF
{
  "inbounds": [{"port": $xray_port, "listen": "127.0.0.1", "protocol": "socks"}],
  "outbounds": [{"protocol": "vmess", "settings": $vmess_config}]
}
EOF

    # 后台启动 xray
    xray run -config "$config_file" &>/dev/null &
    local xray_pid=$!
    sleep 1

    # 测试
    local code
    code=$(curl -x "socks5://127.0.0.1:$xray_port" \
                -o /dev/null -s -w "%{http_code}" \
                --connect-timeout "$TIMEOUT" \
                --max-time "$((TIMEOUT + 2))" \
                "$TEST_URL" 2>/dev/null)

    # 清理
    kill $xray_pid 2>/dev/null
    rm -f "$config_file"

    [[ "$code" == "200" || "$code" == "204" ]]
}

# 测试 trojan 节点
test_trojan() {
    local server="$1"
    local port="$2"
    local password="$3"

    if ! command -v xray &>/dev/null; then
        timeout "$TIMEOUT" bash -c "echo >/dev/tcp/$server/$port" 2>/dev/null
        return $?
    fi

    # 使用 xray 测试 trojan
    timeout "$TIMEOUT" bash -c "
        exec 3<>/dev/tcp/$server/$port
        echo -e 'GET / HTTP/1.1\r\nHost: $server\r\n\r\n' >&3
        head -1 <&3 | grep -q 'HTTP'
    " 2>/dev/null
}

# 测试 vless 节点
test_vless() {
    local server="$1"
    local port="$2"

    # vless 较复杂，简化为 TCP 端口检测
    timeout "$TIMEOUT" bash -c "echo >/dev/tcp/$server/$port" 2>/dev/null
}

# 主测试函数
# 用法: test_node "type" "server" "port" ["password"/"extra_params"]
test_node() {
    local type="$1"
    local server="$2"
    local port="$3"
    shift 3
    local extra="$*"

    case "$type" in
        ss|shadowsocks)
            test_ss "$server" "$port"
            ;;
        hysteria2|hysteria|hy2)
            test_hysteria2 "$server" "$port" "$extra"
            ;;
        vmess)
            test_vmess "$server" "$port"
            ;;
        trojan)
            test_trojan "$server" "$port" "$extra"
            ;;
        vless)
            test_vless "$server" "$port"
            ;;
        *)
            # 未知类型，保留节点
            return 0
            ;;
    esac
}
