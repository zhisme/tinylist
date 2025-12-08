import { useState, useEffect } from 'preact/hooks';
import { subscribers, campaigns } from '../api';

export function Dashboard() {
  const [stats, setStats] = useState({
    totalSubscribers: 0,
    verifiedSubscribers: 0,
    pendingSubscribers: 0,
    totalCampaigns: 0,
    sentCampaigns: 0,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadStats();
  }, []);

  async function loadStats() {
    try {
      setLoading(true);
      const [subsData, campaignsData] = await Promise.all([
        subscribers.list({ per_page: 1 }),
        campaigns.list(),
      ]);

      // Get counts by status
      const [verified, pending] = await Promise.all([
        subscribers.list({ status: 'verified', per_page: 1 }),
        subscribers.list({ status: 'pending', per_page: 1 }),
      ]);

      setStats({
        totalSubscribers: subsData.total || 0,
        verifiedSubscribers: verified.total || 0,
        pendingSubscribers: pending.total || 0,
        totalCampaigns: campaignsData.length || 0,
        sentCampaigns: campaignsData.filter(c => c.status === 'sent').length || 0,
      });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  if (loading) {
    return <div class="text-gray-500">Loading...</div>;
  }

  if (error) {
    return <div class="text-red-500">Error: {error}</div>;
  }

  return (
    <div>
      <h1 class="text-2xl font-bold mb-6">Dashboard</h1>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Subscribers"
          value={stats.totalSubscribers}
          color="blue"
        />
        <StatCard
          title="Verified"
          value={stats.verifiedSubscribers}
          color="green"
        />
        <StatCard
          title="Pending"
          value={stats.pendingSubscribers}
          color="yellow"
        />
        <StatCard
          title="Campaigns Sent"
          value={stats.sentCampaigns}
          subtitle={`of ${stats.totalCampaigns} total`}
          color="purple"
        />
      </div>
    </div>
  );
}

function StatCard({ title, value, subtitle, color }) {
  const colors = {
    blue: 'bg-blue-500',
    green: 'bg-green-500',
    yellow: 'bg-yellow-500',
    purple: 'bg-purple-500',
  };

  return (
    <div class="bg-white rounded-lg shadow p-6">
      <div class={`w-12 h-12 ${colors[color]} rounded-lg flex items-center justify-center mb-4`}>
        <span class="text-white text-xl font-bold">{value}</span>
      </div>
      <h3 class="text-gray-500 text-sm">{title}</h3>
      {subtitle && <p class="text-gray-400 text-xs">{subtitle}</p>}
    </div>
  );
}
