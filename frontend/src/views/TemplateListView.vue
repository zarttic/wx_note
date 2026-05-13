<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { templateApi } from '@/api/client.js'
import { useAuthStore } from '@/stores/auth'
import {
  Search,
  Plus,
  FileText,
  Edit3,
  Trash2,
  Loader2,
  LayoutTemplate,
  Tag,
} from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

// ── State ──────────────────────────────────────────────────────

const templates = ref([])
const categories = ref([])
const selectedCategory = ref('')
const search = ref('')
const isLoading = ref(true)
const deletingId = ref(null)
const confirmDeleteId = ref(null)
const toasts = ref([])

let searchTimer = null

// ── Toast ──────────────────────────────────────────────────────

function showToast(msg, type = 'info') {
  const id = Date.now()
  toasts.value.push({ id, msg, type })
  setTimeout(() => (toasts.value = toasts.value.filter((t) => t.id !== id)), 3500)
}

// ── Data Fetching ──────────────────────────────────────────────

async function fetchTemplates() {
  isLoading.value = true
  try {
    const data = await templateApi.list(selectedCategory.value || undefined)
    templates.value = Array.isArray(data) ? data : []
  } catch (err) {
    const msg = err?.message || '未知错误'
    if (msg.includes('401') || msg.includes('Unauthorized')) {
      authStore.logout()
      router.push('/login')
      return
    }
    showToast(msg, 'error')
    templates.value = []
  } finally {
    isLoading.value = false
  }
}

async function fetchCategories() {
  try {
    const data = await templateApi.categories()
    categories.value = Array.isArray(data) ? data : []
  } catch (err) {
    categories.value = []
  }
}

// ── Lifecycle ──────────────────────────────────────────────────

onMounted(() => {
  if (!authStore.isLoggedIn) {
    router.push('/login')
    return
  }
  fetchCategories()
  fetchTemplates()
})

// ── Watchers ───────────────────────────────────────────────────

watch(selectedCategory, () => {
  fetchTemplates()
})

// ── Search Debounce ────────────────────────────────────────────

function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    fetchTemplates()
  }, 350)
}

// ── Actions ────────────────────────────────────────────────────

function goToEdit(id) {
  router.push(`/templates/${id}/edit`)
}

function goToNew() {
  router.push('/templates/new')
}

function askDelete(id) {
  confirmDeleteId.value = id
}

function cancelDelete() {
  confirmDeleteId.value = null
}

async function confirmDoDelete(id) {
  deletingId.value = id
  try {
    await templateApi.delete(id)
    showToast('模板已删除', 'success')
    confirmDeleteId.value = null
    fetchTemplates()
  } catch (err) {
    showToast(err.message || '删除失败', 'error')
  } finally {
    deletingId.value = null
  }
}

// ── Helpers ────────────────────────────────────────────────────

