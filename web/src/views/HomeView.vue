<template>
  <div class="home">
    <div class="orbs">
      <div class="orb orb-1"></div>
      <div class="orb orb-2"></div>
    </div>

    <div class="home-inner">
      <section class="hero">
        <h1>欢迎回来</h1>
        <p class="hero-subtitle">集中管理 · 统一分发 · 便捷共享</p>
      </section>

      <div class="stats" v-if="categories.length">
        <div class="stat">
          <span class="stat-value">{{ categories.length }}</span>
          <span class="stat-label">分类</span>
        </div>
        <div class="stat-sep"></div>
        <div class="stat">
          <span class="stat-value">{{ totalFiles }}</span>
          <span class="stat-label">文件</span>
        </div>
        <div class="stat-sep"></div>
        <div class="stat">
          <span class="stat-value">{{ totalTools }}</span>
          <span class="stat-label">工具</span>
        </div>
        <div class="stat-sep"></div>
        <div class="stat">
          <span class="stat-value">{{ fmtSize(totalSize) }}</span>
          <span class="stat-label">总大小</span>
        </div>
      </div>

      <section class="gallery" v-if="categories.length">
        <div
          v-for="(cat, i) in categories"
          :key="cat.key"
          class="cat-card"
          :style="{'--delay': `${i * 60}ms`, '--accent-color': cat.color}"
          @click="goCategory(cat.key)"
        >
          <div class="cat-glow"></div>
          <div class="cat-content">
            <div class="cat-header">
              <div class="cat-icon" v-html="icon(cat.icon, { size: 18 })"></div>
              <svg class="cat-arrow" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
            </div>
            <div class="cat-name">{{ cat.name }}</div>
            <div class="cat-desc">{{ cat.description }}</div>
            <div class="cat-footer">
              <div class="cat-tools" v-if="cat.tools && cat.tools.length">
                <span v-for="t in cat.tools.slice(0, 3)" :key="t" class="tool-pill">{{ t }}</span>
                <span v-if="cat.tools.length > 3" class="tool-more">+{{ cat.tools.length - 3 }}</span>
              </div>
              <div class="cat-stats">
                <span class="meta-item" :title="`${cat.fileCount} 个文件`">
                  <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                  {{ cat.fileCount }}
                </span>
                <span v-if="cat.size" class="meta-item" :title="`总大小`">
                  <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                  {{ fmtSize(cat.size) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div v-else-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>加载中…</p>
      </div>

      <div v-else class="empty-state">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.4"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
        <h3>暂无数据</h3>
        <p>在 <code>data/</code> 目录下添加文件夹即可自动出现</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getCategories, type CategoryInfo } from '../api'
import { icon } from '../icons'

const router = useRouter()
const categories = ref<CategoryInfo[]>([])
const loading = ref(true)

const totalFiles = computed(() => categories.value.reduce((s, c) => s + c.fileCount, 0))
const totalTools = computed(() => categories.value.reduce((s, c) => s + (c.tools?.length || 0), 0))
const totalSize = computed(() => categories.value.reduce((s, c) => s + (c.size || 0), 0))

function goCategory(key: string) {
  router.push('/browse/' + key)
}

function fmtSize(b: number): string {
  if (b < 1024) return `${b} B`
  if (b < 1048576) return `${(b / 1024).toFixed(1)} KB`
  if (b < 1073741824) return `${(b / 1048576).toFixed(1)} MB`
  return `${(b / 1073741824).toFixed(2)} GB`
}

onMounted(async () => {
  try {
    const { data } = await getCategories()
    categories.value = data
  } catch {}
  loading.value = false
})
</script>

<style scoped>
.home { position: relative; height: 100%; overflow-y: auto; }

.orbs {
  position: fixed; inset: 0;
  pointer-events: none; overflow: hidden;
  z-index: 0;
}
.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  animation: float 8s ease-in-out infinite;
}
.orb-1 { width: 320px; height: 320px; background: rgba(124, 58, 237, 0.15); top: -100px; right: -80px; }
.orb-2 { width: 280px; height: 280px; background: rgba(168, 85, 247, 0.1); bottom: -80px; left: -60px; animation-delay: 2s; }

.home-inner {
  position: relative; z-index: 1;
  max-width: 1100px;
  margin: 0 auto;
  padding: 20px 24px 32px;
}

