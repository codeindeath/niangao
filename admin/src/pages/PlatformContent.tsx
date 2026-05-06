import { useEffect, useState, useCallback } from 'react';
import type React from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  fetchPlatformExperiences, createPlatformExperience, updatePlatformExperience, batchAIScore,
  togglePublishPlatformExperience, rescorePlatformExperience, importCSVPlatformExperiences,
} from '../api/endpoints';
import type { PlatformExperience, PaginatedData } from '../api/endpoints';
import { DOMAIN_LABELS } from '../api/endpoints';

const EMPTY_FORM: Partial<PlatformExperience> = {
  content: '',
  domain: '',
  sub_domain: '',
  creator_name: '',
  source_label: '',
  score_reason: '',
};

export default function PlatformContent() {
  const [searchParams, setSearchParams] = useSearchParams();

  const currentPage = parseInt(searchParams.get('page') ?? '1', 10);
  const currentDomain = searchParams.get('domain') ?? '';
  const currentInterpretation = searchParams.get('has_interpretation') ?? '';
  const currentSearch = searchParams.get('search') ?? '';

  const [data, setData] = useState<PaginatedData<PlatformExperience> | null>(null);
  const [loading, setLoading] = useState(true);

  // Create modal
  const [creating, setCreating] = useState(false);
  const [createForm, setCreateForm] = useState<Partial<PlatformExperience>>({ ...EMPTY_FORM });
  const [createSaving, setCreateSaving] = useState(false);

  // Edit modal
  const [editing, setEditing] = useState<PlatformExperience | null>(null);
  const [editForm, setEditForm] = useState<Partial<PlatformExperience>>({});
  const [editSaving, setEditSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    const params: Record<string, string> = { page: String(currentPage), page_size: '20' };
    if (currentDomain) params.domain = currentDomain;
    if (currentInterpretation) params.has_interpretation = currentInterpretation;
    if (currentSearch) params.search = currentSearch;

    try {
      const result = await fetchPlatformExperiences(params);
      setData(result);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [currentPage, currentDomain, currentInterpretation, currentSearch]);

  useEffect(() => { load(); }, [load]);

  const updateParams = (updates: Record<string, string>) => {
    const next = new URLSearchParams(searchParams);
    Object.entries(updates).forEach(([k, v]) => {
      if (v) next.set(k, v); else next.delete(k);
    });
    if (!('page' in updates)) next.set('page', '1');
    setSearchParams(next);
  };

  const openCreate = () => {
    setCreateForm({ ...EMPTY_FORM });
    setCreating(true);
  };

  const handleCreate = async () => {
    if (!createForm.content?.trim()) { alert('请填写内容'); return; }
    setCreateSaving(true);
    try {
      await createPlatformExperience(createForm);
      setCreating(false);
      await load();
    } catch {
      alert('创建失败');
    } finally {
      setCreateSaving(false);
    }
  };

  const openEdit = (item: PlatformExperience) => {
    setEditing(item);
    setEditForm({ ...item });
  };

  const handleEditSave = async () => {
    if (!editing) return;
    setEditSaving(true);
    try {
      await updatePlatformExperience(editing.id, editForm);
      setEditing(null);
      await load();
    } catch {
      alert('保存失败');
    } finally {
      setEditSaving(false);
    }
  };

  const [batchAILoading, setBatchAILoading] = useState(false);

  const handleBatchAI = async () => {
    if (!data?.data?.length) {
      alert('当前列表无平台经验');
      return;
    }
    const unscored = data.data.filter((e) => e.quality_score == null);
    if (unscored.length === 0) {
      alert('所有平台经验已有质量分');
      return;
    }
    if (!window.confirm(`将为 ${unscored.length} 条未评分的平台经验执行 AI 打分，继续？`)) return;
    setBatchAILoading(true);
    try {
      const result = await batchAIScore(unscored.map((e) => e.id));
      alert(`完成！成功 ${result.success} 条，失败 ${result.failed} 条`);
      await load();
    } catch {
      alert('批量 AI 打分失败');
    } finally {
      setBatchAILoading(false);
    }
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
          placeholder="搜索内容..."
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
          value={currentInterpretation}
          onChange={(e) => updateParams({ has_interpretation: e.target.value })}
        >
          <option value="">全部</option>
          <option value="true">有解读</option>
          <option value="false">无解读</option>
        </select>

        <button className="btn btn-outline btn-sm" onClick={() => load()} disabled={loading}>
          🔄 刷新
        </button>

        <div style={{ marginLeft: 'auto', display: 'flex', gap: 8 }}>
          <button className="btn btn-outline btn-sm" onClick={async () => {
            const csv = prompt('粘贴 CSV 内容（列：content,domain,sub_domain,creator_name,source_label,score_reason）：');
            if (!csv) return;
            try { const r = await importCSVPlatformExperiences(csv); alert(`导入完成：${JSON.stringify(r)}`); await load(); }
            catch { alert('导入失败'); }
          }}>
            📥 CSV导入
          </button>
          <button className="btn btn-outline btn-sm" onClick={handleBatchAI} disabled={batchAILoading}>
            {batchAILoading ? '⏳ AI分析中...' : '🤖 批量 AI'}
          </button>
          <button className="btn btn-green btn-sm" onClick={openCreate}>
            ＋ 新建
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="card">
        {loading ? (
          <div className="empty-state"><span className="emoji">⏳</span><p>加载中...</p></div>
        ) : !data || !data.data || data.data.length === 0 ? (
          <div className="empty-state"><span className="emoji">📭</span><p>暂无平台内容</p></div>
        ) : (
          <>
            <table>
              <thead>
                <tr>
                  <th style={{ maxWidth: 240 }}>内容</th>
                  <th>来源</th>
                  <th>领域</th>
                  <th>质量分</th>
                  <th>解读</th>
                  <th>👍 点赞</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((item) => (
                  <tr key={item.id}>
                    <td className="text-truncate" style={{ maxWidth: 240 }}>
                      {item.content?.length > 40 ? item.content.slice(0, 40) + '...' : item.content}
                    </td>
                    <td className="text-sm">{item.creator_name ?? '—'}</td>
                    <td>
                      <span className="badge badge-platform">
                        {DOMAIN_LABELS[item.domain] ?? item.domain}
                        {item.sub_domain ? ` · ${item.sub_domain}` : ''}
                      </span>
                    </td>
                    <td className="text-mono">
                      {item.quality_score != null ? (
                        <span className={item.quality_score >= 70 ? 'text-green' : item.quality_score >= 40 ? '' : 'text-red'}>
                          {item.quality_score.toFixed(0)}
                        </span>
                      ) : '—'}
                    </td>
                    <td>
                      {item.has_interpretation ? (
                        <span className="badge badge-approved">有</span>
                      ) : (
                        <span className="badge badge-private">无</span>
                      )}
                    </td>
                    <td className="text-mono">{item.like_count ?? 0}</td>
                    <td>
                      <div className="btn-group">
                        <button className="btn btn-outline btn-sm" onClick={() => openEdit(item)}>
                          ✏️ 编辑
                        </button>
                        <button
                          className="btn btn-outline btn-sm"
                          onClick={async () => {
                            try { await togglePublishPlatformExperience(item.id); await load(); }
                            catch { alert('切换失败'); }
                          }}
                          title={item.status === 'hidden' ? '发布' : '隐藏'}
                        >
                          {item.status === 'hidden' ? '▶ 发布' : '⏸ 隐藏'}
                        </button>
                        <button
                          className="btn btn-outline btn-sm"
                          onClick={async () => {
                            if (!window.confirm('重新AI打分？')) return;
                            try { await rescorePlatformExperience(item.id); await load(); }
                            catch { alert('重打分失败'); }
                          }}
                          title="重新AI打分"
                        >
                          🔄
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

      {/* Create Modal */}
      {creating && (
        <div className="modal-overlay" onClick={() => setCreating(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>＋ 新建平台内容</h3>
            <div className="field">
              <label>内容 *</label>
              <textarea
                value={createForm.content ?? ''}
                onChange={(e) => setCreateForm({ ...createForm, content: e.target.value })}
                rows={4}
              />
            </div>
            <div className="field">
              <label>领域</label>
              <select
                value={createForm.domain ?? ''}
                onChange={(e) => setCreateForm({ ...createForm, domain: e.target.value })}
              >
                <option value="">—</option>
                {Object.entries(DOMAIN_LABELS).map(([k, v]) => (
                  <option key={k} value={k}>{v}</option>
                ))}
              </select>
            </div>
            <div className="field">
              <label>子领域</label>
              <input
                value={createForm.sub_domain ?? ''}
                onChange={(e) => setCreateForm({ ...createForm, sub_domain: e.target.value })}
              />
            </div>
            <div className="field">
              <label>来源名</label>
              <input
                value={createForm.creator_name ?? ''}
                onChange={(e) => setCreateForm({ ...createForm, creator_name: e.target.value })}
              />
            </div>
            <div className="field">
              <label>来源标签</label>
              <input
                value={createForm.source_label ?? ''}
                onChange={(e) => setCreateForm({ ...createForm, source_label: e.target.value })}
              />
            </div>
            <div className="field">
              <label>评分依据</label>
              <textarea
                value={createForm.score_reason ?? ''}
                onChange={(e) => setCreateForm({ ...createForm, score_reason: e.target.value })}
                rows={3}
              />
            </div>
            <div className="modal-actions">
              <button className="btn btn-outline" onClick={() => setCreating(false)}>取消</button>
              <button className="btn btn-green" onClick={handleCreate} disabled={createSaving}>
                {createSaving ? '创建中...' : '创建'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {editing && (
        <div className="modal-overlay" onClick={() => setEditing(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>✏️ 编辑平台内容</h3>
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
              <select
                value={editForm.domain ?? ''}
                onChange={(e) => setEditForm({ ...editForm, domain: e.target.value })}
              >
                <option value="">—</option>
                {Object.entries(DOMAIN_LABELS).map(([k, v]) => (
                  <option key={k} value={k}>{v}</option>
                ))}
              </select>
            </div>
            <div className="field">
              <label>子领域</label>
              <input
                value={editForm.sub_domain ?? ''}
                onChange={(e) => setEditForm({ ...editForm, sub_domain: e.target.value })}
              />
            </div>
            <div className="field">
              <label>来源名</label>
              <input
                value={editForm.creator_name ?? ''}
                onChange={(e) => setEditForm({ ...editForm, creator_name: e.target.value })}
              />
            </div>
            <div className="field">
              <label>来源标签</label>
              <input
                value={editForm.source_label ?? ''}
                onChange={(e) => setEditForm({ ...editForm, source_label: e.target.value })}
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
              <button className="btn btn-green" onClick={handleEditSave} disabled={editSaving}>
                {editSaving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
