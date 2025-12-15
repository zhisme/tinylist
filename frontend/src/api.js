// API base URL - uses the same base path as the frontend
// Set VITE_API_URL to override completely, or VITE_BASE_PATH to change the prefix
const API_BASE = import.meta.env.VITE_API_URL || `${__BASE_PATH__}/api/private`;

async function request(path, options = {}) {
  const url = `${API_BASE}${path}`;

  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include', // Send credentials for cross-origin (CORS) requests
    ...options,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

// Subscribers API
export const subscribers = {
  list: (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return request(`/subscribers${query ? `?${query}` : ''}`);
  },
  get: (id) => request(`/subscribers/${id}`),
  create: (data) => request('/subscribers', { method: 'POST', body: JSON.stringify(data) }),
  delete: (id) => request(`/subscribers/${id}`, { method: 'DELETE' }),
  sendVerification: (id) => request(`/subscribers/${id}/send-verification`, { method: 'POST' }),
};

// Campaigns API
export const campaigns = {
  list: () => request('/campaigns'),
  get: (id) => request(`/campaigns/${id}`),
  create: (data) => request('/campaigns', { method: 'POST', body: JSON.stringify(data) }),
  update: (id, data) => request(`/campaigns/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id) => request(`/campaigns/${id}`, { method: 'DELETE' }),
  send: (id) => request(`/campaigns/${id}/send`, { method: 'POST' }),
  cancel: (id) => request(`/campaigns/${id}/cancel`, { method: 'POST' }),
  journal: (id) => request(`/campaigns/${id}/journal`),
};

// Stats API
export const stats = {
  get: () => request('/stats'),
};

// Settings API
export const settings = {
  getSMTP: () => request('/settings/smtp'),
  updateSMTP: (data) => request('/settings/smtp', { method: 'PUT', body: JSON.stringify(data) }),
  testSMTP: (email) => request('/settings/smtp/test', { method: 'POST', body: JSON.stringify({ email }) }),
};
