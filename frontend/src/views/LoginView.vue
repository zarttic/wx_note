<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Eye, EyeOff, ArrowRight, Loader2 } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const showPassword = ref(false)
const isLoading = ref(false)
const errorMessage = ref('')
const focusedField = ref('')
const mounted = ref(false)

const canSubmit = computed(() => username.value.trim() && password.value.trim() && !isLoading.value)

async function handleLogin() {
  if (!canSubmit.value) return
  errorMessage.value = ''
  isLoading.value = true
  try {
    await authStore.login(username.value.trim(), password.value)
    router.push('/editor')
  } catch (err) {
    errorMessage.value = err.message || '登录失败，请检查用户名和密码'
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  requestAnimationFrame(() => {
    mounted.value = true
  })
})
</script>

<template>
  <div class="login-page">
    <!-- Noise texture overlay -->
    <div class="noise-overlay"></div>

    <!-- Main content -->
    <div class="login-inner">
      <!-- Left: brand -->
      <div class="login-brand" :class="{ 'brand-visible': mounted }">
        <div class="brand-mark">
          <span class="brand-cursor">_</span>
        </div>
        <h1 class="brand-name">wx_note</h1>
        <p class="brand-tagline">微信公众号 Markdown 编辑器</p>
        <div class="brand-meta">
          <span class="meta-dot"></span>
          <span class="meta-text">v0.1.0</span>
        </div>
      </div>

      <!-- Divider -->
      <div class="login-divider" :class="{ 'divider-visible': mounted }"></div>

      <!-- Right: form -->
      <div class="login-form-wrap" :class="{ 'form-visible': mounted }">
        <form class="login-form" @submit.prevent="handleLogin" novalidate>
          <div class="form-heading">
            <h2>登录</h2>
            <p>回到你的写作工作台</p>
          </div>

          <Transition name="error-slide">
            <div v-if="errorMessage" class="form-error">
              {{ errorMessage }}
            </div>
          </Transition>

          <div class="form-fields">
            <div class="field" :class="{ 'field-active': focusedField === 'username' }">
              <label for="login-username">用户名</label>
              <input
                id="login-username"
                v-model="username"
                type="text"
                placeholder="输入用户名"
                autocomplete="username"
                :disabled="isLoading"
                @focus="focusedField = 'username'"
                @blur="focusedField = ''"
              />
            </div>

            <div class="field" :class="{ 'field-active': focusedField === 'password' }">
              <label for="login-password">密码</label>
              <div class="password-wrap">
                <input
                  id="login-password"
                  v-model="password"
                  :type="showPassword ? 'text' : 'password'"
                  placeholder="输入密码"
                  autocomplete="current-password"
                  :disabled="isLoading"
                  @focus="focusedField = 'password'"
                  @blur="focusedField = ''"
                  @keyup.enter="handleLogin"
                />
                <button
                  type="button"
                  class="pw-toggle"
                  tabindex="-1"
                  @click="showPassword = !showPassword"
                >
                  <Eye v-if="!showPassword" :size="15" :stroke-width="1.5" />
                  <EyeOff v-else :size="15" :stroke-width="1.5" />
                </button>
              </div>
            </div>
          </div>

          <button type="submit" class="submit-btn" :disabled="!canSubmit">
            <Loader2 v-if="isLoading" :size="14" class="spin" />
            <template v-else>
              <span>进入工作台</span>
              <ArrowRight :size="14" :stroke-width="2" />
            </template>
          </button>

          <div class="form-footer">
            <span>没有账号？</span>
            <router-link to="/register" class="register-link">创建账号 →</router-link>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── Layout ─────────────────────────────────────────────────── */

.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100dvh;
  background: #09090b;
  position: relative;
  overflow: hidden;
}

.noise-overlay {
  position: fixed;
  inset: 0;
  z-index: 0;
  opacity: 0.03;
  pointer-events: none;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
  background-repeat: repeat;
  background-size: 256px 256px;
}

.login-inner {
  display: flex;
  align-items: stretch;
  width: 100%;
  max-width: 880px;
  min-height: 500px;
  position: relative;
  z-index: 1;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 24px 80px -24px rgba(0, 0, 0, 0.6);
}

/* ── Brand panel ────────────────────────────────────────────── */

.login-brand {
  flex: 1.2;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 48px 56px;
  background: #0c0c0c;
  opacity: 0;
  transform: translateY(12px);
  transition: opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1),
              transform 0.7s cubic-bezier(0.16, 1, 0.3, 1);
}

.login-brand.brand-visible {
  opacity: 1;
  transform: translateY(0);
}

.brand-mark {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--color-accent);
  margin-bottom: 28px;
  letter-spacing: 0.12em;
  display: flex;
  align-items: center;
  gap: 2px;
}

