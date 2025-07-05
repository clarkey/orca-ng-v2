import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from '@/contexts/AuthContext';
import { InstanceProvider } from '@/contexts/InstanceContext';
import { PrivateRoute } from '@/components/PrivateRoute';
import { Login } from '@/pages/Login';
import { Dashboard } from '@/pages/Dashboard';

function App() {
  return (
    <Router>
      <AuthProvider>
        <InstanceProvider>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route
              path="/"
              element={
                <PrivateRoute>
                  <Dashboard />
                </PrivateRoute>
              }
            />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </InstanceProvider>
      </AuthProvider>
    </Router>
  );
}

export default App;