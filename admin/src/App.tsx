import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import type { ReactNode } from 'react';
import Layout from './components/Layout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import ReviewQueue from './pages/ReviewQueue';
import ContentManagement from './pages/ContentManagement';
import UserManagement from './pages/UserManagement';
import PlatformContent from './pages/PlatformContent';

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
                </Routes>
              </Layout>
            </ProtectedRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}
