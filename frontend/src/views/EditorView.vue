<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { defineAsyncComponent } from 'vue'
import 'md-editor-v3/lib/style.css'

const MdEditor = defineAsyncComponent(() =>
  import('md-editor-v3').then(m => m.MdEditor)
)
import { editorApi, articleApi, templateApi, tagApi } from '@/api/client.js'
import { useAuthStore } from '@/stores/auth'
import {
  Save,
  Send,
  ImagePlus,
  X,
  Loader2,
  Clock,
  FileText,
  LayoutTemplate,
  ChevronDown,
  CircleAlert,
  CircleCheck,
  Tag,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// ─── State ─────────────────────────────────────────────────────

const articleId = computed(() => route.params.id || null)
const isNewArticle = computed(() => !articleId.value)
const markdown = ref('')
const articleTitle = ref('')
const lastSavedAt = ref(null)
const isSaving = ref(false)
const editorReady = ref(false)
const autoSaveStatus = ref('idle') // 'idle' | 'saving' | 'saved' | 'error'
const autoSaveError = ref('')
let autoSaveTimer = null
let suppressAutoSave = false // suppress auto-save during initial load

// ─── Template Picker ─────────────────────────────────────────────

const showTemplatePicker = ref(false)
const templates = ref([])
const templateLoading = ref(false)

async function openTemplatePicker() {
  showTemplatePicker.value = true
  templateLoading.value = true
  try {
    const result = await templateApi.list()
    templates.value = Array.isArray(result) ? result : []
  } catch (e) {
    templates.value = []
    const msg = e?.message || '未知错误'
    if (!msg.includes('401')) {
      showToast('加载模板失败：' + msg, 'error')
    }
  } finally {
    templateLoading.value = false
  }
}

function applyTemplate(tpl) {
  markdown.value = tpl.content
  coverPreviewUrl.value = tpl.cover_url
  showTemplatePicker.value = false
  updatePreview()
}

// ─── Preview ────────────────────────────────────────────────────

const previewHtml = ref('')
const previewTitle = ref('')
const previewSummary = ref('')
const isPreviewLoading = ref(false)
let previewTimer = null

// ─── Cover ──────────────────────────────────────────────────────

const coverImage = ref(null)
const coverPreviewUrl = ref('')

// ─── Tags ────────────────────────────────────────────────────────

const articleTags = ref([])
const tagInput = ref('')
const allTags = ref([])
const showTagSuggestions = ref(false)

const tagSuggestions = computed(() => {
  const input = tagInput.value.trim().toLowerCase()
  if (!input) return []
  return allTags.value
    .filter(t => t.name.toLowerCase().includes(input) && !articleTags.value.some(at => at.id === t.id))
    .slice(0, 5)
})

async function addTagFromInput() {
  const name = tagInput.value.trim()
  if (!name) return
  // Check if already added
  if (articleTags.value.some(t => t.name === name)) {
    tagInput.value = ''
    showTagSuggestions.value = false
    return
  }
  // Check if tag exists in allTags
  const existing = allTags.value.find(t => t.name === name)
  if (existing) {
    articleTags.value.push(existing)
  } else {
    // Create new tag
    try {
      const created = await tagApi.create(name)
      allTags.value.push(created)
      articleTags.value.push(created)
    } catch (e) {
      showToast('创建标签失败：' + e.message, 'error')
    }
  }
  tagInput.value = ''
  showTagSuggestions.value = false
}

function selectTagSuggestion(tag) {
  if (!articleTags.value.some(t => t.id === tag.id)) {
    articleTags.value.push(tag)
  }
  tagInput.value = ''
  showTagSuggestions.value = false
}

function removeTag(tagId) {
  articleTags.value = articleTags.value.filter(t => t.id !== tagId)
}

function onTagInputKeydown(e) {
  if (e.key === 'Enter' || e.key === ',') {
    e.preventDefault()
    addTagFromInput()
  }
  if (e.key === 'Backspace' && !tagInput.value && articleTags.value.length > 0) {
    articleTags.value.pop()
  }
}

function onTagInputBlur() {
  // Delay to allow click on suggestion
  setTimeout(() => { showTagSuggestions.value = false }, 150)
}

// ─── Publish ────────────────────────────────────────────────────

const isPublishing = ref(false)
const weConfig = ref({ app_id: '', has_secret: false, default_author: '', last_author: '' })

// ─── Publish Success Modal ──────────────────────────────────────

const showPublishSuccess = ref(false)
const publishResult = ref(null)

// ─── Last Cover Hint ────────────────────────────────────────────

const lastCoverName = ref(localStorage.getItem('wx_note_last_cover_name') || '')

// ─── Toast ──────────────────────────────────────────────────────

const toasts = ref([])
function showToast(msg, type = 'info') {
  const id = Date.now()
  toasts.value.push({ id, msg, type })
  setTimeout(() => (toasts.value = toasts.value.filter((t) => t.id !== id)), 3500)
}

// ─── Computed ───────────────────────────────────────────────────

const displayTitle = computed(() => {
  if (previewTitle.value) return previewTitle.value
  if (articleTitle.value) return articleTitle.value
  return '新建文章'
})

const lastSavedText = computed(() => {
  if (!lastSavedAt.value) return ''
  const d = new Date(lastSavedAt.value)
  const pad = (n) => String(n).padStart(2, '0')
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
})

const autoSaveDisplay = computed(() => {
  if (autoSaveStatus.value === 'saving') return { text: '保存中...', icon: 'loader', color: 'var(--color-text-tertiary)' }
  if (autoSaveStatus.value === 'saved') return { text: `已自动保存 ${lastSavedText.value}`, icon: 'clock', color: 'var(--color-status-success)' }
  if (autoSaveStatus.value === 'error') return { text: '保存失败', icon: 'error', color: 'var(--color-status-error)' }
  // idle — show last saved time if available
  if (lastSavedText.value) return { text: `上次保存 ${lastSavedText.value}`, icon: 'clock', color: 'var(--color-text-tertiary)' }
  return null
})

const canPublish = computed(() => {
  return (
    weConfig.value.app_id &&
    weConfig.value.has_secret &&
    coverImage.value &&
    markdown.value.trim() &&
    !isPublishing.value
  )
})

	const publishHint = computed(() => {
	  if (!weConfig.value?.app_id || !weConfig.value?.has_secret) return '请先配置公众号'
	  if (!coverImage.value) return '请上传封面'
	  if (!markdown.value.trim()) return '请输入内容'
	  return ''
	})

// ─── Preview (debounced) ────────────────────────────────────────

async function updatePreview() {
  if (!markdown.value.trim()) {
    previewHtml.value = ''
    previewTitle.value = ''
    previewSummary.value = ''
    return
  }
  isPreviewLoading.value = true
  try {
    const result = await editorApi.preview(markdown.value)
    previewHtml.value = result.html || ''
    previewTitle.value = result.title || ''
    previewSummary.value = result.summary || ''
  } catch (e) {
    // Silent
  } finally {
    isPreviewLoading.value = false
  }
}

watch(markdown, () => {
  clearTimeout(previewTimer)
  previewTimer = setTimeout(updatePreview, 600)

  // Auto-save with 3s debounce
  clearTimeout(autoSaveTimer)
  if (!markdown.value.trim() || suppressAutoSave) return
  autoSaveTimer = setTimeout(() => {
    doSave({ notify: false, isAuto: true })
  }, 3000)
})

// ─── Article CRUD ────────────────────────────────────────────────

async function loadArticle(id) {
  suppressAutoSave = true
  try {
    const data = await articleApi.get(id)
    markdown.value = data.markdown || ''
    articleTitle.value = data.title || ''
    articleTags.value = data.tags || []
    lastSavedAt.value = data.updated_at || data.created_at || null
    await nextTick()
    await updatePreview()
  } catch (e) {
    showToast('加载文章失败：' + e.message, 'error')
  } finally {
    suppressAutoSave = false
  }
}

async function doSave({ notify = true, isAuto = false } = {}) {
  if (!markdown.value.trim()) {
    if (notify) showToast('内容为空，无需保存', 'info')
    return
  }
  // If a manual save is in progress, skip auto save
  if (isAuto && isSaving.value) return
  isSaving.value = true
  if (isAuto) autoSaveStatus.value = 'saving'
  try {
    const payload = {
      title: previewTitle.value || articleTitle.value || '未命名文章',
      markdown: markdown.value,
      tag_ids: articleTags.value.map(t => t.id),
    }
    if (isNewArticle.value) {
      const result = await articleApi.create(payload)
      articleTitle.value = result.title || payload.title
      lastSavedAt.value = result.created_at || new Date().toISOString()
      if (isAuto) {
        autoSaveStatus.value = 'saved'
        if (result.id) router.replace(`/editor/${result.id}`)
      } else {
        showToast('文章已创建', 'success')
        if (result.id) router.replace(`/editor/${result.id}`)
      }
    } else {
      await articleApi.update(articleId.value, payload)
      lastSavedAt.value = new Date().toISOString()
      if (isAuto) {
        autoSaveStatus.value = 'saved'
      } else {
        showToast('保存成功', 'success')
      }
    }
  } catch (e) {
    if (isAuto) {
      autoSaveStatus.value = 'error'
      autoSaveError.value = e?.message || '未知错误'
    } else {
      showToast('保存失败：' + e.message, 'error')
    }
  } finally {
    isSaving.value = false
  }
}

async function saveArticle() {
  // Cancel pending auto-save when user manually saves
  clearTimeout(autoSaveTimer)
  autoSaveTimer = null
  await doSave({ notify: true, isAuto: false })
  // After manual save succeeds, reflect saved state in auto-save indicator
  if (!isSaving.value && lastSavedAt.value) {
    autoSaveStatus.value = 'saved'
  }
}

// ─── Cover ──────────────────────────────────────────────────────

function handleCoverSelect(e) {
  const file = e.target.files[0]
  if (!file) return
  const validTypes = ['image/jpeg', 'image/png', 'image/webp']
  if (!validTypes.includes(file.type)) {
    showToast('请上传 JPG / PNG 格式的图片', 'error')
    e.target.value = ''
    return
  }
  if (file.size > 2 * 1024 * 1024) {
    compressImage(file).then(compressed => {
      coverImage.value = compressed
      coverPreviewUrl.value = URL.createObjectURL(compressed)
    }).catch(() => {
      coverImage.value = file
      coverPreviewUrl.value = URL.createObjectURL(file)
    })
  } else {
    coverImage.value = file
    coverPreviewUrl.value = URL.createObjectURL(file)
  }
}

function compressImage(file, maxWidth = 900, quality = 0.85) {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => {
      const canvas = document.createElement('canvas')
      let w = img.width, h = img.height
      if (w > maxWidth) { h = Math.round(h * maxWidth / w); w = maxWidth }
      canvas.width = w; canvas.height = h
      const ctx = canvas.getContext('2d')
      ctx.drawImage(img, 0, 0, w, h)
      canvas.toBlob(blob => blob ? resolve(new File([blob], file.name, { type: 'image/jpeg' })) : reject(), 'image/jpeg', quality)
    }
    img.onerror = reject
    img.src = URL.createObjectURL(file)
  })
}

