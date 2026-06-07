<template>
  <div class="browse">
    <!-- Unified Page Header -->
    <PageHeader
      v-if="currentPath"
      :title="currentName"
      :subtitle="coverDesc"
      :icon="coverIcon"
      :accentColor="categoryAccent"
      :showBack="currentPath.includes('/')"
      @back="goUp"
    >
      <template #actions>
        <div class="browse-search">
          <InstantSearch placeholder="搜索当前目录或全局…" />
        </div>
      </template>
    </PageHeader>

    <!-- File browser -->
    <div v-if="!currentFile || isDirectory || fileError" class="browser fade-in">
      <div v-if="fileError" class="file-error-notice">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
        <h3 v-if="fileError.type === 'protected'">文件受保护</h3>
        <h3 v-else-if="fileError.type === 'notfound'">未找到</h3>
        <h3 v-else>出错了</h3>
        <p>{{ fileError.message }}</p>
        <button v-if="fileError.type === 'protected' && !auth.authenticated" class="btn btn-accent" @click="ui.openLogin()" style="margin-top: 14px;">登录解锁</button>
        <button v-else-if="fileError.type === 'notfound'" class="btn btn-ghost" @click="router.push('/')" style="margin-top: 14px;">返回首页</button>
      </div>
      <div v-else>
      <div class="browser-header" v-if="availableTools.length > 1 && isCategoryRoot">
        <div class="browser-tabs">
          <button
            v-for="t in availableTools"
            :key="t.key"
            class="tab"
            :class="{active: activeTool === t.key}"
            @click="switchTool(t.key)"
          >
            <span class="tab-dot" :style="{background: t.color}"></span>
            {{ t.name }}
          </button>
        </div>
      </div>
      <div class="browser-body" v-if="browserPath">
        <FileBrowser
          :root-path="browserPath"
          :active-tool="activeTool"
          :exclude-dirs="isCategoryRoot && !activeTool ? availableTools.map(t => t.key) : []"
        />
      </div>
      </div>
    </div>

    <!-- File viewer (only when actual file selected) -->
    <div v-else class="viewer-wrap fade-in">
      <FileViewer :file="currentFile" :root-path="rootPath" @request-login="ui.openLogin()" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getFile, getCategories, type FileContent, type CategoryInfo } from '../api'
import { icon } from '../icons'
import { useAuthStore } from '../stores/auth'
import { useUIStore } from '../stores/ui'
import FileViewer from '../components/FileViewer.vue'
import FileBrowser from '../components/FileBrowser.vue'
import PageHeader from '../components/PageHeader.vue'
import InstantSearch from '../components/InstantSearch.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const ui = useUIStore()
const currentFile = ref<FileContent | null>(null)
const categories = ref<CategoryInfo[]>([])

const currentPath = computed(() => {
  const p = route.params.pathMatch
  if (!p) return ''
  return Array.isArray(p) ? p.join('/') : p
})

// Detect: are we on a category root? (i.e. just /browse/<cat>)
const isCategoryRoot = computed(() => {
  const p = currentPath.value
  if (!p) return false
  return !p.includes('/')
})

const activeCategory = computed(() => {
  if (!isCategoryRoot.value) return null
  return categories.value.find(c => c.key === currentPath.value) || null
})

const coverDesc = computed(() => {
  if (isCategoryRoot.value && activeCategory.value) {
    return activeCategory.value.description
  }
  return rootPath.value // 父级路径
})
const coverIcon = computed(() => {
  const c = activeCategory.value
  if (c) return icon(c.icon, { size: 22 })
  return icon('folder', { size: 22 })
})
const categoryAccent = computed(() => activeCategory.value?.color || '#7c3aed')

const currentName = computed(() => currentPath.value.split('/').pop() || '')

function goUp() {
  if (!currentPath.value.includes('/')) return
  const parent = currentPath.value.split('/').slice(0, -1).join('/')
  router.push('/browse/' + parent)
}

// Tools (sub-directories) within the category
const availableTools = computed(() => {
  if (!isCategoryRoot.value || !activeCategory.value) return []
  return (activeCategory.value.tools || []).map(key => {
    const found = categories.value.find(c => c.key === key)
    return {
      key,
      name: found?.name || prettifyName(key),
      color: found?.color || '#7c3aed',
    }
  })
})

