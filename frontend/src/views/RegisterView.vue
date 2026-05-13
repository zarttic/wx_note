<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { UserPlus, Eye, EyeOff, Loader2, FileText, Sparkles, Zap, Shield } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const nickname = ref('')
const showPassword = ref(false)
const isLoading = ref(false)
const errorMessage = ref('')
const focusedField = ref('')

const canSubmit = computed(() => username.value.trim() && password.value.trim() && !isLoading.value)

const features = [
  { icon: Sparkles, text: 'Markdown 实时预览' },
  { icon: Zap, text: '一键发布草稿箱' },
  { icon: Shield, text: '多账号安全隔离' },
]

async function handleRegister() {
  if (!canSubmit.value) return
  errorMessage.value = ''
  isLoading.value = true
  try {
    await authStore.register(username.value.trim(), password.value, nickname.value.trim())
    router.push('/editor')
  } catch (err) {
    errorMessage.value = err.message || '注册失败，请稍后重试'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <!-- ─── Left: Branding Panel ─────────────────────────────────── -->
    <div class="auth-branding">
      <div class="branding-bg">
        <div class="bg-gradient-orb orb-1"></div>
        <div class="bg-gradient-orb orb-2"></div>
        <div class="bg-grid"></div>
      </div>

      <div class="branding-content">
        <div class="branding-logo-wrap">
          <div class="branding-logo">
            <FileText :size="24" color="#07c160" :stroke-width="1.8" />
          </div>
          <span class="branding-version">v0.1</span>
        </div>

        <h1 class="branding-title">wx_note</h1>
        <p class="branding-subtitle">微信公众号 Markdown 编辑器</p>

        <div class="branding-features">
          <div v-for="feat in features" :key="feat.text" class="feature-item">
            <div class="feature-icon">
              <component :is="feat.icon" :size="14" :stroke-width="2" />
            </div>
            <span class="feature-text">{{ feat.text }}</span>
          </div>
        </div>

        <div class="branding-footer">
          <span class="footer-line"></span>
          <span class="footer-text">为创作者而生</span>
          <span class="footer-line"></span>
        </div>
      </div>
    </div>

    <!-- ─── Right: Register Form ────────────────────────────────── -->
    <div class="auth-form-side">
      <div class="auth-form-container">
        <div class="auth-form-header">
          <h2 class="auth-form-title">创建账号</h2>
          <p class="auth-form-subtitle">开始使用 wx_note 高效写作</p>
        </div>

        <Transition name="slide-fade">
          <div v-if="errorMessage" class="auth-error">
            <span class="error-icon">!</span>
            {{ errorMessage }}
          </div>
        </Transition>

        <form class="auth-form" @submit.prevent="handleRegister">
          <div class="field-group" :class="{ 'field-focused': focusedField === 'username' }">
            <label class="form-label" for="reg-username">用户名 <span class="required">*</span></label>
            <input
              id="reg-username"
              v-model="username"
              type="text"
              class="form-input"
              placeholder="请输入用户名"
              autocomplete="username"
              :disabled="isLoading"
              @focus="focusedField = 'username'"
              @blur="focusedField = ''"
            />
          </div>

          <div class="field-group" :class="{ 'field-focused': focusedField === 'nickname' }">
            <label class="form-label" for="reg-nickname">昵称 <span class="optional">（可选）</span></label>
            <input
              id="reg-nickname"
              v-model="nickname"
              type="text"
              class="form-input"
              placeholder="请输入昵称"
              autocomplete="nickname"
              :disabled="isLoading"
              @focus="focusedField = 'nickname'"
              @blur="focusedField = ''"
            />
          </div>

          <div class="field-group" :class="{ 'field-focused': focusedField === 'password' }">
            <label class="form-label" for="reg-password">密码 <span class="required">*</span></label>
            <div class="password-input-wrapper">
              <input
                id="reg-password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                class="form-input password-input"
                placeholder="请输入密码（至少6位）"
                autocomplete="new-password"
                :disabled="isLoading"
                @focus="focusedField = 'password'"
                @blur="focusedField = ''"
              />
              <button type="button" class="password-toggle" tabindex="-1" @click="showPassword = !showPassword">
                <Eye v-if="!showPassword" :size="16" :stroke-width="1.8" />
                <EyeOff v-else :size="16" :stroke-width="1.8" />
              </button>
            </div>
          </div>

          <button type="submit" class="btn btn-primary btn-full" :disabled="!canSubmit">
            <Loader2 v-if="isLoading" :size="15" class="animate-spin" />
            <UserPlus v-else :size="15" :stroke-width="2" />
            {{ isLoading ? '注册中...' : '注册' }}
          </button>
        </form>

        <div class="auth-alt-link">
          已有账号？<router-link to="/login" class="auth-link">立即登录</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  min-height: 100dvh;
  width: 100%;
}

.auth-branding {
  flex: 1.4;
  background: #0a0a0f;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px;
  position: relative;
  overflow: hidden;
}

.branding-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.bg-gradient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.35;
}