/* Hero */
.hero {
  display: flex; align-items: baseline; gap: 16px; flex-wrap: wrap;
  margin-bottom: 16px;
  padding: 0 4px;
}
.hero h1 {
  font-size: 20px; font-weight: 700;
  letter-spacing: -0.02em;
  margin: 0;
  background: linear-gradient(135deg, var(--text-primary) 0%, var(--accent-hover) 100%);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  background-clip: text;
}
.hero-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  display: flex; align-items: center; gap: 6px;
  flex-wrap: wrap;
}
.kbd-hint { display: inline-flex; align-items: center; gap: 2px; }
.kbd-hint kbd {
  font-family: var(--font-sans);
  font-size: 10px;
  padding: 1px 5px;
  background: var(--bg-elevated);
  border: 1px solid var(--glass-border);
  border-radius: 4px;
  color: var(--text-tertiary);
  line-height: 1.4;
  font-weight: 500;
}

/* Stats */
.stats {
  display: flex; align-items: center; justify-content: center;
  gap: 16px;
  padding: 10px 16px;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
  backdrop-filter: blur(20px);
  margin-bottom: 14px;
  flex-wrap: wrap;
}
.stat { display: inline-flex; align-items: baseline; gap: 4px; }
.stat-value { font-size: 15px; font-weight: 700; letter-spacing: -0.01em; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-tertiary); }
.stat-sep { width: 1px; height: 14px; background: var(--glass-border); }

@media (max-width: 640px) {
  .stats { gap: 10px; padding: 8px 12px; }
  .stat-sep { display: none; }
}

/* Gallery */
.gallery {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 10px;
}

.cat-card {
  --accent-color: #7c3aed;
  position: relative;
  padding: 14px 16px;
  background: var(--bg-surface);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
  cursor: pointer;
  overflow: hidden;
  transition: all var(--t-base) var(--ease);
  animation: fadeInUp 400ms var(--ease) both;
  animation-delay: var(--delay, 0ms);
}
.cat-card::before {
  content: '';
  position: absolute; inset: 0;
  background: linear-gradient(135deg, transparent 0%, color-mix(in srgb, var(--accent-color) 6%, transparent) 100%);
  opacity: 0;
  transition: opacity var(--t-base) var(--ease);
  pointer-events: none;
}
.cat-card:hover {
  border-color: color-mix(in srgb, var(--accent-color) 40%, var(--glass-border));
  transform: translateY(-1px);
}
.cat-card:hover::before { opacity: 1; }

.cat-glow {
  position: absolute;
  top: -50%; right: -50%;
  width: 160px; height: 160px;
  background: radial-gradient(circle, var(--accent-color) 0%, transparent 70%);
  opacity: 0;
  filter: blur(40px);
  transition: opacity var(--t-base) var(--ease);
  pointer-events: none;
}
.cat-card:hover .cat-glow { opacity: 0.3; }

.cat-content { position: relative; z-index: 1; }

.cat-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 10px;
}
.cat-icon {
  width: 30px; height: 30px; border-radius: 7px;
  display: flex; align-items: center; justify-content: center;
  background: color-mix(in srgb, var(--accent-color) 12%, transparent);
  color: var(--accent-color);
  transition: transform var(--t-base) var(--ease-spring);
}
.cat-card:hover .cat-icon { transform: rotate(-4deg) scale(1.08); }
.cat-arrow {
  color: var(--text-tertiary); opacity: 0;
  transition: all var(--t-base) var(--ease);
}
.cat-card:hover .cat-arrow { opacity: 1; color: var(--accent-color); transform: translateX(2px); }

.cat-name { font-size: 14px; font-weight: 550; line-height: 1.3; }
.cat-desc { font-size: 12px; color: var(--text-secondary); margin-top: 2px; line-height: 1.4;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
  min-height: 34px;
}

.cat-footer {
  display: flex; align-items: center; justify-content: space-between;
  gap: 8px;
  padding-top: 10px;
  border-top: 1px solid var(--glass-border);
}
.cat-tools { display: flex; gap: 3px; flex-wrap: wrap; flex: 1; min-width: 0; }
.tool-pill {
  font-size: 10px;
  padding: 2px 6px;
  background: var(--bg-elevated);
  border-radius: 4px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
  white-space: nowrap;
}
.tool-more { font-size: 10px; padding: 2px 5px; color: var(--text-tertiary); font-family: var(--font-mono); }
.cat-stats { display: flex; gap: 8px; flex-shrink: 0; }
.meta-item {
  display: inline-flex; align-items: center; gap: 3px;
  font-size: 10px; color: var(--text-tertiary); font-family: var(--font-mono);
}
.meta-item svg { opacity: 0.7; }

.loading-state, .empty-state {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 40px 20px;
  text-align: center;
  color: var(--text-tertiary);
  gap: 8px;
  font-size: 13px;
}
.empty-state h3 { font-size: 14px; color: var(--text-secondary); }
.empty-state code { font-family: var(--font-mono); font-size: 12px; padding: 2px 6px; background: var(--bg-elevated); border-radius: 4px; }

@media (max-width: 640px) {
  .home-inner { padding: 16px; }
  .gallery { grid-template-columns: 1fr; }
}
</style>