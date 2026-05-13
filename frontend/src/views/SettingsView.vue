<script setup>
import { ref, onMounted, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { authApi, editorApi } from '@/api/client.js'
import {
  Save,
  ShieldCheck,
  ExternalLink,
  User,
  Key,
  FileText,
  Check,
  AlertCircle,
  Loader2,
  Settings,
} from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

// ─── Auth Guard ────────────────────────────────────────────────

onMounted(async () => {
  if (!authStore.isLoggedIn) {
    router.push('/login')
    return
  }
  await Promise.all([loadConfig(), loadProfile()])
})

// ─── Config State ──────────────────────────────────────────────

const config = reactive({
  app_id: '',
  app_secret: '',
  has_secret: false,
  default_author: '',
})

const isConfigSaving = ref(false)
const isVerifying = ref(false)

const appIdPrefix = computed(() => {
  if (!config.app_id) return ''
  return config.app_id.slice(0, 12)
})

const isConnected = computed(() => config.app_id && config.has_secret)

async function loadConfig() {
  try {
    const cfg = await authApi.getConfig()
    config.app_id = cfg.wechat_app_id || cfg.app_id || ''
    config.has_secret = !!cfg.has_secret
    config.default_author = cfg.default_author || ''
  } catch (err) {
    showToast('加载配置失败：' + (err.message || '未知错误'), 'error')
  }
}

async function saveConfig() {
  isConfigSaving.value = true
  try {
    const payload = {
      app_id: config.app_id,
      default_author: config.default_author,
    }
    if (config.app_secret) {
      payload.app_secret = config.app_secret
    }
    await authApi.updateConfig(payload)
    showToast('配置已保存', 'success')
    config.app_secret = ''
    await loadConfig()
  } catch (err) {
    showToast('保存失败：' + (err.message || '未知错误'), 'error')
  } finally {
    isConfigSaving.value = false
  }
}

async function verifyConnection() {
  isVerifying.value = true
  try {
    await editorApi.verify()
    showToast('连接验证成功', 'success')
    await loadConfig()
  } catch (err) {
    showToast('验证失败：' + (err.message || '未知错误'), 'error')
  } finally {
    isVerifying.value = false
  }
}

// ─── Profile State ─────────────────────────────────────────────

const profile = reactive({
  nickname: '',
  username: '',
})

const isProfileSaving = ref(false)

async function loadProfile() {
  try {
    const data = await authApi.getProfile()
    profile.nickname = data.nickname || ''
    profile.username = data.username || ''
  } catch (err) {
    showToast('加载个人信息失败：' + (err.message || '未知错误'), 'error')
  }
}

async function updateProfile() {
  isProfileSaving.value = true
  try {
    await authApi.updateProfile({ nickname: profile.nickname })
    showToast('个人信息已更新', 'success')
  } catch (err) {
    showToast('更新失败：' + (err.message || '未知错误'), 'error')
  } finally {
    isProfileSaving.value = false
  }
}

// ─── Toast ─────────────────────────────────────────────────────

const toasts = ref([])

function showToast(msg, type = 'info') {
  const id = Date.now()
  toasts.value.push({ id, msg, type })
  setTimeout(() => {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }, 3500)
}
</script>

<template>
  <div class="settings-page">
    <!-- Page Header -->
    <div class="settings-header">
      <div class="settings-header-inner">
        <div class="settings-header-icon">
          <Settings :size="18" :stroke-width="2" />
        </div>
        <div>
          <h1 class="settings-title">设置</h1>
          <p class="settings-subtitle">管理公众号配置与个人信息</p>
        </div>
      </div>
    </div>

    <!-- Two-column asymmetric layout -->
    <div class="settings-body">
      <!-- Left: 60% — Forms -->
      <div class="settings-left">
        <!-- WeChat Config Section -->
        <section class="settings-section">
          <div class="section-header">
            <h2 class="section-title">
              <FileText :size="15" :stroke-width="2" class="section-icon" />
              公众号配置
            </h2>
            <p class="section-desc">配置微信公众号 API 凭据，用于文章发布</p>
          </div>

          <div class="form-fields">
            <div class="field-group">
              <label class="form-label" for="cfg-appid">AppID</label>
              <div class="input-with-icon">
                <Key :size="14" :stroke-width="1.8" class="input-icon" />
                <input
                  id="cfg-appid"
                  v-model="config.app_id"
                  type="text"
                  class="form-input form-input-with-icon"
                  placeholder="wx..."
                />
              </div>
            </div>

            <div class="field-group">
              <label class="form-label" for="cfg-secret">AppSecret</label>
              <div class="input-with-icon">
                <ShieldCheck :size="14" :stroke-width="1.8" class="input-icon" />
                <input
                  id="cfg-secret"
                  v-model="config.app_secret"
                  type="password"
                  class="form-input form-input-with-icon"
                  :placeholder="config.has_secret ? '已配置 — 留空保持不变' : '输入 AppSecret'"
                />
              </div>
            </div>

            <div class="field-group">
              <label class="form-label" for="cfg-author">默认作者</label>
              <div class="input-with-icon">
                <User :size="14" :stroke-width="1.8" class="input-icon" />
                <input
                  id="cfg-author"
                  v-model="config.default_author"
                  type="text"
                  class="form-input form-input-with-icon"
                  placeholder="选填"
                />
              </div>
            </div>
          </div>

          <div class="section-actions">
            <button class="btn btn-primary" :disabled="isConfigSaving" @click="saveConfig">
              <Loader2 v-if="isConfigSaving" :size="14" class="animate-spin" />
              <Save v-else :size="14" :stroke-width="2" />
              {{ isConfigSaving ? '保存中...' : '保存配置' }}
            </button>
            <button class="btn btn-secondary" :disabled="isVerifying" @click="verifyConnection">
              <Loader2 v-if="isVerifying" :size="14" class="animate-spin" />
              <ShieldCheck v-else :size="14" :stroke-width="2" />
              {{ isVerifying ? '验证中...' : '验证连接' }}
            </button>
          </div>
        </section>

        <!-- Divider -->
        <hr class="settings-divider" />

        <!-- Profile Section -->
        <section class="settings-section">
          <div class="section-header">
            <h2 class="section-title">
              <User :size="15" :stroke-width="2" class="section-icon" />
              个人信息
            </h2>
            <p class="section-desc">更新你的账号信息</p>
          </div>

          <div class="form-fields">
            <div class="field-group">
              <label class="form-label" for="profile-nickname">昵称</label>
              <input
                id="profile-nickname"
                v-model="profile.nickname"
                type="text"
                class="form-input"
                placeholder="输入昵称"
              />
            </div>

            <div class="field-group">
              <label class="form-label">用户名</label>
              <input
                :value="profile.username"
                type="text"
                class="form-input form-input-readonly"
                readonly
                tabindex="-1"
              />
            </div>
          </div>

          <div class="section-actions">
            <button class="btn btn-primary" :disabled="isProfileSaving" @click="updateProfile">
              <Loader2 v-if="isProfileSaving" :size="14" class="animate-spin" />
              <Save v-else :size="14" :stroke-width="2" />
              {{ isProfileSaving ? '更新中...' : '更新个人信息' }}
            </button>
          </div>
        </section>
      </div>

      <!-- Right: 40% — Status Cards -->
      <div class="settings-right">
        <!-- Connection Status Card -->
        <div class="status-card">
          <div class="status-card-header">
            <h3 class="status-card-title">连接状态</h3>
          </div>
          <div class="status-card-body">
            <div class="status-row">
              <span
                class="status-dot"
                :class="[
                  isConnected ? 'status-dot-connected' : 'status-dot-unknown',
                ]"
              ></span>
              <span class="status-label">
                {{ isConnected ? '已连接' : '未配置' }}
              </span>
            </div>

            <div v-if="config.app_id" class="status-detail">
              <div class="status-detail-label">AppID</div>
              <div class="status-detail-value font-mono">{{ appIdPrefix }}...</div>
            </div>

            <div v-if="config.has_secret" class="status-detail">
              <div class="status-detail-label">AppSecret</div>
              <div class="status-detail-value">
                <Check :size="13" :stroke-width="2.5" class="text-status-success" />
                已配置
              </div>
            </div>

            <div v-if="config.default_author" class="status-detail">
              <div class="status-detail-label">默认作者</div>
              <div class="status-detail-value">{{ config.default_author }}</div>
            </div>
          </div>
        </div>

        <!-- Quick Links -->
        <div class="status-card">
          <div class="status-card-header">
            <h3 class="status-card-title">快捷链接</h3>
          </div>
          <div class="status-card-body">
            <a
              href="https://mp.weixin.qq.com"
              target="_blank"
              rel="noopener"
              class="quick-link"
            >
              <div class="quick-link-icon">
                <ExternalLink :size="14" :stroke-width="1.8" />
              </div>
              <div class="quick-link-content">
                <div class="quick-link-title">前往微信公众平台</div>
                <div class="quick-link-url">mp.weixin.qq.com</div>
              </div>
            </a>
          </div>
        </div>

        <!-- Tips Card -->
        <div class="status-card status-card-tips">
          <div class="status-card-body">
            <div class="tips-text">
              在微信公众平台获取 AppID 和 AppSecret 后，填入左侧表单并保存。点击"验证连接"确认配置正确。
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Toast Container -->
  <div class="toast-container">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="toast"
      :class="`toast-${toast.type}`"
    >
      {{ toast.msg }}
    </div>
  </div>
