import { Link, useLocation, useNavigate, Outlet } from 'react-router-dom';
import { useState, useEffect, useRef } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { OrcaIcon } from '@/components/OrcaIcon';
import {
  LayoutDashboard,
  Vault,
  Users,
  Layers,
  Settings,
  ListTodo,
  Server,
  Activity,
  ArrowLeft,
  Key,
  Shield,
  Bell,
  Database,
  Globe,
  Search,
  LogOut,
} from 'lucide-react';

export function Layout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [showSettingsMenu, setShowSettingsMenu] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [isMac, setIsMac] = useState(true);
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Check if we're on a settings page
  useEffect(() => {
    setShowSettingsMenu(location.pathname.startsWith('/settings'));
  }, [location.pathname]);

  // Detect platform
  useEffect(() => {
    setIsMac(navigator.platform.toLowerCase().includes('mac'));
  }, []);

  // Keyboard shortcut for search (Cmd/Ctrl + K)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        searchInputRef.current?.focus();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const navigation = {
    main: [
      { name: 'Dashboard', href: '/', icon: LayoutDashboard },
    ],
    cyberark: [
      { name: 'Safes', href: '/safes', icon: Vault },
      { name: 'Users & Groups', href: '/users', icon: Users },
      { name: 'Applications', href: '/applications', icon: Layers },
    ],
    operations: [
      { name: 'Operations Queue', href: '/operations', icon: ListTodo },
      { name: 'Queue Monitoring', href: '/pipeline', icon: Activity },
    ],
    administration: [
      { name: 'Instances', href: '/instances', icon: Server },
      { name: 'Settings', href: '/settings', icon: Settings, isSettings: true },
    ],
  };

  const settingsNavigation = [
    { name: 'General', href: '/settings', icon: Settings },
    { name: 'Access Roles', href: '/settings/access-roles', icon: Key },
    { name: 'SSO Configuration', href: '/settings/sso', icon: Shield },
    { name: 'Notifications', href: '/settings/notifications', icon: Bell },
    { name: 'Database', href: '/settings/database', icon: Database },
    { name: 'API Settings', href: '/settings/api', icon: Globe },
  ];

  const handleLogout = async () => {
    await logout();
  };

  return (
    <div className="min-h-screen bg-white flex">
      {/* Sidebar */}
      <div className="fixed left-0 top-0 h-full w-64 bg-gray-50 border-r border-gray-200 flex flex-col z-10" style={{ pointerEvents: 'auto' }}>
        {/* Logo */}
        <div className="h-16 flex items-center px-6 border-b border-gray-200">
          <OrcaIcon className="h-10 w-10 text-gray-900" />
        </div>

        {/* Search */}
        <div className="px-4 py-4 border-b border-gray-200">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input
              ref={searchInputRef}
              type="text"
              placeholder="Search..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-10 py-2 text-sm border border-gray-300 rounded bg-white text-gray-900 placeholder:text-gray-400 outline-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-700 hover:border-gray-400"
              onKeyDown={(e) => {
                if (e.key === 'Enter' && searchQuery) {
                  navigate(`/search?q=${encodeURIComponent(searchQuery)}`);
                }
                if (e.key === 'Escape') {
                  setSearchQuery('');
                  searchInputRef.current?.blur();
                }
              }}
            />
            <kbd className="absolute right-2 top-1/2 transform -translate-y-1/2 hidden sm:inline-flex h-5 select-none items-center gap-1 rounded border border-gray-200 bg-gray-100 px-1.5 font-mono text-[10px] font-medium text-gray-600">
              <span className="text-xs">{isMac ? '⌘' : 'Ctrl+'}</span>K
            </kbd>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-4 py-4 overflow-y-auto [&_a]:!cursor-pointer [&_button]:!cursor-pointer" style={{ pointerEvents: 'auto' }}>
          <div className="relative h-full overflow-hidden">
            {/* Main Navigation */}
            <div className={cn(
              "absolute inset-0 space-y-6 transition-transform duration-300 ease-in-out",
              showSettingsMenu ? "-translate-x-full pointer-events-none" : "translate-x-0"
            )}>
          {/* Main */}
          <div>
            <ul className="space-y-1">
              {navigation.main.map((item) => (
                <li key={item.name} className="cursor-pointer">
                  <Link
                    to={item.href}
                    style={{ cursor: 'pointer' }}
                    className={cn(
                      "flex items-center px-3 py-2 text-sm font-medium rounded ",
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
                    style={{ cursor: 'pointer' }}
                    className={cn(
                      "flex items-center px-3 py-2 text-sm font-medium rounded ",
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
                    style={{ cursor: 'pointer' }}
                    className={cn(
                      "flex items-center px-3 py-2 text-sm font-medium rounded ",
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
                  {item.isSettings ? (
                    <button
                      type="button"
                      onClick={() => {
                        navigate('/settings');
                        setShowSettingsMenu(true);
                      }}
                      style={{ cursor: 'pointer' }}
                      className={cn(
                        "w-full flex items-center px-3 py-2 text-sm font-medium rounded  hover:cursor-pointer",
                        location.pathname.startsWith('/settings')
                          ? "bg-gray-200 text-gray-900"
                          : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                      )}
                    >
                      <item.icon className="mr-3 h-4 w-4" />
                      {item.name}
                    </button>
                  ) : (
                    <Link
                      to={item.href}
                      className={cn(
                        "flex items-center px-3 py-2 text-sm font-medium rounded  cursor-pointer",
                        location.pathname === item.href
                          ? "bg-gray-200 text-gray-900"
                          : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                      )}
                    >
                      <item.icon className="mr-3 h-4 w-4" />
                      {item.name}
                    </Link>
                  )}
                </li>
              ))}
            </ul>
          </div>
            </div>

            {/* Settings Navigation */}
            <div className={cn(
              "absolute inset-0 space-y-6 transition-transform duration-300 ease-in-out",
              showSettingsMenu ? "translate-x-0" : "translate-x-full pointer-events-none"
            )}>
              {/* Back Button */}
              <div>
                <button
                  onClick={() => {
                    setShowSettingsMenu(false);
                    navigate('/');
                  }}
                  style={{ cursor: 'pointer' }}
                  className="flex items-center px-3 py-2 text-sm font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded  w-full"
                >
                  <ArrowLeft className="mr-3 h-4 w-4" />
                  Back to Main Menu
                </button>
              </div>

              {/* Settings Items */}
              <div>
                <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
                  Settings
                </h3>
                <ul className="space-y-1">
                  {settingsNavigation.map((item) => (
                    <li key={item.name}>
                      <Link
                        to={item.href}
                        style={{ cursor: 'pointer' }}
                        className={cn(
                          "flex items-center px-3 py-2 text-sm font-medium rounded ",
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
            </div>
          </div>
        </nav>

        {/* User section at bottom */}
        <div className="border-t border-gray-200">
          <div className="px-4 py-4">
            <div className="flex items-center justify-between">
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
                size="icon"
                className="text-gray-600 hover:text-gray-900"
                onClick={handleLogout}
                title="Log out"
              >
                <LogOut className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 ml-64">
        <main className="h-full overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}