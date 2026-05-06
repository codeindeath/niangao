import { useState } from 'react';
import type { FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminLogin } from '../api/endpoints';

export default function Login() {
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!username.trim() || !password) {
      setError('请输入用户名和密码');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const { token } = await adminLogin(username.trim(), password);
      localStorage.setItem('admin_token', token);
      navigate('/', { replace: true });
    } catch (err: unknown) {
      const status = (err as { status?: number }).status;
      if (status === 401 || status === 403) {
        setError('用户名或密码错误，或该账号无管理员权限');
      } else {
        setError('登录失败，请稍后重试');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      <form className="login-card" onSubmit={handleSubmit}>
        <h1>
          🍡 <span>年糕</span> 管理后台
        </h1>
        <p className="subtitle">请输入管理员账号登录</p>

        {error && <div className="error">{error}</div>}

        <div className="field">
          <label>用户名</label>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="输入管理员用户名"
            autoFocus
            disabled={loading}
          />
        </div>

        <div className="field">
          <label>密码</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="输入密码"
            disabled={loading}
          />
        </div>

        <button
          type="submit"
          className="btn btn-green"
          disabled={loading}
        >
          {loading ? '登录中...' : '登录'}
        </button>
      </form>
    </div>
  );
}
