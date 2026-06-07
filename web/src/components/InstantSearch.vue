<template>
  <div class="search-container" ref="container">
    <div class="search-shell" :class="{ focused: isFocused }">
      <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
      <input
        v-model="query"
        class="search-input"
        type="text"
        :placeholder="placeholder"
        @focus="onFocus"
        @blur="onBlur"
        @keydown.down.prevent="moveDown"
        @keydown.up.prevent="moveUp"
        @keydown.enter.prevent="onEnter"
      />
      <div class="search-mode" v-if="!results.length">
        <button :class="{active: mode==='name'}" @click="mode='name'">文件名</button>
        <button :class="{active: mode==='content'}" @click="mode='content'">内容</button>
      </div>
      <button class="kbd-hint" @click="ui.openCommand()" title="打开命令面板 (⌘K)">
        <kbd>⌘</kbd><kbd>K</kbd>
      </button>
    </div>

    <!-- Suggestions Dropdown -->
    <transition name="fade">
      <div v-if="showSuggestions && (results.length || loading)" class="suggestions shadow-lg">
        <div v-if="loading" class="suggestion-loading">
          <div class="spinner-mini"></div> 正在搜索...
        </div>
        <template v-else>
          <div
            v-for="(item, i) in results"
            :key="item.path"
            class="suggestion-item"
            :class="{ active: i === activeIndex }"
            @mousedown="selectItem(item)"
          >
            <div class="item-icon" :class="item.type">
              <svg v-if="item.type==='directory'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
              <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            </div>
            <div class="item-text">
              <div class="item-name truncate">{{ item.name }}</div>
              <div class="item-path truncate">{{ item.path }}</div>
            </div>
          </div>
          <div v-if="results.length >= 10" class="suggestion-more" @click="ui.openCommand(query, mode)">
            查看更多结果...
          </div>
        </template>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { search, type SearchResult } from '../api'
import { useUIStore } from '../stores/ui'

defineProps<{ placeholder?: string }>()
const router = useRouter()
const ui = useUIStore()

const query = ref('')
const mode = ref<'name' | 'content'>('name')
const results = ref<SearchResult[]>([])
const loading = ref(false)
const isFocused = ref(false)
const showSuggestions = ref(false)
const activeIndex = ref(-1)
const container = ref<HTMLElement | null>(null)

let timer: any = null

watch(query, (v) => {
  if (timer) clearTimeout(timer)
  if (!v.trim()) {
    results.value = []
    return
  }
  loading.value = true
  timer = setTimeout(async () => {
    try {
      const { data } = await search(v.trim(), mode.value, 0, 10)
      results.value = data.results || []
    } catch {
      results.value = []
    }
    loading.value = false
    activeIndex.value = results.value.length > 0 ? 0 : -1
  }, 300)
})

function onFocus() {
  isFocused.value = true
  showSuggestions.value = true
}


function onBlur() {
  isFocused.value = false
  setTimeout(() => { showSuggestions.value = false }, 200)
}

function selectItem(item: SearchResult) {
  router.push('/browse/' + item.path)
  query.value = ''
  showSuggestions.value = false
}

function moveDown() {
  if (results.value.length === 0) return
  activeIndex.value = (activeIndex.value + 1) % results.value.length
}

function moveUp() {
  if (results.value.length === 0) return
  activeIndex.value = (activeIndex.value - 1 + results.value.length) % results.value.length
}

function onEnter() {
  if (activeIndex.value >= 0 && activeIndex.value < results.value.length) {
    selectItem(results.value[activeIndex.value])
  } else if (query.value.trim()) {
    ui.openCommand(query.value.trim(), mode.value)
  }
}
</script>

<style scoped>
.search-container { position: relative; width: 100%; }
.search-shell {
  display: flex; align-items: center; gap: 8px;
  background: var(--bg-surface);
  border: 1px solid var(--glass-border);
  border-radius: 9px;
  padding: 0 8px 0 12px;
  height: 38px;
  width: 100%;
  color: var(--text-secondary);
  transition: all var(--t-base) var(--ease);
}
.search-shell.focused { border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-muted); background: var(--bg-hover); }

.search-input { flex: 1; min-width: 0; font-size: 13px; color: var(--text-primary); background: transparent; border: none; outline: none; padding: 0; height: 100%; }
.search-mode { display: flex; background: var(--bg-elevated); border-radius: 5px; overflow: hidden; margin-right: 4px; }
.search-mode button { padding: 3px 8px; font-size: 10px; font-weight: 500; color: var(--text-tertiary); }
.search-mode button.active { background: var(--accent); color: white; }

.suggestions {
  position: absolute; top: calc(100% + 8px); left: 0; right: 0;
  background: var(--bg-surface);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
  z-index: 100;
  overflow: hidden;
  backdrop-filter: blur(20px);
  box-shadow: var(--shadow-lg);
}
.suggestion-item {
  display: flex; align-items: center; gap: 12px; padding: 10px 14px;
  cursor: pointer; transition: all var(--t-fast) var(--ease);
}
.suggestion-item:hover, .suggestion-item.active { background: var(--bg-hover); }
.item-icon { width: 28px; height: 28px; border-radius: 6px; display: flex; align-items: center; justify-content: center; background: var(--bg-elevated); flex-shrink: 0; }
.item-icon.directory { color: var(--accent); background: var(--accent-muted); }
.item-text { min-width: 0; flex: 1; }
.item-name { font-size: 13px; font-weight: 500; color: var(--text-primary); }
.item-path { font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono); }

.suggestion-loading { padding: 16px; text-align: center; color: var(--text-tertiary); font-size: 12px; display: flex; align-items: center; justify-content: center; gap: 8px; }
.suggestion-more { padding: 8px; text-align: center; font-size: 11px; color: var(--accent); cursor: pointer; border-top: 1px solid var(--glass-border); }

.spinner-mini { width: 12px; height: 12px; border: 2px solid transparent; border-top-color: currentColor; border-radius: 50%; animation: spin 0.6s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
