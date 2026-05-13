const BASE = '/api'

async function requestJSON(path, options = {}) {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(data.error || data.detail || `Request failed (${res.status})`)
  }
  return data
}

function requestForm(path, body) {
  return fetch(`${BASE}${path}`, { method: 'POST', body }).then(async (res) => {
    const data = await res.json().catch(() => ({}))
    if (!res.ok) throw new Error(data.error || data.detail || `Request failed (${res.status})`)
    return data
  })
}

export const api = {
  getConfig: () => requestJSON('/config'),

  updateConfig: (cfg) => requestJSON('/config', {
    method: 'POST',
    body: JSON.stringify(cfg),
  }),

  verify: () => requestJSON('/verify', { method: 'POST' }),

  preview: (markdown) => requestJSON('/preview', {
    method: 'POST',
    body: JSON.stringify({ markdown }),
  }),

  publish: ({ markdown, cover, author, publish_immediately }) => {
    const form = new FormData()
    form.append('markdown', markdown)
    form.append('cover', cover)
    if (author) form.append('author', author)
    form.append('publish_immediately', String(publish_immediately || false))
    return requestForm('/publish', form)
  },

  uploadImage: (file) => {
    const form = new FormData()
    form.append('file', file)
    return requestForm('/upload-image', form)
  },
}
