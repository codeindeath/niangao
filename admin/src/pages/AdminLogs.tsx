import { useEffect, useState, useCallback } from 'react';
import { fetchLogs } from '../api/endpoints';
import type { AdminLog, PaginatedData } from '../api/endpoints';

export default function AdminLogs() {
  const [data, setData] = useState<PaginatedData<AdminLog> | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const pageSize = 30;
  const [actionTypeFilter, setActionTypeFilter] = useState('');
  const [keywordSearch, setKeywordSearch] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  void setStartDate; void setEndDate; // TODO: date picker UI

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = { page: String(page), page_size: String(pageSize) };
      if (actionTypeFilter) params.action_type = actionTypeFilter;
      if (keywordSearch) params.keyword = keywordSearch;
      if (startDate) params.start_date = startDate;
      if (endDate) params.end_date = endDate;
      const result = await fetchLogs(params);
      setData(result);
    } catch { /* ignore */ }
    setLoading(false);
  }, [page, actionTypeFilter, keywordSearch, startDate, endDate]);

  useEffect(() => { load(); }, [load]);

  const handleFilterChange = (setter: (v: string) => void, value: string) => {
    setter(value);
    setPage(1);
  };

  const totalPages = data ? Math.ceil(data.total / pageSize) : 0;

  const actionLabels: Record<string, string> = {
    review_approve: '审核通过',
    review_reject: '审核拒绝',
    user_enable: '启用用户',
    user_disable: '禁用用户',
    content_delete: '删除内容',
    content_edit: '编辑内容',
    config_update: '更新配置',
    domain_create: '创建领域',
    domain_update: '更新领域',
    platform_create: '创建平台内容',
    platform_edit: '编辑平台内容',
    platform_publish: '发布平台内容',
    platform_hide: '隐藏平台内容',
    platform_rescore: '重新评分',
  };

  const isSensitive = (actionType: string) => {
    const label = actionLabels[actionType] || actionType;
    return label.includes('删除') || label.includes('禁用') || label.includes('封禁');
  };

  const handleExport = () => {
    if (!data?.data) return;
    const headers = ['时间', '操作者', '操作类型', '目标', '详情', '结果'];
    const rows = data.data.map(log => [
      log.created_at?.slice(0, 16).replace('T', ' ') || '',
      log.admin_name || '',
      actionLabels[log.action_type] || log.action_type,
      `${log.target_type ? log.target_type + ': ' : ''}${log.target_id || ''}`,
      log.detail || '',
      log.result === 'success' ? '成功' : log.result,
    ]);
    const csv = [headers, ...rows]
      .map(r => r.map(c => `"${String(c).replace(/"/g, '""')}"`).join(','))
      .join('\n');
    const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `admin-logs-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  if (loading && !data) {
    return <div className="empty-state"><span className="emoji">⏳</span><p>加载中...</p></div>;
  }

  if (!data || !data.data || data.data.length === 0) {
    return <div className="empty-state"><span className="emoji">📭</span><p>暂无操作日志</p></div>;
  }

  return (
    <div className="card">
      <div className="card-header">
        <h3>📋 操作日志</h3>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>
            共 {data.total} 条记录
          </span>
          <button className="btn btn-outline btn-sm" onClick={handleExport}>📥 导出CSV</button>
        </div>
      </div>
      <div style={{ display: 'flex', gap: 10, marginBottom: 14, flexWrap: 'wrap', alignItems: 'center' }}>
        <select
          value={actionTypeFilter}
          onChange={e => handleFilterChange(setActionTypeFilter, e.target.value)}
          style={{ minWidth: 130 }}
        >
          <option value="">全部操作类型</option>
          {Object.entries(actionLabels).map(([key, label]) => (
            <option key={key} value={key}>{label}</option>
          ))}
        </select>
        <input
          type="text"
          placeholder="关键词搜索..."
          value={keywordSearch}
          onChange={e => handleFilterChange(setKeywordSearch, e.target.value)}
          style={{ minWidth: 180 }}
        />
      </div>
      <table>
        <thead>
          <tr>
            <th>时间</th>
            <th>操作者</th>
            <th>操作类型</th>
            <th>目标</th>
            <th>详情</th>
            <th>结果</th>
          </tr>
        </thead>
        <tbody>
          {data.data.map(log => (
            <tr key={log.id} style={isSensitive(log.action_type) ? { background: 'var(--red-light)' } : undefined}>
              <td className="text-sm" style={{ whiteSpace: 'nowrap' }}>
                {log.created_at?.slice(0, 16).replace('T', ' ')}
              </td>
              <td>{log.admin_name}</td>
              <td>
                <span className="tag">{actionLabels[log.action_type] || log.action_type}</span>
              </td>
              <td className="text-sm">
                {log.target_type && <span style={{ color: 'var(--text-secondary)' }}>{log.target_type}: </span>}
                {log.target_id ? log.target_id.slice(0, 8) + '...' : '—'}
              </td>
              <td className="text-sm" style={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {log.detail || '—'}
              </td>
              <td>
                <span className={`badge ${log.result === 'success' ? 'badge-approved' : 'badge-rejected'}`}>
                  {log.result === 'success' ? '成功' : log.result}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {totalPages > 1 && (
        <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 16 }}>
          <button className="btn btn-outline btn-sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
            ← 上一页
          </button>
          <span className="text-sm" style={{ padding: '4px 12px' }}>第 {page} / {totalPages} 页</span>
          <button className="btn btn-outline btn-sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
            下一页 →
          </button>
        </div>
      )}
    </div>
  );
}
