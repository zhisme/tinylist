import Router from 'preact-router';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { Subscribers } from './pages/Subscribers';
import { Campaigns } from './pages/Campaigns';
import { Settings } from './pages/Settings';

export function App() {
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
