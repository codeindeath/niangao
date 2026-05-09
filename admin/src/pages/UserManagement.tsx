import { useEffect, useState, useCallback, useRef } from 'react';
import type React from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  fetchUsers, fetchUserDetail, toggleUserEnabled, batchUpdateUserStatus, exportUsersCSV,
} from '../api/endpoints';
import type { AdminUser, PaginatedData } from '../api/endpoints';
import { DOMAIN_LABELS } from '../api/endpoints';

const AUTH_PROVIDERS: Record<string, string> = {
  '': '全部',
  apple: 'Apple',
  dev: 'Dev',
};

function providerBadge(provider: string) {
  const label = AUTH_PROVIDERS[provider] ?? provider;
  return <span className="badge badge-platform">{label}</span>;
}

export default function UserManagement() {
  const [searchParams, setSearchParams] = useSearchParams();

  const currentPage = parseInt(searchParams.get('page') ?? '1', 10);
  const currentSearch = searchParams.get('search') ?? '';
  const currentProvider = searchParams.get('provider') ?? '';

  const [data, setData] = useState<PaginatedData<AdminUser> | null>(null);
  const [loading, setLoading] = useState(true);
  const [detailUser, setDetailUser] = useState<AdminUser | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [toggling, setToggling] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    const params: Record<string, string> = { page: String(currentPage), page_size: '20' };
    if (currentSearch) params.search = currentSearch;
    if (currentProvider) params.auth_provider = currentProvider;

    try {
      const result = await fetchUsers(params);
      setData(result);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [currentPage, currentSearch, currentProvider]);

  useEffect(() => { load(); }, [load]);

  const updateParams = (updates: Record<string, string>) => {
    const next = new URLSearchParams(searchParams);
    Object.entries(updates).forEach(([k, v]) => {
      if (v) next.set(k, v); else next.delete(k);
    });
    if (!('page' in updates)) next.set('page', '1');
    setSearchParams(next);
  };

  const openDetail = async (user: AdminUser) => {
    setDetailLoading(true);
    try {
      const detail = await fetchUserDetail(user.id);
      setDetailUser(detail);
    } catch {
      alert('获取用户详情失败');
    } finally {
      setDetailLoading(false);
    }
  };

  const handleToggle = async (user: AdminUser) => {
    const newEnabled = !user.is_active;
    const action = newEnabled ? '启用' : '禁用';
    if (newEnabled) {
      if (!window.confirm(`确认${action}用户「${user.nickname}」？`)) return;
    } else {
      const input = window.prompt(`确认禁用「${user.nickname}」？请输入理由（≤200字）：`);
      if (input === null) return;
      if (input.length > 200) { alert('理由不能超过200字'); return; }
    }
    setToggling(true);
    try {
      await toggleUserEnabled(user.id, newEnabled);
      await load();
    } catch {
      alert('操作失败');
    } finally {
      setToggling(false);
    }
  };

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 0;

  const [searchInput, setSearchInput] = useState(currentSearch);
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const doSearch = () => {
    updateParams({ search: searchInput });
  };

  // 搜索防抖 300ms
  useEffect(() => {
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current);
    searchTimerRef.current = setTimeout(() => {
      updateParams({ search: searchInput });
    }, 300);
    return () => {
      if (searchTimerRef.current) clearTimeout(searchTimerRef.current);
    };
  }, [searchInput]);

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') doSearch();
  };

  return (
    <div>
      {/* Filter Bar */}
      <div className="filter-bar">
        <input
          placeholder="搜索昵称..."
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={handleSearchKeyDown}
        />
        <button className="btn btn-outline btn-sm" onClick={doSearch}>🔍 搜索</button>

        <select
          value={currentProvider}
          onChange={(e) => updateParams({ provider: e.target.value })}
        >
          {Object.entries(AUTH_PROVIDERS).map(([k, v]) => (
            <option key={k} value={k}>{v}</option>
          ))}
        </select>

        <button className="btn btn-outline btn-sm" onClick={() => load()} disabled={loading}>
          🔄 刷新
        </button>

        <select
          value={searchParams.get('sort') ?? ''}
          onChange={(e) => updateParams({ sort: e.target.value })}
        >
          <option value="">最新注册</option>
          <option value="activity">按活跃度</option>
        </select>

        <a href={exportUsersCSV()} className="btn btn-outline btn-sm" style={{ textDecoration: 'none' }}>
          📥 CSV
        </a>

        {data && data.data && data.data.length > 0 && (
          <div style={{ marginLeft: 'auto', display: 'flex', gap: 6 }}>
            <button
              className="btn btn-green btn-sm"
              onClick={async () => {
                if (!window.confirm('批量启用所有可见用户？')) return;
                const adminStr = localStorage.getItem('admin_user');
                if (adminStr) {
                  try {
                    const admin = JSON.parse(adminStr);
                    if (admin.id && data.data.some(u => u.id === admin.id)) {
                      alert('不能批量操作自己的账号');
                      return;
                    }
                  } catch { /* ignore parse error */ }
                }
                try {
                  await batchUpdateUserStatus(data.data.map(u => u.id), true);
                  await load();
                } catch { alert('操作失败'); }
              }}
            >
              全部启用
            </button>
            <button
              className="btn btn-red btn-sm"
              onClick={async () => {
                if (!window.confirm('批量禁用所有可见用户？')) return;
                const adminStr = localStorage.getItem('admin_user');
                if (adminStr) {
                  try {
                    const admin = JSON.parse(adminStr);
                    if (admin.id && data.data.some(u => u.id === admin.id)) {
                      alert('不能批量操作自己的账号');
                      return;
                    }
                  } catch { /* ignore parse error */ }
                }
                try {
                  await batchUpdateUserStatus(data.data.map(u => u.id), false);
                  await load();
                } catch { alert('操作失败'); }
              }}
            >
              全部禁用
            </button>
          </div>
        )}
      </div>

      {/* Table */}
      <div className="card">
        {loading ? (
          <div className="empty-state"><span className="emoji">⏳</span><p>加载中...</p></div>
        ) : !data || !data.data || data.data.length === 0 ? (
          <div className="empty-state"><span className="emoji">👤</span><p>暂无用户</p></div>
        ) : (
          <>
            <table>
              <thead>
                <tr>
                  <th>昵称</th>
                  <th>认证方式</th>
                  <th>管理员</th>
                  <th>经验数</th>
                  <th>注册时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((user) => (
                  <tr key={user.id}>
                    <td style={{ fontWeight: 600 }}>
                      {user.nickname}
                      {user.auth_provider === 'apple' && (
                        <div style={{ fontSize: 11, color: 'var(--text-secondary)', fontWeight: 400 }}>
                          🍎 {user.id.slice(0, 8)}****{user.id.slice(-4)}
                        </div>
                      )}
                    </td>
                    <td>{providerBadge(user.auth_provider)}</td>
                    <td>
                      {user.is_active ? (
                        <span className="badge badge-approved">正常</span>
                      ) : (
                        <span className="badge badge-rejected">已禁用</span>
                      )}
                    </td>
                    <td className="text-mono">{user.exp_count ?? 0}</td>
                    <td className="text-mono">
                      {user.created_at ? new Date(user.created_at).toLocaleString('zh-CN', {
                        month: '2-digit', day: '2-digit',
                      }) : '—'}
                    </td>
                    <td>
                      <div className="btn-group">
                        <button
                          className="btn btn-outline btn-sm"
                          onClick={() => openDetail(user)}
                          disabled={detailLoading}
                        >
                          📋 详情
                        </button>
                        <button
                          className={`btn btn-sm ${user.is_active ? 'btn-red' : 'btn-green'}`}
                          onClick={() => handleToggle(user)}
                          disabled={toggling}
                        >
                          {user.is_active ? '禁用' : '启用'}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="pagination">
                <span
                  onClick={() => currentPage > 1 && updateParams({ page: String(currentPage - 1) })}
                  style={{ opacity: currentPage <= 1 ? 0.4 : 1, cursor: currentPage <= 1 ? 'default' : 'pointer' }}
                >
                  ‹ 上一页
                </span>
                {Array.from({ length: totalPages }, (_, i) => i + 1)
                  .filter((p) => {
                    if (totalPages <= 7) return true;
                    if (p === 1 || p === totalPages) return true;
                    if (Math.abs(p - currentPage) <= 1) return true;
                    return false;
                  })
                  .map((p, idx, arr) => {
                    const showEllipsis = idx > 0 && p - arr[idx - 1] > 1;
                    return (
                      <span key={p}>
                        {showEllipsis && <span style={{ cursor: 'default' }}>...</span>}
                        <span
                          className={p === currentPage ? 'active' : ''}
                          onClick={() => updateParams({ page: String(p) })}
                        >
                          {p}
                        </span>
                      </span>
                    );
                  })}
                <span
                  onClick={() => currentPage < totalPages && updateParams({ page: String(currentPage + 1) })}
                  style={{ opacity: currentPage >= totalPages ? 0.4 : 1, cursor: currentPage >= totalPages ? 'default' : 'pointer' }}
                >
                  下一页 ›
                </span>
              </div>
            )}
          </>
        )}
      </div>

      {/* User Detail Modal */}
      {detailUser && (
        <div className="modal-overlay" onClick={() => setDetailUser(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>👤 {detailUser.nickname}</h3>

            <div className="stats-row">
              <div className="stat-item">
                <div className="stat-val">{detailUser.like_received ?? 0}</div>
                <div className="stat-label">被点赞</div>
              </div>
              <div className="stat-item">
                <div className="stat-val">{detailUser.bookmark_received ?? 0}</div>
                <div className="stat-label">被收藏</div>
              </div>
              <div className="stat-item">
                <div className="stat-val">{detailUser.viewed_count ?? 0}</div>
                <div className="stat-label">被浏览</div>
              </div>
              <div className="stat-item">
                <div className="stat-val">{detailUser.liked_count ?? 0}</div>
                <div className="stat-label">发出的赞</div>
              </div>
              <div className="stat-item">
                <div className="stat-val">{detailUser.bookmarked_count ?? 0}</div>
                <div className="stat-label">发出的收藏</div>
              </div>
              <div className="stat-item">
                <div className="stat-val">{detailUser.chat_count ?? 0}</div>
                <div className="stat-label">AI 对话数</div>
              </div>
            </div>

            {detailUser.domain_distribution && Object.keys(detailUser.domain_distribution).length > 0 && (
              <div>
                <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 8 }}>
                  领域分布
                </div>
                <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                  {Object.entries(detailUser.domain_distribution).map(([domain, count]) => (
                    <span key={domain} className="badge badge-platform">
                      {DOMAIN_LABELS[domain] ?? domain}: {count}
                    </span>
                  ))}
                </div>
              </div>
            )}

            <div className="modal-actions">
              <button className="btn btn-outline" onClick={() => setDetailUser(null)}>关闭</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
