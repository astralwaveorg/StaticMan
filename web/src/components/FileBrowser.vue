<template>
  <div class="fb">
    <!-- Sub-header with view toggle (always show) + breadcrumb (when path) -->
    <div class="fb-subbar" v-if="showPath || true">
      <div class="fb-path">
        <a class="path-link" @click="goUp" v-if="parentLabel">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M15 18l-6-6 6-6"/></svg>
        </a>
        <span class="path-sep" v-if="parentLabel">/</span>
        <span class="path-current">{{ currentName }}</span>
      </div>
      <div class="fb-actions">
        <button class="view-toggle" :class="{active: viewMode==='grid'}" @click="viewMode='grid'" title="网格视图">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
        </button>
        <button class="view-toggle" :class="{active: viewMode==='list'}" @click="viewMode='list'" title="列表视图">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/></svg>
        </button>
      </div>
    </div>

    <!-- Grid view -->
    <div v-if="viewMode==='grid'" class="grid">
      <div
        v-for="(item, i) in items"
        :key="item.path"
        class="grid-item-wrap"
        :style="{'--delay': `${i * 30}ms`}"
      >
        <a
          class="grid-item"
          :class="{locked: item.protected, dir: item.type==='directory', bin: item.isBinary}"
          @click.prevent="openItem(item)"
        >
          <div class="grid-icon" :class="{locked: item.protected, dir: item.type==='directory', bin: item.isBinary}">
            <svg v-if="item.protected" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            <svg v-else-if="item.type==='directory'" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
            <svg v-else-if="item.isBinary" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M7 7h10M7 12h10M7 17h6"/></svg>
            <svg v-else width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
          </div>
          <div class="grid-name truncate">{{ item.name }}</div>
          <div class="grid-size">{{ fmtSize(item.size) }}</div>
        </a>
        <div class="grid-actions">
          <button class="grid-action" :title="item.type==='directory' ? '复制路径' : '复制链接'" @click.stop="copyPath(item)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
            <span class="grid-action-label">{{ item.type==='directory' ? '路径' : '链接' }}</span>
          </button>
          <button v-if="item.type!=='directory'" class="grid-action" title="下载" @click.stop="downloadItem(item)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            <span class="grid-action-label">下载</span>
          </button>
        </div>
      </div>
    </div>

    <!-- List view -->
    <div v-else class="list">
      <div class="list-header">
        <span class="col-icon"></span>
        <span class="col-name">名称</span>
        <span class="col-size">大小</span>
        <span class="col-time">修改时间</span>
        <span class="col-actions">操作</span>
      </div>
      <div
        v-for="(item, i) in items"
        :key="item.path"
        class="list-item"
        :class="{locked: item.protected, dir: item.type==='directory', bin: item.isBinary}"
        :style="{'--delay': `${i * 20}ms`}"
        @click="openItem(item)"
      >
        <div class="list-icon" :class="{locked: item.protected, dir: item.type==='directory', bin: item.isBinary}">
          <svg v-if="item.protected" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          <svg v-else-if="item.type==='directory'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
          <svg v-else-if="item.isBinary" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M7 7h10M7 12h10M7 17h6"/></svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
        </div>
        <span class="list-name truncate">{{ item.name }}</span>
        <span class="list-size">{{ fmtSize(item.size) }}</span>
        <span class="list-time">{{ item.modTime || '—' }}</span>
        <div class="list-actions">
          <button v-if="item.type==='directory'" class="row-action" title="复制路径" @click.stop="copyPath(item)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            <span class="row-action-label">路径</span>
          </button>
          <button v-else class="row-action" title="复制 Raw 链接" @click.stop="copyRaw(item)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
            <span class="row-action-label">链接</span>
          </button>
          <button v-if="item.type!=='directory'" class="row-action" title="复制文本" @click.stop="copyContent(item)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M9 5H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h9a2 2 0 0 0 2-2V9"/></svg>
            <span class="row-action-label">复制</span>
          </button>
          <button v-if="item.type!=='directory'" class="row-action" title="下载" @click.stop="downloadItem(item)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            <span class="row-action-label">下载</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Empty -->
    <div v-if="!loading && !items.length" class="empty">
      <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.4"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
      <p>此目录为空</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getLs, getFile, getRawUrl, isLoggedIn, type LsItem } from '../api'

const props = defineProps<{ rootPath: string; activeTool: string; excludeDirs?: string[] }>()
const router = useRouter()

const items = ref<LsItem[]>([])
const loading = ref(false)

// 默认视图：顶层（分类根）= 网格，子级 = 列表
// 也允许 localStorage 覆盖（记住用户偏好）
const storedView = localStorage.getItem('fb_view') as 'grid' | 'list' | null
const defaultView: 'grid' | 'list' = storedView ?? (props.activeTool ? 'list' : 'grid')
const viewMode = ref<'grid' | 'list'>(defaultView)