.brand-cursor {
  animation: blink 1s step-end infinite;
  font-weight: 400;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.brand-name {
  font-family: var(--font-mono);
  font-size: 32px;
  font-weight: 600;
  color: #f5f5f5;
  margin: 0 0 10px;
  letter-spacing: -0.04em;
  line-height: 1;
}

.brand-tagline {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.35);
  margin: 0 0 36px;
  line-height: 1.6;
  max-width: 240px;
}

.brand-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.meta-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--color-accent);
  flex-shrink: 0;
  opacity: 0.7;
}

.meta-text {
  font-family: var(--font-mono);
  font-size: 10px;
  color: rgba(255, 255, 255, 0.2);
  letter-spacing: 0.06em;
}

/* ── Divider ─────────────────────────────────────────────────── */

.login-divider {
  width: 1px;
  background: rgba(0, 0, 0, 0.06);
  align-self: stretch;
  margin: 0;
  opacity: 0;
  transition: opacity 0.5s ease 0.4s;
}

.login-divider.divider-visible {
  opacity: 1;
}

/* ── Form panel ──────────────────────────────────────────────── */

.login-form-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 48px 56px;
  background: linear-gradient(145deg, #fafafa 0%, #f2f2f5 100%);
  opacity: 0;
  transform: translateY(12px);
  transition: opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1) 0.15s,
              transform 0.7s cubic-bezier(0.16, 1, 0.3, 1) 0.15s;
}

.login-form-wrap.form-visible {
  opacity: 1;
  transform: translateY(0);
}

.login-form {
  width: 100%;
}

.form-heading {
  margin-bottom: 36px;
}

.form-heading h2 {
  font-family: var(--font-mono);
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 4px;
  letter-spacing: -0.02em;
}

.form-heading p {
  font-size: 12px;
  color: #9ca3af;
  margin: 0;
}

/* ── Error ───────────────────────────────────────────────────── */

.form-error {
  padding: 10px 14px;
  background: rgba(239, 68, 68, 0.04);
  border: 1px solid rgba(239, 68, 68, 0.1);
  border-radius: 6px;
  font-size: 12px;
  color: #dc2626;
  margin-bottom: 20px;
  line-height: 1.5;
}

.error-slide-enter-active,
.error-slide-leave-active {
  transition: all 0.2s ease;
}

.error-slide-enter-from,
.error-slide-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* ── Fields ──────────────────────────────────────────────────── */

.form-fields {
  display: flex;
  flex-direction: column;
  gap: 0;
  margin-bottom: 28px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-bottom: 20px;
}

.field:last-of-type {
  padding-bottom: 0;
}

.field label {
  font-size: 11px;
  font-weight: 500;
  color: #a1a1aa;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  transition: color 0.2s;
}

.field-active label {
  color: #52525b;
}

.field input {
  width: 100%;
  padding: 10px 0;
  border: none;
  border-bottom: 1.5px solid #e4e4e7;
  border-radius: 0;
  font-size: 14px;
  font-family: var(--font-mono);
  color: #18181b;
  background: transparent;
  outline: none;
  transition: border-color 0.2s;
  -webkit-appearance: none;
}

.field input::placeholder {
  color: #d4d4d8;
  font-family: var(--font-sans);
  font-size: 13px;
}

.field input:focus {
  border-color: var(--color-accent);
}

.field input:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.password-wrap {
  position: relative;
}

.password-wrap input {
  padding-right: 32px;
}

.pw-toggle {
  position: absolute;
  right: 0;
  bottom: 10px;
  background: none;
  border: none;
  color: #d4d4d8;
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s;
}

.pw-toggle:hover {
  color: #71717a;
}

/* ── Submit ──────────────────────────────────────────────────── */

.submit-btn {
  width: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
  background: #18181b;
  color: #fafafa;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  font-family: var(--font-mono);
  cursor: pointer;
  transition: background 0.15s, opacity 0.15s, transform 0.1s;
  letter-spacing: 0.02em;
}

.submit-btn:hover:not(:disabled) {
  background: #27272a;
}

.submit-btn:active:not(:disabled) {
  transform: scale(0.98);
}

.submit-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
  transform: none;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ── Footer ──────────────────────────────────────────────────── */

.form-footer {
  margin-top: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
  color: #a1a1aa;
}

.register-link {
  color: #18181b;
  text-decoration: none;
  font-weight: 600;
  transition: opacity 0.15s;
}

.register-link:hover {
  opacity: 0.6;
}

/* ── Responsive ──────────────────────────────────────────────── */

@media (max-width: 768px) {
  .login-inner {
    flex-direction: column;
    max-width: 400px;
    min-height: auto;
    padding: 40px 0;
  }

  .login-brand {
    padding: 0 28px 32px;
    align-items: center;
    text-align: center;
  }

  .brand-tagline {
    max-width: none;
  }

  .login-divider {
    width: auto;
    height: 1px;
    margin: 0;
    background: rgba(0, 0, 0, 0.06);
  }

  .login-form-wrap {
    padding: 32px 28px 0;
  }

  .form-heading {
    margin-bottom: 28px;
  }
}
</style>
