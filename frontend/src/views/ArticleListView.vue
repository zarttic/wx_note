<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { articleApi } from '@/api/client.js'
import { useAuthStore } from '@/stores/auth'
import {
  Search,
  Plus,
  FileText,
  Edit3,
  Trash2,
  Loader2,
  ChevronLeft,
  ChevronRight,
  Filter,
} from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

// ── State ──────────────────────────────────────────────────────

const articles = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const search = ref('')
const status = ref('')
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

async function fetchArticles() {
  isLoading.value = true
  try {
    const data = await articleApi.list({
      page: page.value,
      pageSize,
      status: status.value || undefined,
      search: search.value || undefined,
    })
    articles.value = data.items || []
    total.value = data.total || 0
  } catch (err) {
    if (err.message.includes('401') || err.message.includes('Unauthorized')) {
      router.push('/login')
      return
    }
    showToast(err.message || '加载失败', 'error')
    articles.value = []
  } finally {
    isLoading.value = false
  }
}

// ── Lifecycle ──────────────────────────────────────────────────

onMounted(() => {
  if (!authStore.isLoggedIn) {
    router.push('/login')
    return
  }
  fetchArticles()
})

// ── Watchers ───────────────────────────────────────────────────

watch(status, () => {
  page.value = 1
  fetchArticles()
})

// ── Search Debounce ────────────────────────────────────────────

function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchArticles()
  }, 350)
}

// ── Actions ────────────────────────────────────────────────────

function goToEditor(id) {
  router.push(`/editor/${id}`)
}

function goToNew() {
  router.push('/editor')
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
    await articleApi.delete(id)
    showToast('文章已删除', 'success')
    confirmDeleteId.value = null
    if (articles.value.length === 1 && page.value > 1) {
      page.value -= 1
    }
    fetchArticles()
  } catch (err) {
    showToast(err.message || '删除失败', 'error')
  } finally {
    deletingId.value = null
  }
}

function prevPage() {
  if (page.value <= 1) return
  page.value -= 1
  fetchArticles()
}

function nextPage() {
  if (page.value * pageSize >= total.value) return
  page.value += 1
  fetchArticles()
}

// ── Helpers ────────────────────────────────────────────────────

function formatDate(str) {
  if (!str) return '--'
  const d = new Date(str)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const totalPages = () => Math.max(1, Math.ceil(total.value / pageSize))
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
        <h1 class="page-title">文章管理</h1>
        <p class="page-subtitle">
          共 <strong>{{ total }}</strong> 篇文章
        </p>
      </div>
      <button class="btn btn-primary" @click="goToNew">
        <Plus :size="15" :stroke-width="2.2" />
        新建文章
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
          placeholder="搜索文章..."
          @input="onSearchInput"
        />
      </div>

      <div class="filter-wrapper">
        <Filter :size="13" class="filter-icon" :stroke-width="1.8" />
        <select v-model="status" class="filter-select">
          <option value="">全部</option>
          <option value="draft">草稿</option>
          <option value="published">已发布</option>
        </select>
      </div>
    </div>

    <!-- Content Area -->
    <div class="content-area">
      <!-- Loading Skeleton -->
      <div v-if="isLoading" class="skeleton-list">
        <div v-for="i in 3" :key="i" class="skeleton-row">
          <div class="skeleton-bar sk-title" />
          <div class="skeleton-bar sk-meta" />
        </div>
      </div>

      <!-- Article Table -->
      <div v-else-if="articles.length > 0" class="table-wrapper">
        <table class="article-table">
          <thead>
            <tr>
              <th class="col-title">标题</th>
              <th class="col-words">字数</th>
              <th class="col-status">状态</th>
              <th class="col-date">更新时间</th>
              <th class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="article in articles"
              :key="article.id"
              class="data-row"
            >
              <!-- Title -->
              <td class="col-title">
                <button
                  class="title-link"
                  @click="goToEditor(article.id)"
                >
                  {{ article.title || '无标题' }}
                </button>
              </td>

              <!-- Word Count -->
              <td class="col-words">
                <span class="word-count">{{ article.word_count || 0 }}</span>
              </td>

              <!-- Status -->
              <td class="col-status">
                <span
                  class="status-badge"
                  :class="
                    article.status === 'published'
                      ? 'badge-published'
                      : 'badge-draft'
                  "
                >
                  {{ article.status === 'published' ? '已发布' : '草稿' }}
                </span>
              </td>

              <!-- Updated At -->
              <td class="col-date">
                <span class="date-text">{{ formatDate(article.updated_at) }}</span>
              </td>

              <!-- Actions -->
              <td class="col-actions">
                <div class="actions-cell">
                  <button
                    class="action-btn"
                    title="编辑"
                    @click="goToEditor(article.id)"
                  >
                    <Edit3 :size="14" :stroke-width="1.8" />
                  </button>

                  <!-- Delete: confirm inline -->
                  <template v-if="confirmDeleteId === article.id">
                    <span class="confirm-text">确认删除？</span>
                    <button
                      class="confirm-btn confirm-yes"
                      :disabled="deletingId === article.id"
                      @click="confirmDoDelete(article.id)"
                    >
                      <Loader2
                        v-if="deletingId === article.id"
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
                    @click="askDelete(article.id)"
                  >
                    <Trash2 :size="14" :stroke-width="1.8" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <FileText :size="40" class="empty-icon" :stroke-width="1.2" />
        <p class="empty-text">暂无文章，点击右上角新建</p>
      </div>
    </div>

    <!-- Pagination -->
    <footer v-if="articles.length > 0 && total > pageSize" class="pagination">
      <button
        class="page-btn"
        :disabled="page <= 1"
        @click="prevPage"
      >
        <ChevronLeft :size="16" :stroke-width="2" />
        上一页
      </button>

      <span class="page-indicator">
        第 <strong>{{ page }}</strong> / <strong>{{ totalPages() }}</strong> 页
      </span>

      <button
        class="page-btn"
        :disabled="page * pageSize >= total"
        @click="nextPage"
      >
        下一页
        <ChevronRight :size="16" :stroke-width="2" />
      </button>
    </footer>
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
  max-width: 1100px;
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
  margin-bottom: 20px;
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
  min-width: 110px;
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

/* ─── Table ──────────────────────────────────────────────────── */

.table-wrapper {
  border: 1px solid var(--color-border-subtle);
  border-radius: 10px;
  overflow: hidden;
}

.article-table {
  width: 100%;
  border-collapse: collapse;
}

.article-table thead {
  background: var(--color-surface-sunken);
}

.article-table th {
  padding: 12px 18px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-align: left;
  letter-spacing: 0.03em;
  white-space: nowrap;
  border-bottom: 1px solid var(--color-border-subtle);
}

.col-title {
  width: auto;
  min-width: 200px;
}

.col-words {
  width: 80px;
  text-align: right;
}

.col-status {
  width: 90px;
}

.col-date {
  width: 160px;
  white-space: nowrap;
}

.col-actions {
  width: 160px;
  text-align: right;
}

.data-row {
  transition: background 0.12s;
}

.data-row:hover {
  background: #fafafa;
}

.data-row:not(:last-child) td {
  border-bottom: 1px solid var(--color-border-subtle);
}

.article-table td {
  padding: 14px 18px;
  font-size: 13px;
  color: var(--color-text-primary);
  vertical-align: middle;
}

/* ─── Title Link ─────────────────────────────────────────────── */

.title-link {
  background: none;
  border: none;
  padding: 0;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-primary);
  cursor: pointer;
  text-align: left;
  font-family: var(--font-sans);
  max-width: 380px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
  transition: color 0.12s;
}

