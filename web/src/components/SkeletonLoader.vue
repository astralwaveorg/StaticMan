<template>
  <div class="skeleton" :class="type">
    <!-- Cards skeleton -->
    <template v-if="type === 'cards'">
      <div v-for="i in count" :key="i" class="skeleton-card">
        <div class="skeleton-glow"></div>
        <div class="skeleton-icon"></div>
        <div class="skeleton-text"></div>
        <div class="skeleton-text short"></div>
      </div>
    </template>

    <!-- List skeleton -->
    <template v-if="type === 'list'">
      <div v-for="i in count" :key="i" class="skeleton-row">
        <div class="skeleton-icon-sm"></div>
        <div class="skeleton-text flex"></div>
        <div class="skeleton-text short"></div>
      </div>
    </template>

    <!-- Code skeleton -->
    <template v-if="type === 'code'">
      <div class="skeleton-code">
        <div v-for="i in count" :key="i" class="skeleton-line" :style="{width: randomWidth(i)}"></div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  type: 'cards' | 'list' | 'code'
  count?: number
}>()

const count = props.count || 6

function randomWidth(seed: number): string {
  const widths = ['40%', '60%', '80%', '55%', '90%', '45%', '70%', '35%', '85%', '50%']
  return widths[seed % widths.length]
}
</script>

<style scoped>
.skeleton {
  display: contents;
}

/* Card skeleton */
.skeleton-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 10px;
}
.skeleton-card {
  position: relative;
  padding: 14px 16px;
  background: var(--bg-surface);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
  overflow: hidden;
}
.skeleton-glow {
  position: absolute;
  top: -50%; right: -50%;
  width: 160px; height: 160px;
  background: radial-gradient(circle, var(--accent) 0%, transparent 70%);
  opacity: 0.05;
  filter: blur(40px);
  pointer-events: none;
}
.skeleton-icon {
  width: 30px; height: 30px;
  border-radius: 7px;
  background: var(--bg-elevated);
  margin-bottom: 10px;
  animation: pulse 1.5s ease-in-out infinite;
}
.skeleton-text {
  height: 12px;
  border-radius: 4px;
  background: var(--bg-elevated);
  margin-bottom: 6px;
  animation: pulse 1.5s ease-in-out infinite;
}
.skeleton-text.short {
  width: 40%;
}

/* List skeleton */
.skeleton-list {
  display: flex;
  flex-direction: column;
}
.skeleton-row {
  display: grid;
  grid-template-columns: 24px 1fr 90px 150px 220px;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--glass-border);
}
.skeleton-icon-sm {
  width: 16px; height: 16px;
  border-radius: 4px;
  background: var(--bg-elevated);
  animation: pulse 1.5s ease-in-out infinite;
}
.skeleton-row .skeleton-text {
  height: 11px;
}
.skeleton-row .skeleton-text.flex {
  flex: 1;
}

/* Code skeleton */
.skeleton-code {
  padding: 16px 20px;
  background: var(--code-bg);
  border-radius: var(--radius);
}
.skeleton-line {
  height: 14px;
  border-radius: 3px;
  background: var(--bg-elevated);
  margin-bottom: 8px;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

@media (max-width: 768px) {
  .skeleton-cards {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  }
  .skeleton-row {
    grid-template-columns: 22px 1fr auto;
    padding: 7px 10px;
  }
}
</style>
