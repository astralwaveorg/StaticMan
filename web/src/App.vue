<template>
  <div class="app" :data-theme="theme">
    <!-- Floating top bar (glass) -->
    <header class="topbar">
      <div class="topbar-inner">
        <!-- Brand: logo + name + tagline -->
        <router-link to="/" class="brand">
          <div class="brand-mark">
            <img src="/logo.svg" alt="M" class="brand-icon" />
          </div>
          <div class="brand-text">
            <span class="brand-title">配置共享管理系统</span>
            <span class="brand-sub">MagicHub</span>
          </div>
        </router-link>

        <!-- Center: search + Ctrl K -->
        <div class="topbar-center">
          <button class="search-box" @click="openCommand" title="搜索 (Ctrl+K)">
            <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
            <span class="search-placeholder">搜索文件名、跳转分类…</span>
            <span class="kbd-hint"><kbd>⌘</kbd><kbd>K</kbd></span>
          </button>
        </div>

        <!-- Right: actions -->
        <div class="topbar-right">
          <nav class="crumbs" v-if="crumbs.length > 1">
            <template v-for="(c, i) in crumbs" :key="c.path">
              <a class="crumb" v-if="i < crumbs.length - 1" @click="navigateTo(c.path)">{{ c.name }}</a>
              <span class="crumb-sep" v-if="i < crumbs.length - 1">/</span>
              <span class="crumb-current" v-else>{{ c.name }}</span>
            </template>
          </nav>

          <button class="icon-btn theme-btn" @click="toggleTheme" :title="theme==='dark'?'亮色':'暗色'">
            <svg v-if="theme==='dark'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
            <svg v-else width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
          </button>

          <button v-if="!auth.authenticated" class="login-btn" @click="showLogin=true">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            <span>登录</span>
          </button>
          <button v-else class="user-btn" @click="auth.logout()">
            <div class="user-dot"></div>
            <span>已解锁</span>
          </button>
        </div>
      </div>
    </header>

    <!-- Main -->
    <main class="main">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" :key="route.fullPath" />
        </transition>
      </router-view>
    </main>

    <!-- Command palette -->
    <CommandPalette v-model:open="commandOpen" />

    <!-- Login modal -->
    <Teleport to="body">
      <transition name="fade">
        <div v-if="showLogin" class="modal-backdrop" @click.self="showLogin=false">
          <div class="modal">
            <div class="modal-mark">
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            </div>
            <h3>登录解锁</h3>
            <p>输入密码后可查看受保护文件</p>
            <input v-model="password" type="password" class="input" placeholder="访问密码" @keydown.enter="doLogin" ref="passwordInput" />
            <div class="modal-row">
              <button class="btn btn-ghost" @click="showLogin=false">取消</button>
              <button class="btn btn-primary" @click="doLogin" :disabled="loginLoading">
                <span v-if="loginLoading" class="spinner-mini"></span>
                {{ loginLoading ? '验证中' : '登录' }}
              </button>
            </div>
          </div>
        </div>
      </transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { getBreadcrumbs, type Breadcrumb } from './api'
import CommandPalette from './components/CommandPalette.vue'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const theme = ref((localStorage.getItem('magichub_theme') as 'dark' | 'light') || 'dark')
const showLogin = ref(false)
const password = ref('')
const loginLoading = ref(false)
const passwordInput = ref<HTMLInputElement | null>(null)
const commandOpen = ref(false)
const crumbs = ref<Breadcrumb[]>([])

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  localStorage.setItem('magichub_theme', theme.value)
}

async function doLogin() {
  if (!password.value) return
  loginLoading.value = true
  const ok = await auth.login(password.value)
  loginLoading.value = false
  if (ok) { showLogin.value = false; password.value = '' }
  else { alert('密码错误') }
}

function openCommand() { commandOpen.value = true }

async function loadCrumbs() {
  const p = route.params.pathMatch
  const s = Array.isArray(p) ? p.join('/') : (p || '')
  if (!s) {
    crumbs.value = []
    return
  }
  try {
    const { data } = await getBreadcrumbs(s)
    crumbs.value = data
  } catch {
    crumbs.value = []
  }
}

function navigateTo(path: string) {
  router.push(path === '' ? '/' : '/browse/' + path)
}

watch(() => route.fullPath, loadCrumbs, { immediate: true })

function handleKey(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    commandOpen.value = true
  }
  if (e.key === 'Escape' && commandOpen.value) {
    commandOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKey)
})
onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKey)
})

watch(showLogin, (v) => { if (v) nextTick(() => passwordInput.value?.focus()) })
</script>

<style scoped>
.app { height: 100vh; display: flex; flex-direction: column; }

/* ── Topbar ── */
.topbar {
  height: var(--header-height);
  flex-shrink: 0;
  background: var(--glass-bg);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border-bottom: 1px solid var(--glass-border);
  position: relative;
  z-index: 50;
}
.topbar-inner {
  height: 100%;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 20px;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

/* Brand */
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: var(--text-primary);
  flex-shrink: 0;
  user-select: none;
}
.brand-mark {
  width: 30px; height: 30px;
  border-radius: 8px;
  background: linear-gradient(135deg, #7c3aed 0%, #a855f7 100%);
  display: flex; align-items: center; justify-content: center;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.25);
}
.brand-icon { width: 17px; height: 17px; filter: brightness(2.2); }
.brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}
.brand-title {
  font-size: 13px;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--text-primary);
}
.brand-sub {
  font-size: 10px;
  color: var(--text-tertiary);
  letter-spacing: 0.04em;
  font-weight: 500;
}

