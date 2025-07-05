import { ReactNode } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import { useInstance } from '@/contexts/InstanceContext';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
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
          <svg 
            className="h-8 w-8 text-gray-900" 
            viewBox="0 0 572.958 572.958" 
            fill="currentColor"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path d="M519.749,295.946c-4.684-1.567-9.47-5.386-12.321-9.474c-28.014-40.184-59.364-77.014-100.768-104.244c-30.131-19.816-63.681-28.997-99.132-32.457c-33.333-3.251-66.41-0.539-99.381,4.402c-19.959,2.991-29.539-1.873-38.523-20.004c-9.935-20.045-13.138-41.755-15.177-63.689c-0.918-9.857-3.77-11.779-12.367-6.863c-15.381,8.797-21.759,23.607-22.13,40.013c-0.53,23.26,0.44,46.639,2.301,69.841c1.269,15.802-1.016,28.756-11.518,41.861c-19.188,23.938-37.997,48.246-54.358,74.199c-1.791-1.248-3.586-2.497-5.377-3.745c-2.832-3.708-6.006-7.687-8.78-11.934c-3.725-5.7-7.124-6.72-12.199-1.289c-15.37,16.462-21.028,48.343-13.342,70.803c0.563,4.578,2.175,9.013,4.831,12.787c-10.22,23.354-18.311,47.459-21.077,73.269c-1.269,11.84,0.596,24.044,1.367,36.063c0.441,6.821,1.901,13.578,2.326,20.396c0.824,13.284,16.503,27.932,29.576,25.863c8.482-1.343,17.997-3.269,24.586-8.209c26.553-19.914,56.798-33.235,85.129-49.919c13.317-7.842,25.541-17.083,36.679-28.241c5.207-3.971,9.723-8.834,13.269-14.342c47.964,4.537,100-13.64,127.606-46.088c5.92-6.96,10.559-15.39,14.162-23.835c3.369-7.903,0.383-11.126-8.246-11.179c-14.827-0.09-29.657-0.025-44.484-0.025v1.058c-6.609,0-13.219,0.024-19.829,0.036c6.1-2.027,11.999-4.479,17.572-7.674c7.397-2.999,14.272-7.687,21.143-13.791c14.798-13.149,31.808-24.826,49.678-33.292c20.355-9.641,40.159-11.62,59.237-8.054c3.484,1.098,7.002,2.371,10.571,3.831c2.521,1.032,5.129,1.42,7.716,1.293c11.493,4.292,22.685,10.441,33.533,17.968c1.991,1.384,3.925,2.853,5.818,4.378c-0.144,1.359-0.286,2.718-0.429,4.076c-0.11,0.053-0.225,0.11-0.335,0.163c-1.905,12.191-11.399,18.675-18.964,26.585c-9.755,10.2-18.307,20.976-14.843,36.361c1.122,4.989,3.558,9.739,5.806,14.395c0.702,1.448,2.864,3.202,4.17,3.063s3.166-2.342,3.428-3.872c1.938-11.31,11.729-10.771,19.339-14.716c7.055-3.66,14.117-2.632,20.515-5.047c3.725-1.408,5.744-7.328,8.616-11.653c0.192,0.151,0.384,0.307,0.575,0.449c4.031,10.44,13.652,12.342,23.771,10.563c15.275-2.689,27.169,4.521,38.977,11.722c4.234,2.582,6.76,7.972,11.647,14.043c1.302-5.451,2.836-8.902,2.845-12.358C573.059,332.368,554.805,307.684,519.749,295.946z"/>
          </svg>
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