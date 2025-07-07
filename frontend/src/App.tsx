import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { AuthProvider } from '@/contexts/AuthContext';
import { InstanceProvider } from '@/contexts/InstanceContext';
import { PrivateRoute } from '@/components/PrivateRoute';
import { QueryErrorBoundary } from '@/components/QueryErrorBoundary';
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

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: (failureCount, error: any) => {
        // Don't retry on 4xx errors
        if (error?.response?.status >= 400 && error?.response?.status < 500) {
          return false;
        }
        return failureCount < 3;
      },
      staleTime: 5 * 60 * 1000, // 5 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
    },
    mutations: {
      onError: (error: any) => {
        // Global mutation error handler
        console.error('Mutation error:', error);
        // You could show a toast notification here
      },
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <QueryErrorBoundary>
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
      </QueryErrorBoundary>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}

export default App;