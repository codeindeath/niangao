import { useEffect, useState } from 'react';
import {
  fetchUserStats, fetchExperienceStats, fetchInteractionStats,
  fetchReviewStats, fetchDomainDistribution, fetchAIStats, fetchRetention,
} from '../api/endpoints';

type StatMap = Record<string, any>;

export default function Statistics() {
  const [users, setUsers] = useState<StatMap | null>(null);
  const [experiences, setExperiences] = useState<StatMap | null>(null);
  const [interactions, setInteractions] = useState<StatMap | null>(null);
  const [reviews, setReviews] = useState<StatMap | null>(null);
  const [domains, setDomains] = useState<StatMap | null>(null);
  const [ai, setAi] = useState<StatMap | null>(null);
  const [retention, setRetention] = useState<StatMap | null>(null);
  const [loading, setLoading] = useState(true);
  const [days, setDays] = useState(7);

  const load = async () => {
    setLoading(true);
    try {
      const [u, e, i, r, d, a, rt] = await Promise.all([
        fetchUserStats(days),
        fetchExperienceStats(days),
        fetchInteractionStats(days),
        fetchReviewStats(days),
        fetchDomainDistribution(),
        fetchAIStats(days),
        fetchRetention(days),
      ]);
      setUsers(u);
      setExperiences(e);
      setInteractions(i);
      setReviews(r);
      setDomains(d);
      setAi(a);
      setRetention(rt);
    } catch { /* ignore */ }
    setLoading(false);
  };

  useEffect(() => { load(); }, [days]);

  const StatCard = ({ title, data, render }: { title: string; data: StatMap | null; render: (d: StatMap) => React.ReactNode }) => (
    <div className="card" style={{ minWidth: 280 }}>
      <h3 style={{ marginBottom: 12 }}>{title}</h3>
      {data ? render(data) : <div className="text-sm" style={{ color: 'var(--text-secondary)' }}>—</div>}
    </div>
  );

  const renderTrends = (data: any) => {
    const items = data?.data || data?.users || data?.experiences || [];
    if (!Array.isArray(items) || items.length === 0) {
      return <div className="text-sm" style={{ color: 'var(--text-secondary)' }}>暂无数据</div>;
    }
    const maxVal = Math.max(...items.map((t: any) => t.count || 0), 1);
    return (
      <div>
        {items.map((item: any, i: number) => (
          <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
            <span className="text-sm" style={{ width: 50, textAlign: 'right' }}>{item.date?.slice(5) || item.day || i}</span>
            <div style={{
              height: 18,
              width: `${Math.max((item.count / maxVal) * 100, 2)}%`,
              background: 'var(--green)',
              borderRadius: 4,
              minWidth: 2,
            }} />
            <span className="text-sm" style={{ fontWeight: 600, minWidth: 30 }}>{item.count}</span>
          </div>
        ))}
      </div>
    );
  };

  const renderDomainDist = (data: any) => {
    const items = data?.stats || data?.domain_distribution || data?.data || [];
    if (!Array.isArray(items) || items.length === 0) {
      return <div className="text-sm" style={{ color: 'var(--text-secondary)' }}>暂无数据</div>;
    }
    const maxVal = Math.max(...items.map((t: any) => t.count || 0), 1);
    return (
      <div>
        {items.map((item: any, i: number) => (
          <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
            <span className="text-sm" style={{ width: 60 }}>{item.domain || item.name}</span>
            <div style={{
              height: 18,
              width: `${Math.max((item.count / maxVal) * 100, 2)}%`,
              background: 'var(--orange)',
              borderRadius: 4,
              minWidth: 2,
            }} />
            <span className="text-sm" style={{ fontWeight: 600 }}>{item.count}</span>
          </div>
        ))}
      </div>
    );
  };

  const renderSummary = (data: any) => {
    const entries = Object.entries(data).filter(([k]) => !['data', 'days', 'trends'].includes(k));
    if (entries.length === 0) return <div className="text-sm" style={{ color: 'var(--text-secondary)' }}>暂无</div>;
    return (
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 6 }}>
        {entries.map(([k, v]) => (
          <div key={k}>
            <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>{k}: </span>
            <span style={{ fontWeight: 600 }}>{typeof v === 'number' ? v.toLocaleString() : String(v)}</span>
          </div>
        ))}
      </div>
    );
  };

  if (loading) return <div className="empty-state"><span className="emoji">⏳</span><p>加载中...</p></div>;

  return (
    <>
      <div className="toolbar">
        <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>统计天数:</span>
        {[7, 14, 30].map(d => (
          <button
            key={d}
            className={`btn btn-sm ${days === d ? 'btn-green' : 'btn-outline'}`}
            onClick={() => setDays(d)}
          >{d}天</button>
        ))}
        <button className="btn btn-outline btn-sm" onClick={load}>🔄 刷新</button>
      </div>

      <div className="stats-grid">
        <StatCard title="👥 用户增长" data={users} render={renderTrends} />
        <StatCard title="📝 经验增长" data={experiences} render={renderTrends} />
        <StatCard title="💬 互动统计" data={interactions} render={renderSummary} />
        <StatCard title="✅ 审核统计" data={reviews} render={renderSummary} />
        <StatCard title="📊 领域分布" data={domains} render={renderDomainDist} />
        <StatCard title="🤖 AI 统计" data={ai} render={renderSummary} />
        <StatCard title="📈 留存率" data={retention} render={renderSummary} />
      </div>
    </>
  );
}
