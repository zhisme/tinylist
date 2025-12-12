// API base URL - defaults to /tinylist/api/private for K8s deployment
// Frontend nginx proxies /tinylist/api/* to backend at /api/*
// Set VITE_API_URL at build time for different configurations
const API_BASE = import.meta.env.VITE_API_URL || '/tinylist/api/private';

// Auth credentials storage
let authCredentials = null;

export function setAuthCredentials(username, password) {
  if (username && password) {
    authCredentials = btoa(`${username}:${password}`);
    sessionStorage.setItem('tinylist_auth', authCredentials);
  } else {
    authCredentials = null;
    sessionStorage.removeItem('tinylist_auth');
  }
}

export function getStoredAuth() {
  if (!authCredentials) {
    authCredentials = sessionStorage.getItem('tinylist_auth');
  }
  return authCredentials;
}

export function clearAuth() {
  authCredentials = null;
  sessionStorage.removeItem('tinylist_auth');
}

async function request(path, options = {}) {
  const url = `${API_BASE}${path}`;
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  // Add auth header if credentials exist
  const auth = getStoredAuth();
  if (auth) {
    headers['Authorization'] = `Basic ${auth}`;
  }

  const response = await fetch(url, {
    headers,
    ...options,
  });

  if (response.status === 401) {
    clearAuth();
    const error = new Error('Unauthorized');
    error.status = 401;
    throw error;
  }

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

// Validate credentials
export async function validateCredentials(username, password) {
  const auth = btoa(`${username}:${password}`);
  const response = await fetch(`${API_BASE}/stats`, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Basic ${auth}`,
    },
  });
  return response.ok;
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

// Stats API (to be implemented on backend)
export const stats = {
  get: () => request('/stats'),
};

// Settings API
export const settings = {
  getSMTP: () => request('/settings/smtp'),
  updateSMTP: (data) => request('/settings/smtp', { method: 'PUT', body: JSON.stringify(data) }),
  testSMTP: (email) => request('/settings/smtp/test', { method: 'POST', body: JSON.stringify({ email }) }),
};