/* Center search */
.topbar-center {
  flex: 1;
  display: flex;
  justify-content: center;
  min-width: 0;
  max-width: 520px;
  margin: 0 auto;
}
.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-elevated);
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  padding: 0 12px;
  height: 34px;
  width: 100%;
  max-width: 360px;
  color: var(--text-secondary);
  transition: all var(--t-base) var(--ease);
}
.search-box:hover {
  border-color: var(--accent);
  background: var(--bg-hover);
}
.search-box:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-muted);
}
.search-box .icon { color: var(--text-tertiary); flex-shrink: 0; }
.search-placeholder {
  flex: 1; min-width: 0;
  font-size: 13px;
  color: var(--text-tertiary);
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.kbd-hint {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}
.kbd-hint kbd {
  font-family: var(--font-sans);
  font-size: 10px;
  padding: 1px 5px;
  background: var(--bg-base);
  border: 1px solid var(--glass-border);
  border-radius: 4px;
  color: var(--text-tertiary);
  line-height: 1.4;
  font-weight: 500;
}

/* Right actions */
.topbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.crumbs {
  display: flex; align-items: center; gap: 4px;
  font-size: 12px;
  margin-right: 4px;
}
.crumb {
  color: var(--text-secondary);
  padding: 4px 8px;
  border-radius: 5px;
  cursor: pointer;
  max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  transition: all var(--t-fast) var(--ease);
}
.crumb:hover { background: var(--bg-hover); color: var(--text-primary); }
.crumb-sep { color: var(--text-tertiary); }
.crumb-current {
  color: var(--text-primary); font-weight: 500;
  padding: 4px 6px;
  max-width: 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

.icon-btn {
  background: none; border: none; color: var(--text-secondary);
  cursor: pointer; padding: 7px; border-radius: 7px;
  display: flex; align-items: center; justify-content: center;
  transition: all var(--t-fast) var(--ease);
}
.icon-btn:hover { background: var(--bg-hover); color: var(--text-primary); }

.btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 7px 12px; border-radius: 7px;
  font-size: 13px; font-weight: 500; cursor: pointer;
  border: 1px solid transparent;
  transition: all var(--t-fast) var(--ease);
}
.btn-accent { background: var(--accent); color: #fff; }
.btn-accent:hover { background: var(--accent-hover); }
.btn-ghost { background: transparent; color: var(--text-secondary); border-color: var(--glass-border); }
.btn-ghost:hover { background: var(--bg-hover); color: var(--text-primary); }

.login-btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 7px 12px; border-radius: 7px;
  background: var(--accent); color: white;
  font-size: 13px; font-weight: 500;
  transition: all var(--t-fast) var(--ease);
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.25);
}
.login-btn:hover { background: var(--accent-hover); }

.user-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 10px; border-radius: 16px;
  background: var(--bg-elevated);
  color: var(--text-primary); font-size: 12px; font-weight: 500;
  border: 1px solid var(--glass-border);
  transition: all var(--t-fast) var(--ease);
}
.user-btn:hover { background: var(--bg-hover); }
.user-dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: var(--success);
  box-shadow: 0 0 6px var(--success);
}

.main { flex: 1; overflow: hidden; min-height: 0; }

/* ── Modal ── */
.modal-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.7);
  display: flex; align-items: center; justify-content: center;
  z-index: 100; backdrop-filter: blur(8px);
}
.modal {
  background: var(--bg-surface);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-xl);
  padding: 28px;
  width: 360px; max-width: 90vw;
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-slow) var(--ease);
}
.modal-mark {
  width: 44px; height: 44px; border-radius: 11px;
  background: var(--accent-muted);
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px; color: var(--accent);
}
.modal h3 { font-size: 17px; font-weight: 600; margin-bottom: 4px; }
.modal p { font-size: 13px; color: var(--text-secondary); margin-bottom: 18px; }
.input {
  width: 100%; padding: 10px 12px;
  background: var(--bg-base);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-size: 14px;
  transition: all var(--t-fast) var(--ease);
}
.input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-muted); outline: none; }
.modal-row { display: flex; justify-content: flex-end; gap: 8px; margin-top: 18px; }
.spinner-mini { width: 12px; height: 12px; border: 2px solid transparent; border-top-color: currentColor; border-radius: 50%; animation: spin 0.6s linear infinite; display: inline-block; }

/* Page transitions */
.page-enter-active { transition: all var(--t-slow) var(--ease); }
.page-leave-active { transition: all var(--t-base) var(--ease); }
.page-enter-from { opacity: 0; transform: translateY(8px); }
.page-leave-to { opacity: 0; transform: translateY(-4px); }
.fade-enter-active, .fade-leave-active { transition: opacity var(--t-base) ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

/* Responsive */
@media (max-width: 768px) {
  .topbar-inner { padding: 0 12px; gap: 8px; }
  .brand-sub { display: none; }
  .crumbs { display: none; }
  .topbar-center { max-width: none; }
  .search-box { max-width: none; }
  .search-placeholder { font-size: 12px; }
  .kbd-hint { display: none; }
}
@media (max-width: 480px) {
  .topbar { height: 48px; }
  .brand-title { font-size: 12px; }
  .login-btn span { display: none; }
  .login-btn { padding: 7px; }
}
</style>