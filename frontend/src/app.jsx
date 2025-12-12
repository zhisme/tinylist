import Router from 'preact-router';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { Subscribers } from './pages/Subscribers';
import { Campaigns } from './pages/Campaigns';
import { Settings } from './pages/Settings';

export function App() {
  // Browser handles Basic Auth automatically via 401 + WWW-Authenticate header
  // No need for custom login page or auth state management
  return (
    <Layout>
      <Router>
        <Dashboard path="/" />
        <Subscribers path="/subscribers" />
        <Campaigns path="/campaigns" />
        <Settings path="/settings" />
      </Router>
    </Layout>
  );
}
