import { ReactNode } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
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
} from 'lucide-react';

interface LayoutProps {
  children: ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { instances, currentInstanceId, setCurrentInstanceId, currentInstance } = useInstance();

  const navigation = [
    { name: 'Dashboard', href: '/', icon: LayoutDashboard },
    { name: 'Safes', href: '/safes', icon: Vault },
    { name: 'Access', href: '/access', icon: Key },
    { name: 'Users', href: '/users', icon: Users },
    { name: 'Applications', href: '/applications', icon: Layers },
    { name: 'Settings', href: '/settings', icon: Settings },
  ];

  const handleLogout = async () => {
    await logout();
  };

  return (
    <div className="min-h-screen bg-white flex">
      {/* Sidebar */}
      <div className="w-64 bg-gray-50 border-r border-gray-200 flex flex-col">
        {/* Logo */}
        <div className="h-16 flex items-center px-6 border-b border-gray-200">
          <OrcaIcon className="h-8 w-8 text-gray-900" />
          <span className="ml-3 text-xl font-semibold text-gray-900">ORCA</span>
        </div>

        {/* Instance Selector */}
        <div className="px-4 py-4 border-b border-gray-200">
          <label className="block text-xs font-medium text-gray-600 mb-2">
            CYBERARK INSTANCE
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
        <nav className="flex-1 px-4 py-4">
          <ul className="space-y-1">
            {navigation.map((item) => (
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
        </nav>

        {/* User section */}
        <div className="p-4 border-t border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
                <span className="text-sm font-medium text-gray-700">
                  {user?.username?.[0]?.toUpperCase()}
                </span>
              </div>
              <span className="ml-3 text-sm font-medium text-gray-700">
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
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col">
        {/* Top bar */}
        <header className="h-16 bg-white border-b border-gray-200 px-6 flex items-center justify-between">
          <h1 className="text-xl font-medium text-gray-900">
            {navigation.find(item => item.href === location.pathname)?.name || 'Dashboard'}
          </h1>
          <Button
            variant="ghost"
            size="sm"
            className="text-gray-600 hover:text-gray-900 font-normal"
            onClick={() => navigate('/support')}
          >
            Support
          </Button>
        </header>

        {/* Page content */}
        <main className="flex-1 overflow-auto">
          {children}
        </main>
      </div>
    </div>
  );
}