<template>
  <div class="page-header" :style="{'--accent': accentColor}">
    <div class="header-glow"></div>
    <div class="header-content">
      <button class="header-back" v-if="showBack" @click="$emit('back')" title="返回上一级">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M15 18l-6-6 6-6"/></svg>
      </button>
      <div v-if="icon" class="header-icon" v-html="icon" :style="{color: accentColor}"></div>
      <div class="header-text">
        <h1 class="header-title">{{ title }}</h1>
        <p class="header-subtitle" :title="subtitleFull || subtitle">{{ subtitle }}</p>
      </div>
      <div class="header-actions">
        <slot name="actions"></slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string;
  subtitle?: string;
  subtitleFull?: string;
  icon?: string;
  accentColor?: string;
  showBack?: boolean;
}>()
defineEmits(['back'])
</script>

<style scoped>
.page-header {
  position: relative;
  padding: 12px 24px;
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  overflow: visible;
  flex-shrink: 0;
}
.header-glow {
  position: absolute;
  top: -100px; left: 50%; transform: translateX(-50%);
  width: 400px; height: 180px;
  background: radial-gradient(ellipse, var(--accent, var(--accent-glow)) 0%, transparent 60%);
  opacity: 0.12;
  filter: blur(50px);
  pointer-events: none;
}
.header-content {
  position: relative; z-index: 1;
  display: flex; align-items: center; gap: 16px;
  width: 100%;
}
.header-back {
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
.header-back:hover { background: var(--bg-hover); color: var(--text-primary); border-color: var(--accent); }
.header-icon {
  width: 36px; height: 36px;
  border-radius: 9px;
  background: color-mix(in srgb, var(--accent) 12%, transparent);
  display: flex; align-items: center; justify-content: center;
  border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent);
  flex-shrink: 0;
}
.header-text { text-align: left; min-width: 0; overflow: hidden; }
.header-title {
  font-size: 18px; font-weight: 700;
  letter-spacing: -0.02em;
  margin: 0;
  line-height: 1.1;
  background: linear-gradient(135deg, var(--text-primary) 0%, color-mix(in srgb, var(--accent, var(--text-primary)) 80%, var(--text-primary)) 100%);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  background-clip: text;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.header-subtitle { font-size: 12px; color: var(--text-secondary); margin: 1px 0 0; transition: all 0.3s ease; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.header-actions { margin-left: auto; display: flex; align-items: center; gap: 12px; }

@media (max-width: 768px) {
  .page-header { padding: 10px 10px; }
  .header-content { gap: 8px; flex-wrap: wrap; }
  .header-text { flex: 1 1 auto; min-width: 120px; }
  .header-actions { margin-left: 0; flex: 1 1 100%; }
  .header-actions > * { width: 100%; }
  .header-icon { width: 28px; height: 28px; border-radius: 7px; }
  .header-title { font-size: 15px; }
  .header-subtitle { font-size: 11px; }
  .header-back { width: 26px; height: 26px; }
}
</style>