function clearCover() {
  coverImage.value = null
  coverPreviewUrl.value = ''
}

// ─── Publish ────────────────────────────────────────────────────

async function handlePublish() {
  if (!canPublish.value) {
    if (!weConfig.value.app_id || !weConfig.value.has_secret) {
      showToast('请先配置微信公众号凭据', 'error')
      return
    }
    if (!coverImage.value) {
      showToast('请上传封面图片', 'error')
      return
    }
    return
  }
  isPublishing.value = true
  try {
    const result = await editorApi.publish({
      markdown: markdown.value,
      cover: coverImage.value,
      author: weConfig.value.last_author || weConfig.value.default_author,
    })
    if (result.ok) {
      // Save cover filename to localStorage
      if (coverImage.value && coverImage.value.name) {
        localStorage.setItem('wx_note_last_cover_name', coverImage.value.name)
        lastCoverName.value = coverImage.value.name
      }
      publishResult.value = result
      showPublishSuccess.value = true
    }
  } catch (e) {
    showToast('发布失败：' + e.message, 'error')
  } finally {
    isPublishing.value = false
  }
}

// ─── md-editor image upload ─────────────────────────────────────

async function handleEditorUploadImage(files, callback) {
  try {
    const urls = []
    for (const file of files) {
      const result = await editorApi.uploadImage(file)
      urls.push(result.url)
    }
    callback(urls)
  } catch (e) {
    showToast('图片上传失败：' + e.message, 'error')
  }
}

