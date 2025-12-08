import { useState, useEffect } from 'preact/hooks';
import { settings } from '../api.js';

export function Settings() {
  const [smtpConfig, setSmtpConfig] = useState({
    host: '',
    port: 587,
    username: '',
    password: '',
    from_email: '',
    from_name: '',
    tls: true,
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testEmail, setTestEmail] = useState('');
  const [message, setMessage] = useState(null);

  useEffect(() => {
    loadSettings();
  }, []);

  async function loadSettings() {
    try {
      const data = await settings.getSMTP();
      setSmtpConfig(data);
    } catch (err) {
      setMessage({ type: 'error', text: 'Failed to load settings: ' + err.message });
    } finally {
      setLoading(false);
    }
  }

  function handleChange(field, value) {
    setSmtpConfig(prev => ({ ...prev, [field]: value }));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setSaving(true);
    setMessage(null);
    try {
      await settings.updateSMTP(smtpConfig);
      setMessage({ type: 'success', text: 'Settings saved successfully' });
      // Reload to get masked password
      loadSettings();
    } catch (err) {
      setMessage({ type: 'error', text: 'Failed to save settings: ' + err.message });
    } finally {
      setSaving(false);
    }
  }

  async function handleTestEmail(e) {
    e.preventDefault();
    if (!testEmail) {
      setMessage({ type: 'error', text: 'Please enter an email address for testing' });
      return;
    }
    setTesting(true);
    setMessage(null);
    try {
      await settings.testSMTP(testEmail);
      setMessage({ type: 'success', text: 'Test email sent successfully! Check your inbox.' });
    } catch (err) {
      setMessage({ type: 'error', text: 'Failed to send test email: ' + err.message });
    } finally {
      setTesting(false);
    }
  }

  const isConfigured = smtpConfig.host && smtpConfig.from_email;

  if (loading) {
    return (
      <div class="flex items-center justify-center h-64">
        <div class="text-gray-500">Loading settings...</div>
      </div>
    );
  }

  return (
    <div>
      <h1 class="text-2xl font-bold mb-6">Settings</h1>

      {!isConfigured && (
        <div class="mb-4 p-4 rounded-lg bg-yellow-50 text-yellow-800 border border-yellow-200">
          <strong>SMTP not configured.</strong> Configure your SMTP settings below to enable email sending.
        </div>
      )}

      {message && (
        <div class={`mb-4 p-4 rounded-lg ${message.type === 'error' ? 'bg-red-50 text-red-800 border border-red-200' : 'bg-green-50 text-green-800 border border-green-200'}`}>
          {message.text}
        </div>
      )}

      <div class="bg-white rounded-lg shadow p-6 max-w-2xl">
        <h2 class="text-lg font-semibold mb-4">SMTP Configuration</h2>
        <form onSubmit={handleSubmit}>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium mb-1">SMTP Host</label>
              <input
                type="text"
                value={smtpConfig.host}
                onInput={(e) => handleChange('host', e.target.value)}
                class="w-full border rounded px-3 py-2"
                placeholder="smtp.gmail.com"
              />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Port</label>
              <input
                type="number"
                value={smtpConfig.port}
                onInput={(e) => handleChange('port', parseInt(e.target.value) || 587)}
                class="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Username</label>
              <input
                type="text"
                value={smtpConfig.username}
                onInput={(e) => handleChange('username', e.target.value)}
                class="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Password</label>
              <input
                type="password"
                value={smtpConfig.password}
                onInput={(e) => handleChange('password', e.target.value)}
                class="w-full border rounded px-3 py-2"
                placeholder={smtpConfig.password === '***' ? '(unchanged)' : ''}
              />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">From Email</label>
              <input
                type="email"
                value={smtpConfig.from_email}
                onInput={(e) => handleChange('from_email', e.target.value)}
                class="w-full border rounded px-3 py-2"
                placeholder="newsletter@example.com"
              />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">From Name</label>
              <input
                type="text"
                value={smtpConfig.from_name}
                onInput={(e) => handleChange('from_name', e.target.value)}
                class="w-full border rounded px-3 py-2"
                placeholder="Newsletter"
              />
            </div>
          </div>
          <div class="mt-4">
            <label class="flex items-center gap-2">
              <input
                type="checkbox"
                checked={smtpConfig.tls}
                onChange={(e) => handleChange('tls', e.target.checked)}
                class="rounded"
              />
              <span class="text-sm">Use TLS</span>
            </label>
          </div>
          <div class="mt-6">
            <button
              type="submit"
              disabled={saving}
              class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:opacity-50"
            >
              {saving ? 'Saving...' : 'Save Settings'}
            </button>
          </div>
        </form>
      </div>

      <div class="bg-white rounded-lg shadow p-6 mt-6 max-w-2xl">
        <h2 class="text-lg font-semibold mb-4">Test Email</h2>
        <p class="text-gray-600 text-sm mb-4">
          Send a test email to verify your SMTP configuration is working correctly.
        </p>
        <form onSubmit={handleTestEmail} class="flex gap-4">
          <input
            type="email"
            value={testEmail}
            onInput={(e) => setTestEmail(e.target.value)}
            class="flex-1 border rounded px-3 py-2"
            placeholder="Enter email address"
          />
          <button
            type="submit"
            disabled={testing || !smtpConfig.host}
            class="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 disabled:opacity-50"
          >
            {testing ? 'Sending...' : 'Send Test Email'}
          </button>
        </form>
        {!smtpConfig.host && (
          <p class="text-yellow-600 text-sm mt-2">
            Configure SMTP settings above before sending a test email.
          </p>
        )}
      </div>
    </div>
  );
}
