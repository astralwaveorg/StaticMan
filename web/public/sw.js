const CACHE_NAME = 'staticman-v2'
const STATIC_ASSETS = [
  '/logo.svg',
  '/logo-192.png',
  '/logo-512.png',
]

// 安装：缓存核心资源
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS)
    })
  )
  self.skipWaiting()
})

// 激活：清理旧缓存
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys
          .filter((key) => key !== CACHE_NAME)
          .map((key) => caches.delete(key))
      )
    })
  )
  self.clients.claim()
})

// 辅助：判断响应是否有效
function isValidResponse(response) {
  return response && response.status === 200 && response.type === 'basic'
}

// 拦截请求
self.addEventListener('fetch', (event) => {
  const { request } = event
  const url = new URL(request.url)

  // API 请求：网络优先（从不缓存）
  if (url.pathname.startsWith('/api/') || url.pathname.startsWith('/raw/')) {
    event.respondWith(
      fetch(request)
        .catch(() => caches.match(request))
    )
    return
  }

  // index.html 和根路径：网络优先，确保标题等动态内容最新
  if (url.pathname === '/' || url.pathname === '/index.html') {
    event.respondWith(
      fetch(request)
        .then((response) => {
          if (isValidResponse(response)) {
            const clone = response.clone()
            caches.open(CACHE_NAME).then((cache) => {
              cache.put(request, clone)
            })
          }
          return response
        })
        .catch(() => caches.match(request))
    )
    return
  }

  // 静态资源：缓存优先，后台更新
  event.respondWith(
    caches.match(request).then((cached) => {
      const networkFetch = fetch(request).then((response) => {
        if (isValidResponse(response)) {
          const clone = response.clone()
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(request, clone)
          })
        }
        return response
      }).catch(() => cached)

      return cached || networkFetch
    })
  )
})