// ─── Init ───────────────────────────────────────────────────────

async function loadWeConfig() {
  try {
    const cfg = await authStore.fetchConfig()
    weConfig.value = {
      app_id: cfg.wechat_app_id || '',
      has_secret: cfg.has_secret || false,
      default_author: cfg.default_author || '',
      last_author: cfg.last_author || '',
    }
  } catch (e) {
    // Config may not be available
  }
}

onMounted(async () => {
  if (!authStore.isLoggedIn) {
    router.push('/login')
    return
  }
  await loadWeConfig()
  try {
    const tagData = await tagApi.list()
    allTags.value = Array.isArray(tagData) ? tagData : []
  } catch (e) {
    allTags.value = []
  }
  if (articleId.value) {
    await loadArticle(articleId.value)
  } else {
    await nextTick()
    await updatePreview()
  }
  // Delay editor loading to avoid SSR/hydration issues
  setTimeout(() => { editorReady.value = true }, 100)
})

</script>

<template>
  <div class="editor-layout">
    <!-- ─── Left: Editor ─────────────────────────────────────────── -->
    <div class="editor-pane">
      <!-- Toolbar -->
      <div class="editor-toolbar">
        <div class="toolbar-left">
          <FileText :size="14" :stroke-width="1.8" class="toolbar-icon" />
          <span class="toolbar-title">{{ displayTitle }}</span>
        </div>
        <div class="toolbar-right">
          <span v-if="autoSaveDisplay" class="toolbar-autosave" :style="{ color: autoSaveDisplay.color }">
            <Loader2 v-if="autoSaveDisplay.icon === 'loader'" :size="11" class="animate-spin" />
            <Clock v-else-if="autoSaveDisplay.icon === 'clock'" :size="11" :stroke-width="2" />
            <CircleAlert v-else-if="autoSaveDisplay.icon === 'error'" :size="11" :stroke-width="2" />
            {{ autoSaveDisplay.text }}
            <span v-if="autoSaveStatus === 'error' && autoSaveError" class="autosave-error-detail">{{ autoSaveError }}</span>
          </span>
          <button class="btn btn-ghost btn-sm" @click="openTemplatePicker">
            <LayoutTemplate :size="13" :stroke-width="2" />
            模板
            <ChevronDown :size="11" :stroke-width="2" />
          </button>
          <button class="btn btn-secondary btn-sm" :disabled="isSaving" @click="saveArticle">
            <Loader2 v-if="isSaving" :size="12" class="animate-spin" />
            <Save v-else :size="12" :stroke-width="2" />
            {{ isSaving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>

      <!-- md-editor -->
      <div class="editor-container">
        <template v-if="editorReady">
          <MdEditor
            v-model="markdown"
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
            @on-upload-img="handleEditorUploadImage"
          />
        </template>
        <div v-else class="editor-loading">
          <Loader2 :size="16" class="animate-spin" />
          <span>加载编辑器...</span>
        </div>
      </div>
    </div>

    <!-- ─── Right: Preview + Publish ─────────────────────────────── -->
    <div class="preview-pane">
      <!-- Phone Preview -->
      <div class="phone-frame">
        <div class="phone-screen">
          <div class="phone-content">
            <div v-if="previewTitle" class="wx-article">
              <h1>{{ previewTitle }}</h1>
            </div>
            <div v-if="previewHtml" class="wx-article" v-html="previewHtml" />
            <div v-else class="preview-placeholder">
              <FileText :size="28" :stroke-width="1.2" class="placeholder-icon" />
              <p>开始输入 Markdown 内容<br />预览将在此处显示</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Publish Bar -->
      <div class="publish-bar">
        <div class="publish-bar-inner">
          <!-- Cover -->
          <div class="cover-section">
            <input type="file" accept="image/jpeg,image/png,image/webp" class="hidden" id="cover-upload" @change="handleCoverSelect" />
            <label for="cover-upload" class="cover-upload">
              <template v-if="coverPreviewUrl">
                <img :src="coverPreviewUrl" class="cover-thumb" />
              </template>
              <template v-else>
                <ImagePlus :size="16" :stroke-width="1.5" />
                <span>封面</span>
              </template>
            </label>
            <button v-if="coverImage" class="cover-clear" @click="clearCover">
              <X :size="10" :stroke-width="2.5" />
            </button>
            <div v-if="!coverImage && lastCoverName" class="cover-hint">上次使用：{{ lastCoverName }}</div>
          </div>

          <!-- Summary -->
          <div class="summary-section">
            <div class="summary-label">摘要</div>
            <div v-if="previewSummary" class="summary-text">{{ previewSummary }}</div>
            <div v-else class="summary-empty">添加正文内容以自动生成摘要</div>
          </div>

          <!-- Tags -->
          <div class="tag-section">
            <div class="tag-label">
              <Tag :size="11" :stroke-width="2" />
              标签
            </div>
            <div class="tag-input-wrapper">
              <span
                v-for="tag in articleTags"
                :key="tag.id"
                class="editor-tag-badge"
              >
                {{ tag.name }}
                <button class="tag-remove-btn" @click="removeTag(tag.id)">
                  <X :size="9" :stroke-width="2.5" />
                </button>
              </span>
              <input
                v-model="tagInput"
                type="text"
                class="tag-input"
                placeholder="输入标签..."
                @keydown="onTagInputKeydown"
                @focus="showTagSuggestions = true"
                @blur="onTagInputBlur"
              />
              <div v-if="showTagSuggestions && tagSuggestions.length > 0" class="tag-suggestions">
                <div
                  v-for="suggestion in tagSuggestions"
                  :key="suggestion.id"
                  class="tag-suggestion-item"
                  @mousedown.prevent="selectTagSuggestion(suggestion)"
                >
                  {{ suggestion.name }}
                </div>
              </div>
            </div>
          </div>

          <!-- Publish -->
          <button class="btn btn-primary publish-btn" :disabled="!canPublish" @click="handlePublish">
            <Loader2 v-if="isPublishing" :size="13" class="animate-spin" />
            <Send v-else :size="13" :stroke-width="2" />
            {{ isPublishing ? '发布中...' : '发布草稿' }}
          </button>
          <span v-if="!canPublish && !isPublishing && publishHint" class="publish-hint">{{ publishHint }}</span>
        </div>
      </div>
    </div>

    <!-- Template Picker Modal -->
    <div v-if="showTemplatePicker" class="modal-overlay" @click.self="showTemplatePicker = false">
      <div class="modal template-picker-modal">
        <div class="modal-header">
          <h3 class="modal-title">选择模板</h3>
          <button class="btn btn-ghost btn-sm" @click="showTemplatePicker = false">
            <X :size="14" :stroke-width="2" />
          </button>
        </div>
        <div class="modal-body">
          <div v-if="templateLoading" class="template-loading">
            <Loader2 :size="16" class="animate-spin" />
            <span>加载中...</span>
          </div>
          <div v-else-if="!templates || templates.length === 0" class="template-empty">
            <LayoutTemplate :size="24" :stroke-width="1.2" class="empty-icon" />
            <p>暂无模板</p>
            <router-link to="/templates/new" class="btn btn-primary btn-sm" @click="showTemplatePicker = false">
              去创建模板
            </router-link>
          </div>
          <div v-else class="template-grid">
            <div
              v-for="tpl in templates"
              :key="tpl.id"
              class="template-card"
              @click="applyTemplate(tpl)"
            >
              <div class="tpl-name">{{ tpl.name }}</div>
              <div class="tpl-category">{{ tpl.category }}</div>
              <div class="tpl-preview">{{ (tpl.content || '').slice(0, 80) }}...</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Publish Success Modal -->
    <div v-if="showPublishSuccess" class="modal-overlay" @click.self="showPublishSuccess = false">
      <div class="modal publish-success-modal">
        <div class="modal-header">
          <h3 class="modal-title">发布成功</h3>
          <button class="btn btn-ghost btn-sm" @click="showPublishSuccess = false">
            <X :size="14" :stroke-width="2" />
          </button>
        </div>
        <div class="modal-body">
          <div class="publish-success-content">
            <div class="success-icon-row">
              <CircleCheck :size="32" :stroke-width="1.5" class="success-icon" />
            </div>
            <p class="success-article-title">{{ previewTitle || articleTitle || '未命名文章' }}</p>
            <p v-if="publishResult && publishResult.draft_media_id" class="success-draft-id">
              草稿 ID：{{ publishResult.draft_media_id.slice(0, 16) }}
            </p>
            <a href="https://mp.weixin.qq.com" target="_blank" rel="noopener" class="success-link">
              前往微信公众平台查看 &rarr;
            </a>
          </div>
          <div class="success-actions">
            <button class="btn btn-secondary" @click="showPublishSuccess = false">
              继续编辑
            </button>
            <button class="btn btn-primary" @click="showPublishSuccess = false; router.push('/articles')">
              返回列表
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Toast -->
    <div class="toast-container">
      <div v-for="toast in toasts" :key="toast.id" class="toast" :class="`toast-${toast.type}`">
        {{ toast.msg }}
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ─── Layout ──────────────────────────────────────────────────── */

.editor-layout {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
  background: var(--color-surface-sunken);
}

/* ─── Left: Editor ─────────────────────────────────────────────── */

.editor-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-right: 1px solid var(--color-border);
  min-width: 0;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border-subtle);
  flex-shrink: 0;
  height: 44px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.toolbar-icon {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

.toolbar-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 280px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.toolbar-autosave {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}

.autosave-error-detail {
  font-size: 10px;
  color: var(--color-text-tertiary);
  margin-left: 2px;
}

.editor-container {
  flex: 1;
  overflow: hidden;
}

.editor-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 100%;
  color: var(--color-text-tertiary);
  font-size: 13px;
  background: var(--color-surface);
}

.editor-container :deep(.md-editor) {
  height: 100%;
  border: none;
  border-radius: 0;
}

.editor-container :deep(.md-editor-toolbar) {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border-subtle);
}

