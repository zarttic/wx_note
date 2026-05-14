<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from './stores/auth'
import {
  FileText,
  PenLine,
  List,
  Image,
  Settings,
  LogOut,
  User,
  LayoutTemplate,
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const isPublicPage = computed(() => route.meta.public)
const isLoggedIn = computed(() => authStore.isLoggedIn)

const navItems = [
  { path: '/editor', label: '编辑器', icon: PenLine },
  { path: '/articles', label: '文章管理', icon: List },
  { path: '/media', label: '素材库', icon: Image },
  { path: '/templates', label: '模板管理', icon: LayoutTemplate },
  { path: '/settings', label: '设置', icon: Settings },
]

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- Top Navigation (hidden on public pages) -->
    <header
      v-if="!isPublicPage && isLoggedIn"
      class="flex items-center justify-between px-5 h-12 bg-white border-b border-border flex-shrink-0"
    >
      <!-- Logo -->
      <div class="flex items-center gap-2.5">
        <div class="w-7 h-7 rounded-lg bg-accent flex items-center justify-center">
          <FileText :size="14" color="#fff" :stroke-width="2.5" />
        </div>
        <span class="text-sm font-semibold tracking-tight text-text-primary">wx_note</span>
      </div>

      <!-- Nav -->
      <nav class="flex items-center gap-1">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[13px] font-medium transition-colors"
          :class="
            route.path.startsWith(item.path) && item.path !== '/editor'
              ? 'bg-accent-subtle text-accent'
              : route.path === '/editor' && item.path === '/editor'
                ? 'bg-accent-subtle text-accent'
                : 'text-text-secondary hover:bg-surface-sunken hover:text-text-primary'
          "
        >
          <component :is="item.icon" :size="14" :stroke-width="2" />
          {{ item.label }}
        </router-link>
      </nav>

      <!-- User -->
      <div class="flex items-center gap-3">
        <div class="flex items-center gap-2 text-[12px] text-text-secondary">
          <div class="w-6 h-6 rounded-full bg-surface-sunken flex items-center justify-center">
            <User :size="12" :stroke-width="2" />
          </div>
          <span class="font-medium">{{ authStore.user?.nickname || authStore.user?.username }}</span>
        </div>
        <button
          class="btn btn-ghost btn-sm text-text-tertiary hover:text-text-primary"
          @click="handleLogout"
        >
          <LogOut :size="14" :stroke-width="2" />
        </button>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 overflow-hidden">
      <router-view />
    </main>
  </div>
</template>
