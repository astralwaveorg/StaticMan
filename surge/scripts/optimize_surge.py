#!/usr/bin/env python3
"""
Surge 配置优化脚本
优化 proxy group 和 rule 顺序，解决：
1. Gemini/AI 服务优先走住宅代理
2. YouTube/Google 流畅度优化
3. Rule 顺序优化（特定规则先于通用规则）
4. Apple Push / Google FCM 直连优化
"""

import sys

# 统一的 Rule 段替换内容
NEW_RULE_SECTION = '''[Rule]
# === 1. 基础系统与局域网 ===
RULE-SET,LAN,DIRECT
RULE-SET,SYSTEM,DIRECT
DEST-PORT,123,DIRECT
AND,((DOMAIN-SUFFIX,github.com), (DEST-PORT,22)),DIRECT
IP-CIDR,100.64.0.0/10,DIRECT,no-resolve
IP-CIDR,203.248.84.18/32,DIRECT,no-resolve
IP-CIDR,103.38.83.78/32,DIRECT,no-resolve
IP-CIDR,38.147.173.222/32,DIRECT,no-resolve
IP-CIDR,116.62.189.89/32,DIRECT,no-resolve
DOMAIN-SUFFIX,push.apple.com,DIRECT
DOMAIN-SUFFIX,akadns.net,🚀 节点选择
DOMAIN-KEYWORD,apple.com.edgekey.net,🚀 节点选择
IP-CIDR,17.249.0.0/16,🚀 节点选择,no-resolve
IP-CIDR,17.252.0.0/16,🚀 节点选择,no-resolve
IP-CIDR,17.57.144.0/22,🚀 节点选择,no-resolve
IP-CIDR,17.188.128.0/18,🚀 节点选择,no-resolve
IP-CIDR,17.188.20.0/23,🚀 节点选择,no-resolve
IP-CIDR6,2620:149:a44::/48,🚀 节点选择,no-resolve
IP-CIDR6,2403:300:a42::/48,🚀 节点选择,no-resolve
IP-CIDR6,2403:300:a51::/48,🚀 节点选择,no-resolve
IP-CIDR6,2a01:b740:a42::/48,🚀 节点选择,no-resolve
DOMAIN-SET,https://cdn.jsdelivr.net/gh/Loyalsoldier/surge-rules@release/private.txt,DIRECT
# === 2. 广告拦截 ===
RULE-SET,https://list.magichub.top/d/surge/rules/reject.list,REJECT
RULE-SET,https://surge.bojin.co/geosite/balanced/category-media-cn@ads,REJECT
# === 3. 国内直连 ===
RULE-SET,https://list.magichub.top/d/surge/rules/mydomains.list,DIRECT
RULE-SET,https://list.magichub.top/d/surge/rules/direct.list,DIRECT
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/DouYin/DouYin.list,DIRECT
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Tencent/Tencent_All.list,DIRECT
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Tencent/Tencent_All_No_Resolve.list,DIRECT,no-resolve
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/AliPay/AliPay.list,DIRECT
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Alibaba/Alibaba_All.list,DIRECT
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Alibaba/Alibaba_All_No_Resolve.list,DIRECT,no-resolve
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/WeChat/WeChat.list,DIRECT
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/DingTalk/DingTalk.list,DIRECT
RULE-SET,https://ruleset.skk.moe/List/ip/domestic.conf,DIRECT,no-resolve
RULE-SET,https://cdn.jsdelivr.net/gh/Loyalsoldier/surge-rules@release/cncidr.txt,DIRECT,no-resolve
RULE-SET,https://surge.bojin.co/geosite/balanced/category-media-cn,DIRECT
RULE-SET,https://surge.bojin.co/geosite/balanced/category-social-media-cn,DIRECT
# === 4. AI 服务分流（特定优先于通用）===
RULE-SET,https://list.magichub.top/d/surge/rules/ai/ai-china.list,DIRECT
RULE-SET,https://list.magichub.top/d/surge/rules/ai/openai.list,"🤖 OpenAI"
RULE-SET,https://list.magichub.top/d/surge/rules/ai/anthropic.list,"🤖 Anthropic"
RULE-SET,https://list.magichub.top/d/surge/rules/ai/gemini.list,"🤖 国际 AI"
RULE-SET,https://list.magichub.top/d/surge/rules/ai/perplexity.list,"🤖 国际 AI"
RULE-SET,https://list.magichub.top/d/surge/rules/ai/xai.list,"🤖 国际 AI"
RULE-SET,https://list.magichub.top/d/surge/rules/ai/github-copilot.list,"📟 开发运维"
RULE-SET,https://list.magichub.top/d/surge/rules/ai/ai-global.list,"🤖 国际 AI"
RULE-SET,https://raw.githubusercontent.com/viewer12/OverseasAI.list/main/rule/Surge/OverseasAI/OverseasAI.list,"🤖 国际 AI"
# === 5. 社交通讯 ===
RULE-SET,https://list.magichub.top/d/surge/rules/social/telegram.list,"📱 Telegram"
RULE-SET,https://list.magichub.top/d/surge/rules/social/facebook.list,"📱 社交通讯"
RULE-SET,https://list.magichub.top/d/surge/rules/social/x.list,"📱 社交通讯"
RULE-SET,https://surge.bojin.co/geosite/balanced/whatsapp,"📱 社交通讯"
RULE-SET,https://surge.bojin.co/geosite/balanced/discord,"📱 社交通讯"
RULE-SET,https://surge.bojin.co/geosite/balanced/slack,"📟 开发运维"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Line/Line.list,"📱 社交通讯"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/TikTok/TikTok.list,"🗣️ TikTok"
# === 6. 视频流媒体 ===
RULE-SET,https://list.magichub.top/d/surge/rules/google/youtube.list,"🎬 YouTube"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Netflix/Netflix.list,"📺 国际媒体"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Disney/Disney.list,"📺 国际媒体"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Spotify/Spotify.list,"📺 国际媒体"
RULE-SET,https://surge.bojin.co/geosite/balanced/twitch,"📺 国际媒体"
RULE-SET,https://list.magichub.top/d/surge/rules/apple/apple-tvplus.list,"📺 国际媒体"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/GlobalMedia/GlobalMedia.list,"📺 国际媒体"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/GlobalMedia/GlobalMedia_All.list,"📺 国际媒体"
# === 7. Google 服务 ===
RULE-SET,https://list.magichub.top/d/surge/rules/google/google.list,"🔍 Google"
RULE-SET,https://list.magichub.top/d/surge/rules/google/googlefcm.list,DIRECT
RULE-SET,https://list.magichub.top/d/surge/rules/google/google-scholar.list,"🚀 节点选择"
RULE-SET,https://list.magichub.top/d/surge/rules/google/google-play.list,"🚀 节点选择"
# === 8. 开发工具 ===
RULE-SET,https://list.magichub.top/d/surge/rules/devtools/github.list,"📟 开发运维"
RULE-SET,https://list.magichub.top/d/surge/rules/devtools/docker.list,"📟 开发运维"
RULE-SET,https://surge.bojin.co/geosite/balanced/npmjs,"📟 开发运维"
RULE-SET,https://surge.bojin.co/geosite/balanced/python,"📟 开发运维"
RULE-SET,https://surge.bojin.co/geosite/balanced/notion,"📟 开发运维"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Stackexchange/Stackexchange.list,"📟 开发运维"
# === 9. 其他特定服务 ===
RULE-SET,https://surge.bojin.co/geosite/balanced/reddit,"🚀 节点选择"
RULE-SET,https://surge.bojin.co/geosite/balanced/pinterest,"🚀 节点选择"
RULE-SET,https://list.magichub.top/d/surge/rules/cloudflare/cloudflare-cn.list,DIRECT
RULE-SET,https://list.magichub.top/d/surge/rules/cloudflare/cloudflare-proxy.list,"🚀 节点选择"
RULE-SET,https://list.magichub.top/d/surge/rules/amazon.list,"🚀 节点选择"
RULE-SET,https://surge.bojin.co/geosite/balanced/paypal,"🚀 节点选择"
RULE-SET,https://list.magichub.top/d/surge/rules/microsoft.list,"Ⓜ️ 微软服务"
RULE-SET,https://list.magichub.top/d/surge/rules/apple/apple-cn.list,DIRECT
RULE-SET,https://list.magichub.top/d/surge/rules/apple/apple-intelligence.list,"🍎 苹果服务"
RULE-SET,https://list.magichub.top/d/surge/rules/apple/apple.list,"🍎 苹果服务"
RULE-SET,https://list.magichub.top/d/surge/rules/apple/apple-update.list,DIRECT
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/AppleID/AppleID.list,"🍎 苹果服务"
RULE-SET,https://surge.bojin.co/geosite/balanced/speedtest,DIRECT
RULE-SET,https://surge.bojin.co/geosite/balanced/medium,"🚀 节点选择"
RULE-SET,https://list.magichub.top/d/surge/rules/proxy.list,"🚀 节点选择"
RULE-SET,https://list.magichub.top/d/surge/rules/airport.list,"🚀 节点选择"
RULE-SET,https://raw.githubusercontent.com/bunizao/TutuBetterRules/tutu/RuleList/DOMAlN/Mail.list,"🚀 节点选择"
RULE-SET,https://list.magichub.top/d/surge/rules/devtools/steam.list,"🚀 节点选择"
# === 10. 第三方大范围白名单 ===
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Direct/Direct.list,DIRECT
DOMAIN-SET,https://cdn.jsdelivr.net/gh/Loyalsoldier/surge-rules@release/direct.txt,DIRECT
RULE-SET,https://ruleset.skk.moe/List/non_ip/direct.conf,DIRECT
# === 11. 第三方大范围黑名单 ===
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/GlobalScholar/GlobalScholar.list,"🚀 节点选择"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Global/Global_All.list,"🚀 节点选择"
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Proxy/Proxy_All.list,"🚀 节点选择"
DOMAIN-SET,https://cdn.jsdelivr.net/gh/Loyalsoldier/surge-rules@release/proxy.txt,"🚀 节点选择"
# === 12. No-Resolve IP 兜底 ===
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/GlobalMedia/GlobalMedia_All_No_Resolve.list,"📺 国际媒体",no-resolve
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Global/Global_All_No_Resolve.list,"🚀 节点选择",no-resolve
RULE-SET,https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Surge/Proxy/Proxy_All_No_Resolve.list,"🚀 节点选择",no-resolve
# === 13. 最终兜底 ===
GEOIP,CN,DIRECT,no-resolve
FINAL,🌐 漏网之鱼,dns-failed
'''

