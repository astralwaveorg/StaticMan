<template>
  <div class="viewer">
    <!-- Floating top bar -->
    <div class="viewer-bar glass">
      <div class="file-path-row">
        <div class="file-type-icon" :class="{protected: file.protected, binary: file.isBinary, dir: file.type==='directory'}">
          <svg v-if="file.isBinary" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M7 7h10M7 12h10M7 17h6"/></svg>
          <svg v-else-if="file.protected" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          <svg v-else-if="file.type==='directory'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
        </div>
        <span class="file-name">{{ file.name }}</span>
        <span v-if="file.isBinary" class="badge badge-bin">二进制</span>
        <span v-if="file.protected" class="badge badge-warn">受保护</span>
        <span class="file-size">{{ fmtSize(file.size) }}</span>
      </div>
      <div class="file-actions">
        <button class="action-btn" :class="{active: isFavorite(file.path)}" @click="toggleFavorite({path: file.path, name: file.name, type: 'file'})" :title="isFavorite(file.path) ? '取消收藏' : '收藏'">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
          <span class="action-text">{{ isFavorite(file.path) ? '已收藏' : '收藏' }}</span>
        </button>
        <button class="action-btn" :class="{copied: copied==='raw'}" @click="copyRaw" :disabled="file.type==='directory' || (file.protected && !isLoggedIn())" :title="file.protected && !isLoggedIn() ? '需要登录' : '复制 Raw URL'">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
          <span class="action-text">{{ copied==='raw' ? '已复制' : 'Raw' }}</span>
        </button>
        <button class="action-btn" :class="{copied: copied==='path'}" @click="copyPath" title="复制路径">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
          <span class="action-text">{{ copied==='path' ? '已复制' : '路径' }}</span>
        </button>
        <a v-if="file.type !== 'directory'" :href="rawUrl" target="_blank" rel="noopener" class="action-btn" :class="{disabled: file.protected && !isLoggedIn()}" :tabindex="file.protected && !isLoggedIn() ? -1 : 0" :aria-disabled="file.protected && !isLoggedIn()" :title="file.protected && !isLoggedIn() ? '需要登录' : '新窗口打开'">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
        </a>
        <a v-if="file.type !== 'directory'" :href="rawUrl" :download="file.name" class="action-btn action-accent" :class="{disabled: file.protected && !isLoggedIn()}" :tabindex="file.protected && !isLoggedIn() ? -1 : 0" :aria-disabled="file.protected && !isLoggedIn()" :title="file.protected && !isLoggedIn() ? '需要登录' : '下载'">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
        </a>
      </div>
    </div>

    <!-- Body -->
    <div class="viewer-body">
      <!-- Binary file -->
      <div v-if="file.isBinary" class="binary-notice">
        <div class="binary-mark">
          <svg width="44" height="44" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <path d="M7 7h10M7 12h10M7 17h6"/>
            <line x1="3" y1="3" x2="21" y2="21" stroke-dasharray="2 2" opacity="0.3"/>
          </svg>
        </div>
        <h3>无法预览此文件</h3>
        <p>{{ file.name }} 是二进制文件</p>
        <div class="binary-actions">
          <a :href="rawUrl" :download="file.name" class="btn btn-accent">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            下载
          </a>
          <button class="btn btn-ghost" @click="copyRaw">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
            复制 Raw URL
          </button>
        </div>
        <div class="binary-info">
          <div><span>大小</span><strong>{{ fmtSize(file.size) }}</strong></div>
          <div v-if="file.modTime"><span>修改时间</span><strong>{{ file.modTime }}</strong></div>
          <div><span>路径</span><strong class="path-text">{{ file.path }}</strong></div>
        </div>
      </div>

      <!-- Text preview -->
      <div v-else-if="file.content" class="text-view">
        <div class="code-wrap">
          <div class="line-gutter">
            <div v-for="(_, i) in contentLines" :key="i" class="ln">{{ i + 1 }}</div>
          </div>
          <pre class="code-block"><code :class="`language-${file.language}`" v-html="highlighted"></code></pre>
        </div>
        <div v-if="file.truncated" class="trunc-notice">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          文件过大，已截断至前 1000 行
          <a :href="rawUrl" :download="file.name" class="dl-link">下载完整文件</a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { FileContent } from '../api'
