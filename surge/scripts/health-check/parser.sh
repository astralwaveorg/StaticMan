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

        # 提取基础类型（去掉协议修饰符）
        local base_type="${type%%+*}"
        base_type="${base_type%% *}"

        # 提取额外参数（password 等）
        local extra=""
        for i in "${!parts[@]}"; do
            if [[ $i -lt 3 ]]; then continue; fi
            local param=$(echo "${parts[$i]}" | xargs)
            if [[ "$param" =~ ^password= ]]; then
                extra="${param#password=}"
                extra="${extra//\"/}"
            fi
        done

        # 构造 JSON
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
