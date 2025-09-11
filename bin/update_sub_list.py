import requests
import yaml
import os

# --- 配置 ---
# GitHub仓库中的文件夹路径
CLASH_DIR = "clash"
SURGE_DIR = "surge"
# 新生成的文件名
CLASH_FILE = "list.yaml"
SURGE_FILE = "list.ini"

# 下载链接
CLASH_URL = "https://api-huacloud.net/sub?target=clash&insert=true&emoji=true&udp=true&clash.doh=true&new_name=true&filename=Flower_Trojan.yaml&url=https%3A%2F%2Fapi.xmancdn.net%2Fosubscribe.php%3Fsid%3D177358%26token%3DUdRCu8VV"
# 注意: Surge 的下载链接未提供，这里使用一个占位符。请替换为实际链接。
SURGE_URL = "https://example.com/your-surge-config.ini"

# 过滤关键词黑名单
# 名称中包含这些关键词的节点将被过滤掉
CLASH_BLACKLIST = ["Traffic", "Expire"]
SURGE_BLACKLIST = ["Traffic", "Expire"]


# --- 辅助函数 ---
def download_content(url):
    """从给定的 URL 下载内容并返回文本。"""
    try:
        response = requests.get(url, timeout=15)
        response.raise_for_status()  # 检查请求是否成功
        return response.text
    except requests.exceptions.RequestException as e:
        print(f"Error downloading {url}: {e}")
        return None


def ensure_directory_exists(directory):
    """确保指定的目录存在，如果不存在则创建。"""
    if not os.path.exists(directory):
        os.makedirs(directory)
        print(f"Created directory: {directory}")


# --- Clash 配置文件处理 ---
def update_clash_config():
    """下载、解析、过滤并保存 Clash 配置。"""
    print("Starting to update Clash config...")

    raw_content = download_content(CLASH_URL)
    if not raw_content:
        print("Failed to download Clash content. Aborting Clash update.")
        return

    try:
        data = yaml.safe_load(raw_content)
    except yaml.YAMLError as e:
        print(f"Error parsing YAML content: {e}")
        return

    # 过滤代理节点
    if "proxies" in data and isinstance(data["proxies"], list):
        filtered_proxies = []
        for proxy in data["proxies"]:
            if "name" in proxy and isinstance(proxy["name"], str):
                # 检查代理名称是否包含黑名单中的关键词
                if not any(keyword in proxy["name"] for keyword in CLASH_BLACKLIST):
                    filtered_proxies.append(proxy)

        # 创建新的 YAML 数据结构，只包含过滤后的代理
        new_clash_data = {"proxies": filtered_proxies}

        output_path = os.path.join(CLASH_DIR, CLASH_FILE)
        ensure_directory_exists(CLASH_DIR)

        try:
            with open(output_path, "w", encoding="utf-8") as f:
                yaml.dump(new_clash_data, f, allow_unicode=True, sort_keys=False)
            print(f"Successfully saved filtered Clash config to {output_path}")
        except IOError as e:
            print(f"Error writing file to {output_path}: {e}")
    else:
        print("No 'proxies' section found in the downloaded Clash config.")


# --- Surge 配置文件处理 ---
def update_surge_config():
    """下载、解析、过滤并保存 Surge 配置。"""
    print("Starting to update Surge config...")

    raw_content = download_content(SURGE_URL)
    if not raw_content:
        print("Failed to download Surge content. Aborting Surge update.")
        return

    try:
        # Surge INI 文件没有标准解析器，使用字符串分割来处理
        proxy_section_start = raw_content.find("[Proxy]")
        if proxy_section_start == -1:
            print("No [Proxy] section found in Surge config.")
            return

        proxy_section = raw_content[proxy_section_start:].split("\n")

        filtered_proxies = []
        for line in proxy_section:
            line = line.strip()
            # 过滤空行、注释行和黑名单行
            if (
                line
                and not line.startswith(";")
                and not any(keyword in line for keyword in SURGE_BLACKLIST)
            ):
                filtered_proxies.append(line)

        # 将过滤后的代理重新组装成 [Proxy] 部分
        new_surge_content = "[Proxy]\n" + "\n".join(filtered_proxies) + "\n"

        output_path = os.path.join(SURGE_DIR, SURGE_FILE)
        ensure_directory_exists(SURGE_DIR)

        try:
            with open(output_path, "w", encoding="utf-8") as f:
                f.write(new_surge_content)
            print(f"Successfully saved filtered Surge config to {output_path}")
        except IOError as e:
            print(f"Error writing file to {output_path}: {e}")

    except Exception as e:
        print(f"Error processing Surge content: {e}")


# --- 主程序入口 ---
if __name__ == "__main__":
    update_clash_config()
    update_surge_config()