const path = computed(() => props.rootPath)
const showPath = computed(() => !!path.value)
const currentName = computed(() => {
  if (!path.value) return ''
  return path.value.split('/').pop() || ''
})
const parentLabel = computed(() => {
  if (!path.value || !path.value.includes('/')) return ''
  return path.value.split('/').slice(0, -1).pop() || ''
})

async function load(p: string) {
  if (!p) {
    items.value = []
    loading.value = false
    return
  }
  loading.value = true
  try {
    const { data } = await getLs(p)
    // 后端 items 可能为 null，统一兜底成空数组
    let raw: any[] = data?.items || []
    // 过滤掉已注册为 tools 的子目录
    if (props.excludeDirs && props.excludeDirs.length) {
      const excludeSet = new Set(props.excludeDirs)
      raw = raw.filter((it: any) => !(it.type === 'directory' && excludeSet.has(it.name)))
    }
    items.value = raw
  } catch {
    items.value = []
  }
  loading.value = false
}

function openItem(item: LsItem) {
  router.push('/browse/' + item.path)
}

function goUp() {
  if (!path.value.includes('/')) return
  const parent = path.value.split('/').slice(0, -1).join('/')
  router.push('/browse/' + parent)
}

function fmtSize(b: number): string {
  if (!b) return '—'
  if (b < 1024) return `${b} B`
  if (b < 1048576) return `${(b / 1024).toFixed(1)} KB`
  return `${(b / 1048576).toFixed(1)} MB`
}

async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.removeChild(ta)
    return true
  }
}

// 复制当前项的相对路径
function copyPath(item: LsItem) {
  copyToClipboard(item.path)
}

// 复制文件原始 Raw URL（受保护时若已登录会带 ?key=）
function copyRaw(item: LsItem) {
  const url = getRawUrl(item.path, item.protected, true)
  if (item.protected && !isLoggedIn()) {
    copyToClipboard(item.path)
  } else {
    copyToClipboard(url)
  }
}

// 复制文件内容（仅文件）
async function copyContent(item: LsItem) {
  if (item.type === 'directory') return
  if (item.protected && !isLoggedIn()) {
    alert('此文件受保护，请先登录后再复制内容')
    return
  }
  try {
    const { data } = await getFile(item.path)
    await copyToClipboard(data.content || '')
  } catch (e: any) {
    if (e?.response?.status === 403) {
      alert('此文件受保护，请先登录')
    } else {
      alert('复制失败')
    }
  }
}

// 触发下载（通过 a 链接）
function downloadItem(item: LsItem) {
  if (item.type === 'directory') return
  const url = getRawUrl(item.path, item.protected, true)
  const a = document.createElement('a')
  a.href = url
  a.download = item.name
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

watch(viewMode, (v) => localStorage.setItem('fb_view', v))
watch(() => [path.value, props.excludeDirs], () => load(path.value), { immediate: true })
</script>

<style scoped>
.fb { display: flex; flex-direction: column; min-height: 100%; flex: 1; }

.fb-subbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 16px;
  border-bottom: 1px solid var(--glass-border);
  gap: 8px;
  flex-shrink: 0;
}
.fb-path {
  display: flex; align-items: center; gap: 6px;
  font-size: 13px;
  font-family: var(--font-mono);
  min-width: 0; flex: 1;
}
.path-link {
  color: var(--text-secondary);
  padding: 3px 8px;
  border-radius: 6px;
  cursor: pointer;
  max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  transition: all var(--t-fast) var(--ease);
}
.path-link:hover { background: var(--bg-hover); color: var(--text-primary); }
.path-sep { color: var(--text-tertiary); }
.path-current { color: var(--text-primary); font-weight: 500; }

.fb-actions { display: flex; gap: 2px; }
.view-toggle {
  width: 28px; height: 28px;
  border-radius: 6px;
  color: var(--text-tertiary);
  display: flex; align-items: center; justify-content: center;
  transition: all var(--t-fast) var(--ease);
}
.view-toggle:hover { background: var(--bg-hover); color: var(--text-primary); }
.view-toggle.active { background: var(--bg-hover); color: var(--accent); }