.editor-container :deep(.md-editor-content) {
  background: var(--color-surface);
}

.editor-container :deep(.md-editor-input) {
  background: var(--color-surface);
  color: var(--color-text-primary);
}

.editor-container :deep(.cm-editor) {
  background: var(--color-surface);
}

.editor-container :deep(.cm-content) {
  font-family: 'Geist Mono', ui-monospace, monospace;
  font-size: 14px;
  line-height: 1.8;
}

.editor-container :deep(.cm-gutters) {
  background: var(--color-surface);
  border-right: 1px solid var(--color-border-subtle);
}

/* ─── Right: Preview ──────────────────────────────────────────── */

.preview-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f3f4f6;
  min-width: 0;
}

.phone-frame {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 16px 12px;
  overflow: hidden;
}

.phone-screen {
  width: 375px;
  background: var(--color-surface);
  border-radius: 20px;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.06), 0 24px 48px -12px rgba(0, 0, 0, 0.12);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  flex: 1;
  max-height: calc(100dvh - 200px);
}

.phone-content {
  padding: 16px 16px 24px;
  overflow-y: auto;
  flex: 1;
}

.preview-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  gap: 12px;
}

.placeholder-icon {
  color: #d1d5db;
}

.preview-placeholder p {
  font-size: 13px;
  color: var(--color-text-tertiary);
  text-align: center;
  line-height: 1.7;
  margin: 0;
}

