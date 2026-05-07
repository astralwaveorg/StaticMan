#!/bin/bash
# INI 节点文件解析器
# 解析格式: 名称 = 类型, 服务器, 端口, 参数...
# 输出: JSON 格式

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

        # 提取基础类型（保留完整类型用于判断是否 shadow-tls）
        local base_type="${type%% *}"

        # 提取额外参数（password, sni, shadow-tls-sni 等）
        local password=""
        local sni=""
        for i in "${!parts[@]}"; do
            if [[ $i -lt 3 ]]; then continue; fi
            local param=$(echo "${parts[$i]}" | xargs)

            # password
            if [[ "$param" =~ ^password= ]]; then
                password="${param#password=}"
                password="${password//\"/}"
            fi

            # sni (可能是 sni= 或 sni= 或 sni=)
            if [[ "$param" =~ ^sni= ]]; then
                sni="${param#sni=}"
                sni="${sni//\"/}"
            fi

            # shadow-tls-sni
            if [[ "$param" =~ ^shadow-tls-sni= ]]; then
                sni="${param#shadow-tls-sni=}"
                sni="${sni//\"/}"
            fi
        done

        # 判断是否是 shadow-tls 类型
        if [[ "$type" == *"shadow-tls"* || "$type" == *"ss+"* ]]; then
            base_type="ss+shadow-tls"
        fi

        # 构造 JSON（合并 password 和 sni 到 extra 字段，格式: password|sni）
        local extra="$password"
        if [[ -n "$sni" ]]; then
            extra="${extra}|${sni}"
        fi

        if [[ -n "$extra" ]]; then
            cat <<EOF
{"name":"$name","type":"$base_type","server":"$server","port":"$port","extra":"$extra"}
EOF
        else
            cat <<EOF
{"name":"$name","type":"$base_type","server":"$server","port":"$port"}
EOF
        fi

        return 0
    fi

    return 1
}

# 解析 URI 格式（备用）
build_test_uri() {
    local type="$1"
    local server="$2"
    local port="$3"
    shift 3

    case "$type" in
        ss)
            echo "socks5://$server:$port"
            ;;
        hysteria2|hysteria)
            local password="$1"
            echo "hysteria2://$password@$server:$port?insecure=1"
            ;;
        vmess)
            # vmess URI 比较复杂，需要 full JSON
            echo ""
            ;;
        *)
            echo ""
            ;;
    esac
}