# Google DNS hints 要插入 [Host] 段的内容
GOOGLE_DNS_HINTS = '''*.googlevideo.com = server:1.1.1.1
*.google.com = server:1.1.1.1
*.youtube.com = server:1.1.1.1
*.googleapis.com = server:1.1.1.1
*.gstatic.com = server:1.1.1.1
*.googleusercontent.com = server:1.1.1.1
*.ggpht.com = server:1.1.1.1
*.ytimg.com = server:1.1.1.1
'''


def fix_proxy_groups(content):
    """修正 Proxy Group 策略"""
    # 1) 🤖 国际 AI: smart -> fallback，住宅代理优先
    content = content.replace(
        '🤖 国际 AI = smart, "🇯🇵 日本机房 01", "🇯🇵 日本机房 02", "🇭🇰 香港机房 01", "🇭🇰 香港机房 02", "🇺🇸 美国机房 01", "🇺🇸 美国机房 02", "🇺🇸 美国家宽 01", "🇺🇸 美国家宽 02"',
        '🤖 国际 AI = fallback, "🇺🇸 美国家宽 01", "🇺🇸 美国家宽 02", "🇺🇸 美国机房 01", "🇺🇸 美国机房 02", "🇯🇵 日本机房 01", "🇯🇵 日本机房 02", "🇭🇰 香港机房 01", "🇭🇰 香港机房 02"'
    )

    # 2) 🎬 YouTube: fallback -> url-test，增加 Google 专属测试 URL
    content = content.replace(
        '🎬 YouTube = fallback, "🇺🇸 美国机房 01", "🇺🇸 美国机房 02", "🇺🇸 美国家宽 01", "🇺🇸 美国家宽 02", "🇯🇵 日本机房 01", "🇯🇵 日本机房 02", "🇭🇰 香港机房 01", "🇭🇰 香港机房 02", "🇺🇸 美国", "🚀 节点选择"',
        '🎬 YouTube = url-test, "🇺🇸 美国机房 01", "🇺🇸 美国机房 02", "🇺🇸 美国家宽 01", "🇺🇸 美国家宽 02", "🇯🇵 日本机房 01", "🇯🇵 日本机房 02", "🇭🇰 香港机房 01", "🇭🇰 香港机房 02", "🇺🇸 美国", "🚀 节点选择", url=http://www.google.com/generate_204, interval=300, tolerance=50, evaluate-before-use=true'
    )

    # 3) 🔍 Google: fallback -> url-test，增加 Google 专属测试 URL
    content = content.replace(
        '🔍 Google = fallback, "🇺🇸 美国机房 01", "🇺🇸 美国机房 02", "🇺🇸 美国家宽 01", "🇺🇸 美国家宽 02", "🇯🇵 日本机房 01", "🇯🇵 日本机房 02", "🇭🇰 香港机房 01", "🇭🇰 香港机房 02", "🇺🇸 美国", "🚀 节点选择"',
        '🔍 Google = url-test, "🇺🇸 美国机房 01", "🇺🇸 美国机房 02", "🇺🇸 美国家宽 01", "🇺🇸 美国家宽 02", "🇯🇵 日本机房 01", "🇯🇵 日本机房 02", "🇭🇰 香港机房 01", "🇭🇰 香港机房 02", "🇺🇸 美国", "🚀 节点选择", url=http://www.google.com/generate_204, interval=300, tolerance=50, evaluate-before-use=true'
    )

    return content