import { getRawUrl, isLoggedIn } from '../api'
import { useToast } from '../composables/useToast'
import { useFavorites } from '../composables/useFavorites'
import hljs from 'highlight.js'

const props = defineProps<{ file: FileContent }>()
const toast = useToast()
const { isFavorite, toggleFavorite } = useFavorites()

const copied = ref<'raw' | 'path' | null>(null)

const highlighted = computed(() => {
  if (!props.file.content) return ''
  try {
    return hljs.highlight(props.file.content, { language: props.file.language || 'plaintext' }).value
  } catch {
    return hljs.highlightAuto(props.file.content).value
  }
})

const contentLines = computed(() => props.file.content?.split('\n') || [])
const rawUrl = computed(() => getRawUrl(props.file.path, props.file.protected, true))

function fmtSize(b: number): string {
  if (!b && b !== 0) return '—'
  if (b < 1024) return `${b} B`
  if (b < 1048576) return `${(b/1024).toFixed(1)} KB`
  return `${(b/1048576).toFixed(1)} MB`
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    return true
  }
}

async function copyRaw() {
  if (props.file.protected && !isLoggedIn()) {
    toast.error('此文件受保护，请先登录')
    return
  }
  const url = getRawUrl(props.file.path, props.file.protected, true)
  if (await copyToClipboard(url)) {
    copied.value = 'raw'
    toast.success('Raw URL 已复制')
    setTimeout(() => { if (copied.value === 'raw') copied.value = null }, 1500)
  }
}

async function copyPath() {
  if (await copyToClipboard(props.file.path)) {
    copied.value = 'path'
    toast.success('路径已复制')
    setTimeout(() => { if (copied.value === 'path') copied.value = null }, 1500)
  }
}
</script>

<style scoped>
.viewer { display: flex; flex-direction: column; height: 100%; min-width: 0; }

