import { useState, useEffect } from 'preact/hooks';
import { subscribers } from '../api';

export function Subscribers() {
  const [data, setData] = useState({ data: [], total: 0, page: 1, per_page: 20 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [statusFilter, setStatusFilter] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);

  useEffect(() => {
    loadSubscribers();
  }, [data.page, statusFilter]);

  async function loadSubscribers() {
    try {
      setLoading(true);
      const params = { page: data.page, per_page: data.per_page };
      if (statusFilter) params.status = statusFilter;
      const result = await subscribers.list(params);
      setData(result);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(id) {
    if (!confirm('Are you sure you want to delete this subscriber?')) return;
    try {
      await subscribers.delete(id);
      loadSubscribers();
    } catch (err) {
      alert('Failed to delete: ' + err.message);
    }
  }

  async function handleAdd(email, name) {
    try {
      await subscribers.create({ email, name });
      setShowAddModal(false);
      loadSubscribers();
    } catch (err) {
      alert('Failed to add: ' + err.message);
    }
  }

  return (
    <div>
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold">Subscribers</h1>
        <button
          onClick={() => setShowAddModal(true)}
          class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Add Subscriber
        </button>
      </div>

      {/* Filters */}
      <div class="mb-4">
        <select
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value);
            setData(d => ({ ...d, page: 1 }));
          }}
          class="border rounded px-3 py-2"
        >
          <option value="">All statuses</option>
          <option value="pending">Pending</option>
          <option value="verified">Verified</option>
          <option value="unsubscribed">Unsubscribed</option>
        </select>
      </div>

      {error && <div class="text-red-500 mb-4">Error: {error}</div>}

      {/* Table */}
      <div class="bg-white rounded-lg shadow overflow-hidden">
        <table class="w-full">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-3 text-left text-sm font-medium text-gray-500">Email</th>
              <th class="px-4 py-3 text-left text-sm font-medium text-gray-500">Name</th>
              <th class="px-4 py-3 text-left text-sm font-medium text-gray-500">Status</th>
              <th class="px-4 py-3 text-left text-sm font-medium text-gray-500">Created</th>
              <th class="px-4 py-3 text-left text-sm font-medium text-gray-500">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y">
            {loading ? (
              <tr><td colspan="5" class="px-4 py-8 text-center text-gray-500">Loading...</td></tr>
            ) : data.data.length === 0 ? (
              <tr><td colspan="5" class="px-4 py-8 text-center text-gray-500">No subscribers found</td></tr>
            ) : (
              data.data.map(sub => (
                <tr key={sub.id}>
                  <td class="px-4 py-3">{sub.email}</td>
                  <td class="px-4 py-3">{sub.name || '-'}</td>
                  <td class="px-4 py-3">
                    <StatusBadge status={sub.status} />
                  </td>
                  <td class="px-4 py-3 text-sm text-gray-500">
                    {new Date(sub.created_at).toLocaleDateString()}
                  </td>
                  <td class="px-4 py-3">
                    <button
                      onClick={() => handleDelete(sub.id)}
                      class="text-red-500 hover:text-red-700"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {data.total_pages > 1 && (
        <div class="mt-4 flex justify-center gap-2">
          <button
            onClick={() => setData(d => ({ ...d, page: d.page - 1 }))}
            disabled={data.page <= 1}
            class="px-3 py-1 border rounded disabled:opacity-50"
          >
            Previous
          </button>
          <span class="px-3 py-1">
            Page {data.page} of {data.total_pages}
          </span>
          <button
            onClick={() => setData(d => ({ ...d, page: d.page + 1 }))}
            disabled={data.page >= data.total_pages}
            class="px-3 py-1 border rounded disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}

      {/* Add Modal */}
      {showAddModal && (
        <AddSubscriberModal
          onClose={() => setShowAddModal(false)}
          onAdd={handleAdd}
        />
      )}
    </div>
  );
}

function StatusBadge({ status }) {
  const colors = {
    pending: 'bg-yellow-100 text-yellow-800',
    verified: 'bg-green-100 text-green-800',
    unsubscribed: 'bg-gray-100 text-gray-800',
  };

  return (
    <span class={`px-2 py-1 rounded-full text-xs ${colors[status] || 'bg-gray-100'}`}>
      {status}
    </span>
  );
}

function AddSubscriberModal({ onClose, onAdd }) {
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');

  function handleSubmit(e) {
    e.preventDefault();
    onAdd(email, name);
  }

  return (
    <div class="fixed inset-0 bg-black/50 flex items-center justify-center">
      <div class="bg-white rounded-lg p-6 w-full max-w-md">
        <h2 class="text-xl font-bold mb-4">Add Subscriber</h2>
        <form onSubmit={handleSubmit}>
          <div class="mb-4">
            <label class="block text-sm font-medium mb-1">Email</label>
            <input
              type="email"
              value={email}
              onInput={(e) => setEmail(e.target.value)}
              required
              class="w-full border rounded px-3 py-2"
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium mb-1">Name (optional)</label>
            <input
              type="text"
              value={name}
              onInput={(e) => setName(e.target.value)}
              class="w-full border rounded px-3 py-2"
            />
          </div>
          <div class="flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              class="px-4 py-2 border rounded"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
            >
              Add
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