def fix_hosts(content):
    """在 [Host] 段追加 Google/YouTube DNS 解析优化"""
    # 防止重复插入
    if '*.googlevideo.com = server:1.1.1.1' in content:
        return content

    content = content.replace(
        '[Host]\n',
        '[Host]\n' + GOOGLE_DNS_HINTS
    )
    return content


def process_conf(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # 1. 修正 Proxy Groups
    content = fix_proxy_groups(content)

    # 2. 替换整个 Rule 段
    rule_start = content.find('[Rule]\n')
    host_start = content.find('[Host]\n')
    if rule_start == -1 or host_start == -1:
        print(f'[ERROR] Could not find [Rule] or [Host] in {filepath}')
        return False

    content = content[:rule_start] + NEW_RULE_SECTION + '\n' + content[host_start:]

    # 3. 追加 Host DNS 优化
    content = fix_hosts(content)

    with open(filepath, 'w') as f:
        f.write(content)

    print(f'[OK] {filepath}')
    return True


def expand_google_list(filepath):
    """扩展 google.list，补充常见缺失域名"""
    with open(filepath, 'r') as f:
        content = f.read()

    additions = '''DOMAIN-SUFFIX,googleusercontent.com
DOMAIN-SUFFIX,1e100.net
DOMAIN-SUFFIX,googletagmanager.com
DOMAIN-SUFFIX,google-analytics.com
DOMAIN-SUFFIX,doubleclick.net
DOMAIN-SUFFIX,googleadservices.com
DOMAIN-SUFFIX,googlesyndication.com
DOMAIN-SUFFIX,google.co.jp
DOMAIN-SUFFIX,google.co.uk
DOMAIN-SUFFIX,google.co.hk
DOMAIN-SUFFIX,google.com.hk
DOMAIN-SUFFIX,google.com.tw
DOMAIN-SUFFIX,google.com.sg
DOMAIN-SUFFIX,google.de
DOMAIN-SUFFIX,google.fr
DOMAIN-SUFFIX,google.co.in
'''

    # 防止重复追加
    if 'googleusercontent.com' in content:
        print(f'[SKIP] {filepath} already expanded')
        return

    with open(filepath, 'a') as f:
        f.write('\n' + additions)

    print(f'[OK] {filepath}')


if __name__ == '__main__':
    for conf in ['surge/iOS.conf', 'surge/Macmini.conf', 'surge/macOS.conf']:
        process_conf(conf)

    expand_google_list('surge/rules/google/google.list')
    print('All done.')