function formatDate(str) {
  if (!str) return '--'
  const d = new Date(str)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function truncateContent(content, len = 100) {
  if (!content) return ''
  const stripped = content.replace(/[#*`>\-\[\]()]/g, '').trim()
  return stripped.length > len ? stripped.slice(0, len) + '...' : stripped
}

const filteredTemplates = () => {
  if (!search.value.trim()) return templates.value
  const q = search.value.trim().toLowerCase()
  return templates.value.filter(
    (t) =>
      (t.title || t.name || '').toLowerCase().includes(q) ||
      (t.category || '').toLowerCase().includes(q)
  )
}
</script>

<template>
  <div class="page-root">
    <!-- Toast Container -->
    <div class="toast-container">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="toast"
        :class="`toast-${t.type}`"
      >
        {{ t.msg }}
      </div>
    </div>

    <!-- Page Header -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">模板管理</h1>
        <p class="page-subtitle">
          共 <strong>{{ templates.length }}</strong> 个模板
        </p>
      </div>
      <button class="btn btn-primary" @click="goToNew">
        <Plus :size="15" :stroke-width="2.2" />
        新建模板
      </button>
    </header>

    <!-- Toolbar -->
    <div class="toolbar">
      <div class="search-wrapper">
        <Search :size="15" class="search-icon" :stroke-width="1.8" />
        <input
          v-model="search"
          type="text"
          class="search-input"
          placeholder="搜索模板..."
          @input="onSearchInput"
        />
      </div>

      <div class="filter-wrapper">
        <Tag :size="13" class="filter-icon" :stroke-width="1.8" />
        <select v-model="selectedCategory" class="filter-select">
          <option value="">全部分类</option>
          <option v-for="cat in categories" :key="cat" :value="cat">
            {{ cat }}
          </option>
        </select>
      </div>
    </div>

    <!-- Content Area -->
    <div class="content-area">
      <!-- Loading Skeleton -->
      <div v-if="isLoading" class="card-grid">
        <div v-for="i in 6" :key="i" class="skeleton-card">
          <div class="skeleton-bar sk-card-title" />
          <div class="skeleton-bar sk-card-meta" />
          <div class="skeleton-bar sk-card-content" />
          <div class="skeleton-bar sk-card-content short" />
        </div>
      </div>

      <!-- Template Cards Grid -->
      <div v-else-if="filteredTemplates().length > 0" class="card-grid">
        <div
          v-for="template in filteredTemplates()"
          :key="template.id"
          class="template-card"
          @click="goToEdit(template.id)"
        >
          <!-- Card Header -->
          <div class="card-header">
            <div class="card-title-row">
              <LayoutTemplate :size="14" :stroke-width="1.8" class="card-icon" />
              <span class="card-title">{{ template.title || template.name || '未命名模板' }}</span>
            </div>
            <span v-if="template.category" class="category-badge">
              {{ template.category }}
            </span>
          </div>

          <!-- Card Content Preview -->
          <p class="card-preview">
            {{ truncateContent(template.content, 100) }}
          </p>

          <!-- Card Footer -->
          <div class="card-footer">
            <span class="card-date">{{ formatDate(template.updated_at) }}</span>
            <div class="card-actions" @click.stop>
              <button
                class="action-btn"
                title="编辑"
                @click="goToEdit(template.id)"
              >
                <Edit3 :size="14" :stroke-width="1.8" />
              </button>

              <template v-if="confirmDeleteId === template.id">
                <span class="confirm-text">确认删除？</span>
                <button
                  class="confirm-btn confirm-yes"
                  :disabled="deletingId === template.id"
                  @click="confirmDoDelete(template.id)"
                >
                  <Loader2
                    v-if="deletingId === template.id"
                    :size="12"
                    class="animate-spin"
                  />
                  <span v-else>确认</span>
                </button>
                <button
                  class="confirm-btn confirm-no"
                  @click="cancelDelete"
                >
                  取消
                </button>
              </template>

              <button
                v-else
                class="action-btn action-danger"
                title="删除"
                @click="askDelete(template.id)"
              >
                <Trash2 :size="14" :stroke-width="1.8" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <FileText :size="40" class="empty-icon" :stroke-width="1.2" />
        <p class="empty-text">暂无模板，点击右上角新建</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ─── Layout ─────────────────────────────────────────────────── */

.page-root {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 36px 40px 28px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

/* ─── Header ─────────────────────────────────────────────────── */

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 28px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.page-subtitle {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin: 0;
  font-weight: 400;
}

.page-subtitle strong {
  color: var(--color-text-primary);
  font-weight: 600;
}

/* ─── Toolbar ────────────────────────────────────────────────── */

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
  flex-shrink: 0;
}

.search-wrapper {
  position: relative;
  flex: 1;
  max-width: 380px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-tertiary);
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 9px 14px 9px 36px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  font-family: var(--font-sans);
  color: var(--color-text-primary);
  background: var(--color-surface);
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.search-input:focus {
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px var(--color-accent-subtle);
}

.search-input::placeholder {
  color: var(--color-text-tertiary);
}

.filter-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.filter-icon {
  position: absolute;
  left: 11px;
  color: var(--color-text-tertiary);
  pointer-events: none;
}

.filter-select {
  padding: 9px 12px 9px 32px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  font-family: var(--font-sans);
  color: var(--color-text-primary);
  background: var(--color-surface);
  outline: none;
  cursor: pointer;
  transition: border-color 0.15s;
  appearance: none;
  min-width: 130px;
}

.filter-select:focus {
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px var(--color-accent-subtle);
}

/* ─── Content Area ───────────────────────────────────────────── */

.content-area {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

/* ─── Card Grid ──────────────────────────────────────────────── */

.card-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

@media (max-width: 1024px) {
  .card-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .card-grid {
    grid-template-columns: 1fr;
  }
}

/* ─── Template Card ──────────────────────────────────────────── */

.template-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: 12px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s, transform 0.15s;
}

.template-card:hover {
  border-color: #d1d5db;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
  transform: translateY(-1px);
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.card-icon {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

.card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.category-badge {
  flex-shrink: 0;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  background: var(--color-accent-subtle);
  color: var(--color-accent);
  letter-spacing: 0.02em;
  white-space: nowrap;
}

.card-preview {
  font-size: 13px;
  color: var(--color-text-secondary);
  line-height: 1.7;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  flex: 1;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 10px;
  border-top: 1px solid var(--color-border-subtle);
  flex-shrink: 0;
}

.card-date {
  font-size: 11px;
  color: var(--color-text-tertiary);
  font-variant-numeric: tabular-nums;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--color-text-tertiary);
  cursor: pointer;
  transition: background 0.12s, color 0.12s;
}

.action-btn:hover {
  background: var(--color-surface-sunken);
  color: var(--color-text-primary);
}

.action-danger:hover {
  background: rgba(239, 68, 68, 0.06);
  color: var(--color-status-error);
}

/* ─── Delete Confirmation ────────────────────────────────────── */

.confirm-text {
  font-size: 12px;
  color: var(--color-status-error);
  font-weight: 500;
  margin-left: 4px;
}

.confirm-btn {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 500;
  border: none;
  cursor: pointer;
  font-family: var(--font-sans);
  transition: background 0.12s, opacity 0.12s;
}

.confirm-yes {
  background: var(--color-status-error);
  color: #fff;
}

.confirm-yes:hover:not(:disabled) {
  background: #dc2626;
}

.confirm-yes:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.confirm-no {
  background: var(--color-surface-sunken);
  color: var(--color-text-secondary);
}

.confirm-no:hover {
  background: #e5e7eb;
}

/* ─── Skeleton Cards ─────────────────────────────────────────── */

.skeleton-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: 12px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-bar {
  border-radius: 4px;
  background: linear-gradient(90deg, #f3f4f6 25%, #e5e7eb 50%, #f3f4f6 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s ease-in-out infinite;
}

.sk-card-title {
  height: 16px;
  width: 60%;
}

.sk-card-meta {
  height: 12px;
  width: 30%;
}

.sk-card-content {
  height: 13px;
  width: 100%;
}

.sk-card-content.short {
  width: 70%;
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

/* ─── Empty State ────────────────────────────────────────────── */

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  gap: 14px;
}

.empty-icon {
  color: #d1d5db;
}

.empty-text {
  font-size: 14px;
  color: var(--color-text-tertiary);
  margin: 0;
}

/* ─── Responsive ─────────────────────────────────────────────── */

@media (max-width: 768px) {
  .page-root {
    padding: 24px 16px 20px;
  }

  .page-title {
    font-size: 20px;
  }

  .toolbar {
    flex-wrap: wrap;
  }

  .search-wrapper {
    max-width: 100%;
  }
}
</style>
