import { useEffect, useState } from 'react';
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from 'recharts';
import { fetchDashboard, fetchTrends, fetchAIStatus } from '../api/endpoints';
import type { Dashboard, Trends, Trend, AIStatus } from '../api/endpoints';
import { useNavigate } from 'react-router-dom';

export default function DashboardPage() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [trends, setTrends] = useState<Trends | null>(null);
  const [aiStatus, setAIStatus] = useState<AIStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const [dash, trendData, ai] = await Promise.all([
          fetchDashboard(),
          fetchTrends(7),
          fetchAIStatus(),
        ]);
        if (!cancelled) {
          setDashboard(dash);
          setTrends(trendData);
          setAIStatus(ai);
        }
      } catch {
        // silently ignore
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => { cancelled = true; };
  }, []);

  if (loading) {
    return (
      <div className="empty-state">
        <span className="emoji">⏳</span>
        <p>加载中...</p>
      </div>
    );
  }

  if (!dashboard) {
    return (
      <div className="empty-state">
        <span className="emoji">⚠️</span>
        <p>数据加载失败，请刷新重试</p>
      </div>
    );
  }

  const chartData = trends?.users?.map((u: Trend, i: number) => ({
    date: u.date.slice(5),
    users: u.count,
    exps: trends.experiences[i]?.count ?? 0,
  })) ?? [];

  const approvalRate = (dashboard.today_approved + dashboard.today_rejected) > 0
    ? ((dashboard.today_approved / (dashboard.today_approved + dashboard.today_rejected)) * 100).toFixed(1)
    : '—';

  const userVsYesterday = dashboard.yesterday_new_users != null
    ? (dashboard.today_new_users - dashboard.yesterday_new_users)
    : null;
  const expVsYesterday = dashboard.yesterday_new_exps != null
    ? (dashboard.today_new_exps - dashboard.yesterday_new_exps)
    : null;

  return (
    <>
      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card green">
          <div className="label">总用户数</div>
          <div className="value">{dashboard.total_users.toLocaleString()}</div>
          <div className="sub">
            ↑ {dashboard.today_new_users} 今日新增
            {userVsYesterday !== null && (
              <span style={{ color: userVsYesterday >= 0 ? 'var(--green)' : 'var(--red)', marginLeft: 6 }}>
                {userVsYesterday >= 0 ? '↑' : '↓'} vs昨日 {Math.abs(userVsYesterday)}
              </span>
            )}
          </div>
        </div>
        <div className="stat-card">
          <div className="label">总经验数</div>
          <div className="value">{dashboard.total_experiences.toLocaleString()}</div>
          <div className="sub">
            ↑ {dashboard.today_new_exps} 今日新增
            {expVsYesterday !== null && (
              <span style={{ color: expVsYesterday >= 0 ? 'var(--green)' : 'var(--red)', marginLeft: 6 }}>
                {expVsYesterday >= 0 ? '↑' : '↓'} vs昨日 {Math.abs(expVsYesterday)}
              </span>
            )}
          </div>
        </div>
        <div className="stat-card">
          <div className="label">今日活跃用户</div>
          <div className="value">{dashboard.today_active_users.toLocaleString()}</div>
          <div className="sub">DAU</div>
        </div>
        <div className="stat-card">
          <div className="label">今日 AI 对话</div>
          <div className="value">{dashboard.today_ai_chats.toLocaleString()}</div>
          <div className="sub">AI 互动次数</div>
        </div>
        <div className="stat-card red">
          <div className="label">待审核</div>
          <div className="value">{dashboard.pending_reviews}</div>
          <div className="sub">⏱ 待处理</div>
        </div>
        <div className="stat-card green">
          <div className="label">今日审核通过</div>
          <div className="value">{dashboard.today_approved}</div>
          <div className="sub">通过率 {approvalRate}%</div>
        </div>
      </div>

      {/* Row: Chart + AI Status */}
      <div className="row">
        <div className="card">
          <div className="card-header">
            <h3>📈 近 7 天趋势</h3>
            <span className="text-sm text-green">新增 · 用户/经验</span>
          </div>
          {chartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={200}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e8e4df" />
                <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#7a7670' }} />
                <YAxis tick={{ fontSize: 11, fill: '#7a7670' }} />
                <Tooltip
                  contentStyle={{
                    background: '#fff',
                    border: '1px solid #e8e4df',
                    borderRadius: 8,
                    fontSize: 12,
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="users"
                  stroke="#4a7c59"
                  strokeWidth={2}
                  dot={{ r: 3, fill: '#4a7c59' }}
                  name="用户"
                />
                <Line
                  type="monotone"
                  dataKey="exps"
                  stroke="#e67e22"
                  strokeWidth={2}
                  dot={{ r: 3, fill: '#e67e22' }}
                  name="经验"
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <div className="empty-state" style={{ padding: 20 }}>
              <p>暂无趋势数据</p>
            </div>
          )}
        </div>

        {/* AI Status — enhanced */}
        <div className="card">
          <div className="card-header">
            <h3>🤖 AI 服务状态</h3>
            <span className={`badge ${aiStatus?.healthy ? 'badge-approved' : 'badge-rejected'}`} style={{ fontSize: 12, padding: '4px 10px' }}>
              {aiStatus?.healthy ? '● 正常' : '● 异常'}
            </span>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, fontSize: 13 }}>
            <div>
              <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>模型</span>
              <br />{aiStatus?.model ?? '—'}
            </div>
            <div>
              <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>延迟</span>
              <br />{aiStatus ? `${aiStatus.latency_ms.toFixed(0)}ms` : '—'}
            </div>
            {aiStatus?.tier_stats && (
              <>
                <div>
                  <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>审核</span>
                  <br />
                  <span className="text-green">{aiStatus.tier_stats.review.today} 今日</span>
                  <span className="text-sm"> / {aiStatus.tier_stats.review.total} 总计</span>
                </div>
                <div>
                  <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>对话</span>
                  <br />
                  <span className="text-green">{aiStatus.tier_stats.chat.today} 今日</span>
                  <span className="text-sm"> / {aiStatus.tier_stats.chat.total} 总计</span>
                </div>
                <div>
                  <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>解读生成</span>
                  <br />{aiStatus.tier_stats.interpretation.today} / {aiStatus.tier_stats.interpretation.total}
                </div>
              </>
            )}
            {aiStatus?.daily_cost && (
              <div>
                <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>预估费用</span>
                <br />
                <span style={{ fontWeight: 600 }}>¥{aiStatus.daily_cost.today_estimated.toFixed(2)}</span>
                <span className="text-sm"> / ¥{aiStatus.daily_cost.month_estimated.toFixed(2)} 月</span>
              </div>
            )}
            {aiStatus?.error_msg && (
              <div style={{ gridColumn: '1 / -1' }}>
                <span className="text-sm text-red">{aiStatus.error_msg}</span>
              </div>
            )}
            {aiStatus?.batch_tasks && aiStatus.batch_tasks.length > 0 && (
              <div style={{ gridColumn: '1 / -1', marginTop: 8 }}>
                <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>最近批任务</span>
                {aiStatus.batch_tasks.slice(0, 3).map((t, i) => (
                  <div key={i} className="text-sm" style={{ padding: '2px 0' }}>
                    {t.action_type} — {t.result} — {t.created_at?.slice(0, 16)}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Review Queue Preview — real data */}
      <div className="card" style={{ marginBottom: 20, cursor: 'pointer' }} onClick={() => navigate('/reviews')}>
        <div className="card-header">
          <h3>🔍 待审核队列 · {dashboard.pending_reviews} 条</h3>
          <span className="btn btn-outline btn-sm">查看全部 →</span>
        </div>
        {dashboard.review_preview && dashboard.review_preview.length > 0 ? (
          <div>
            {dashboard.review_preview.map((r) => (
              <div
                key={r.id}
                style={{
                  padding: '8px 16px',
                  borderBottom: '1px solid var(--border)',
                  fontSize: 13,
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                }}
              >
                <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', marginRight: 12 }}>
                  <span style={{ color: 'var(--text-secondary)', marginRight: 8 }}>
                    [{r.domain}]
                  </span>
                  {r.content}
                </span>
                <span className="text-sm" style={{ color: 'var(--text-secondary)', whiteSpace: 'nowrap' }}>
                  {r.submitted_at?.slice(0, 10)}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <div className="empty-state" style={{ padding: 20 }}>
            <span className="emoji">🎉</span>
            <p>暂无待审核内容</p>
          </div>
        )}
      </div>
    </>
  );
}
