#!/bin/bash
# INI 节点文件解析器
# 解析格式: 名称 = 类型, 服务器, 端口, 参数...

parse_node() {
    local line="$1"

    # 跳过空行和注释
    [[ -z "$line" || "$line" =~ ^# ]] && return 1

    # 提取节点名称和剩余部分
    if [[ "$line" =~ ^(.+)\ =\ (.+) ]]; then
        local name="${BASH_REMATCH[1]}"
        local rest="${BASH_REMATCH[2]}"

        # 按逗号分割
        IFS=',' read -ra parts <<< "$rest"

        local type=$(echo "${parts[0]}" | xargs)
        local server=$(echo "${parts[1]}" | xargs)
        local port=$(echo "${parts[2]}" | xargs)

        # 提取类型（去掉可能的协议前缀）
        type="${type%%+*}"  # ss+shadow-tls -> ss

        # 输出 JSON 格式
        cat <<EOF
{"name":"$name","type":"$type","server":"$server","port":"$port","raw":"$line"}
EOF
        return 0
    fi

    return 1
}

# 解析 URI 格式（用于测试工具）
build_test_uri() {
    local type="$1"
    local server="$2"
    local port="$3"
    shift 3
    local params="$@"

    case "$type" in
        ss)
            # 简单 SOCKS5 测试用 curl
            echo "socks5://$server:$port"
            ;;
        vmess)
            # 构造 vmess URI
            echo "vmess://$(echo "$params" | base64 -w0)@$server:$port"
            ;;
        hysteria2|hysteria)
            echo "hysteria2://$params@$server:$port"
            ;;
        *)
            echo ""
            ;;
    esac
}