.title-link:hover {
  color: #07c160;
}

/* ─── Word Count ─────────────────────────────────────────────── */

.word-count {
  font-size: 12px;
  color: var(--color-text-secondary);
  font-variant-numeric: tabular-nums;
  display: block;
  text-align: right;
}

/* ─── Status Badge ───────────────────────────────────────────── */

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.02em;
  white-space: nowrap;
}

.badge-draft {
  background: #f3f4f6;
  color: #6b7280;
}

.badge-published {
  background: rgba(7, 193, 96, 0.08);
  color: #059669;
}

/* ─── Date ───────────────────────────────────────────────────── */

.date-text {
  font-size: 12px;
  color: var(--color-text-tertiary);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

/* ─── Actions ────────────────────────────────────────────────── */

.actions-cell {
  display: flex;
  align-items: center;
  justify-content: flex-end;
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

/* ─── Skeleton ───────────────────────────────────────────────── */

.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.skeleton-row {
  padding: 18px 20px;
  border-bottom: 1px solid var(--color-border-subtle);
}

.skeleton-row:last-child {
  border-bottom: none;
}

.skeleton-bar {
  border-radius: 4px;
  background: linear-gradient(90deg, #f3f4f6 25%, #e5e7eb 50%, #f3f4f6 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s ease-in-out infinite;
}

.sk-title {
  height: 14px;
  width: 45%;
  margin-bottom: 10px;
}

.sk-meta {
  height: 11px;
  width: 25%;
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

/* ─── Pagination ─────────────────────────────────────────────── */

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding-top: 20px;
  flex-shrink: 0;
}

.page-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 14px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
  color: var(--color-text-primary);
  font-size: 13px;
  font-weight: 500;
  font-family: var(--font-sans);
  cursor: pointer;
  transition: background 0.12s, border-color 0.12s;
}

.page-btn:hover:not(:disabled) {
  background: var(--color-surface-sunken);
  border-color: #d1d5db;
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-indicator {
  font-size: 13px;
  color: var(--color-text-secondary);
  font-variant-numeric: tabular-nums;
}

.page-indicator strong {
  color: var(--color-text-primary);
  font-weight: 600;
}

/* ─── Responsive ─────────────────────────────────────────────── */

@media (max-width: 768px) {
  .page-root {
    padding: 24px 16px 20px;
  }

  .page-title {
    font-size: 20px;
  }

  .col-words,
  .col-status {
    display: none;
  }

  .toolbar {
    flex-wrap: wrap;
  }

  .search-wrapper {
    max-width: 100%;
  }
}
</style>
