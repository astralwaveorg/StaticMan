# 文件访问控制规则引擎（Rule Engine）落地实施文档

本文档定义了 StaticMan 访问控制系统的技术实现细节，旨在替换现有的硬编码系统文件过滤及简单的受保护路径列表。

## 一、 核心目标

1.  **多级规则**：支持全局配置 (`password.yaml`) 与目录级配置 (`.encrypt`)。
2.  **双重策略**：
    - `hide` (隐藏)：文件在所有列表、搜索、树状图中不可见，直接访问返回 404。
    - `protect` (保护)：文件可见（带锁），但访问内容需认证，未认证访问返回 403。
3.  **模式匹配**：支持 Glob (`*.key`, `**/private/*`) 和正则 (`regex:.*secret.*`)。
4.  **高性能**：规则预编译，匹配结果二级缓存（LRU + 目录规则缓存）。

---

## 二、 语法与规则优先级

### 2.1 语法规范
- `#`：注释。
- `path/to/dir/`：以 `/` 结尾仅匹配目录。
- `*.ext`：匹配特定后缀。
- `**/name`：匹配任意深度的文件/目录。
- `/top/level`：以 `/` 开头匹配相对于数据根目录的绝对路径。
- `regex:pattern`：使用正则表达式。
- 默认前缀为 `protect`，在 `.encrypt` 中可显式使用 `hide pattern` 声明隐藏。

### 2.2 冲突解决优先级（由高到低）
1.  **策略优先**：`hide` 规则总是覆盖 `protect` 规则。
2.  **深度优先**：深层目录的 `.encrypt` 规则优先于父目录及全局规则。
3.  **具体度优先**：精确匹配 > Glob 匹配 > 正则匹配。

---

## 三、 数据结构 (Go 实现)

### 3.1 规则模型
```go
type RuleType string

const (
	Hide    RuleType = "hide"
	Protect RuleType = "protect"
)

type Rule struct {
	Type     RuleType
	Pattern  string         // 原始模式
	Regexp   *regexp.Regexp // 编译后的正则
	IsAbs    bool           // 是否从根开始
	IsDirOnly bool          // 是否仅目录
	Source   string         // 规则来源 (如 ".encrypt" 或 "global")
}
```

### 3.2 匹配结果
```go
type MatchResult struct {
	Hidden    bool
	Protected bool
	MatchedBy *Rule
}
```

---

## 四、 核心逻辑实现

### 4.1 路径标准化
所有进入引擎的路径必须：
1. 使用 `path.Clean()` 清理。
2. 统一转换为 Unix 风格 `/`。
3. 去除首尾斜杠（除根目录外）。

### 4.2 规则匹配流程
对于给定路径 `path`：
1. **收集规则链**：
   - 查找全局规则 (`password.yaml`)。
   - 向上递归收集所有 `.encrypt`：`dir(path)` -> `parent(dir(path))` -> ... -> `root`。
2. **倒序遍历**（从子目录到全局）：
   - 首先检查是否有任何匹配的 `hide` 规则 -> 若匹配，立即返回 `Hidden: true`。
   - 其次检查是否有任何匹配的 `protect` 规则 -> 若匹配，记录 `Protected: true` 并继续寻找（因为可能被更深层的 `hide` 覆盖）。
3. **返回默认值**：若均无匹配，则为公开文件。

### 4.3 缓存策略
- **LRU 缓存**：缓存 `path -> MatchResult`，容量 10k。
- **目录规则缓存**：缓存每个 `.encrypt` 文件的编译结果及 `ModTime`。
- **失效机制**：
  - 全局配置更新 -> 清空所有缓存。
  - 访问目录时检测 `.encrypt` 的 `ModTime` -> 若变更，清理受影响的分支缓存。

---

## 五、 代码集成点

### 5.1 internal/config/rules.go (新增)
实现 `RuleEngine` 结构体及 `Match(path string, isDir bool)` 方法。

### 5.2 internal/config/config.go
- 在 `Config` 结构体中集成 `RuleEngine`。
- `Load` 时初始化引擎。
- 改造 `IsProtected(path)` 为调用引擎：`return engine.Match(path, false).Protected`。

### 5.3 internal/handler/handler.go
- **`handleLs` / `buildTree`**：
  ```go
  res := h.cfg.Match(childPath, entry.IsDir())
  if res.Hidden { continue }
  item.Protected = res.Protected
  ```
- **`handleFile` / `handleRaw`**：
  ```go
  res := h.cfg.Match(path, info.IsDir())
  if res.Hidden { return 404 }
  if res.Protected && !auth { return 403 }
  ```
- **`handleSearch`**：
  在 `walkDir` 循环中第一步执行 `Match`，若 `Hidden` 则跳过该分支。

---

## 六、 验收测试用例 (Test Matrix)

| 测试场景 | 规则配置 | 路径 | 预期结果 |
| :--- | :--- | :--- | :--- |
| 全局隐藏 | `hide: [".git"]` | `.git/config` | 404 / 列表不可见 |
| 扩展名保护 | `protect: ["*.key"]` | `ssh/id_rsa.key` | 列表可见带锁 / 访问 403 |
| 目录级覆盖 | 全局 `hide: ["*.tmp"]`；`.encrypt` 中 `protect *.tmp` | `cache/test.tmp` | **隐藏** (Hide 优先) |
| 深度优先级 | `/A/.encrypt` 隐藏所有；`/A/B/.encrypt` 保护所有 | `/A/B/file.txt` | **保护** (深层优先) |
| 正则匹配 | `regex:.*(pass\|secret).*` | `my_password.txt` | 保护 |
| 向后兼容 | `protected: [{path: "Surge"}]` | `Surge/config` | 保护 |

---

## 七、 实施计划

1. **第一阶段 (开发)**：
   - 编写 `rules.go` 核心引擎及其单元测试。
   - 完成 Glob 到正则的转换函数。
2. **第二阶段 (集成)**：
   - 在 `Config` 中接入引擎，替换 `IsSystemFile` 硬编码。
   - 改造 `handler.go` 中的所有调用点。
3. **第三阶段 (验证)**：
   - 运行上述测试矩阵。
   - 性能压测：在 1000 级目录下进行搜索。
4. **第四阶段 (交付)**：
   - 更新文档，并在 `magicdata` 仓库中演示 `.encrypt` 用法。

---

## 八、 风险与补救
- **性能风险**：若目录极深导致频繁 `Stat` `.encrypt`，则引入 30s 缓存 TTL，不每次检查磁盘。
- **误屏蔽风险**：增加 `/api/debug/rules?path=xxx` 接口，方便管理员查询某个文件为何被隐藏或保护。