/* ─── Publish Bar ─────────────────────────────────────────────── */

.publish-bar {
  flex-shrink: 0;
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
  padding: 14px 20px;
}

.publish-bar-inner {
  display: flex;
  align-items: center;
  gap: 16px;
  max-width: 600px;
  margin: 0 auto;
}

.cover-section {
  position: relative;
  flex-shrink: 0;
}

.cover-upload {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  width: 64px;
  height: 64px;
  border: 1.5px dashed var(--color-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s;
  color: var(--color-text-tertiary);
  font-size: 11px;
  overflow: hidden;
  flex-shrink: 0;
}

.cover-upload:hover {
  border-color: var(--color-accent);
  background: var(--color-accent-subtle);
  color: var(--color-accent);
}

.cover-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 8px;
}

.cover-clear {
  position: absolute;
  top: -5px;
  right: -5px;
  width: 18px;
  height: 18px;
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

.cover-hint {
  font-size: 10px;
  color: var(--color-text-tertiary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 64px;
  text-align: center;
  margin-top: 2px;
}

.summary-section {
  flex: 1;
  min-width: 0;
}

.summary-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--color-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 4px;
}

.summary-text {
  font-size: 12px;
  color: var(--color-text-secondary);
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.summary-empty {
  font-size: 12px;
  color: var(--color-text-tertiary);
  font-style: italic;
}

.publish-btn {
  flex-shrink: 0;
  padding: 9px 20px;
  border-radius: 10px;
}

.publish-hint {
  font-size: 11px;
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

/* ─── Tag Input ──────────────────────────────────────────────── */

.tag-section {
  flex-shrink: 0;
  min-width: 0;
}

.tag-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 500;
  color: var(--color-text-tertiary);
  letter-spacing: 0.04em;
  margin-bottom: 4px;
}

.tag-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
  min-height: 28px;
  flex-wrap: wrap;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.tag-input-wrapper:focus-within {
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px var(--color-accent-subtle);
}

.editor-tag-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  background: #f3f4f6;
  color: #374151;
  white-space: nowrap;
  letter-spacing: 0.02em;
}

.tag-remove-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  padding: 0;
  cursor: pointer;
  color: #9ca3af;
  transition: color 0.12s;
  height: 12px;
  width: 12px;
}

