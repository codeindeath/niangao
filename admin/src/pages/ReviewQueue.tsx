import { useEffect, useState, useRef, useCallback } from 'react';
import {
  fetchReviews, approveReview, rejectReview, batchReview, retryReview,
} from '../api/endpoints';
import type { ReviewItem, PaginatedData } from '../api/endpoints';

// ── Helpers ──
const ts = (d: string) => {
  const dt = new Date(d);
  const now = Date.now();
  const diff = now - dt.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return '刚刚';
  if (mins < 60) return `${mins}分钟前`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}小时前`;
  const days = Math.floor(hrs / 24);
  return `${days}天前`;
};

const DOMAIN_LABELS: Record<string, string> = {
  vitality: '生命',
  living: '生活',
  work: '工作',
  relationship: '关系',
  cognition: '认知',
  meaning: '意义',
};

const REJECT_REASONS = [
  '内容不符合社区规范',
  '非经验内容（纯知识/定义）',
  '内容过于空泛',
  '涉及广告或推广',
  '内容重复',
  '其他',
];

// ── Component ──
export default function ReviewQueuePage() {
  const [items, setItems] = useState<ReviewItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(true);

  // Selection
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [selectAll, setSelectAll] = useState(false);

  // Reject modal
  const [rejectTarget, setRejectTarget] = useState<ReviewItem | null>(null);
  const [rejectReason, setRejectReason] = useState('');
  const [rejectCustomReason, setRejectCustomReason] = useState('');
  const [rejecting, setRejecting] = useState(false);

  // AI detail modal
  const [detailTarget, setDetailTarget] = useState<ReviewItem | null>(null);

  // Batch
  const [batchAction, setBatchAction] = useState<'approve' | 'reject' | null>(null);
  const [batchReason, setBatchReason] = useState('');
  const [batchProcessing, setBatchProcessing] = useState(false);

  // Toast
  const [toast, setToast] = useState<{ msg: string; type: 'ok' | 'err' } | null>(null);
  const toastTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const showToast = (msg: string, type: 'ok' | 'err' = 'ok') => {
    setToast({ msg, type });
    if (toastTimer.current) clearTimeout(toastTimer.current);
    toastTimer.current = setTimeout(() => setToast(null), 3000);
  };

  // Infinite scroll
  const sentinelRef = useRef<HTMLDivElement>(null);

  // ── Load ──
  const loadPage = useCallback(async (p: number, append: boolean) => {
    try {
      const res = await fetchReviews({ page: String(p), page_size: '20' }) as unknown as PaginatedData<ReviewItem>;
      const data = res.data || [];
      if (append) {
        setItems(prev => [...prev, ...data]);
      } else {
        setItems(data);
      }
      setTotal(res.total || 0);
      setHasMore(data.length === 20);
      setError(null);
    } catch {
      setError('加载审核队列失败');
      showToast('加载失败，请重试', 'err');
    }
  }, []);

  useEffect(() => {
    setLoading(true);
    loadPage(1, false).finally(() => setLoading(false));
  }, [loadPage]);

  // Infinite scroll observer
  useEffect(() => {
    if (!sentinelRef.current) return;
    const obs = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting && hasMore && !loadingMore) {
        setLoadingMore(true);
        const next = page + 1;
        setPage(next);
        loadPage(next, true).finally(() => setLoadingMore(false));
      }
    }, { rootMargin: '200px' });
    obs.observe(sentinelRef.current);
    return () => obs.disconnect();
  }, [hasMore, loadingMore, page, loadPage]);

  // ── Select ──
  const toggleSelect = (id: string) => {
    setSelected(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
    setSelectAll(false);
  };

  const toggleSelectAll = () => {
    if (selectAll) {
      setSelected(new Set());
      setSelectAll(false);
    } else {
      setSelected(new Set(items.map(i => i.id)));
      setSelectAll(true);
    }
  };

  // ── Actions ──
  const handleApprove = async (item: ReviewItem) => {
    try {
      await approveReview(item.id);
      setItems(prev => prev.filter(i => i.id !== item.id));
      setSelected(prev => { const n = new Set(prev); n.delete(item.id); return n; });
      showToast('已通过');
    } catch (e: any) {
      if (e?.status === 409) {
        setItems(prev => prev.filter(i => i.id !== item.id));
        showToast('该经验已被审核');
      } else {
        showToast('操作失败', 'err');
      }
    }
  };

  const handleRejectClick = (item: ReviewItem) => {
    setRejectTarget(item);
    setRejectReason('');
    setRejectCustomReason('');
  };

  const handleRejectConfirm = async () => {
    if (!rejectTarget) return;
    const reason = rejectReason === '其他' ? rejectCustomReason.trim() : rejectReason;
    if (!reason) { showToast('请填写拒绝理由', 'err'); return; }
    setRejecting(true);
    try {
      await rejectReview(rejectTarget.id, reason);
      setItems(prev => prev.filter(i => i.id !== rejectTarget.id));
      setSelected(prev => { const n = new Set(prev); n.delete(rejectTarget.id); return n; });
      showToast('已拒绝');
      setRejectTarget(null);
    } catch (e: any) {
      if (e?.status === 409) {
        setItems(prev => prev.filter(i => i.id !== rejectTarget.id));
        showToast('该经验已被审核');
        setRejectTarget(null);
      } else {
        showToast('操作失败', 'err');
      }
    } finally {
      setRejecting(false);
    }
  };

  const handleRetry = async (item: ReviewItem) => {
    try {
      await retryReview(item.id);
      setItems(prev => prev.filter(i => i.id !== item.id));
      setSelected(prev => { const n = new Set(prev); n.delete(item.id); return n; });
      showToast('已退回重审');
    } catch {
      showToast('操作失败', 'err');
    }
  };

  // Batch
  const handleBatchApprove = () => {
    setBatchAction('approve');
    setBatchReason('');
  };

  const handleBatchReject = () => {
    setBatchAction('reject');
    setBatchReason('');
  };

  const handleBatchConfirm = async () => {
    if (batchAction === 'reject' && !batchReason.trim()) {
      showToast('请填写批量拒绝理由', 'err');
      return;
    }
    setBatchProcessing(true);
    try {
      const ids = Array.from(selected);
      await batchReview(ids, batchAction!, batchReason || undefined);
      setItems(prev => prev.filter(i => !selected.has(i.id)));
      setSelected(new Set());
      setSelectAll(false);
      showToast(`${batchAction === 'approve' ? '已通过' : '已拒绝'} ${ids.length} 条`);
    } catch {
      showToast('批量操作失败', 'err');
    } finally {
      setBatchProcessing(false);
      setBatchAction(null);
    }
  };

  // ── Render ──
  const verdcitBadge = (item: ReviewItem) => {
    if (item.hard_policy_result) {
      try {
        const hp = typeof item.hard_policy_result === 'string'
          ? JSON.parse(item.hard_policy_result) : item.hard_policy_result;
        if (hp && (hp as any).passed === false) {
          return <span className="badge badge-rejected" title="硬策略拒绝">硬拒绝</span>;
        }
      } catch { /* ignore */ }
    }
    if (item.ai_verdict === 'approved') {
      return <span className="badge badge-approved" title={`AI评分: ${item.ai_score ?? '--'}`}>
        AI通过 {item.ai_score != null ? `· ${item.ai_score}` : ''}
      </span>;
    }
    if (item.ai_verdict === 'rejected') {
      return <span className="badge badge-rejected">AI拒绝</span>;
    }
    return <span className="badge badge-pending">待判定</span>;
  };

  if (loading) {
    return <div className="empty-state"><span className="emoji">⏳</span><p>加载中...</p></div>;
  }

  return (
    <div>
      {/* Header */}
      <div className="filter-bar" style={{ marginBottom: 16 }}>
        <span style={{ fontWeight: 600 }}>
          待审核：{total} 条
        </span>
        <span style={{ marginLeft: 'auto', fontSize: 12, color: 'var(--text-secondary)' }}>
          按提交时间从旧到新排序
        </span>
      </div>

      {/* Error */}
      {error && (
        <div style={{ background: 'var(--red-light)', color: 'var(--red)', padding: '10px 14px', borderRadius: 8, marginBottom: 14, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>{error}</span>
          <button className="btn btn-sm btn-outline" onClick={() => { setError(null); setLoading(true); loadPage(1, false).finally(() => setLoading(false)); }}>
            重试
          </button>
        </div>
      )}

      {/* Empty */}
      {!loading && items.length === 0 && (
        <div className="empty-state">
          <span className="emoji">🎉</span>
          <p>暂无待审核内容</p>
          <p style={{ fontSize: 13, color: 'var(--text-secondary)' }}>所有经验已处理完毕</p>
        </div>
      )}

      {/* Table */}
      {items.length > 0 && (
        <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
          <table style={{ marginBottom: 0 }}>
            <thead>
              <tr>
                <th style={{ width: 40 }}>
                  <input type="checkbox" checked={selectAll} onChange={toggleSelectAll} />
                </th>
                <th style={{ width: '40%' }}>内容</th>
                <th>领域</th>
                <th>AI判定</th>
                <th>提交时间</th>
                <th style={{ width: 180 }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {items.map(item => (
                <tr
                  key={item.id}
                  style={{
                    background: item.hard_policy_result ? 'var(--red-light)' : item.ai_verdict === 'rejected' ? 'var(--orange-light)' : undefined,
                  }}
                >
                  <td>
                    <input
                      type="checkbox"
                      checked={selected.has(item.id)}
                      onChange={() => toggleSelect(item.id)}
                    />
                  </td>
                  <td>
                    <div style={{ fontWeight: 500, marginBottom: 2 }}>{item.content}</div>
                    <div style={{ fontSize: 11, color: 'var(--text-secondary)' }}>
                      {item.author_name}
                      {item.source_type === 'platform' && <span className="badge badge-platform" style={{ marginLeft: 6 }}>官</span>}
                    </div>
                  </td>
                  <td>
                    <span className="badge">{DOMAIN_LABELS[item.domain] || item.domain}</span>
                    {item.sub_domain && <div style={{ fontSize: 11, color: 'var(--text-secondary)', marginTop: 2 }}>{item.sub_domain}</div>}
                  </td>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
                      {verdcitBadge(item)}
                    </div>
                    {item.ai_interpretation && (
                      <button
                        className="btn btn-sm btn-outline"
                        onClick={() => setDetailTarget(detailTarget?.id === item.id ? null : item)}
                      >
                        {detailTarget?.id === item.id ? '收起详情' : '查看详情'}
                      </button>
                    )}
                  </td>
                  <td style={{ fontSize: 12, color: 'var(--text-secondary)' }}>{ts(item.submitted_at)}</td>
                  <td>
                    <div className="btn-group">
                      <button className="btn btn-sm btn-green" onClick={() => handleApprove(item)}>
                        通过
                      </button>
                      <button className="btn btn-sm btn-red" onClick={() => handleRejectClick(item)}>
                        拒绝
                      </button>
                      <button className="btn btn-sm btn-outline" onClick={() => handleRetry(item)}>
                        退回
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {/* Infinite scroll sentinel */}
          <div ref={sentinelRef} style={{ padding: 20, textAlign: 'center' }}>
            {loadingMore && <span style={{ color: 'var(--text-secondary)' }}>加载更多...</span>}
            {!hasMore && items.length > 0 && (
              <span style={{ color: 'var(--text-secondary)', fontSize: 13 }}>— 已加载全部待审核内容 —</span>
            )}
          </div>
        </div>
      )}

      {/* ── Floating Batch Bar ── */}
      {selected.size > 0 && (
        <div style={{
          position: 'fixed', bottom: 24, left: '50%', transform: 'translateX(-50%)',
          background: 'var(--sidebar)', color: '#fff', borderRadius: 12,
          padding: '12px 24px', display: 'flex', alignItems: 'center', gap: 16,
          boxShadow: '0 4px 20px rgba(0,0,0,0.25)', zIndex: 100,
        }}>
          <span>已选 {selected.size} 条</span>
          <button className="btn btn-green" onClick={handleBatchApprove} disabled={batchProcessing}>
            批量通过
          </button>
          <button className="btn btn-red" onClick={handleBatchReject} disabled={batchProcessing}>
            批量拒绝
          </button>
          <button className="btn btn-outline" style={{ color: '#fff', borderColor: 'rgba(255,255,255,0.3)' }}
            onClick={() => { setSelected(new Set()); setSelectAll(false); }}>
            取消
          </button>
        </div>
      )}

      {/* ── Reject Modal ── */}
      {rejectTarget && (
        <div className="modal-overlay" onClick={() => setRejectTarget(null)}>
          <div className="modal" onClick={e => e.stopPropagation()} style={{ maxWidth: 480 }}>
            <h3>拒绝经验</h3>
            <div style={{ marginBottom: 12, padding: 10, background: 'var(--bg)', borderRadius: 8, fontSize: 13 }}>
              {rejectTarget.content}
            </div>
            <div className="field">
              <label>拒绝理由（必填）</label>
              <select
                value={rejectReason}
                onChange={e => { setRejectReason(e.target.value); if (e.target.value !== '其他') setRejectCustomReason(''); }}
              >
                <option value="">选择理由...</option>
                {REJECT_REASONS.map(r => <option key={r} value={r}>{r}</option>)}
              </select>
            </div>
            {rejectReason === '其他' && (
              <div className="field">
                <label>自定义理由</label>
                <textarea
                  value={rejectCustomReason}
                  onChange={e => setRejectCustomReason(e.target.value)}
                  placeholder="请输入具体理由（≤500字）"
                  rows={3}
                  maxLength={500}
                />
              </div>
            )}
            <div style={{
              marginTop: 8, padding: '8px 12px', background: '#fef9e7',
              borderRadius: 6, fontSize: 12, color: '#b7950b',
            }}>
              ⚠️ 拒绝后用户无法修改重新提交，仅可在个人中心查看
            </div>
            <div className="modal-actions">
              <button className="btn btn-outline" onClick={() => setRejectTarget(null)}>取消</button>
              <button
                className="btn btn-red"
                disabled={!rejectReason || (rejectReason === '其他' && !rejectCustomReason.trim()) || rejecting}
                onClick={handleRejectConfirm}
              >
                {rejecting ? '处理中...' : '确认拒绝'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Batch Modal ── */}
      {batchAction && (
        <div className="modal-overlay" onClick={() => setBatchAction(null)}>
          <div className="modal" onClick={e => e.stopPropagation()} style={{ maxWidth: 440 }}>
            <h3>{batchAction === 'approve' ? '批量通过' : '批量拒绝'}</h3>
            <p style={{ color: 'var(--text-secondary)', marginBottom: 16 }}>
              确认{batchAction === 'approve' ? '通过' : '拒绝'}选中的 {selected.size} 条经验？
            </p>
            {batchAction === 'reject' && (
              <div className="field">
                <label>统一拒绝理由（必填）</label>
                <textarea
                  value={batchReason}
                  onChange={e => setBatchReason(e.target.value)}
                  placeholder="请输入统一理由"
                  rows={3}
                  maxLength={500}
                />
              </div>
            )}
            <div className="modal-actions">
              <button className="btn btn-outline" onClick={() => setBatchAction(null)}>取消</button>
              <button
                className={`btn ${batchAction === 'approve' ? 'btn-green' : 'btn-red'}`}
                disabled={batchAction === 'reject' && !batchReason.trim() || batchProcessing}
                onClick={handleBatchConfirm}
              >
                {batchProcessing ? '处理中...' : `确认${batchAction === 'approve' ? '通过' : '拒绝'}`}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── AI Detail Modal ── */}
      {detailTarget && (
        <div className="modal-overlay" onClick={() => setDetailTarget(null)}>
          <div className="modal" onClick={e => e.stopPropagation()} style={{ maxWidth: 640 }}>
            <h3>AI 审核详情</h3>
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', gap: 12, marginBottom: 12 }}>
                <span className={`badge ${detailTarget.ai_verdict === 'approved' ? 'badge-approved' : 'badge-rejected'}`}>
                  AI判定：{detailTarget.ai_verdict === 'approved' ? '通过' : detailTarget.ai_verdict === 'rejected' ? '拒绝' : '未判定'}
                </span>
                {detailTarget.ai_score != null && (
                  <span className="badge">综合分：{detailTarget.ai_score}</span>
                )}
              </div>
              {detailTarget.ai_score_detail && (
                <div style={{ fontSize: 13, lineHeight: 1.6 }}>
                  <strong>六维评分明细：</strong>
                  <pre style={{ background: 'var(--bg)', padding: 10, borderRadius: 6, fontSize: 12, marginTop: 8, whiteSpace: 'pre-wrap' }}>
                    {(() => {
                      try {
                        const d = typeof detailTarget.ai_score_detail === 'string'
                          ? JSON.parse(detailTarget.ai_score_detail) : detailTarget.ai_score_detail;
                        return JSON.stringify(d, null, 2);
                      } catch { return String(detailTarget.ai_score_detail); }
                    })()}
                  </pre>
                </div>
              )}
              {detailTarget.ai_interpretation && (
                <div style={{ marginTop: 12 }}>
                  <strong>AI 解读：</strong>
                  <p style={{ fontSize: 13, lineHeight: 1.6, marginTop: 4 }}>{detailTarget.ai_interpretation}</p>
                </div>
              )}
              {detailTarget.hard_policy_result && (
                <div style={{ marginTop: 12 }}>
                  <strong>硬策略检查：</strong>
                  <pre style={{ background: 'var(--red-light)', padding: 10, borderRadius: 6, fontSize: 12, marginTop: 4 }}>
                    {(() => {
                      try {
                        const hp = typeof detailTarget.hard_policy_result === 'string'
                          ? JSON.parse(detailTarget.hard_policy_result) : detailTarget.hard_policy_result;
                        return JSON.stringify(hp, null, 2);
                      } catch { return String(detailTarget.hard_policy_result); }
                    })()}
                  </pre>
                </div>
              )}
            </div>
            <div className="modal-actions">
              <button className="btn btn-outline" onClick={() => setDetailTarget(null)}>关闭</button>
            </div>
          </div>
        </div>
      )}

      {/* ── Toast ── */}
      {toast && (
        <div style={{
          position: 'fixed', top: 24, right: 24, zIndex: 200,
          padding: '10px 20px', borderRadius: 8, fontSize: 14,
          background: toast.type === 'ok' ? 'var(--green)' : 'var(--red)',
          color: '#fff', boxShadow: '0 2px 12px rgba(0,0,0,0.15)',
        }}>
          {toast.msg}
        </div>
      )}
    </div>
  );
}
