import { useState, useEffect } from 'preact/hooks';
import Router from 'preact-router';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { Subscribers } from './pages/Subscribers';
import { Campaigns } from './pages/Campaigns';
import { Settings } from './pages/Settings';
import { Login } from './pages/Login';
import { getStoredAuth, clearAuth } from './api';

export function App() {
  const [authenticated, setAuthenticated] = useState(null); // null = checking, true/false = known

  useEffect(() => {
    // If we have stored credentials, assume authenticated
    setAuthenticated(!!getStoredAuth());
  }, []);

  const handleLogin = () => {
    setAuthenticated(true);
  };

  const handleLogout = () => {
    clearAuth();
    setAuthenticated(false);
  };

  if (authenticated === null) {
    return (
      <div class="min-h-screen flex items-center justify-center bg-gray-100">
        <div class="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (!authenticated) {
    return <Login onLogin={handleLogin} />;
  }

  return (
    <Layout onLogout={handleLogout}>
      <Router>
        <Dashboard path="/" />
        <Subscribers path="/subscribers" />
        <Campaigns path="/campaigns" />
        <Settings path="/settings" />
      </Router>
    </Layout>
  );
}