.viewer-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 14px; gap: 10px; flex-shrink: 0;
  border-radius: var(--radius);
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.file-path-row { display: flex; align-items: center; gap: 8px; min-width: 0; flex: 1; }
.file-type-icon {
  width: 26px; height: 26px; border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  background: var(--bg-elevated); color: var(--text-tertiary); flex-shrink: 0;
}
.file-type-icon.protected { background: rgba(251,191,36,0.08); color: var(--warning); }
.file-type-icon.binary { background: rgba(168,85,247,0.08); color: #a855f7; }
.file-type-icon.dir { background: var(--accent-muted); color: var(--accent); }
.file-name {
  font-family: var(--font-mono);
  font-size: 13px; font-weight: 500;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

.badge { font-size: 10px; padding: 2px 7px; border-radius: 10px; font-weight: 500; flex-shrink: 0; }
.badge-warn { background: rgba(251,191,36,0.08); color: var(--warning); }
.badge-bin { background: rgba(168,85,247,0.08); color: #a855f7; }
.file-size { font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono); flex-shrink: 0; }

.file-actions { display: flex; align-items: center; gap: 4px; flex-shrink: 0; }
.action-btn {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 5px 8px; border-radius: 6px;
  font-size: 11px; font-weight: 500; cursor: pointer;
  background: var(--bg-elevated); color: var(--text-secondary);
  border: 1px solid var(--glass-border);
  transition: all var(--t-fast) var(--ease);
  text-decoration: none; white-space: nowrap;
}
.action-btn:hover { background: var(--bg-hover); color: var(--text-primary); }
.action-btn:disabled, .action-btn.disabled { opacity: 0.4; cursor: not-allowed; pointer-events: none; }
.action-btn.copied { background: var(--success); color: white; border-color: var(--success); }
.action-accent { background: var(--accent); color: white; border-color: var(--accent); }
.action-accent:hover { background: var(--accent-hover); }

.viewer-body { flex: 1; overflow: auto; min-height: 0; }

/* Binary */
.binary-notice {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 100%; padding: 40px 20px; text-align: center;
  max-width: 460px; margin: 0 auto;
}
.binary-mark {
  width: 80px; height: 80px;
  border-radius: 18px;
  background: rgba(168,85,247,0.08);
  color: #a855f7;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}
.binary-notice h3 { font-size: 16px; font-weight: 600; margin-bottom: 4px; }
.binary-notice p { font-size: 13px; color: var(--text-secondary); margin-bottom: 20px; }
.binary-actions { display: flex; gap: 8px; margin-bottom: 24px; }

.btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 8px 16px; border-radius: 8px;
  font-size: 13px; font-weight: 500; cursor: pointer;
  transition: all var(--t-fast) var(--ease);
  text-decoration: none;
}
.btn-accent { background: var(--accent); color: white; }
.btn-accent:hover { background: var(--accent-hover); }
.btn-ghost { background: var(--bg-elevated); color: var(--text-secondary); border: 1px solid var(--glass-border); }
.btn-ghost:hover { background: var(--bg-hover); color: var(--text-primary); }

.binary-info {
  width: 100%;
  display: flex; flex-direction: column; gap: 6px;
  padding: 14px 16px;
  background: var(--bg-surface);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
  font-size: 12px;
}
.binary-info > div { display: flex; justify-content: space-between; align-items: center; }
.binary-info span { color: var(--text-tertiary); }
.binary-info strong { font-weight: 500; color: var(--text-primary); font-family: var(--font-mono); }
.binary-info .path-text { font-size: 11px; opacity: 0.85; }

/* Text view */
.text-view { display: flex; flex-direction: column; min-height: 100%; background: var(--bg-surface); border: 1px solid var(--glass-border); border-radius: var(--radius-lg); overflow: hidden; }
.code-wrap { display: flex; flex: 1; min-height: 0; }
.line-gutter {
  padding: 16px 0;
  min-width: 48px;
  text-align: right;
  user-select: none;
  border-right: 1px solid var(--glass-border);
  background: var(--bg-elevated);
  position: sticky; left: 0; z-index: 1;
  overflow: hidden;
}
.ln {
  font-family: var(--font-mono);
  font-size: 12px; line-height: 1.65;
  padding: 0 12px 0 8px; color: var(--text-tertiary);
}
.code-block {
  margin: 0; padding: 16px 20px;
  background: var(--code-bg);
  font-family: var(--font-mono); font-size: 13px; line-height: 1.65;
  overflow-x: auto; flex: 1;
}

.trunc-notice {
  display: flex; align-items: center; justify-content: center; gap: 6px;
  padding: 10px; text-align: center; color: var(--warning); font-size: 12px;
  background: rgba(251,191,36,0.04); border-top: 1px solid rgba(251,191,36,0.08);
  flex-wrap: wrap;
}
.dl-link { color: var(--accent); text-decoration: underline; }

@media (max-width: 768px) {
  .viewer-bar {
    position: sticky;
    bottom: 0;
    top: auto;
    margin-bottom: 0;
    border-radius: 0;
    border-top: 1px solid var(--glass-border);
    background: var(--glass-bg);
    backdrop-filter: blur(20px);
    padding: 8px 10px;
    z-index: 10;
    flex-direction: row;
    align-items: center;
    padding-bottom: max(8px, env(safe-area-inset-bottom));
  }
  .file-actions {
    justify-content: space-around;
    width: 100%;
  }
  .action-btn {
    flex-direction: column;
    gap: 2px;
    padding: 6px 8px;
    font-size: 10px;
    min-height: 44px;
    min-width: 44px;
    justify-content: center;
  }
  .action-text { display: block !important; }
  .action-btn.active { color: var(--warning); background: rgba(251,191,36,0.08); border-color: rgba(251,191,36,0.2); }
  .line-gutter { display: none; }
  .code-block {
    font-size: 14px;
    padding: 12px 14px;
    line-height: 1.7;
    -webkit-overflow-scrolling: touch;
    overscroll-behavior-x: contain;
  }
}
</style>