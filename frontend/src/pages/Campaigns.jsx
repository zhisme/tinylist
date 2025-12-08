import { useState, useEffect } from 'preact/hooks';
import { campaigns } from '../api';

export function Campaigns() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingCampaign, setEditingCampaign] = useState(null);
  const [journalCampaign, setJournalCampaign] = useState(null);

  useEffect(() => {
    loadCampaigns();
  }, []);

  async function loadCampaigns() {
    try {
      setLoading(true);
      const result = await campaigns.list();
      setData(result);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate(campaignData) {
    try {
      await campaigns.create(campaignData);
      setShowCreateModal(false);
      loadCampaigns();
    } catch (err) {
      alert('Failed to create: ' + err.message);
    }
  }

  async function handleUpdate(id, campaignData) {
    try {
      await campaigns.update(id, campaignData);
      setEditingCampaign(null);
      loadCampaigns();
    } catch (err) {
      alert('Failed to update: ' + err.message);
    }
  }

  async function handleDelete(id) {
    if (!confirm('Are you sure you want to delete this campaign?')) return;
    try {
      await campaigns.delete(id);
      loadCampaigns();
    } catch (err) {
      alert('Failed to delete: ' + err.message);
    }
  }

  async function handleSend(id) {
    if (!confirm('Are you sure you want to send this campaign to all subscribers?')) return;
    try {
      await campaigns.send(id);
      alert('Campaign sending started!');
      loadCampaigns();
    } catch (err) {
      alert('Failed to send: ' + err.message);
    }
  }

  async function handleCancel(id) {
    if (!confirm('Are you sure you want to cancel this campaign?')) return;
    try {
      await campaigns.cancel(id);
      alert('Campaign cancellation requested');
      loadCampaigns();
    } catch (err) {
      alert('Failed to cancel: ' + err.message);
    }
  }

  return (
    <div>
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold">Campaigns</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Create Campaign
        </button>
      </div>

      {error && <div class="text-red-500 mb-4">Error: {error}</div>}

      {loading ? (
        <div class="text-gray-500">Loading...</div>
      ) : data.length === 0 ? (
        <div class="bg-white rounded-lg shadow p-8 text-center text-gray-500">
          No campaigns yet. Create your first one!
        </div>
      ) : (
        <div class="space-y-4">
          {data.map(campaign => (
            <CampaignCard
              key={campaign.id}
              campaign={campaign}
              onEdit={() => setEditingCampaign(campaign)}
              onDelete={() => handleDelete(campaign.id)}
              onSend={() => handleSend(campaign.id)}
              onCancel={() => handleCancel(campaign.id)}
              onJournal={() => setJournalCampaign(campaign)}
            />
          ))}
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <CampaignModal
          onClose={() => setShowCreateModal(false)}
          onSave={handleCreate}
        />
      )}

      {/* Edit Modal */}
      {editingCampaign && (
        <CampaignModal
          campaign={editingCampaign}
          onClose={() => setEditingCampaign(null)}
          onSave={(data) => handleUpdate(editingCampaign.id, data)}
        />
      )}

      {/* Journal Modal */}
      {journalCampaign && (
        <JournalModal
          campaign={journalCampaign}
          onClose={() => setJournalCampaign(null)}
        />
      )}
    </div>
  );
}

function CampaignCard({ campaign, onEdit, onDelete, onSend, onCancel, onJournal }) {
  const statusColors = {
    draft: 'bg-gray-100 text-gray-800',
    sending: 'bg-blue-100 text-blue-800',
    sent: 'bg-green-100 text-green-800',
    failed: 'bg-red-100 text-red-800',
    cancelled: 'bg-orange-100 text-orange-800',
  };

  return (
    <div class="bg-white rounded-lg shadow p-4">
      <div class="flex justify-between items-start">
        <div class="flex-1">
          <div class="flex items-center gap-2 mb-2">
            <h3 class="font-semibold">{campaign.subject}</h3>
            <span class={`px-2 py-0.5 rounded-full text-xs ${statusColors[campaign.status]}`}>
              {campaign.status}
            </span>
          </div>
          <p class="text-gray-500 text-sm line-clamp-2">{campaign.body_text}</p>
          {campaign.status !== 'draft' && (
            <div class="mt-2 text-sm text-gray-500">
              Sent: {campaign.sent_count} / Failed: {campaign.failed_count} / Total: {campaign.total_count}
            </div>
          )}
        </div>
        <div class="flex gap-2 ml-4">
          {campaign.status === 'draft' && (
            <>
              <button
                onClick={onEdit}
                class="text-blue-500 hover:text-blue-700 text-sm"
              >
                Edit
              </button>
              <button
                onClick={onSend}
                class="bg-green-500 text-white px-3 py-1 rounded text-sm hover:bg-green-600"
              >
                Send
              </button>
              <button
                onClick={onDelete}
                class="text-red-500 hover:text-red-700 text-sm"
              >
                Delete
              </button>
            </>
          )}
          {campaign.status === 'sending' && (
            <button
              onClick={onCancel}
              class="bg-red-500 text-white px-3 py-1 rounded text-sm hover:bg-red-600"
            >
              Cancel
            </button>
          )}
          <button
            onClick={onJournal}
            class="text-gray-500 hover:text-gray-700 text-sm"
          >
            Journal
          </button>
        </div>
      </div>
    </div>
  );
}

function CampaignModal({ campaign, onClose, onSave }) {
  const [subject, setSubject] = useState(campaign?.subject || '');
  const [bodyText, setBodyText] = useState(campaign?.body_text || '');
  const [bodyHtml, setBodyHtml] = useState(campaign?.body_html || '');

  function handleSubmit(e) {
    e.preventDefault();
    onSave({
      subject,
      body_text: bodyText,
      body_html: bodyHtml || null,
    });
  }

  return (
    <div class="fixed inset-0 bg-black/50 flex items-center justify-center overflow-auto py-8">
      <div class="bg-white rounded-lg p-6 w-full max-w-2xl mx-4">
        <h2 class="text-xl font-bold mb-4">
          {campaign ? 'Edit Campaign' : 'Create Campaign'}
        </h2>
        <form onSubmit={handleSubmit}>
          <div class="mb-4">
            <label class="block text-sm font-medium mb-1">Subject</label>
            <input
              type="text"
              value={subject}
              onInput={(e) => setSubject(e.target.value)}
              required
              class="w-full border rounded px-3 py-2"
              placeholder="Newsletter #1"
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium mb-1">
              Body (Plain Text) <span class="text-gray-400">- Use {'{{name}}'} and {'{{email}}'} for personalization</span>
            </label>
            <textarea
              value={bodyText}
              onInput={(e) => setBodyText(e.target.value)}
              required
              rows={6}
              class="w-full border rounded px-3 py-2 font-mono text-sm"
              placeholder="Hi {{name}},&#10;&#10;Welcome to our newsletter..."
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium mb-1">
              Body (HTML) <span class="text-gray-400">- Optional</span>
            </label>
            <textarea
              value={bodyHtml}
              onInput={(e) => setBodyHtml(e.target.value)}
              rows={6}
              class="w-full border rounded px-3 py-2 font-mono text-sm"
              placeholder="<p>Hi {{name}},</p>&#10;<p>Welcome to our newsletter...</p>"
            />
          </div>
          <div class="flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              class="px-4 py-2 border rounded hover:bg-gray-100"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
            >
              {campaign ? 'Save' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function JournalModal({ campaign, onClose }) {
  const [journal, setJournal] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadJournal();
  }, [campaign.id]);

  async function loadJournal() {
    try {
      setLoading(true);
      const result = await campaigns.journal(campaign.id);
      setJournal(result);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  const eventTypeStyles = {
    info: 'bg-blue-100 text-blue-800',
    warning: 'bg-yellow-100 text-yellow-800',
    error: 'bg-red-100 text-red-800',
    success: 'bg-green-100 text-green-800',
  };

  function formatTime(timestamp) {
    return new Date(timestamp).toLocaleString();
  }

  return (
    <div class="fixed inset-0 bg-black/50 flex items-center justify-center overflow-auto py-8">
      <div class="bg-white rounded-lg p-6 w-full max-w-xl mx-4">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-xl font-bold">Campaign Journal</h2>
          <button
            onClick={onClose}
            class="text-gray-500 hover:text-gray-700 text-xl"
          >
            &times;
          </button>
        </div>
        <p class="text-gray-500 text-sm mb-4">{campaign.subject}</p>

        {error && <div class="text-red-500 mb-4">Error: {error}</div>}

        {loading ? (
          <div class="text-gray-500 text-center py-8">Loading...</div>
        ) : journal.length === 0 ? (
          <div class="text-gray-500 text-center py-8">No journal entries yet.</div>
        ) : (
          <div class="space-y-2 max-h-96 overflow-y-auto">
            {journal.map(entry => (
              <div key={entry.id} class="border-l-4 border-gray-200 pl-3 py-2">
                <div class="flex items-center gap-2 mb-1">
                  <span class={`px-2 py-0.5 rounded text-xs ${eventTypeStyles[entry.event_type]}`}>
                    {entry.event_type}
                  </span>
                  <span class="text-xs text-gray-400">{formatTime(entry.created_at)}</span>
                </div>
                <p class="text-sm">{entry.message}</p>
              </div>
            ))}
          </div>
        )}

        <div class="mt-4 flex justify-end">
          <button
            onClick={onClose}
            class="px-4 py-2 border rounded hover:bg-gray-100"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
