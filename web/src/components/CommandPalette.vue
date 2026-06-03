<template>
  <Teleport to="body">
    <transition name="cmd">
      <div v-if="open" class="cmd-backdrop" @click.self="close">
        <div class="cmd-panel" @keydown.escape="close">
          <div class="cmd-search">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
            <input
              ref="inputRef"
              v-model="query"
              @keydown.down.prevent="move(1)"
              @keydown.up.prevent="move(-1)"
              @keydown.enter="execute"
              placeholder="搜索文件、跳转到分类…"
              class="cmd-input"
            />
            <div class="cmd-mode">
              <button :class="{active: mode==='name'}" @click="mode='name'">文件名</button>
              <button :class="{active: mode==='content'}" @click="mode='content'">内容</button>
            </div>
          </div>

          <div class="cmd-results" v-if="results.length">
            <div
              v-for="(r, i) in results"
              :key="r.path"
              class="cmd-item"
              :class="{active: i === activeIdx}"
              @mouseenter="activeIdx = i"
              @click="executeItem(r)"
            >
              <div class="cmd-icon" :class="{locked: r.protected, dir: r.type==='directory', bin: r.isBinary}">
                <svg v-if="r.protected" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                <svg v-else-if="r.type==='directory'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
                <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              </div>
              <div class="cmd-info">
                <div class="cmd-name">{{ r.name }}</div>
                <div class="cmd-path">{{ r.path }}</div>
              </div>
              <span v-if="r.matches" class="cmd-badge">{{ r.matches.length }} 匹配</span>
              <svg class="cmd-arrow" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
            </div>
          </div>

          <div class="cmd-empty" v-else-if="query.trim() && !loading">
            <p>没有找到 "{{ query }}"</p>
          </div>

          <div class="cmd-hint" v-if="!query.trim()">
            <p>开始输入以搜索 · 支持文件名和文件内容</p>
          </div>

          <div v-if="loading" class="cmd-loading">
            <div class="spinner"></div>
          </div>

          <footer class="cmd-footer">
            <span><kbd>↑↓</kbd> 切换</span>
            <span><kbd>↵</kbd> 打开</span>
            <span><kbd>esc</kbd> 关闭</span>
          </footer>
        </div>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { searchFiles, type SearchResult } from '../api'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ 'update:open': [v: boolean] }>()

const router = useRouter()
const query = ref('')
const mode = ref<'name' | 'content'>('name')
const results = ref<SearchResult[]>([])
const activeIdx = ref(0)
const loading = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)

function close() { emit('update:open', false) }

function move(delta: number) {
  if (!results.value.length) return
  activeIdx.value = (activeIdx.value + delta + results.value.length) % results.value.length
}

function executeItem(r: SearchResult) {
  router.push('/browse/' + r.path)
  close()
}

function execute() {
  if (results.value[activeIdx.value]) {
    executeItem(results.value[activeIdx.value])
  } else if (query.value.trim()) {
    // trigger search
  }
}

let debounceTimer: number | null = null
watch(query, (q) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  if (!q.trim()) {
    results.value = []
    return
  }
  debounceTimer = window.setTimeout(async () => {
    loading.value = true
    try {
      const { data } = await searchFiles(q.trim(), mode.value)
      results.value = data
      activeIdx.value = 0
    } catch {}
    loading.value = false
  }, 250)
})

watch(mode, () => {
  if (query.value.trim()) {
    query.value = query.value // trigger
  }
})

watch(() => props.open, async (v) => {
  if (v) {
    query.value = ''
    results.value = []
    activeIdx.value = 0
    await nextTick()
    inputRef.value?.focus()
  }
})
</script>

<style scoped>
.cmd-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(12px);
  z-index: 200;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 12vh 20px 20px;
}
.cmd-panel {
  width: 100%; max-width: 580px;
  background: var(--bg-surface);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  display: flex; flex-direction: column;
  max-height: 70vh;
  animation: scaleIn var(--t-base) var(--ease);
}
.cmd-search {
  display: flex; align-items: center; gap: 10px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--glass-border);
  color: var(--text-tertiary);
}
.cmd-input {
  flex: 1; border: none; background: none; outline: none;
  color: var(--text-primary);
  font-size: 15px;
  min-width: 0;
}
.cmd-input::placeholder { color: var(--text-tertiary); }
.cmd-mode {
  display: flex; gap: 0;
  border-radius: 6px;
  overflow: hidden;
  background: var(--bg-elevated);
  flex-shrink: 0;
}
.cmd-mode button {
  padding: 4px 10px;
  font-size: 11px; font-weight: 500;
  color: var(--text-tertiary);
  transition: all var(--t-fast) var(--ease);
  white-space: nowrap;
}
.cmd-mode button:hover { color: var(--text-primary); }
.cmd-mode button.active { background: var(--accent); color: white; }

.cmd-results { flex: 1; overflow-y: auto; padding: 6px; }
.cmd-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background var(--t-fast) var(--ease);
}
.cmd-item.active { background: var(--bg-elevated); }
.cmd-icon {
  width: 28px; height: 28px; border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  background: var(--bg-base);
  color: var(--text-tertiary); flex-shrink: 0;
}
.cmd-icon.locked { color: var(--warning); background: rgba(251,191,36,0.08); }
.cmd-icon.dir { color: var(--accent); }
.cmd-icon.bin { color: #a855f7; background: rgba(168,85,247,0.08); }
.cmd-info { flex: 1; min-width: 0; }
.cmd-name { font-size: 13px; font-weight: 500; }
.cmd-path { font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono); }
.cmd-badge {
  font-size: 10px; padding: 2px 7px;
  background: var(--accent-muted); color: var(--accent);
  border-radius: 10px; font-weight: 500;
}
.cmd-arrow { color: var(--text-tertiary); opacity: 0; transition: opacity var(--t-fast) var(--ease); }
.cmd-item.active .cmd-arrow { opacity: 1; }

.cmd-empty, .cmd-hint {
  padding: 32px 20px; text-align: center;
  color: var(--text-tertiary); font-size: 13px;
}
.cmd-loading {
  display: flex; justify-content: center; padding: 20px;
}

.cmd-footer {
  display: flex; gap: 14px;
  padding: 8px 14px;
  border-top: 1px solid var(--glass-border);
  font-size: 11px; color: var(--text-tertiary);
}
.cmd-footer kbd {
  font-family: var(--font-sans);
  font-size: 10px;
  padding: 1px 5px;
  background: var(--bg-elevated);
  border: 1px solid var(--glass-border);
  border-radius: 3px;
  margin-right: 4px;
}

.cmd-enter-active, .cmd-leave-active { transition: opacity var(--t-base) ease; }
.cmd-enter-active .cmd-panel { animation: scaleIn var(--t-base) var(--ease); }
.cmd-leave-active .cmd-panel { transition: transform var(--t-fast) ease, opacity var(--t-fast) ease; transform: scale(0.95); opacity: 0; }
.cmd-enter-from, .cmd-leave-to { opacity: 0; }

/* Mobile */
@media (max-width: 640px) {
  .cmd-backdrop { padding: 8vh 12px 12px; }
  .cmd-panel { max-height: 75vh; }
  .cmd-search { padding: 12px; gap: 8px; }
  .cmd-input { font-size: 14px; }
  .cmd-mode button { padding: 4px 8px; font-size: 10px; }
  .cmd-footer { gap: 10px; font-size: 10px; }
}
</style>