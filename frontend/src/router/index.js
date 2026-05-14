import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/',
    redirect: '/editor',
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/RegisterView.vue'),
    meta: { public: true },
  },
  {
    path: '/editor',
    name: 'editor-new',
    component: () => import('../views/EditorView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/editor/:id',
    name: 'editor-edit',
    component: () => import('../views/EditorView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/articles',
    name: 'articles',
    component: () => import('../views/ArticleListView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('../views/SettingsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/media',
    name: 'media',
    component: () => import('../views/MediaListView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/templates',
    name: 'templates',
    component: () => import('../views/TemplateListView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/templates/new',
    name: 'template-new',
    component: () => import('../views/TemplateEditView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/templates/:id/edit',
    name: 'template-edit',
    component: () => import('../views/TemplateEditView.vue'),
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  if (!authStore.user && authStore.token) {
    try {
      await authStore.fetchProfile()
    } catch {
      authStore.logout()
    }
  }

  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    next('/login')
    return
  }

  if (to.meta.public && authStore.isLoggedIn) {
    next('/editor')
    return
  }

  next()
})

export default router
