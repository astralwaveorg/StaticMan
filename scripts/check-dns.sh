#!/bin/bash
echo "=== DNS 解析状态 ==="
echo ""
echo "files.magichub.top:"
dig +short files.magichub.top 2>/dev/null || host files.magichub.top 2>/dev/null | grep "has address"

echo ""
echo "magichub.top:"
dig +short magichub.top 2>/dev/null || host magichub.top 2>/dev/null | grep "has address"

echo ""
echo "www.magichub.top:"
dig +short www.magichub.top 2>/dev/null || host www.magichub.top 2>/dev/null | grep "has address"

echo ""
echo "如果 magichub.top 已解析到 38.147.173.222，请运行:"
echo "  /opt/magichub/setup-ssl.sh"
