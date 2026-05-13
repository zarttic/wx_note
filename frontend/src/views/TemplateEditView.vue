<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { defineAsyncComponent } from 'vue'
import 'md-editor-v3/lib/style.css'

const MdEditor = defineAsyncComponent(() =>
  import('md-editor-v3').then(m => m.MdEditor)
)

import { templateApi } from '@/api/client.js'
import { useAuthStore } from '@/stores/auth'
import {
  Save,
  ArrowLeft,
  Loader2,
  ImagePlus,
  X,
  LayoutTemplate,
  Tag,
  Type,
  FileText,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// ─── State ─────────────────────────────────────────────────────

const templateId = computed(() => route.params.id || null)
const isNew = computed(() => !templateId.value)

const name = ref('')
const category = ref('')
const content = ref('')
const coverFile = ref(null)
const coverPreviewUrl = ref('')

const categories = ref([])
const categoryInput = ref('')

const isSaving = ref(false)
const isLoading = ref(false)
const editorReady = ref(false)
const toasts = ref([])

// ─── Toast ─────────────────────────────────────────────────────

function showToast(msg, type = 'info') {
  const id = Date.now()
  toasts.value.push({ id, msg, type })
  setTimeout(() => (toasts.value = toasts.value.filter((t) => t.id !== id)), 3500)
}

// ─── Categories ────────────────────────────────────────────────

async function fetchCategories() {
  try {
    const data = await templateApi.categories()
    categories.value = data.categories || data || []
  } catch (err) {
    categories.value = []
  }
}

// ─── Load Template (edit mode) ─────────────────────────────────

async function loadTemplate(id) {
  isLoading.value = true
  try {
    const data = await templateApi.get(id)
    name.value = data.title || data.name || ''
    category.value = data.category || ''
    content.value = data.content || ''
    coverPreviewUrl.value = data.cover_url || ''
    await nextTick()
  } catch (err) {
    showToast('加载模板失败：' + err.message, 'error')
  } finally {
    isLoading.value = false
  }
}

// ─── Cover Upload ──────────────────────────────────────────────

function handleCoverSelect(e) {
  const file = e.target.files[0]
  if (!file) return
  coverFile.value = file
  coverPreviewUrl.value = URL.createObjectURL(file)
}

function clearCover() {
  coverFile.value = null
  coverPreviewUrl.value = ''
}

// ─── Save ──────────────────────────────────────────────────────

async function handleSave() {
  if (!name.value.trim()) {
    showToast('请输入模板名称', 'error')
    return
  }
  if (!content.value.trim()) {
    showToast('请输入模板内容', 'error')
    return
  }

  isSaving.value = true
  try {
    const payload = {
      name: name.value.trim(),
      category: category.value.trim(),
      content: content.value,
      cover_url: coverPreviewUrl.value,
    }

    if (isNew.value) {
      await templateApi.create(payload)
      showToast('模板已创建', 'success')
    } else {
      await templateApi.update(templateId.value, payload)
      showToast('模板已更新', 'success')
    }
    router.push('/templates')
  } catch (err) {
    showToast('保存失败：' + err.message, 'error')
  } finally {
    isSaving.value = false
  }
}

function handleCancel() {
  router.push('/templates')
}

// ─── Init ──────────────────────────────────────────────────────

onMounted(async () => {
  if (!authStore.isLoggedIn) {
    router.push('/login')
    return
  }
  await fetchCategories()
  if (templateId.value) {
    await loadTemplate(templateId.value)
  }
  setTimeout(() => { editorReady.value = true }, 100)
})
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

    <!-- Loading -->
    <div v-if="isLoading" class="loading-state">
      <Loader2 :size="20" class="animate-spin" />
      <span>加载中...</span>
    </div>

    <!-- Edit Form -->
    <template v-else>
      <!-- Page Header -->
      <header class="page-header">
        <div class="header-left">
          <button class="back-btn" @click="handleCancel">
            <ArrowLeft :size="16" :stroke-width="2" />
          </button>
          <h1 class="page-title">{{ isNew ? '新建模板' : '编辑模板' }}</h1>
        </div>
        <div class="header-actions">
          <button class="btn btn-secondary" @click="handleCancel">
            取消
          </button>
          <button class="btn btn-primary" :disabled="isSaving" @click="handleSave">
            <Loader2 v-if="isSaving" :size="13" class="animate-spin" />
            <Save v-else :size="13" :stroke-width="2" />
            {{ isSaving ? '保存中...' : '保存' }}
          </button>
        </div>
      </header>

      <!-- Form Body -->
      <div class="form-body">
        <!-- Top Fields Row -->
        <div class="form-row">
          <!-- Name -->
          <div class="form-group">
            <label class="form-label">
              <Type :size="12" :stroke-width="2" class="label-icon" />
              名称 <span class="required">*</span>
            </label>
            <input
              v-model="name"
              type="text"
              class="form-input"
              placeholder="输入模板名称"
              maxlength="64"
            />
          </div>

          <!-- Category -->
          <div class="form-group">
            <label class="form-label">
              <Tag :size="12" :stroke-width="2" class="label-icon" />
              分类
            </label>
            <input
              v-model="category"
              type="text"
              class="form-input"
              placeholder="输入或选择分类"
              maxlength="32"
              list="category-suggestions"
            />
            <datalist id="category-suggestions">
              <option v-for="cat in categories" :key="cat" :value="cat" />
            </datalist>
          </div>
        </div>

        <!-- Cover Upload -->
        <div class="form-group">
          <label class="form-label">
            <ImagePlus :size="12" :stroke-width="2" class="label-icon" />
            封面图
          </label>
          <div class="cover-upload-area">
            <input
              type="file"
              accept="image/*"
              class="hidden"
              id="template-cover-upload"
              @change="handleCoverSelect"
            />
            <div v-if="coverPreviewUrl" class="cover-preview-wrapper">
              <img :src="coverPreviewUrl" class="cover-preview-img" />
              <button class="cover-clear" @click="clearCover">
                <X :size="12" :stroke-width="2.5" />
              </button>
            </div>
            <label v-else for="template-cover-upload" class="cover-upload-label">
              <ImagePlus :size="18" :stroke-width="1.5" />
              <span>点击上传封面图</span>
            </label>
          </div>
        </div>

        <!-- Content Editor -->
        <div class="form-group editor-group">
          <label class="form-label">
            <FileText :size="12" :stroke-width="2" class="label-icon" />
            内容 <span class="required">*</span>
          </label>
          <div class="editor-wrapper">
            <template v-if="editorReady">
              <MdEditor
                v-model="content"
                theme="light"
                preview-theme="wechat"
                language="zh-CN"
                :toolbars="[
                  'bold', 'italic', 'strikethrough', 'heading', '|',
                  'quote', 'unordered-list', 'ordered-list', '|',
                  'link', 'image', 'table', 'code', 'code-block', '|',
                  'preview', 'fullscreen', '=',
                  'undo', 'redo',
                ]"
                :preview="false"
                :html-preview="false"
              />
            </template>
            <div v-else class="editor-loading">
              <Loader2 :size="16" class="animate-spin" />
              <span>加载编辑器...</span>
            </div>
          </div>
        </div>
      </div>
    </template>
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
  max-width: 900px;
  margin: 0 auto;
  width: 100%;
}

