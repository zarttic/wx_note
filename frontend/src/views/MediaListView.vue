<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { mediaApi, editorApi } from '@/api/client.js'
import { useAuthStore } from '@/stores/auth'
import {
  Image,
  Trash2,
  Copy,
  Loader2,
  ChevronLeft,
  ChevronRight,
  Upload,
} from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

// ── State ──────────────────────────────────────────────────────

const mediaList = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 40
const isLoading = ref(true)
const deletingId = ref(null)
const confirmDeleteId = ref(null)

const isUploading = ref(false)
const uploadProgress = ref({ done: 0, total: 0 })
const fileInputRef = ref(null)

function triggerUpload() {
  fileInputRef.value?.click()
}

async function handleFileUpload(e) {
  const files = Array.from(e.target.files || [])
  if (files.length === 0) return

  isUploading.value = true
  uploadProgress.value = { done: 0, total: files.length }

  for (let i = 0; i < files.length; i++) {
    try {
      await editorApi.uploadImage(files[i])
      uploadProgress.value.done = i + 1
    } catch (err) {
      showToast(`上传失败：${files[i].name}`, 'error')
    }
  }

  isUploading.value = false
  uploadProgress.value = { done: 0, total: 0 }
  e.target.value = ''
  fetchMedia()
}

const toasts = ref([])

// ── Toast ──────────────────────────────────────────────────────

function showToast(msg, type = 'info') {
  const id = Date.now()
  toasts.value.push({ id, msg, type })
  setTimeout(() => (toasts.value = toasts.value.filter((t) => t.id !== id)), 3500)
}

// ── Data Fetching ──────────────────────────────────────────────

async function fetchMedia() {
  isLoading.value = true
  try {
    const data = await mediaApi.list(page.value, pageSize)
    mediaList.value = data.items || []
    total.value = data.total || 0
  } catch (err) {
    if (err.message.includes('401') || err.message.includes('Unauthorized')) {
      router.push('/login')
      return
    }
    showToast(err.message || '加载失败', 'error')
    mediaList.value = []
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
  fetchMedia()
})

// ── Actions ────────────────────────────────────────────────────

function copyUrl(url) {
  navigator.clipboard.writeText(url).then(() => {
    showToast('已复制图片链接', 'success')
  }).catch(() => {
    showToast('复制失败', 'error')
  })
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
    await mediaApi.delete(id)
    showToast('素材已删除', 'success')
    confirmDeleteId.value = null
    if (mediaList.value.length === 1 && page.value > 1) {
      page.value -= 1
    }
    fetchMedia()
  } catch (err) {
    showToast(err.message || '删除失败', 'error')
  } finally {
    deletingId.value = null
  }
}

function prevPage() {
  if (page.value <= 1) return
  page.value -= 1
  fetchMedia()
}

function nextPage() {
  if (page.value * pageSize >= total.value) return
  page.value += 1
  fetchMedia()
}

// ── Helpers ────────────────────────────────────────────────────

