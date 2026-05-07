import { useEffect, useState } from 'react';
import {
  fetchDomains, fetchDomainStats, updateDomain, toggleDomainActive,
  createDomain, addSubDomain, reorderDomains,
} from '../api/endpoints';
import type { DomainItem } from '../api/endpoints';

export default function DomainManagement() {
  const [domains, setDomains] = useState<DomainItem[]>([]);
  const [stats, setStats] = useState<{ domain: string; count: number }[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<DomainItem | null>(null);
  const [editForm, setEditForm] = useState({ display_name: '', icon: '' });
  const [creating, setCreating] = useState(false);
  const [createForm, setCreateForm] = useState({ name: '', display_name: '', icon: '' });
  const [addSub, setAddSub] = useState<{ parent: string; name: string; display_name: string } | null>(null);
  const [dragIdx, setDragIdx] = useState<number | null>(null);

  const load = async () => {
    setLoading(true);
    try {
      const [d, s] = await Promise.all([fetchDomains(), fetchDomainStats()]);
      setDomains(d.domains || []);
      setStats(s.stats || []);
    } catch { /* ignore */ }
    setLoading(false);
  };

  useEffect(() => { load(); }, []);

  const getCount = (name: string) => stats.find(s => s.domain === name)?.count ?? 0;

  const handleEdit = (d: DomainItem) => {
    setEditing(d);
    setEditForm({ display_name: d.display_name, icon: d.icon });
  };

  const saveEdit = async () => {
    if (!editing) return;
    try {
      await updateDomain(editing.name, editForm);
      setEditing(null);
      await load();
    } catch { alert('保存失败'); }
  };

  const handleToggle = async (name: string, active: boolean) => {
    try {
      await toggleDomainActive(name, !active);
      await load();
    } catch { alert('操作失败'); }
  };

  const handleReorder = async () => {
    const names = domains.map(d => d.name);
    try {
      await reorderDomains({ names });
      await load();
    } catch { alert('排序失败'); }
  };

  // Drag-and-drop handlers
  const handleDragStart = (idx: number) => {
    setDragIdx(idx);
  };

  const handleDragOver = (e: React.DragEvent, idx: number) => {
    e.preventDefault();
    if (dragIdx === null || dragIdx === idx) return;
    const reordered = [...domains];
    const [moved] = reordered.splice(dragIdx, 1);
    reordered.splice(idx, 0, moved);
    setDomains(reordered);
    setDragIdx(idx);
  };

  const handleDragEnd = () => {
    if (dragIdx !== null) {
      const names = domains.map(d => d.name);
      reorderDomains({ names }).catch(() => {}).finally(() => load());
    }
    setDragIdx(null);
  };

  const saveCreate = async () => {
    if (!createForm.name.trim() || !createForm.display_name.trim()) return;
    try {
      await createDomain({
        name: createForm.name.trim().toLowerCase(),
        display_name: createForm.display_name.trim(),
        icon: createForm.icon.trim() || 'folder',
      });
      setCreating(false);
      setCreateForm({ name: '', display_name: '', icon: '' });
      await load();
    } catch { alert('创建失败'); }
  };

  const saveSub = async () => {
    if (!addSub || !addSub.name.trim()) return;
    try {
      await addSubDomain(addSub.parent, { name: addSub.name.trim(), display_name: addSub.display_name.trim() });
      setAddSub(null);
      await load();
    } catch { alert('添加子领域失败'); }
  };

  if (loading) return <div className="empty-state"><span className="emoji">⏳</span><p>加载中...</p></div>;

  return (
    <>
      <div className="toolbar">
        <button className="btn btn-green btn-sm" onClick={() => setCreating(true)}>＋ 新建领域</button>
        <button className="btn btn-outline btn-sm" onClick={handleReorder}>🔄 保存排序</button>
      </div>

      <div className="card">
        <table>
          <thead>
            <tr>
              <th>排序</th>
              <th>名称</th>
              <th>显示名</th>
              <th>图标</th>
              <th>子领域</th>
              <th>经验数</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {domains.map((d, idx) => (
              <tr
                key={d.name}
                draggable
                onDragStart={() => handleDragStart(idx)}
                onDragOver={(e) => handleDragOver(e, idx)}
                onDragEnd={handleDragEnd}
                style={{
                  cursor: 'grab',
                  opacity: dragIdx === idx ? 0.5 : 1,
                  borderLeft: dragIdx === idx ? '3px solid var(--green)' : undefined,
                }}
              >
                <td>{d.sort_order}</td>
                <td style={{ fontWeight: 600 }}>{d.name}</td>
                <td>{d.display_name}</td>
                <td>{d.icon}</td>
                <td>
                  {d.sub_domains?.map(s => (
                    <span key={s.name} className="tag">{s.display_name || s.name}</span>
                  ))}
                  <button
                    className="btn btn-outline btn-sm"
                    style={{ fontSize: 11, padding: '2px 6px' }}
                    onClick={() => setAddSub({ parent: d.name, name: '', display_name: '' })}
                  >＋</button>
                </td>
                <td>{getCount(d.name)}</td>
                <td>
                  <span className={`badge ${d.active ? 'badge-approved' : 'badge-rejected'}`}>
                    {d.active ? '启用' : '禁用'}
                  </span>
                </td>
                <td style={{ display: 'flex', gap: 4 }}>
                  <button className="btn btn-outline btn-sm" onClick={() => handleEdit(d)}>✏️</button>
                  <button className="btn btn-outline btn-sm" onClick={() => handleToggle(d.name, d.active)}>
                    {d.active ? '禁用' : '启用'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Edit Modal */}
      {editing && (
        <div className="modal-overlay" onClick={() => setEditing(null)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <h3>编辑领域: {editing.name}</h3>
            <div className="field">
              <label>显示名</label>
              <input value={editForm.display_name} onChange={e => setEditForm({ ...editForm, display_name: e.target.value })} />
            </div>
            <div className="field">
              <label>图标</label>
              <input value={editForm.icon} onChange={e => setEditForm({ ...editForm, icon: e.target.value })} />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 16 }}>
              <button className="btn btn-outline btn-sm" onClick={() => setEditing(null)}>取消</button>
              <button className="btn btn-green btn-sm" onClick={saveEdit}>保存</button>
            </div>
          </div>
        </div>
      )}

      {/* Create Modal */}
      {creating && (
        <div className="modal-overlay" onClick={() => setCreating(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <h3>新建领域</h3>
            <div className="field">
              <label>英文名 (key)</label>
              <input value={createForm.name} onChange={e => setCreateForm({ ...createForm, name: e.target.value })} placeholder="如: health" />
            </div>
            <div className="field">
              <label>显示名</label>
              <input value={createForm.display_name} onChange={e => setCreateForm({ ...createForm, display_name: e.target.value })} placeholder="如: 健康" />
            </div>
            <div className="field">
              <label>图标</label>
              <input value={createForm.icon} onChange={e => setCreateForm({ ...createForm, icon: e.target.value })} placeholder="默认: folder" />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 16 }}>
              <button className="btn btn-outline btn-sm" onClick={() => setCreating(false)}>取消</button>
              <button className="btn btn-green btn-sm" onClick={saveCreate}>创建</button>
            </div>
          </div>
        </div>
      )}

      {/* Add Sub Modal */}
      {addSub && (
        <div className="modal-overlay" onClick={() => setAddSub(null)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <h3>添加子领域 → {addSub.parent}</h3>
            <div className="field">
              <label>英文名 (key)</label>
              <input value={addSub.name} onChange={e => setAddSub({ ...addSub, name: e.target.value })} placeholder="如: sleep" />
            </div>
            <div className="field">
              <label>显示名</label>
              <input value={addSub.display_name} onChange={e => setAddSub({ ...addSub, display_name: e.target.value })} placeholder="如: 睡眠" />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 16 }}>
              <button className="btn btn-outline btn-sm" onClick={() => setAddSub(null)}>取消</button>
              <button className="btn btn-green btn-sm" onClick={saveSub}>添加</button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