</template>

<style scoped>
/* ─── Page Layout ─────────────────────────────────────────────── */

.settings-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.settings-header {
  padding: 28px 40px 0;
  flex-shrink: 0;
}

.settings-header-inner {
  display: flex;
  align-items: center;
  gap: 14px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--color-border-subtle);
}

.settings-header-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: var(--color-surface-sunken);
  border: 1px solid var(--color-border-subtle);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.settings-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 2px;
  letter-spacing: -0.01em;
}

.settings-subtitle {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin: 0;
}

/* ─── Two-column body ─────────────────────────────────────────── */

.settings-body {
  flex: 1;
  display: flex;
  gap: 32px;
  padding: 28px 40px 40px;
  overflow-y: auto;
  min-height: 0;
}

.settings-left {
  flex: 0 0 60%;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.settings-right {
  flex: 0 0 calc(40% - 32px);
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ─── Section ─────────────────────────────────────────────────── */

.settings-section {
  padding: 8px 0 32px;
}

.section-header {
  margin-bottom: 24px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 4px;
  display: flex;
  align-items: center;
  gap: 8px;
  letter-spacing: -0.01em;
}

.section-icon {
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.section-desc {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin: 0;
}

.form-fields {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-bottom: 24px;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* ─── Input with icon ─────────────────────────────────────────── */

.input-with-icon {
  position: relative;
}

.input-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-tertiary);
  pointer-events: none;
}

.form-input-with-icon {
  padding-left: 38px;
}

.form-input-readonly {
  background: var(--color-surface-sunken);
  color: var(--color-text-tertiary);
  cursor: default;
}

.form-input-readonly:focus {
  border-color: var(--color-border);
  box-shadow: none;
}

/* ─── Section Actions ─────────────────────────────────────────── */

.section-actions {
  display: flex;
  gap: 10px;
}

/* ─── Divider ─────────────────────────────────────────────────── */

.settings-divider {
  border: none;
  height: 1px;
  background: var(--color-border-subtle);
  margin: 0;
}

/* ─── Status Card ─────────────────────────────────────────────── */

.status-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: 12px;
  overflow: hidden;
}

.status-card-header {
  padding: 14px 18px;
  border-bottom: 1px solid var(--color-border-subtle);
}

.status-card-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.status-card-body {
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot-connected {
  background: var(--color-status-success);
}

.status-dot-unknown {
  background: var(--color-text-tertiary);
}

.status-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-primary);
}