/* ─── Loading State ──────────────────────────────────────────── */

.loading-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--color-text-tertiary);
  font-size: 14px;
}

/* ─── Header ─────────────────────────────────────────────────── */

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: background 0.12s, color 0.12s;
}

.back-btn:hover {
  background: var(--color-surface-sunken);
  color: var(--color-text-primary);
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* ─── Form Body ──────────────────────────────────────────────── */

.form-body {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-bottom: 20px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

@media (max-width: 640px) {
  .form-row {
    grid-template-columns: 1fr;
  }
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  gap: 5px;
  letter-spacing: 0.02em;
}

.label-icon {
  color: var(--color-text-tertiary);
}

.required {
  color: var(--color-status-error);
}

.form-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  font-family: var(--font-sans);
  color: var(--color-text-primary);
  background: var(--color-surface);
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.form-input:focus {
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px var(--color-accent-subtle);
}

.form-input::placeholder {
  color: var(--color-text-tertiary);
}

/* ─── Cover Upload ───────────────────────────────────────────── */

.cover-upload-area {
  position: relative;
}

.cover-preview-wrapper {
  position: relative;
  display: inline-block;
  max-width: 280px;
}

.cover-preview-img {
  max-width: 100%;
  max-height: 180px;
  border-radius: 10px;
  display: block;
  border: 1px solid var(--color-border-subtle);
}

.cover-clear {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #111827;
  color: #fff;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s;
  padding: 0;
}

.cover-clear:hover {
  background: var(--color-status-error);
}

.cover-upload-label {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 200px;
  height: 120px;
  border: 1.5px dashed var(--color-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s;
  color: var(--color-text-tertiary);
  font-size: 12px;
}

.cover-upload-label:hover {
  border-color: var(--color-accent);
  background: var(--color-accent-subtle);
  color: var(--color-accent);
}

/* ─── Editor ──────────────────────────────────────────────────── */

.editor-group {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.editor-wrapper {
  flex: 1;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
  min-height: 400px;
  display: flex;
  flex-direction: column;
}

.editor-wrapper :deep(.md-editor) {
  height: 100%;
  border: none;
  border-radius: 0;
  display: flex;
  flex-direction: column;
  flex: 1;
}

.editor-wrapper :deep(.md-editor-toolbar) {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border-subtle);
  flex-shrink: 0;
}

.editor-wrapper :deep(.md-editor-content) {
  background: var(--color-surface);
  flex: 1;
  display: flex;
}

.editor-wrapper :deep(.md-editor-input) {
  background: var(--color-surface);
  color: var(--color-text-primary);
  flex: 1;
}

.editor-wrapper :deep(.cm-editor) {
  background: var(--color-surface);
}

.editor-wrapper :deep(.cm-content) {
  font-family: 'Geist Mono', ui-monospace, monospace;
  font-size: 14px;
  line-height: 1.85;
}

.editor-wrapper :deep(.cm-gutters) {
  background: var(--color-surface);
  border-right: 1px solid var(--color-border-subtle);
}

.editor-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 100%;
  min-height: 300px;
  color: var(--color-text-tertiary);
  font-size: 13px;
  background: var(--color-surface);
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
