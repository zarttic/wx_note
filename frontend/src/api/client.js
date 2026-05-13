const BASE_URL = '/api';

async function request(url, options = {}) {
  const token = localStorage.getItem('token');

  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const config = {
    ...options,
    headers,
  };

  const response = await fetch(`${BASE_URL}${url}`, config);

  if (response.status === 401) {
    localStorage.removeItem('token');
    window.location.href = '/login';
    throw new Error('401 Unauthorized');
  }

  let data = null
  try {
    data = await response.json()
  } catch (e) {
    // Response body is not valid JSON
  }

  if (!response.ok) {
    throw new Error(data?.error || `Request failed (${response.status})`);
  }

  return data;
}

export const authApi = {
  login(username, password) {
    return request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
  },

  register(username, password, nickname) {
    return request('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, password, nickname }),
    });
  },

  logout() {
    return request('/auth/logout', {
      method: 'POST',
    });
  },

  getProfile() {
    return request('/user/profile');
  },

  updateProfile(data) {
    return request('/user/profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  getConfig() {
    return request('/user/config');
  },

  updateConfig(data) {
    return request('/user/config', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },
};

export const articleApi = {
  list({ page, pageSize, status, search } = {}) {
    const params = new URLSearchParams();
    if (page) params.set('page', page);
    if (pageSize) params.set('page_size', pageSize);
    if (status) params.set('status', status);
    if (search) params.set('search', search);
    const query = params.toString();
    return request(`/articles${query ? `?${query}` : ''}`);
  },

  get(id) {
    return request(`/articles/${id}`);
  },

  create(data) {
    return request('/articles', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  update(id, data) {
    return request(`/articles/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  delete(id) {
    return request(`/articles/${id}`, {
      method: 'DELETE',
    });
  },
};

export const editorApi = {
  preview(markdown) {
    return request('/editor/preview', {
      method: 'POST',
      body: JSON.stringify({ markdown }),
    });
  },

  uploadImage(file) {
    const formData = new FormData();
    formData.append('file', file);
    return request('/editor/upload-image', {
      method: 'POST',
      body: formData,
      headers: {},
    });
  },

  verify() {
    return request('/editor/verify', {
      method: 'POST',
    });
  },

  publish({ markdown, cover, author }) {
    const formData = new FormData();
    formData.append('markdown', markdown);
    if (cover) formData.append('cover', cover);
    if (author) formData.append('author', author);
    return request('/editor/publish', {
      method: 'POST',
      body: formData,
      headers: {},
    });
  },
};

export const templateApi = {
  list(category) {
    const params = new URLSearchParams();
    if (category) params.set('category', category);
    const query = params.toString();
    return request(`/templates${query ? `?${query}` : ''}`);
  },

  get(id) {
    return request(`/templates/${id}`);
  },

  create(data) {
    return request('/templates', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  update(id, data) {
    return request(`/templates/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  delete(id) {
    return request(`/templates/${id}`, {
      method: 'DELETE',
    });
  },

  categories() {
    return request('/templates/categories/all');
  },
};
