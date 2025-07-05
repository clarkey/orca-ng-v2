import { useAuth } from '@/contexts/AuthContext';
import { useInstance } from '@/contexts/InstanceContext';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Layout } from '@/components/Layout';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

export function Dashboard() {
  const { user } = useAuth();
  const { currentInstance, isOverviewMode } = useInstance();

  return (
    <Layout>
      <div className="p-6">
        <div className="max-w-7xl mx-auto">
          {/* Instance Header */}
          {!isOverviewMode && currentInstance && (
            <div className="mb-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-medium text-gray-900">{currentInstance.name}</h2>
                  <p className="text-sm text-gray-600">{currentInstance.url}</p>
                </div>
                <div className="flex items-center gap-2">
                  <div className={cn(
                    "w-3 h-3 rounded-full",
                    currentInstance.status === 'connected' ? "bg-green-500" :
                    currentInstance.status === 'disconnected' ? "bg-gray-400" : "bg-red-500"
                  )} />
                  <span className="text-sm text-gray-600 capitalize">{currentInstance.status}</span>
                </div>
              </div>
            </div>
          )}

          {/* Overview Mode Content */}
          {isOverviewMode && (
            <div className="mb-8">
              <h2 className="text-2xl font-semibold text-gray-900 mb-4">CyberArk Overview</h2>
              <div className="grid gap-4 md:grid-cols-3">
                <Card className="border-gray-200">
                  <CardHeader>
                    <CardTitle className="text-gray-900">Total Safes</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-3xl font-semibold text-gray-900">247</p>
                    <p className="text-sm text-gray-600">Across all instances</p>
                  </CardContent>
                </Card>
                <Card className="border-gray-200">
                  <CardHeader>
                    <CardTitle className="text-gray-900">Active Users</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-3xl font-semibold text-gray-900">1,842</p>
                    <p className="text-sm text-gray-600">Unique users</p>
                  </CardContent>
                </Card>
                <Card className="border-gray-200">
                  <CardHeader>
                    <CardTitle className="text-gray-900">Access Roles</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-3xl font-semibold text-gray-900">89</p>
                    <p className="text-sm text-gray-600">Defined roles</p>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            <Card className="border-gray-200">
              <CardHeader>
                <CardTitle className="text-gray-900">CyberArk Instances</CardTitle>
                <CardDescription className="text-gray-600">
                  Manage connected CyberArk environments
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-gray-500">
                  No instances configured yet
                </p>
              </CardContent>
            </Card>

            <Card className="border-gray-200">
              <CardHeader>
                <CardTitle className="text-gray-900">Safes</CardTitle>
                <CardDescription className="text-gray-600">
                  View and manage CyberArk safes
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-gray-500">
                  Connect a CyberArk instance to view safes
                </p>
              </CardContent>
            </Card>

            <Card className="border-gray-200">
              <CardHeader>
                <CardTitle className="text-gray-900">Access Roles</CardTitle>
                <CardDescription className="text-gray-600">
                  Configure logical safe access roles
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-gray-500">
                  No access roles defined yet
                </p>
              </CardContent>
            </Card>
          </div>

          {user?.is_admin && (
            <div className="mt-8">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">Administration</h2>
              <div className="grid gap-4 md:grid-cols-2">
                <Card className="border-gray-200">
                  <CardHeader>
                    <CardTitle className="text-gray-900">Users</CardTitle>
                    <CardDescription className="text-gray-600">
                      Manage ORCA users and permissions
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button 
                      variant="outline" 
                      size="sm"
                      className="border-gray-300 text-gray-700 hover:bg-gray-50"
                    >
                      Manage Users
                    </Button>
                  </CardContent>
                </Card>

                <Card className="border-gray-200">
                  <CardHeader>
                    <CardTitle className="text-gray-900">System Settings</CardTitle>
                    <CardDescription className="text-gray-600">
                      Configure system-wide settings
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button 
                      variant="outline" 
                      size="sm"
                      className="border-gray-300 text-gray-700 hover:bg-gray-50"
                    >
                      View Settings
                    </Button>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}