.orb-1 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, #07c160 0%, transparent 70%);
  top: -10%;
  right: -5%;
  animation: float-orb 8s ease-in-out infinite;
}

.orb-2 {
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, #1a6b4a 0%, transparent 70%);
  bottom: -8%;
  left: -3%;
  animation: float-orb 10s ease-in-out infinite reverse;
}

.bg-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255,255,255,0.02) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255,255,255,0.02) 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: radial-gradient(ellipse at center, black 30%, transparent 80%);
  -webkit-mask-image: radial-gradient(ellipse at center, black 30%, transparent 80%);
}

@keyframes float-orb {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(20px, -20px); }
}

.branding-content {
  max-width: 420px;
  position: relative;
  z-index: 1;
}

.branding-logo-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 32px;
}

.branding-logo {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  background: rgba(7, 193, 96, 0.08);
  border: 1px solid rgba(7, 193, 96, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 40px rgba(7, 193, 96, 0.1);
}

.branding-version {
  font-family: var(--font-mono);
  font-size: 11px;
  color: rgba(255,255,255,0.25);
  background: rgba(255,255,255,0.05);
  padding: 3px 8px;
  border-radius: 6px;
  border: 1px solid rgba(255,255,255,0.06);
}

.branding-title {
  font-family: var(--font-mono);
  font-size: 36px;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 12px;
  letter-spacing: -0.03em;
  line-height: 1;
}

.branding-subtitle {
  font-size: 15px;
  color: rgba(255, 255, 255, 0.45);
  margin: 0 0 40px;
  font-weight: 400;
  line-height: 1.6;
}

.branding-features {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 48px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.05);
  border-radius: 10px;
  transition: all 0.3s ease;
}

.feature-item:hover {
  background: rgba(255,255,255,0.05);
  border-color: rgba(255,255,255,0.08);
  transform: translateX(4px);
}

.feature-icon {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: rgba(7, 193, 96, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #07c160;
  flex-shrink: 0;
}

.feature-text {
  font-size: 13px;
  color: rgba(255,255,255,0.6);
  font-weight: 500;
}

.branding-footer {
  display: flex;
  align-items: center;
  gap: 16px;
}

.footer-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.08), transparent);
}

.footer-text {
  font-size: 11px;
  color: rgba(255,255,255,0.2);
  white-space: nowrap;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.auth-form-side {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 40px;
  background: var(--color-surface);
}

.auth-form-container {
  width: 100%;
  max-width: 360px;
}

.auth-form-header {
  margin-bottom: 36px;
}

.auth-form-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 6px;
  letter-spacing: -0.02em;
}

.auth-form-subtitle {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin: 0;
}

.auth-error {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  background: rgba(239, 68, 68, 0.06);
  border: 1px solid rgba(239, 68, 68, 0.15);
  border-radius: 8px;
  font-size: 13px;
  color: var(--color-status-error);
  line-height: 1.5;
  margin-bottom: 20px;
}

.error-icon {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: rgba(239, 68, 68, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}

.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.25s ease;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 4px 0 20px;
  transition: all 0.2s ease;
}

.field-group:last-of-type {
  padding-bottom: 24px;
}

.field-focused .form-input {
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px var(--color-accent-subtle);
}

.required {
  color: var(--color-status-error);
}

.optional {
  color: var(--color-text-tertiary);
  font-weight: 400;
}

.password-input-wrapper {
  position: relative;
}

.password-input {
  padding-right: 44px;
}

.password-toggle {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--color-text-tertiary);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s;
  border-radius: 4px;
}

.password-toggle:hover {
  color: var(--color-text-primary);
}

.btn-full {
  width: 100%;
  padding: 11px 16px;
  font-size: 14px;
  font-weight: 500;
  border-radius: 10px;
  position: relative;
  overflow: hidden;
}

.btn-full::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(255,255,255,0.1) 0%, transparent 50%);
  pointer-events: none;
}

.auth-alt-link {
  margin-top: 28px;
  text-align: center;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.auth-link {
  color: #07c160;
  font-weight: 500;
  text-decoration: none;
  transition: opacity 0.15s;
}

.auth-link:hover {
  opacity: 0.8;
}

@media (max-width: 900px) {
  .auth-branding {
    display: none;
  }
}

@media (max-width: 768px) {
  .auth-branding {
    display: none;
  }

  .auth-form-side {
    padding: 32px 24px 48px;
  }

  .auth-form-header {
    margin-bottom: 28px;
  }

  .auth-form-title {
    font-size: 22px;
  }
}
</style>
