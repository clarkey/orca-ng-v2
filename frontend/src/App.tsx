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
import Safes from '@/pages/Safes';
import UsersAndGroups from '@/pages/Users';
import Applications from '@/pages/Applications';
import AccessRoles from '@/pages/AccessRoles';
import SettingsGeneral from '@/pages/SettingsGeneral';
import SettingsSSO from '@/pages/SettingsSSO';
import SettingsNotifications from '@/pages/SettingsNotifications';
import SettingsDatabase from '@/pages/SettingsDatabase';
import SettingsAPI from '@/pages/SettingsAPI';

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
              <Route path="/safes" element={<Safes />} />
              <Route path="/users" element={<UsersAndGroups />} />
              <Route path="/applications" element={<Applications />} />
              <Route path="/operations" element={<Operations />} />
              <Route path="/operations/:id" element={<OperationDetail />} />
              <Route path="/pipeline" element={<PipelineDashboard />} />
              <Route path="/instances" element={<Instances />} />
              <Route path="/settings" element={<SettingsGeneral />} />
              <Route path="/settings/access-roles" element={<AccessRoles />} />
              <Route path="/settings/sso" element={<SettingsSSO />} />
              <Route path="/settings/notifications" element={<SettingsNotifications />} />
              <Route path="/settings/database" element={<SettingsDatabase />} />
              <Route path="/settings/api" element={<SettingsAPI />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </InstanceProvider>
      </AuthProvider>
    </Router>
  );
}

export default App;