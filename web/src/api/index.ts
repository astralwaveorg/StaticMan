import axios from 'axios'

const API_BASE = '/api'

const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  const token = getToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (r) => r,
  (e) => { if (e.response?.status === 401) clearToken(); return Promise.reject(e) }
)

function getToken(): string | null { return localStorage.getItem('staticman_token') }
function setToken(t: string) {
  localStorage.setItem('staticman_token', t)
  document.cookie = `staticman_token=${t};path=/;max-age=${7*24*3600};samesite=lax`
}
function clearToken() {
  localStorage.removeItem('staticman_token')
  document.cookie = 'staticman_token=;path=/;max-age=0'
}

// ─── Types ──────────────────────────────────────────────
export interface TreeNode {
  name: string
  path: string
  type: 'file' | 'directory'
  protected?: boolean
  size?: number
  modTime?: string
  isBinary?: boolean
  children?: TreeNode[]
}

export interface CategoryInfo {
  key: string
  name: string
  icon: string
  description: string
  color: string
  fileCount: number
  size: number
  tools?: string[]
}

export interface LsItem {
  name: string
  path: string
  type: 'file' | 'directory'
  size: number
  modTime: string
  protected: boolean
  isBinary: boolean
  language?: string
}

export interface Breadcrumb {
  name: string
  path: string
}

export interface FileContent {
  name: string
  path: string
  type?: 'file' | 'directory'
  content: string
  language: string
  protected: boolean
  size: number
  truncated: boolean
  isBinary: boolean
  modTime?: string
  description?: string
}

export interface MatchLine {
  line: number
  text: string
}

export interface SearchResult {
  path: string
  name: string
  type: 'file' | 'directory'
  protected: boolean
  isBinary?: boolean
  language?: string
  size?: number
  matches?: MatchLine[]
}

// ─── API Calls ──────────────────────────────────────────
export const getTree = (maxDepth = 3) =>
  api.get<TreeNode>('/tree', { params: { maxDepth } })

export const getCategories = () =>
  api.get<CategoryInfo[]>('/categories')

export const getLs = (path: string = '') =>
  api.get<{ path: string; items: LsItem[]; total: number }>('/ls', { params: { path } })

export const getBreadcrumbs = (path: string) =>
  api.get<Breadcrumb[]>('/breadcrumbs', { params: { path } })

export const getFile = (p: string) =>
  api.get<FileContent>(`/file/${p}`)

export const searchFiles = (q: string, t: 'name' | 'content') =>
  api.get<SearchResult[]>('/search', { params: { q, t } })

export const authenticate = (p: string) =>
  api.post<{ token: string }>('/auth', { password: p })

export const getHealth = () =>
  api.get('/health')

// ─── URL Helpers ────────────────────────────────────────
// Raw URL: 服务于 /raw/<path>?key=JWT (受保护文件) 或 /<path> (公开文件)
export function getRawUrl(path: string, isProtected: boolean, useRawPrefix = true): string {
  const basePath = useRawPrefix ? `/raw/${path}` : `/${path}`
  const origin = typeof window !== 'undefined' ? window.location.origin : ''
  const url = `${origin}${basePath}`
  if (isProtected) {
    const t = getToken()
    if (t) return `${url}?key=${t}`
  }
  return url
}

// 直接路径（不含 origin）用于 API 客户端
export function getRawPath(path: string, isProtected: boolean, useRawPrefix = true): string {
  const basePath = useRawPrefix ? `/raw/${path}` : `/${path}`
  if (isProtected) {
    const t = getToken()
    if (t) return `${basePath}?key=${t}`
  }
  return basePath
}

export const isLoggedIn = () => !!getToken()
export { getToken, setToken, clearToken }
export default api