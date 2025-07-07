import { Link, useLocation, useNavigate, Outlet } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import { useInstance } from '@/contexts/InstanceContext';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { OrcaIcon } from '@/components/OrcaIcon';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  LayoutDashboard,
  Shield,
  Lock,
  Vault,
  Database,
  FolderLock,
  Key,
  Users,
  Layers,
  Settings,
  ListTodo,
  Server,
} from 'lucide-react';

export function Layout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { instances, currentInstanceId, setCurrentInstanceId, currentInstance } = useInstance();

  const navigation = {
    main: [
      { name: 'Dashboard', href: '/', icon: LayoutDashboard },
    ],
    cyberark: [
      { name: 'Safes', href: '/safes', icon: Vault },
      { name: 'Access Roles', href: '/access', icon: Key },
      { name: 'Users & Groups', href: '/users', icon: Users },
      { name: 'Applications', href: '/applications', icon: Layers },
    ],
    operations: [
      { name: 'Operations Queue', href: '/operations', icon: ListTodo },
      { name: 'Pipeline Monitor', href: '/pipeline', icon: Database },
    ],
    administration: [
      { name: 'Instances', href: '/instances', icon: Server },
      { name: 'Settings', href: '/settings', icon: Settings },
    ],
  };

  const handleLogout = async () => {
    await logout();
  };

  return (
    <div className="min-h-screen bg-white flex">
      {/* Sidebar */}
      <div className="w-64 bg-gray-50 border-r border-gray-200 flex flex-col">
        {/* Logo */}
        <div className="h-16 flex items-center px-6 border-b border-gray-200">
          <OrcaIcon className="h-10 w-10 text-gray-900" />
        </div>

        {/* Instance Selector */}
        <div className="px-4 py-4 border-b border-gray-200">
          <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
            CyberArk Instance
          </label>
          <Select value={currentInstanceId} onValueChange={setCurrentInstanceId}>
            <SelectTrigger className="w-full bg-white border-gray-300">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {instances.map((instance) => (
                <SelectItem key={instance.id} value={instance.id}>
                  <div className="flex items-center">
                    <div className={cn(
                      "w-2 h-2 rounded-full mr-2",
                      instance.type === 'overview' ? "bg-blue-500" : 
                      instance.status === 'connected' ? "bg-green-500" :
                      instance.status === 'disconnected' ? "bg-gray-400" : "bg-red-500"
                    )} />
                    {instance.name}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-4 py-4 space-y-6">
          {/* Main */}
          <div>
            <ul className="space-y-1">
              {navigation.main.map((item) => (
                <li key={item.name}>
                  <Link
                    to={item.href}
                    className={cn(
                      "flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors",
                      location.pathname === item.href
                        ? "bg-gray-200 text-gray-900"
                        : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                    )}
                  >
                    <item.icon className="mr-3 h-4 w-4" />
                    {item.name}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* CyberArk Management */}
          <div>
            <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
              CyberArk Management
            </h3>
            <ul className="space-y-1">
              {navigation.cyberark.map((item) => (
                <li key={item.name}>
                  <Link
                    to={item.href}
                    className={cn(
                      "flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors",
                      location.pathname === item.href
                        ? "bg-gray-200 text-gray-900"
                        : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                    )}
                  >
                    <item.icon className="mr-3 h-4 w-4" />
                    {item.name}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Operations & Processing */}
          <div>
            <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
              Operations & Processing
            </h3>
            <ul className="space-y-1">
              {navigation.operations.map((item) => (
                <li key={item.name}>
                  <Link
                    to={item.href}
                    className={cn(
                      "flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors",
                      location.pathname === item.href
                        ? "bg-gray-200 text-gray-900"
                        : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                    )}
                  >
                    <item.icon className="mr-3 h-4 w-4" />
                    {item.name}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Administration */}
          <div>
            <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
              Administration
            </h3>
            <ul className="space-y-1">
              {navigation.administration.map((item) => (
                <li key={item.name}>
                  <Link
                    to={item.href}
                    className={cn(
                      "flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors",
                      location.pathname === item.href
                        ? "bg-gray-200 text-gray-900"
                        : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                    )}
                  >
                    <item.icon className="mr-3 h-4 w-4" />
                    {item.name}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </nav>

      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col">
        {/* Top bar */}
        <header className="h-16 bg-white border-b border-gray-200 px-6 flex items-center justify-between">
          <h1 className="text-xl font-medium text-gray-900">
            {(() => {
              // Find current page in all navigation sections
              for (const section of Object.values(navigation)) {
                const item = section.find(item => item.href === location.pathname);
                if (item) return item.name;
              }
              return 'Dashboard';
            })()}
          </h1>
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="sm"
              className="text-gray-600 hover:text-gray-900 font-normal"
              onClick={() => navigate('/support')}
            >
              Support
            </Button>
            <div className="h-8 w-px bg-gray-200" />
            <div className="flex items-center gap-3">
              <div className="flex items-center">
                <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
                  <span className="text-sm font-medium text-gray-700">
                    {user?.username?.[0]?.toUpperCase()}
                  </span>
                </div>
                <span className="ml-2 text-sm font-medium text-gray-700">
                  {user?.username}
                </span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="text-gray-600 hover:text-gray-900"
                onClick={handleLogout}
              >
                Log out
              </Button>
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}