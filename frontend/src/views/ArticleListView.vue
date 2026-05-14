<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { articleApi, tagApi } from '@/api/client.js'
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
  Tag,
  AlertTriangle,
  X,
  GripVertical,
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
const selectedTagId = ref('')
const tags = ref([])
const isLoading = ref(true)
const deletingId = ref(null)
const deleteTarget = ref(null)
const toasts = ref([])

let searchTimer = null

// ── Drag & Drop ──────────────────────────────────────────────────

const dragIndex = ref(null)
const dragOverIndex = ref(null)

function onDragStart(index) {
  dragIndex.value = index
}

function onDragOver(e, index) {
  e.preventDefault()
  dragOverIndex.value = index
}

function onDragLeave() {
  dragOverIndex.value = null
}

function onDrop(e, index) {
  e.preventDefault()
  dragOverIndex.value = null
  if (dragIndex.value === null || dragIndex.value === index) {
    dragIndex.value = null
    return
  }
  const list = [...articles.value]
  const [moved] = list.splice(dragIndex.value, 1)
  list.splice(index, 0, moved)
  dragIndex.value = null
  articles.value = list
  saveArticleReorder(list)
}

function onDragEnd() {
  dragIndex.value = null
  dragOverIndex.value = null
}

async function saveArticleReorder(items) {
  try {
    await articleApi.reorder(items.map((a, i) => ({ id: a.id, sort_order: i })))
  } catch (e) {
    showToast('排序保存失败', 'error')
  }
}

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
      tag_id: selectedTagId.value || undefined,
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

onMounted(async () => {
  if (!authStore.isLoggedIn) {
    router.push('/login')
    return
  }
  try {
    const tagData = await tagApi.list()
    tags.value = Array.isArray(tagData) ? tagData : []
  } catch (e) {
    tags.value = []
  }
  fetchArticles()
})

// ── Watchers ───────────────────────────────────────────────────

watch(status, () => {
  page.value = 1
  fetchArticles()
})

watch(selectedTagId, () => {
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

function askDelete(article) {
  deleteTarget.value = article
}

function cancelDelete() {
  deleteTarget.value = null
}

async function confirmDoDelete() {
  const id = deleteTarget.value?.id
  if (!id) return
  deletingId.value = id
  try {
    await articleApi.delete(id)
    showToast('文章已删除', 'success')
    deleteTarget.value = null
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

      <div class="filter-wrapper">
        <Tag :size="13" class="filter-icon" :stroke-width="1.8" />
        <select v-model="selectedTagId" class="filter-select">
          <option value="">全部标签</option>
          <option v-for="tag in tags" :key="tag.id" :value="tag.id">{{ tag.name }}</option>
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
              <th class="col-drag"></th>
              <th class="col-title">标题</th>
              <th class="col-tags">标签</th>
              <th class="col-words">字数</th>
              <th class="col-status">状态</th>
              <th class="col-date">更新时间</th>
              <th class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(article, index) in articles"
              :key="article.id"
              class="data-row"
              :class="{ 'drag-over': dragOverIndex === index, 'dragging': dragIndex === index }"
              draggable="true"
              @dragstart="onDragStart(index)"
              @dragover="onDragOver($event, index)"
              @dragleave="onDragLeave"
              @drop="onDrop($event, index)"
              @dragend="onDragEnd"
            >
              <td class="col-drag">
                <div class="drag-handle">
                  <GripVertical :size="14" :stroke-width="1.8" />
                </div>
              </td>
              <!-- Title -->
              <td class="col-title">
                <button
                  class="title-link"
                  @click="goToEditor(article.id)"
                >
                  {{ article.title || '无标题' }}
                </button>
              </td>

              <!-- Tags -->
              <td class="col-tags">
                <div class="article-tags">
                  <span
                    v-for="tag in (article.tags || [])"
                    :key="tag.id"
                    class="tag-badge"
                  >
                    {{ tag.name }}
                  </span>
                </div>
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

                  <button
                    class="action-btn action-danger"
                    title="删除"
                    @click="askDelete(article)"
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

    <!-- Delete Confirmation Modal -->
    <div v-if="deleteTarget" class="modal-overlay" @click.self="cancelDelete">
      <div class="modal delete-modal">
        <div class="modal-header">
          <h3 class="modal-title">确认删除</h3>
          <button class="btn btn-ghost btn-sm" @click="cancelDelete">
            <X :size="14" :stroke-width="2" />
          </button>
        </div>
        <div class="modal-body">
          <div class="delete-warning">
            <AlertTriangle :size="20" :stroke-width="1.8" class="warning-icon" />
            <p>确定要删除文章 <strong>「{{ deleteTarget.title || '无标题' }}」</strong>吗？此操作无法撤销。</p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="cancelDelete">取消</button>
          <button
            class="btn btn-danger"
            :disabled="deletingId === deleteTarget.id"
            @click="confirmDoDelete"
          >
            <Loader2 v-if="deletingId === deleteTarget.id" :size="14" class="animate-spin" />
            <Trash2 v-else :size="14" :stroke-width="2" />
            {{ deletingId === deleteTarget.id ? '删除中...' : '确认删除' }}
          </button>
        </div>
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

.col-tags {
  width: 160px;
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

/* ─── Tag Badge ─────────────────────────────────────────────────── */

.article-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  background: #f3f4f6;
  color: #374151;
  white-space: nowrap;
  letter-spacing: 0.02em;
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

/* ─── Delete Confirmation Modal ────────────────────────────────── */

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fade-in 0.15s ease-out;
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.delete-modal {
  background: var(--color-surface);
  border-radius: 14px;
  width: 400px;
  max-width: 90vw;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  animation: modal-in 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modal-in {
  from { transform: translateY(-8px) scale(0.97); opacity: 0; }
  to { transform: translateY(0) scale(1); opacity: 1; }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 22px;
  border-bottom: 1px solid var(--color-border-subtle);
}

.modal-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.modal-body {
  padding: 22px;
}

.delete-warning {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}

.warning-icon {
  color: #f59e0b;
  flex-shrink: 0;
  margin-top: 1px;
}

.delete-warning p {
  font-size: 14px;
  color: var(--color-text-secondary);
  line-height: 1.6;
  margin: 0;
}

.delete-warning strong {
  color: var(--color-text-primary);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 22px;
  border-top: 1px solid var(--color-border-subtle);
}

.btn-danger {
  background: var(--color-status-error);
  color: #fff;
  border: none;
  padding: 8px 18px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: var(--font-sans);
  transition: background 0.15s, opacity 0.15s;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

.btn-danger:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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
  .col-tags,
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

/* ─── Drag & Drop ─────────────────────────────────────────────── */

.col-drag {
  width: 36px;
  padding: 14px 8px !important;
}

.drag-handle {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-tertiary);
  opacity: 0;
  cursor: grab;
  transition: opacity 0.15s;
  padding: 2px;
  border-radius: 4px;
}

.data-row:hover .drag-handle {
  opacity: 1;
}

.drag-handle:active {
  cursor: grabbing;
}

.data-row.dragging {
  opacity: 0.5;
}

.data-row.drag-over {
  background: var(--color-accent-subtle) !important;
}
</style>
