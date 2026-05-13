import { ref, computed } from 'vue';
import { defineStore } from 'pinia';
import { authApi } from '../api/client.js';

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null);
  const token = ref(localStorage.getItem('token') || null);

  const isLoggedIn = computed(() => !!token.value);

  async function login(username, password) {
    const data = await authApi.login(username, password);
    token.value = data.token;
    localStorage.setItem('token', data.token);
    user.value = data.user;
  }

  async function register(username, password, nickname) {
    const data = await authApi.register(username, password, nickname);
    token.value = data.token;
    localStorage.setItem('token', data.token);
    user.value = data.user;
  }

  function logout() {
    // Fire-and-forget server logout
    authApi.logout().catch(() => {})
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }

  async function fetchProfile() {
    try {
      const data = await authApi.getProfile();
      user.value = data;
    } catch {
      await logout();
      throw new Error('获取用户信息失败');
    }
  }

  async function fetchConfig() {
    return await authApi.getConfig();
  }

  return {
    user,
    token,
    isLoggedIn,
    login,
    register,
    logout,
    fetchProfile,
    fetchConfig,
  };
});
