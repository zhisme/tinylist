import Router from 'preact-router';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { Subscribers } from './pages/Subscribers';
import { Campaigns } from './pages/Campaigns';
import { Settings } from './pages/Settings';

// Base path injected by Vite at build time (configurable via VITE_BASE_PATH)
const BASE = __BASE_PATH__;

export function App() {
  // Browser handles Basic Auth automatically via 401 + WWW-Authenticate header
  // No need for custom login page or auth state management
  return (
    <Layout basePath={BASE}>
      <Router>
        <Dashboard path={`${BASE}/`} />
        <Dashboard path={`${BASE}/stats`} />
        <Subscribers path={`${BASE}/subscribers`} />
        <Campaigns path={`${BASE}/campaigns`} />
        <Settings path={`${BASE}/settings`} />
      </Router>
    </Layout>
  );
}
