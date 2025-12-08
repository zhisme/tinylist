import { useState } from 'preact/hooks';

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
  const [saving, setSaving] = useState(false);

  function handleChange(field, value) {
    setSmtpConfig(prev => ({ ...prev, [field]: value }));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setSaving(true);
    // TODO: Implement settings save API when backend supports it
    alert('Settings API not yet implemented on backend');
    setSaving(false);
  }

  return (
    <div>
      <h1 class="text-2xl font-bold mb-6">Settings</h1>

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
                onInput={(e) => handleChange('port', parseInt(e.target.value))}
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

      <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mt-6 max-w-2xl">
        <p class="text-yellow-800 text-sm">
          <strong>Note:</strong> SMTP settings are currently configured via config.yaml file.
          The settings API will be available in a future update.
        </p>
      </div>
    </div>
  );
}