const activeTool = ref<string>('')

const rootPath = computed(() => {
  if (isCategoryRoot.value) {
    return activeTool.value
      ? `${currentPath.value}/${activeTool.value}`
      : currentPath.value
  }
  return currentPath.value
})

const browserPath = computed(() => {
  if (isCategoryRoot.value && !activeTool.value) {
    return currentPath.value
  }
  return rootPath.value
})

function switchTool(key: string) {
  activeTool.value = key
  router.push('/browse/' + currentPath.value + '/' + key)
}

function prettifyName(name: string) {
  return name.replace(/[-_]/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
}

const isDirectory = computed(() => currentFile.value?.type === 'directory')
const fileError = ref<{ type: 'protected' | 'notfound' | 'other'; message: string } | null>(null)

async function loadFile(path: string) {
  if (!path) {
    currentFile.value = null
    fileError.value = null
    return
  }
  if (!path.includes('/')) {
    currentFile.value = null
    fileError.value = null
    return
  }
  fileError.value = null
  try {
    const { data } = await getFile(path)
    currentFile.value = data
  } catch (e: any) {
    currentFile.value = null
    if (e?.response?.status === 403) {
      fileError.value = { type: 'protected', message: '此文件受保护，请登录后查看' }
    } else if (e?.response?.status === 404) {
      fileError.value = { type: 'notfound', message: '文件不存在' }
    } else {
      fileError.value = { type: 'other', message: '加载失败' }
    }
  }
}

async function loadCategories() {
  try {
    const { data } = await getCategories()
    categories.value = data
  } catch {}
}

watch(currentPath, (p) => {
  if (p.includes('/')) {
    const parts = p.split('/')
    if (parts.length === 2 && isCategoryRoot.value === false) {
      activeTool.value = parts[1]
    }
  } else {
    activeTool.value = ''
  }
  loadFile(p)
}, { immediate: true })

watch(() => route.fullPath, loadCategories, { immediate: true })
</script>

<style scoped>
.browse { height: 100%; overflow-y: auto; display: flex; flex-direction: column; align-items: center; }

.browse-search {
  flex: 0 1 auto;
  min-width: 280px;
  width: clamp(280px, 400px, 500px);
}

.viewer-wrap {
  flex: 1;
  padding: 12px 24px 24px;
  max-width: 1280px;
  margin: 0 auto;
  width: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.browser {
  flex: 1;
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 24px 24px;
  width: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.browser-header {
  margin-bottom: 10px;
  flex-shrink: 0;
}
.browser-tabs {
  display: flex; gap: 4px;
  padding: 4px;
  background: var(--bg-elevated);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
  width: fit-content;
  max-width: 100%;
  overflow-x: auto;
}
.tab {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 5px 12px;
  font-size: 12px; font-weight: 500;
  border-radius: 5px;
  color: var(--text-secondary);
  white-space: nowrap;
  transition: all var(--t-fast) var(--ease);
}
.tab:hover { color: var(--text-primary); background: var(--bg-hover); }
.tab.active {
  background: var(--bg-surface);
  color: var(--text-primary);
  box-shadow: var(--shadow-sm);
}
.tab-dot {
  width: 6px; height: 6px; border-radius: 50%;
}

.browser-body {
  background: var(--bg-surface);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  min-height: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.file-error-notice {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 60px 20px; text-align: center;
  flex: 1; gap: 8px;
}
.file-error-notice svg { color: var(--warning); }
.file-error-notice h3 { font-size: 18px; font-weight: 600; }
.file-error-notice p { color: var(--text-secondary); font-size: 13px; max-width: 400px; }
.file-error-notice .btn { padding: 8px 20px; font-size: 13px; }

@media (max-width: 768px) {
  .browse-search { flex: 1 1 100%; min-width: 0; max-width: none; order: 2; margin-right: 0; margin-top: 8px; }
  .browser { padding: 0 10px 16px; }
  .viewer-wrap { padding: 8px 10px 16px; }
  .browser-body { border-radius: var(--radius); }
  .tab { padding: 4px 8px; font-size: 11px; }
  .tab-dot { width: 5px; height: 5px; }
}
</style>