function formatDate(str) {
  if (!str) return '--'
  const d = new Date(str)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

function truncateFilename(name, maxLen = 18) {
  if (!name) return '未命名'
  if (name.length <= maxLen) return name
  const ext = name.lastIndexOf('.')
  if (ext > 0 && name.length - ext <= 5) {
    const base = name.slice(0, maxLen - (name.length - ext) - 1)
    return base + '...' + name.slice(ext)
  }
  return name.slice(0, maxLen - 3) + '...'
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
        <h1 class="page-title">素材库</h1>
        <p class="page-subtitle">
          共 <strong>{{ total }}</strong> 个素材
        </p>
      </div>
      <div class="header-right">
        <input
          ref="fileInputRef"
          type="file"
          accept="image/*"
          multiple
          class="hidden"
          @change="handleFileUpload"
        />
        <button class="btn btn-primary" :disabled="isUploading" @click="triggerUpload">
          <Loader2 v-if="isUploading" :size="15" class="animate-spin" />
          <Upload v-else :size="15" :stroke-width="2.2" />
          {{ isUploading ? `上传中 ${uploadProgress.done}/${uploadProgress.total}` : '上传图片' }}
        </button>
      </div>
    </header>

    <!-- Content Area -->
    <div class="content-area">
      <!-- Loading Skeleton -->
      <div v-if="isLoading" class="skeleton-grid">
        <div v-for="i in 6" :key="i" class="skeleton-card">
          <div class="skeleton-bar sk-image" />
          <div class="skeleton-bar sk-name" />
        </div>
      </div>

      <!-- Media Grid -->
      <div v-else-if="mediaList.length > 0" class="media-grid">
        <div
          v-for="item in mediaList"
          :key="item.id"
          class="media-card"
        >
          <!-- Image -->
          <div class="media-image-wrapper" @click="copyUrl(item.url)">
            <img :src="item.url" :alt="item.filename" class="media-image" loading="lazy" />
            <div class="media-overlay">
              <button class="overlay-btn" title="复制链接" @click.stop="copyUrl(item.url)">
                <Copy :size="14" :stroke-width="1.8" />
              </button>
              <template v-if="confirmDeleteId === item.id">
                <span class="confirm-text">确认删除？</span>
                <button
                  class="confirm-btn confirm-yes"
                  :disabled="deletingId === item.id"
                  @click.stop="confirmDoDelete(item.id)"
                >
                  <Loader2 v-if="deletingId === item.id" :size="12" class="animate-spin" />
                  <span v-else>确认</span>
                </button>
                <button class="confirm-btn confirm-no" @click.stop="cancelDelete">
                  取消
                </button>
              </template>
              <button
                v-else
                class="overlay-btn overlay-danger"
                title="删除"
                @click.stop="askDelete(item.id)"
              >
                <Trash2 :size="14" :stroke-width="1.8" />
              </button>
            </div>
          </div>

          <!-- Info -->
          <div class="media-info">
            <span class="media-filename">{{ truncateFilename(item.filename) }}</span>
            <span class="media-date">{{ formatDate(item.created_at) }}</span>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <Image :size="40" class="empty-icon" :stroke-width="1.2" />
        <p class="empty-text">暂无素材，编辑器中上传图片后会自动保存到素材库</p>
      </div>
    </div>

    <!-- Pagination -->
    <footer v-if="mediaList.length > 0 && total > pageSize" class="pagination">
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

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
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

/* ─── Content Area ───────────────────────────────────────────── */

.content-area {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

/* ─── Media Grid ─────────────────────────────────────────────── */

.media-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

@media (max-width: 900px) {
  .media-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 600px) {
  .media-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* ─── Media Card ─────────────────────────────────────────────── */

.media-card {
  background: var(--color-surface);
  border-radius: 10px;
  border: 1px solid var(--color-border-subtle);
  overflow: hidden;
  transition: box-shadow 0.15s, border-color 0.15s;
}

.media-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
  border-color: var(--color-border);
}

/* ─── Image Wrapper ──────────────────────────────────────────── */

.media-image-wrapper {
  position: relative;
  aspect-ratio: 1;
  background: #f9fafb;
  cursor: pointer;
  overflow: hidden;
}

.media-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.2s;
}

.media-image-wrapper:hover .media-image {
  transform: scale(1.02);
}

.media-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.15s;
}

.media-image-wrapper:hover .media-overlay {
  opacity: 1;
}

.overlay-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: rgba(255, 255, 255, 0.9);
  color: var(--color-text-primary);
  cursor: pointer;
  transition: background 0.12s;
}

.overlay-btn:hover {
  background: #fff;
}

.overlay-danger {
  color: var(--color-status-error);
}

.overlay-danger:hover {
  background: rgba(239, 68, 68, 0.1);
}

/* ─── Delete Confirmation ────────────────────────────────────── */

.confirm-text {
  font-size: 12px;
  color: #fff;
  font-weight: 500;
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
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.confirm-no:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* ─── Info ───────────────────────────────────────────────────── */

.media-info {
  padding: 8px 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
}

.media-filename {
  font-size: 12px;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
  flex: 1;
}

.media-date {
  font-size: 11px;
  color: var(--color-text-tertiary);
  white-space: nowrap;
  flex-shrink: 0;
}

/* ─── Skeleton ───────────────────────────────────────────────── */

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

@media (max-width: 900px) {
  .skeleton-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 600px) {
  .skeleton-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.skeleton-card {
  border-radius: 10px;
  overflow: hidden;
}

.sk-image {
  height: 120px;
  width: 100%;
  margin-bottom: 0;
  border-radius: 0;
}

.sk-name {
  height: 12px;
  width: 60%;
  margin: 8px 10px;
  border-radius: 4px;
}

.skeleton-bar {
  border-radius: 4px;
  background: linear-gradient(90deg, #f3f4f6 25%, #e5e7eb 50%, #f3f4f6 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s ease-in-out infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
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
  text-align: center;
  line-height: 1.6;
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
}
</style>
