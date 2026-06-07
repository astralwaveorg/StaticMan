<template>
  <div class="home">
    <div class="orbs">
      <div class="orb orb-1"></div>
      <div class="orb orb-2"></div>
    </div>

    <div class="home-inner">
      <PageHeader
        :title="siteTitleCN"
        :subtitle="greeting"
        :subtitleFull="currentTimeFull"
      >
        <template #actions>
          <div class="home-search">
            <InstantSearch placeholder="搜索文件名、内容或跳转分类…" />
          </div>
        </template>
      </PageHeader>

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
              <div class="cat-name">{{ cat.name }}</div>
            </div>
            <div class="cat-footer">
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

      <div v-else-if="loading" class="gallery">
        <SkeletonLoader type="cards" :count="8" />
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
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { getCategories, getConfig, type CategoryInfo } from '../api'
import { icon } from '../icons'
import PageHeader from '../components/PageHeader.vue'
import InstantSearch from '../components/InstantSearch.vue'
import SkeletonLoader from '../components/SkeletonLoader.vue'

const router = useRouter()
const categories = ref<CategoryInfo[]>([])
const loading = ref(true)

// 站点配置
const siteTitleCN = ref('StaticMan')
const siteTitleEN = ref('')
const siteDesc = ref('')
const siteLogo = ref('/logo.svg')

const totalFiles = computed(() => categories.value.reduce((s, c) => s + c.fileCount, 0))
const totalTools = computed(() => categories.value.reduce((s, c) => s + (c.tools?.length || 0), 0))
const totalSize = computed(() => categories.value.reduce((s, c) => s + (c.size || 0), 0))

// 智能问候语
const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 5) return '🌙 夜深了，还在折腾规则？'
  if (hour < 11) return '🌅 早上好，开启高效的一天'
  if (hour < 14) return '🍛 午安，记得休息一下'
  if (hour < 18) return '☕ 下午好，工作辛苦了'
  if (hour < 22) return '🌆 晚上好，准备更新配置吗？'
  return '🌃 晚安，祝好梦'
})

// 当前时间全称
const currentTimeFull = ref('')
let timeTimer: ReturnType<typeof setInterval> | null = null

function updateTime() {
  const now = new Date()
  currentTimeFull.value = new Intl.DateTimeFormat('zh-CN', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    weekday: 'long',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(now)
}

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
  updateTime()
  timeTimer = setInterval(updateTime, 1000)
  try {
    const { data } = await getCategories()
    categories.value = data || []
  } catch {
    categories.value = []
  }

  try {
    const { data } = await getConfig()
    if (data.title_cn) siteTitleCN.value = data.title_cn
    if (data.title_en) siteTitleEN.value = data.title_en
    if (data.description) siteDesc.value = data.description
    if (data.logo) siteLogo.value = data.logo
  } catch {}

  loading.value = false
})

onBeforeUnmount(() => {
  if (timeTimer) {
    clearInterval(timeTimer)
    timeTimer = null
  }
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
  max-width: 1280px;
  margin: 0 auto;
  padding: 20px 24px 32px;
}

/* 搜索框在首页的特殊宽度处理 */
.home-search {
  flex: 0 1 auto;
  min-width: 320px;
  width: clamp(320px, 450px, 600px);
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

/* 头部：图标+名称 居左 */
.cat-header {
  display: flex; align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}
.cat-icon {
  width: 30px; height: 30px; border-radius: 7px;
  display: flex; align-items: center; justify-content: center;
  background: color-mix(in srgb, var(--accent-color) 12%, transparent);
  color: var(--accent-color);
  transition: transform var(--t-base) var(--ease-spring);
  flex-shrink: 0;
}
.cat-card:hover .cat-icon { transform: rotate(-4deg) scale(1.08); }
.cat-name { font-size: 14px; font-weight: 550; line-height: 1.3; }

.cat-footer {
  display: flex; align-items: center; justify-content: flex-end;
  gap: 8px;
  padding-top: 10px;
  border-top: 1px solid var(--glass-border);
}
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

@media (max-width: 768px) {
  .home-inner { padding: 12px 10px 20px; }
  .orbs { display: none; }
  .gallery { grid-template-columns: repeat(2, 1fr); gap: 8px; }
  .home-search { width: 100%; max-width: 100%; flex: 1 1 100%; min-width: 0; margin-top: 8px; box-sizing: border-box; }
  .stats { margin-bottom: 10px; padding: 6px 10px; }
  .stat-value { font-size: 13px; }
  .stat-label { font-size: 10px; }

  /* 移动端卡片精简 */
  .cat-card { padding: 10px 12px; }
  .cat-header { margin-bottom: 6px; }
  .cat-icon { width: 24px; height: 24px; border-radius: 5px; }
  .cat-icon :deep(svg) { width: 14px !important; height: 14px !important; }
  .cat-name { font-size: 13px; font-weight: 550; }
  .cat-footer { padding-top: 6px; }
  .cat-stats { width: 100%; justify-content: space-between; }
  .meta-item { font-size: 10px; }
  .cat-glow { display: none; }
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-20px); }
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
