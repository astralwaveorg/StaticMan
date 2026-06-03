<template>
  <div class="fb">
    <!-- Sub-header with breadcrumb within this view -->
    <div class="fb-subbar" v-if="showPath">
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
      <a
        v-for="(item, i) in items"
        :key="item.path"
        class="grid-item"
        :class="{locked: item.protected, dir: item.type==='directory'}"
        :style="{'--delay': `${i * 30}ms`}"
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
    </div>

    <!-- List view -->
    <div v-else class="list">
      <div class="list-header">
        <span class="col-name">名称</span>
        <span class="col-size">大小</span>
        <span class="col-time">修改时间</span>
      </div>
      <a
        v-for="(item, i) in items"
        :key="item.path"
        class="list-item"
        :class="{locked: item.protected, dir: item.type==='directory'}"
        :style="{'--delay': `${i * 20}ms`}"
        @click.prevent="openItem(item)"
      >
        <div class="list-icon" :class="{locked: item.protected, dir: item.type==='directory', bin: item.isBinary}">
          <svg v-if="item.protected" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          <svg v-else-if="item.type==='directory'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
          <svg v-else-if="item.isBinary" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M7 7h10M7 12h10M7 17h6"/></svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
        </div>
        <span class="list-name truncate">{{ item.name }}</span>
        <span class="list-size">{{ fmtSize(item.size) }}</span>
        <span class="list-time">{{ item.modTime || '—' }}</span>
      </a>
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
import { getLs, type LsItem } from '../api'

const props = defineProps<{ rootPath: string; activeTool: string }>()
const router = useRouter()

const items = ref<LsItem[]>([])
const loading = ref(false)
const viewMode = ref<'grid' | 'list'>((localStorage.getItem('fb_view') as 'grid' | 'list') || 'grid')

const path = computed(() => props.rootPath)
// 当 activeTool 为空时（即处在分类根），不显示内部 subbar（由 cover 显示分类名）
const showPath = computed(() => !!path.value && !!props.activeTool)
const currentName = computed(() => {
  if (!path.value) return ''
  return path.value.split('/').pop() || ''
})
const parentLabel = computed(() => {
  if (!path.value || !path.value.includes('/')) return ''
  return path.value.split('/').slice(0, -1).pop() || ''
})

async function load(p: string) {
  loading.value = true
  try {
    const { data } = await getLs(p)
    items.value = data.items
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

watch(viewMode, (v) => localStorage.setItem('fb_view', v))
watch(() => path.value, (p) => load(p), { immediate: true })
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
  width: 26px; height: 26px;
  border-radius: 5px;
  color: var(--text-tertiary);
  display: flex; align-items: center; justify-content: center;
  transition: all var(--t-fast) var(--ease);
}
.view-toggle:hover { background: var(--bg-hover); color: var(--text-primary); }
.view-toggle.active { background: var(--bg-hover); color: var(--accent); }

/* Grid */
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 6px;
  padding: 12px;
}
.grid-item {
  display: flex; flex-direction: column; align-items: center;
  padding: 14px 6px 8px;
  background: var(--bg-elevated);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
  text-decoration: none; color: inherit;
  cursor: pointer;
  transition: all var(--t-base) var(--ease);
  animation: fadeInUp 350ms var(--ease) both;
  animation-delay: var(--delay, 0ms);
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
.grid-item:hover .grid-icon { transform: scale(1.1); }
.grid-icon.dir { color: var(--accent); background: color-mix(in srgb, var(--accent) 12%, var(--bg-surface)); }
.grid-icon.locked { color: var(--warning); background: rgba(251,191,36,0.1); }
.grid-icon.bin { color: #a855f7; background: rgba(168,85,247,0.1); }

.grid-name { font-size: 12px; font-weight: 500; max-width: 100%; }
.grid-size { font-size: 10px; color: var(--text-tertiary); margin-top: 2px; font-family: var(--font-mono); }

/* List */
.list { display: flex; flex-direction: column; }
.list-header, .list-item {
  display: grid;
  grid-template-columns: 24px 1fr 90px 140px;
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
.list-item {
  border-bottom: 1px solid var(--glass-border);
  text-decoration: none; color: inherit;
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
.list-name { font-size: 13px; }
.list-size { font-size: 12px; color: var(--text-tertiary); font-family: var(--font-mono); text-align: right; }
.list-time { font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono); text-align: right; }

/* States */
.empty, .loading {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 60px 20px; gap: 10px;
  color: var(--text-tertiary); font-size: 13px;
}

@media (max-width: 768px) {
  .grid { grid-template-columns: repeat(auto-fill, minmax(90px, 1fr)); gap: 6px; padding: 12px; }
  .grid-item { padding: 12px 6px 8px; }
  .grid-icon { width: 36px; height: 36px; }
  .list-header, .list-item { grid-template-columns: 24px 1fr 70px; padding: 8px 12px; }
  .col-time, .list-time { display: none; }
  .list-size { font-size: 11px; }
}
</style>