.tag-remove-btn:hover {
  color: #374151;
}

.tag-input {
  flex: 1;
  min-width: 60px;
  border: none;
  outline: none;
  font-size: 12px;
  font-family: var(--font-sans);
  color: var(--color-text-primary);
  background: transparent;
  padding: 0;
  line-height: 1.4;
}

.tag-input::placeholder {
  color: var(--color-text-tertiary);
}

.tag-suggestions {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  z-index: 100;
  margin-top: 4px;
  overflow: hidden;
}

.tag-suggestion-item {
  padding: 6px 12px;
  font-size: 12px;
  color: var(--color-text-primary);
  cursor: pointer;
  transition: background 0.12s;
}

.tag-suggestion-item:hover {
  background: #f3f4f6;
}

/* ─── Toast ───────────────────────────────────────────────────── */

.toast-container {
  position: fixed;
  top: 16px;
  right: 16px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.toast {
  padding: 12px 18px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  color: #fff;
  animation: toast-in 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  max-width: 380px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  display: flex;
  align-items: center;
  gap: 8px;
}

.toast-success { background: #111; }
.toast-success::before { content: ''; width: 6px; height: 6px; background: var(--color-status-success); border-radius: 50%; flex-shrink: 0; }

.toast-error { background: #111; }
.toast-error::before { content: ''; width: 6px; height: 6px; background: var(--color-status-error); border-radius: 50%; flex-shrink: 0; }

.toast-info { background: #111; }
.toast-info::before { content: ''; width: 6px; height: 6px; background: var(--color-status-info); border-radius: 50%; flex-shrink: 0; }

@keyframes toast-in {
  from { transform: translateY(-8px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

/* ─── Publish Success Modal ───────────────────────────────────── */

.publish-success-modal {
  width: 400px;
  display: flex;
  flex-direction: column;
}

.publish-success-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding-bottom: 8px;
}

.success-icon-row {
  display: flex;
  justify-content: center;
}

.success-icon {
  color: var(--color-status-success);
}

.success-article-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
  text-align: center;
  margin: 0;
}

.success-draft-id {
  font-size: 12px;
  color: var(--color-text-tertiary);
  font-variant-numeric: tabular-nums;
  margin: 0;
}

.success-link {
  font-size: 13px;
  color: var(--color-accent);
  text-decoration: none;
  transition: opacity 0.15s;
}

.success-link:hover {
  opacity: 0.8;
}

.success-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--color-border-subtle);
  margin-top: 8px;
}

.success-actions .btn {
  min-width: 100px;
}

/* ─── Responsive ──────────────────────────────────────────────── */

@media (max-width: 900px) {
  .editor-layout { flex-direction: column; }
  .editor-pane { flex: none; height: 50%; border-right: none; border-bottom: 1px solid var(--color-border); }
  .preview-pane { flex: none; height: 50%; }
  .phone-screen { width: 320px; }
}
</style>
