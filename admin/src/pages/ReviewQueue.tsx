import { useEffect, useState } from 'react';
import { fetchReviews } from '../api/endpoints';
import type { ReviewItem, PaginatedData } from '../api/endpoints';
import { DOMAIN_LABELS } from '../api/endpoints';

export default function ReviewQueue() {
  const [data, setData] = useState<PaginatedData<ReviewItem> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    fetchReviews({ page: '1', page_size: '20' })
      .then((result) => {
        setData(result);
        setLoading(false);
      })
      .catch((err) => {
        console.error('REVIEW_QUEUE_ERROR:', err);
        setError(err?.message || String(err));
        setLoading(false);
      });
  }, []);

  if (loading) {
    return (
      <div className="empty-state">
        <span className="emoji">⏳</span>
        <p>加载中...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="empty-state">
        <span className="emoji">⚠️</span>
        <p>加载失败: {error}</p>
      </div>
    );
  }

  if (!data || !data.data || data.data.length === 0) {
    return (
      <div className="empty-state">
        <span className="emoji">📭</span>
        <p>暂无审核内容</p>
      </div>
    );
  }

  return (
    <table>
      <thead>
        <tr>
          <th>内容</th>
          <th>领域</th>
          <th>状态</th>
          <th>提交者</th>
        </tr>
      </thead>
      <tbody>
        {data.data.map((item) => (
          <tr key={item.id}>
            <td>{item.content}</td>
            <td>{DOMAIN_LABELS[item.domain] ?? item.domain}</td>
            <td>{item.review_status}</td>
            <td>{item.author_name}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
