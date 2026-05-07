import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import type { ReactNode } from 'react';
import Layout from './components/Layout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import ReviewQueue from './pages/ReviewQueue';
import ContentManagement from './pages/ContentManagement';
import UserManagement from './pages/UserManagement';
import PlatformContent from './pages/PlatformContent';
import DomainManagement from './pages/DomainManagement';
import SystemConfigPage from './pages/SystemConfig';
import AdminLogs from './pages/AdminLogs';
import Statistics from './pages/Statistics';

function ProtectedRoute({ children }: { children: ReactNode }) {
  const token = localStorage.getItem('admin_token');
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <BrowserRouter basename="/admin">
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/*"
          element={
            <ProtectedRoute>
              <Layout>
                <Routes>
                  <Route path="/" element={<Dashboard />} />
                  <Route path="/reviews" element={<ReviewQueue />} />
                  <Route path="/content" element={<ContentManagement />} />
                  <Route path="/users" element={<UserManagement />} />
                  <Route path="/platform" element={<PlatformContent />} />
                  <Route path="/domains" element={<DomainManagement />} />
                  <Route path="/config" element={<SystemConfigPage />} />
                  <Route path="/logs" element={<AdminLogs />} />
                  <Route path="/stats" element={<Statistics />} />
                </Routes>
              </Layout>
            </ProtectedRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}
