<template>
  <div class="browse">
    <!-- Compact cover for current directory -->
    <div v-if="currentPath" class="cover" :style="{'--accent': categoryAccent}">
      <div class="cover-glow"></div>
      <div class="cover-content">
        <button class="cover-back" v-if="currentPath.includes('/')" @click="goUp" :title="`返回上一级`">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M15 18l-6-6 6-6"/></svg>
        </button>
        <div class="cover-icon" v-html="coverIcon" :style="{color: categoryAccent}"></div>
        <div class="cover-text">
          <h1 class="cover-title">{{ currentName }}</h1>
          <p class="cover-desc">{{ coverDesc }}</p>
        </div>
      </div>
    </div>

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
      <div class="browser-body">
        <FileBrowser :root-path="browserPath" :active-tool="activeTool" />
      </div>
      </div>
    </div>

    <!-- File viewer (only when actual file selected) -->
    <div v-else class="viewer-wrap fade-in">
      <FileViewer :file="currentFile" :root-path="rootPath" />
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

// Default tool: first one if available
const activeTool = ref<string>('')

// root path for file browser
const rootPath = computed(() => {
  if (isCategoryRoot.value) {
    return activeTool.value
      ? `${currentPath.value}/${activeTool.value}`
      : currentPath.value
  }
  return currentPath.value
})

// path passed to FileBrowser — excludes the category root when no tool is selected,
// to avoid showing the category's own tools in FileBrowser (they're shown as tabs)
const browserPath = computed(() => {
  if (isCategoryRoot.value && !activeTool.value) {
    return '' // FileBrowser with empty path shows nothing; tabs handle the display
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
  // 单层路径是分类根：交由 FileBrowser 显示工具列表
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
  // detect active tool from path
  if (p.includes('/')) {
    const parts = p.split('/')
    if (parts.length === 2 && isCategoryRoot.value === false) {
      // /browse/<cat>/<tool>
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

/* ── Compact cover ── */
.cover {
  position: relative;
  padding: 12px 24px;
  width: fit-content;
  margin: 0 auto;
  overflow: visible;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}
.cover-glow {
  position: absolute;
  top: -100px; left: 50%; transform: translateX(-50%);
  width: 400px; height: 180px;
  background: radial-gradient(ellipse, var(--accent) 0%, transparent 60%);
  opacity: 0.12;
  filter: blur(50px);
  pointer-events: none;
}
.cover-content {
  position: relative; z-index: 1;
  display: flex; align-items: center; gap: 10px;
  justify-content: center;
  flex-direction: row;
  width: fit-content;
}
.cover-back {
  display: flex; align-items: center; justify-content: center;
  width: 30px; height: 30px;
  border-radius: 8px;
  background: var(--bg-elevated);
  color: var(--text-secondary);
  border: 1px solid var(--glass-border);
  cursor: pointer;
  transition: all var(--t-fast) var(--ease);
  flex-shrink: 0;
}
.cover-back:hover { background: var(--bg-hover); color: var(--text-primary); border-color: var(--accent); }
.cover-icon {
  width: 36px; height: 36px;
  border-radius: 9px;
  background: color-mix(in srgb, var(--accent) 12%, transparent);
  display: flex; align-items: center; justify-content: center;
  border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent);
  flex-shrink: 0;
}
.cover-text { text-align: left; }
.cover-title {
  font-size: 18px; font-weight: 700;
  letter-spacing: -0.02em;
  margin: 0;
  line-height: 1.1;
  background: linear-gradient(135deg, var(--text-primary) 0%, color-mix(in srgb, var(--accent) 80%, var(--text-primary)) 100%);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  background-clip: text;
}
.cover-desc { font-size: 12px; color: var(--text-secondary); margin: 1px 0 0; }

.viewer-wrap {
  flex: 1;
  padding: 12px 24px 24px;
  max-width: 1100px;
  margin: 0 auto;
  width: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.browser {
  flex: 1;
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 24px 32px;
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
  .cover { padding: 12px 16px; }
  .cover-icon { width: 32px; height: 32px; }
  .cover-title { font-size: 16px; }
  .cover-desc { font-size: 11px; }
  .browser { padding: 0 12px 24px; }
  .viewer-wrap { padding: 8px 12px 16px; }
  .tab { padding: 4px 10px; font-size: 11px; }
}
</style>
