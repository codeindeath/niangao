import { NavLink, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import type { ReactNode } from 'react';
import { fetchDashboard } from '../api/endpoints';
import type { Dashboard } from '../api/endpoints';

function getPageTitle(pathname: string): string {
  if (pathname === '/' || pathname === '') return '📊 总览仪表盘';
  if (pathname.startsWith('/reviews')) return '🔍 内容审核';
  if (pathname.startsWith('/content')) return '📝 内容管理';
  if (pathname.startsWith('/users')) return '👥 用户管理';
  if (pathname.startsWith('/platform')) return '📋 平台内容';
  return '管理后台';
}

export default function Layout({ children }: { children: ReactNode }) {
  const navigate = useNavigate();
  const [pendingCount, setPendingCount] = useState<number | null>(null);

  useEffect(() => {
    fetchDashboard()
      .then((d: Dashboard) => setPendingCount(d.pending_reviews))
      .catch(() => {});
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('admin_token');
    navigate('/login');
  };

  return (
    <>
      {/* Sidebar */}
      <aside className="sidebar">
        <div className="sidebar-logo">
          🍡 <span>年糕</span> 管理后台
        </div>
        <nav>
          <NavLink
            to="/"
            end
            className={({ isActive }) => isActive ? 'active' : ''}
          >
            <span className="icon">📊</span> 总览
          </NavLink>
          <NavLink
            to="/reviews"
            className={({ isActive }) => isActive ? 'active' : ''}
          >
            <span className="icon">🔍</span> 内容审核
            {pendingCount != null && pendingCount > 0 && (
              <span className="badge">{pendingCount}</span>
            )}
          </NavLink>
          <NavLink
            to="/content"
            className={({ isActive }) => isActive ? 'active' : ''}
          >
            <span className="icon">📝</span> 内容管理
          </NavLink>
          <NavLink
            to="/platform"
            className={({ isActive }) => isActive ? 'active' : ''}
          >
            <span className="icon">📋</span> 平台内容
          </NavLink>
          <NavLink
            to="/users"
            className={({ isActive }) => isActive ? 'active' : ''}
          >
            <span className="icon">👥</span> 用户管理
          </NavLink>
        </nav>
        <div className="sidebar-footer">
          <div>v1.0 · admin@niangao</div>
          <button
            onClick={handleLogout}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--sidebar-text)',
              cursor: 'pointer',
              fontSize: 11,
              marginTop: 6,
              fontFamily: 'inherit',
              padding: 0,
            }}
          >
            退出登录
          </button>
        </div>
      </aside>

      {/* Main */}
      <div className="main">
        <div className="topbar">
          <h1>
            {typeof window !== 'undefined'
              ? getPageTitle(window.location.pathname)
              : '管理后台'}
          </h1>
          <div className="topbar-right">
            <span className="dot"></span> 服务正常
          </div>
        </div>
        <div className="content">{children}</div>
      </div>
    </>
  );
}
