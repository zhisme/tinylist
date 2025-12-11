import { useState } from 'preact/hooks';
import { validateCredentials, setAuthCredentials } from '../api';

export function Login({ onLogin }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const valid = await validateCredentials(username, password);
      if (valid) {
        setAuthCredentials(username, password);
        onLogin();
      } else {
        setError('Invalid username or password');
      }
    } catch (err) {
      setError('Connection failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="min-h-screen flex items-center justify-center bg-gray-100">
      <div class="max-w-md w-full bg-white rounded-lg shadow-md p-8">
        <div class="text-center mb-8">
          <h1 class="text-2xl font-bold text-gray-900">TinyList</h1>
          <p class="text-gray-600 mt-2">Sign in to admin panel</p>
        </div>

        <form onSubmit={handleSubmit}>
          {error && (
            <div class="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-md text-sm">
              {error}
            </div>
          )}

          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1" for="username">
              Username
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onInput={(e) => setUsername(e.target.value)}
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              required
              autoFocus
            />
          </div>

          <div class="mb-6">
            <label class="block text-sm font-medium text-gray-700 mb-1" for="password">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onInput={(e) => setPassword(e.target.value)}
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            class="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  );
}
