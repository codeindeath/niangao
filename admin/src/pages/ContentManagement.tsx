import { useEffect, useState, useCallback } from 'react';
import type React from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  fetchExperiences, updateExperience, deleteExperience,
  unpublishExperience, hardDeleteExperience, updateReviewStatus, exportExperiencesCSV,
} from '../api/endpoints';
import type { Experience, PaginatedData } from '../api/endpoints';
import { DOMAIN_LABELS } from '../api/endpoints';

const SOURCE_TYPES: Record<string, string> = {
  '': '全部来源',
  user: '用户',
  platform: '平台',
};

const REVIEW_STATUSES: Record<string, string> = {
  '': '全部状态',
  pending: '待审核',
  approved: '已通过',
  rejected: '已拒绝',
};

function reviewBadge(status: string) {
  if (status === 'pending') return <span className="badge badge-pending">待审核</span>;
  if (status === 'approved') return <span className="badge badge-approved">已通过</span>;
  if (status === 'rejected') return <span className="badge badge-rejected">已拒绝</span>;
  return <span className="badge">{status}</span>;
}

export default function ContentManagement() {
  const [searchParams, setSearchParams] = useSearchParams();

  const currentPage = parseInt(searchParams.get('page') ?? '1', 10);
  const currentSearch = searchParams.get('search') ?? '';
  const currentDomain = searchParams.get('domain') ?? '';
  const currentSourceType = searchParams.get('source_type') ?? '';
  const currentStatus = searchParams.get('status') ?? '';

  const [data, setData] = useState<PaginatedData<Experience> | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<Experience | null>(null);
  const [editForm, setEditForm] = useState<Partial<Experience>>({});
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    const params: Record<string, string> = { page: String(currentPage), page_size: '20' };
    if (currentSearch) params.search = currentSearch;
    if (currentDomain) params.domain = currentDomain;
    if (currentSourceType) params.source_type = currentSourceType;
    if (currentStatus) params.review_status = currentStatus;

    try {
      const result = await fetchExperiences(params);
      setData(result);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [currentPage, currentSearch, currentDomain, currentSourceType, currentStatus]);

  useEffect(() => { load(); }, [load]);

  const updateParams = (updates: Record<string, string>) => {
    const next = new URLSearchParams(searchParams);
    Object.entries(updates).forEach(([k, v]) => {
      if (v) next.set(k, v); else next.delete(k);
    });
    if (!('page' in updates)) next.set('page', '1');
    setSearchParams(next);
  };

  const openEdit = (item: Experience) => {
    setEditing(item);
    setEditForm({ ...item });
  };

  const handleEditSave = async () => {
    if (!editing) return;
    setSaving(true);
    try {
      await updateExperience(editing.id, editForm);
      setEditing(null);
      await load();
    } catch {
      alert('保存失败');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('确认软删除该内容？')) return;
    try {
      await deleteExperience(id);
      await load();
    } catch {
      alert('删除失败');
    }
  };

  const handleUnpublish = async (id: string) => {
    if (!window.confirm('确认下架该内容？将从推荐池移除。')) return;
    try { await unpublishExperience(id); await load(); }
    catch { alert('下架失败'); }
  };

  const handleHardDelete = async (id: string) => {
    if (!window.confirm('⚠️ 确认永久删除？此操作不可恢复！')) return;
    try { await hardDeleteExperience(id); await load(); }
    catch { alert('永久删除失败（可能已入池无法硬删）'); }
  };

  const handleReviewStatusChange = async (id: string, status: string) => {
    if (!window.confirm(`确认将审核状态改为「${status}」？`)) return;
    try { await updateReviewStatus(id, status); await load(); }
    catch { alert('修改失败'); }
  };

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 0;

  const [searchInput, setSearchInput] = useState(currentSearch);

  const doSearch = () => {
    updateParams({ search: searchInput });
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') doSearch();
  };

  return (
    <div>
      {/* Filter Bar */}
      <div className="filter-bar">
        <input
          placeholder="搜索内容关键词..."
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={handleSearchKeyDown}
        />
        <button className="btn btn-outline btn-sm" onClick={doSearch}>🔍 搜索</button>

        <select
          value={currentDomain}
          onChange={(e) => updateParams({ domain: e.target.value })}
        >
          <option value="">全部领域</option>
          {Object.entries(DOMAIN_LABELS).map(([k, v]) => (
            <option key={k} value={k}>{v}</option>
          ))}
        </select>

        <select
          value={currentSourceType}
          onChange={(e) => updateParams({ source_type: e.target.value })}
        >
          {Object.entries(SOURCE_TYPES).map(([k, v]) => (
            <option key={k} value={k}>{v}</option>
          ))}
        </select>

        <select
          value={currentStatus}
          onChange={(e) => updateParams({ status: e.target.value })}
        >
          {Object.entries(REVIEW_STATUSES).map(([k, v]) => (
            <option key={k} value={k}>{v}</option>
          ))}
        </select>

        <button className="btn btn-outline btn-sm" onClick={() => load()} disabled={loading}>
          🔄 刷新
        </button>

        <a href={exportExperiencesCSV()} className="btn btn-outline btn-sm" style={{ textDecoration: 'none' }}>
          📥 CSV导出
        </a>
      </div>

      {/* Table */}
      <div className="card">
        {loading ? (
          <div className="empty-state"><span className="emoji">⏳</span><p>加载中...</p></div>
        ) : !data || !data.data || data.data.length === 0 ? (
          <div className="empty-state"><span className="emoji">📭</span><p>暂无内容</p></div>
        ) : (
          <>
            <table>
              <thead>
                <tr>
                  <th style={{ maxWidth: 260 }}>内容</th>
                  <th>作者</th>
                  <th>领域</th>
                  <th>审核状态</th>
                  <th>👍 点赞</th>
                  <th>🔖 收藏</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((item) => (
                  <tr key={item.id}>
                    <td className="text-truncate" style={{ maxWidth: 260 }}>
                      {item.content?.length > 40 ? item.content.slice(0, 40) + '...' : item.content}
                    </td>
                    <td className="text-sm">{item.creator_name ?? '—'}</td>
                    <td>
                      <span className="badge badge-platform">
                        {DOMAIN_LABELS[item.domain] ?? item.domain}
                        {item.sub_domain ? ` · ${item.sub_domain}` : ''}
                      </span>
                    </td>
                    <td>{reviewBadge(item.review_status)}</td>
                    <td className="text-mono">{item.likes ?? 0}</td>
                    <td className="text-mono">{item.bookmarks ?? 0}</td>
                    <td className="text-mono">
                      {item.created_at ? new Date(item.created_at).toLocaleString('zh-CN', {
                        month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit',
                      }) : '—'}
                    </td>
                    <td>
                      <div className="btn-group">
                        <button className="btn btn-outline btn-sm" onClick={() => openEdit(item)}>✏️ 编辑</button>
                        <button className="btn btn-red btn-sm" onClick={() => handleDelete(item.id)}>🗑 删除</button>
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

      {/* Edit Modal */}
      {editing && (
        <div className="modal-overlay" onClick={() => setEditing(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>✏️ 编辑内容</h3>
            <div className="field">
              <label>内容</label>
              <textarea
                value={editForm.content ?? ''}
                onChange={(e) => setEditForm({ ...editForm, content: e.target.value })}
                rows={4}
              />
            </div>
            <div className="field">
              <label>领域</label>
              <input
                value={editForm.domain ?? ''}
                onChange={(e) => setEditForm({ ...editForm, domain: e.target.value })}
              />
            </div>
            <div className="field">
              <label>子领域</label>
              <input
                value={editForm.sub_domain ?? ''}
                onChange={(e) => setEditForm({ ...editForm, sub_domain: e.target.value })}
              />
            </div>
            <div className="field">
              <label>来源类型</label>
              <select
                value={editForm.source_type ?? ''}
                onChange={(e) => setEditForm({ ...editForm, source_type: e.target.value })}
              >
                <option value="">—</option>
                <option value="user">用户</option>
                <option value="platform">平台</option>
              </select>
            </div>
            <div className="field">
              <label>作者名</label>
              <input
                value={editForm.creator_name ?? ''}
                onChange={(e) => setEditForm({ ...editForm, creator_name: e.target.value })}
              />
            </div>
            <div className="field">
              <label>评分依据</label>
              <textarea
                value={editForm.score_reason ?? ''}
                onChange={(e) => setEditForm({ ...editForm, score_reason: e.target.value })}
                rows={3}
              />
            </div>
            <div className="modal-actions">
              <button className="btn btn-outline" onClick={() => setEditing(null)}>取消</button>
              <select
                onChange={(e) => e.target.value && handleReviewStatusChange(editing.id, e.target.value)}
                style={{ marginLeft: 'auto' }}
              >
                <option value="">改审核状态...</option>
                <option value="approved">通过</option>
                <option value="rejected">拒绝</option>
                <option value="pending">待审</option>
                <option value="private">私密</option>
              </select>
              <button className="btn btn-outline btn-sm" onClick={() => handleUnpublish(editing.id)}>⬇ 下架</button>
              <button className="btn btn-red btn-sm" onClick={() => handleHardDelete(editing.id)}>🗑 永久删除</button>
              <button className="btn btn-green" onClick={handleEditSave} disabled={saving}>
                {saving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