.status-detail {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--color-border-subtle);
}

.status-detail:last-child {
  padding-bottom: 0;
  border-bottom: none;
}

.status-detail-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--color-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  flex-shrink: 0;
}

.status-detail-value {
  font-size: 12px;
  color: var(--color-text-secondary);
  text-align: right;
  display: flex;
  align-items: center;
  gap: 4px;
}

/* ─── Quick Link ──────────────────────────────────────────────── */

.quick-link {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 10px;
  border: 1px solid var(--color-border-subtle);
  text-decoration: none;
  transition: background 0.15s, border-color 0.15s;
}

.quick-link:hover {
  background: var(--color-surface-sunken);
  border-color: var(--color-border);
}

.quick-link-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: var(--color-accent-subtle);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-accent);
  flex-shrink: 0;
}

.quick-link-content {
  min-width: 0;
}

.quick-link-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-primary);
  margin: 0 0 2px;
}

.quick-link-url {
  font-size: 11px;
  color: var(--color-text-tertiary);
  font-family: var(--font-mono);
}

/* ─── Tips ────────────────────────────────────────────────────── */

.status-card-tips {
  background: var(--color-surface-sunken);
  border-style: dashed;
}

.tips-text {
  font-size: 12px;
  color: var(--color-text-tertiary);
  line-height: 1.75;
  margin: 0;
}

/* ─── Responsive ──────────────────────────────────────────────── */

@media (max-width: 900px) {
  .settings-body {
    flex-direction: column;
  }

  .settings-left,
  .settings-right {
    flex: none;
    width: 100%;
  }
}

@media (max-width: 640px) {
  .settings-header {
    padding: 20px 20px 0;
  }

  .settings-body {
    padding: 20px 20px 32px;
  }

  .section-actions {
    flex-direction: column;
  }
}
</style>
