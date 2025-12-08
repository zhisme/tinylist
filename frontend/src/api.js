// API base URL - defaults to relative path for K8s ingress routing
// Set VITE_API_URL at build time for different configurations
const API_BASE = import.meta.env.VITE_API_URL || '/api/private';

async function request(path, options = {}) {
  const url = `${API_BASE}${path}`;
  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
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
