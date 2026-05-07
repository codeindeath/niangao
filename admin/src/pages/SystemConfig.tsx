import { useEffect, useState } from 'react';
import { fetchConfig, updateConfig, fetchConfigDefaults, fetchSensitiveWords, addSensitiveWord, deleteSensitiveWord } from '../api/endpoints';
import type { SystemConfig } from '../api/endpoints';

export default function SystemConfigPage() {
  const [config, setConfig] = useState<SystemConfig | null>(null);
  const [defaults, setDefaults] = useState<SystemConfig | null>(null);
  const [words, setWords] = useState<{ id: number; word: string }[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [newWord, setNewWord] = useState('');
  const [editError, setEditError] = useState('');

  const load = async () => {
    setLoading(true);
    try {
      const [c, d, w] = await Promise.all([fetchConfig(), fetchConfigDefaults(), fetchSensitiveWords()]);
      setConfig(c);
      setDefaults(d);
      setWords(w);
    } catch { /* ignore */ }
    setLoading(false);
  };

  useEffect(() => { load(); }, []);

  const saveConfig = async (key: string) => {
    // 类型校验：根据默认值 typeof 判断，Number/Boolean 需匹配类型
    if (defaults?.[key] != null) {
      const expectedType = typeof defaults[key];
      if (expectedType === 'number') {
        const num = Number(editValue);
        if (editValue.trim() === '' || isNaN(num)) {
          alert('值类型不匹配：该配置项期望数字类型');
          return;
        }
      } else if (expectedType === 'boolean') {
        if (editValue !== 'true' && editValue !== 'false') {
          alert('值类型不匹配：该配置项期望布尔类型 (true/false)');
          return;
        }
      }
    }

    // 核心配置二次确认
    const coreConfigs = ['review_mode', 'registration_enabled'];
    if (coreConfigs.includes(key)) {
      if (!window.confirm(`⚠️ 确定要修改核心配置 "${key}" 吗？此操作可能影响系统运行。`)) return;
    }

    try {
      await updateConfig(key, editValue);
      setEditingKey(null);
      setEditError('');
      await load();
    } catch { alert('保存失败'); }
  };

  const addWord = async () => {
    if (!newWord.trim()) return;
    try {
      await addSensitiveWord(newWord.trim());
      setNewWord('');
      await load();
    } catch { alert('添加失败'); }
  };

  const removeWord = async (id: number) => {
    if (!window.confirm('删除此敏感词？')) return;
    try {
      await deleteSensitiveWord(id);
      await load();
    } catch { alert('删除失败'); }
  };

  if (loading) return <div className="empty-state"><span className="emoji">⏳</span><p>加载中...</p></div>;
  if (!config) return <div className="empty-state"><span className="emoji">⚠️</span><p>加载失败</p></div>;

  const entries = Object.entries(config);

  return (
    <div style={{ display: 'flex', gap: 20, flexWrap: 'wrap' }}>
      {/* Config Keys */}
      <div className="card" style={{ flex: 2, minWidth: 400 }}>
        <h3>⚙️ 系统配置</h3>
        <table>
          <thead>
            <tr>
              <th>Key</th>
              <th>当前值</th>
              <th>默认值</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {entries.map(([key, value]) => (
              <tr key={key}>
                <td style={{ fontWeight: 600 }}>{key}</td>
                <td>
                  {editingKey === key ? (
                    <div>
                      <input
                        value={editValue}
                        onChange={e => { setEditValue(e.target.value); setEditError(''); }}
                        style={{ width: '100%', borderColor: editError ? 'var(--red)' : undefined }}
                        autoFocus
                      />
                      {editError && (
                        <div style={{ color: 'var(--red)', fontSize: 12, marginTop: 2 }}>{editError}</div>
                      )}
                    </div>
                  ) : (
                    <code style={{ fontSize: 12 }}>{typeof value === 'object' ? JSON.stringify(value) : String(value)}</code>
                  )}
                </td>
                <td>
                  <code className="text-sm" style={{ color: 'var(--text-secondary)' }}>
                    {defaults?.[key] != null ? String(defaults[key]) : '—'}
                  </code>
                </td>
                <td>
                  {editingKey === key ? (
                    <div style={{ display: 'flex', gap: 4 }}>
                      <button className="btn btn-green btn-sm" onClick={() => saveConfig(key)}>💾</button>
                      <button className="btn btn-outline btn-sm" onClick={() => setEditingKey(null)}>✕</button>
                    </div>
                  ) : (
                    <button className="btn btn-outline btn-sm" onClick={() => { setEditingKey(key); setEditValue(String(value)); setEditError(''); }}>
                      ✏️
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Sensitive Words */}
      <div className="card" style={{ flex: 1, minWidth: 300 }}>
        <h3>🚫 敏感词管理</h3>
        <div style={{ display: 'flex', gap: 6, marginBottom: 12 }}>
          <input
            value={newWord}
            onChange={e => setNewWord(e.target.value)}
            placeholder="输入敏感词..."
            onKeyDown={e => e.key === 'Enter' && addWord()}
            style={{ flex: 1 }}
          />
          <button className="btn btn-green btn-sm" onClick={addWord}>添加</button>
        </div>
        <table>
          <thead>
            <tr><th>ID</th><th>敏感词</th><th>操作</th></tr>
          </thead>
          <tbody>
            {words.map(w => (
              <tr key={w.id}>
                <td>{w.id}</td>
                <td style={{ color: 'var(--red)' }}>{w.word}</td>
                <td>
                  <button className="btn btn-outline btn-sm" onClick={() => removeWord(w.id)}>🗑</button>
                </td>
              </tr>
            ))}
            {words.length === 0 && (
              <tr><td colSpan={3} style={{ textAlign: 'center', color: 'var(--text-secondary)' }}>暂无敏感词</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
