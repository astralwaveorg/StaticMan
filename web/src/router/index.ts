import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: () => import('../views/HomeView.vue') },
    { path: '/browse/:pathMatch(.*)*', name: 'browse-path', component: () => import('../views/BrowseView.vue') },
  ],
})

export default router