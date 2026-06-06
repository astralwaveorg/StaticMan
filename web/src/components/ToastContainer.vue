<template>
  <Teleport to="body">
    <div class="toast-container">
      <TransitionGroup name="toast">
        <div
          v-for="t in toasts"
          :key="t.id"
          class="toast"
          :class="t.type"
        >
          <span class="toast-icon">
            <svg v-if="t.type === 'success'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M20 6L9 17l-5-5"/></svg>
            <svg v-else-if="t.type === 'error'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
            <svg v-else-if="t.type === 'warning'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
          </span>
          <span class="toast-message">{{ t.message }}</span>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { toasts } from '../composables/useToast'
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: 16px;
  right: 16px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: var(--radius);
  font-size: 13px;
  font-weight: 500;
  background: var(--bg-surface);
  border: 1px solid var(--glass-border);
  box-shadow: var(--shadow-lg);
  pointer-events: auto;
  min-width: 240px;
  max-width: 400px;
  backdrop-filter: blur(20px);
}

.toast-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.toast-message {
  flex: 1;
  min-width: 0;
  word-break: break-word;
}

.toast.success { color: var(--success); border-color: rgba(52,211,153,0.2); background: rgba(52,211,153,0.06); }
.toast.error   { color: var(--danger);  border-color: rgba(248,113,113,0.2); background: rgba(248,113,113,0.06); }
.toast.warning { color: var(--warning); border-color: rgba(251,191,36,0.2);  background: rgba(251,191,36,0.06); }
.toast.info    { color: var(--accent);  border-color: var(--accent-muted);   background: var(--accent-muted); }

/* Animations */
.toast-enter-active,
.toast-leave-active {
  transition: all 300ms var(--ease);
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(30px) scale(0.96);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(30px) scale(0.96);
}

@media (max-width: 640px) {
  .toast-container {
    top: auto;
    bottom: 16px;
    right: 12px;
    left: 12px;
    align-items: center;
  }
  .toast {
    width: 100%;
    max-width: none;
    justify-content: center;
  }
  .toast-enter-from {
    opacity: 0;
    transform: translateY(30px) scale(0.96);
  }
  .toast-leave-to {
    opacity: 0;
    transform: translateY(30px) scale(0.96);
  }
}
</style>
