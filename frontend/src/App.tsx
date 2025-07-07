import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from '@/contexts/AuthContext';
import { InstanceProvider } from '@/contexts/InstanceContext';
import { PrivateRoute } from '@/components/PrivateRoute';
import { Layout } from '@/components/Layout';
import { Login } from '@/pages/Login';
import { Dashboard } from '@/pages/Dashboard';
import Operations from '@/pages/OperationsTable';
import OperationDetail from '@/pages/OperationDetail';
import PipelineDashboard from '@/pages/PipelineDashboard';
import Instances from '@/pages/Instances';

function App() {
  return (
    <Router>
      <AuthProvider>
        <InstanceProvider>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route
              element={
                <PrivateRoute>
                  <Layout />
                </PrivateRoute>
              }
            >
              <Route path="/" element={<Dashboard />} />
              <Route path="/operations" element={<Operations />} />
              <Route path="/operations/:id" element={<OperationDetail />} />
              <Route path="/pipeline" element={<PipelineDashboard />} />
              <Route path="/instances" element={<Instances />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </InstanceProvider>
      </AuthProvider>
    </Router>
  );
}

export default App;