/* Grid */
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 8px;
  padding: 14px;
}
.grid-item-wrap {
  display: flex; flex-direction: column; gap: 4px;
  animation: fadeInUp 350ms var(--ease) both;
  animation-delay: var(--delay, 0ms);
}
.grid-item {
  display: flex; flex-direction: column; align-items: center;
  padding: 16px 6px 10px;
  background: var(--bg-elevated);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
  text-decoration: none; color: inherit;
  cursor: pointer;
  transition: all var(--t-base) var(--ease);
  text-align: center;
}
.grid-item:hover {
  transform: translateY(-2px);
  border-color: var(--accent);
  background: var(--bg-hover);
}
.grid-item.dir:hover { border-color: var(--accent); }
.grid-item.locked { background: rgba(251,191,36,0.04); }
.grid-icon {
  width: 44px; height: 44px;
  border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  background: var(--bg-surface);
  color: var(--text-tertiary);
  margin-bottom: 8px;
  transition: all var(--t-base) var(--ease);
}
.grid-item:hover .grid-icon { transform: scale(1.08); }
.grid-icon.dir { color: var(--accent); background: color-mix(in srgb, var(--accent) 12%, var(--bg-surface)); }
.grid-icon.locked { color: var(--warning); background: rgba(251,191,36,0.1); }
.grid-icon.bin { color: #a855f7; background: rgba(168,85,247,0.1); }

.grid-name { font-size: 12px; font-weight: 500; max-width: 100%; word-break: break-word; line-height: 1.3; }
.grid-size { font-size: 10px; color: var(--text-tertiary); margin-top: 2px; font-family: var(--font-mono); }

/* Grid actions - 常驻显示 */
.grid-actions {
  display: flex; gap: 4px; justify-content: center;
  opacity: 1;
}
.grid-action {
  display: inline-flex; align-items: center; gap: 3px;
  padding: 2px 6px;
  font-size: 10px; color: var(--text-tertiary);
  background: var(--bg-elevated);
  border: 1px solid var(--glass-border);
  border-radius: 4px;
  transition: all var(--t-fast) var(--ease);
}
.grid-action:hover { color: var(--text-primary); border-color: var(--accent); background: var(--bg-hover); }

/* List */
.list { display: flex; flex-direction: column; }
.list-header, .list-item {
  display: grid;
  grid-template-columns: 24px 1fr 90px 150px 220px;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 13px;
}
.list-header {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
  border-bottom: 1px solid var(--glass-border);
  padding-top: 10px; padding-bottom: 10px;
}
.list-header > span { display: block; }
.list-header .col-icon { width: 24px; }
.list-header .col-size { text-align: right; padding-right: 0; }
.list-header .col-time { text-align: right; padding-right: 0; }
.list-header .col-actions { text-align: right; padding-right: 0; }
.list-item {
  border-bottom: 1px solid var(--glass-border);
  cursor: pointer;
  transition: background var(--t-fast) var(--ease);
  animation: fadeIn 250ms var(--ease) both;
  animation-delay: var(--delay, 0ms);
}
.list-item:last-child { border-bottom: none; }
.list-item:hover { background: var(--bg-hover); }
.list-icon {
  display: flex; align-items: center; justify-content: center;
  color: var(--text-tertiary);
}
.list-icon.dir { color: var(--accent); }
.list-icon.locked { color: var(--warning); }
.list-icon.bin { color: #a855f7; }
.list-name { font-size: 13px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.list-size { font-size: 12px; color: var(--text-tertiary); font-family: var(--font-mono); text-align: right; }
.list-time { font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono); text-align: right; }

/* List row actions - 常驻显示 */
.list-actions {
  display: flex; gap: 4px; justify-content: flex-end;
  opacity: 1;
}
.row-action {
  display: inline-flex; align-items: center; gap: 3px;
  padding: 3px 8px;
  font-size: 11px; color: var(--text-secondary);
  background: var(--bg-elevated);
  border: 1px solid var(--glass-border);
  border-radius: 5px;
  transition: all var(--t-fast) var(--ease);
}
.row-action:hover { color: var(--text-primary); border-color: var(--accent); background: var(--bg-hover); }

/* States */
.empty, .loading {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 60px 20px; gap: 10px;
  color: var(--text-tertiary); font-size: 13px;
}

@media (max-width: 768px) {
  .grid { grid-template-columns: repeat(auto-fill, minmax(80px, 1fr)); gap: 6px; padding: 10px; }
  .grid-item { padding: 10px 4px 8px; }
  .grid-icon { width: 32px; height: 32px; border-radius: 8px; }
  .grid-icon :deep(svg) { width: 18px !important; height: 18px !important; }
  .grid-name { font-size: 11px; }
  .grid-size { font-size: 9px; }
  .list-header, .list-item { grid-template-columns: 22px 1fr auto; padding: 7px 10px; }
  .col-size, .list-size { display: none; }
  .col-time, .list-time { display: none; }
  .col-actions { display: none; }
  .list-header .col-actions { display: block; text-align: right; }
  .list-actions { gap: 2px; }
  .grid-action-label, .row-action-label { display: none; }
  .grid-action, .row-action { padding: 5px 7px; }
  .fb-subbar { padding: 6px 10px; }
  .path-link { max-width: 80px; font-size: 11px; }
}
</style>