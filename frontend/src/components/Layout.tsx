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
  RefreshCw,
  Globe,
  Search,
  LogOut,
  ChevronLeft,
  ChevronRight,
  ShieldCheck,
} from 'lucide-react';

const STORAGE_KEY_SIDEBAR_COLLAPSED = 'orca-sidebar-collapsed';

interface NavItemProps {
  item: {
    name: string;
    href: string;
    icon: React.ComponentType<{ className?: string }>;
    isSettings?: boolean;
  };
  isActive: boolean;
  isCollapsed: boolean;
  onClick?: () => void;
}

function NavItem({ item, isActive, isCollapsed, onClick }: NavItemProps) {
  const className = cn(
    "flex items-center px-3 py-2 text-sm font-medium rounded whitespace-nowrap",
    isActive
      ? "bg-gray-200 text-gray-900"
      : "text-gray-600 hover:bg-gray-100 hover:text-gray-900",
    isCollapsed && "justify-center"
  );

  const content = (
    <>
      <item.icon className={cn("h-4 w-4 flex-shrink-0", !isCollapsed && "mr-3")} />
      {!isCollapsed && item.name}
    </>
  );

  if (item.isSettings) {
    return (
      <button
        type="button"
        onClick={onClick}
        style={{ cursor: 'pointer' }}
        title={isCollapsed ? item.name : undefined}
        className={cn(className, "w-full hover:cursor-pointer")}
      >
        {content}
      </button>
    );
  }

  return (
    <Link
      to={item.href}
      style={{ cursor: 'pointer' }}
      title={isCollapsed ? item.name : undefined}
      className={cn(className, "cursor-pointer")}
    >
      {content}
    </Link>
  );
}

export function Layout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [showSettingsMenu, setShowSettingsMenu] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [isMac, setIsMac] = useState(true);
  const [isCollapsed, setIsCollapsed] = useState(() => {
    const saved = localStorage.getItem(STORAGE_KEY_SIDEBAR_COLLAPSED);
    return saved === 'true';
  });
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Check if we're on a settings page
  useEffect(() => {
    setShowSettingsMenu(location.pathname.startsWith('/settings'));
  }, [location.pathname]);

  // Detect platform
  useEffect(() => {
    setIsMac(navigator.platform.toLowerCase().includes('mac'));
  }, []);

  // Save collapsed state to localStorage
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY_SIDEBAR_COLLAPSED, isCollapsed.toString());
  }, [isCollapsed]);

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
      { name: 'Activity', href: '/activity', icon: Activity },
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
    { name: 'Certificate Authorities', href: '/settings/certificates', icon: ShieldCheck },
    { name: 'SSO Configuration', href: '/settings/sso', icon: Shield },
    { name: 'Notifications', href: '/settings/notifications', icon: Bell },
    { name: 'Database', href: '/settings/database', icon: Database },
    { name: 'API Settings', href: '/settings/api', icon: Globe },
  ];

  const handleLogout = async () => {
    await logout();
  };

  return (
    <div className="min-h-screen bg-white">
      {/* Sidebar */}
      <div className={cn(
        "fixed left-0 top-0 bottom-0 bg-gray-50 border-r border-gray-200 flex flex-col z-10 transition-all duration-300",
        isCollapsed ? "w-16" : "w-64"
      )} style={{ pointerEvents: 'auto' }}>
        {/* Logo */}
        <div className="h-16 flex items-center justify-between px-4 border-b border-gray-200">
          <div className={cn("flex items-center", isCollapsed && "justify-center w-full")}>
            <OrcaIcon className="h-10 w-10 text-gray-900" />
          </div>
          {!isCollapsed ? (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setIsCollapsed(true)}
              className="text-gray-600 hover:text-gray-900"
              title="Collapse sidebar"
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
          ) : (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setIsCollapsed(false)}
              className="absolute right-0 top-4 translate-x-1/2 bg-gray-50 border border-gray-200 shadow-sm hover:bg-gray-100 z-20"
              title="Expand sidebar"
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          )}
        </div>

        {/* Search */}
        {!isCollapsed ? (
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
        ) : (
          <div className="px-4 py-4 border-b border-gray-200 flex justify-center">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => {
                setIsCollapsed(false);
                setTimeout(() => searchInputRef.current?.focus(), 100);
              }}
              className="text-gray-600 hover:text-gray-900"
              title="Search"
            >
              <Search className="h-4 w-4" />
            </Button>
          </div>
        )}

        {/* Navigation */}
        <nav className={cn("flex-1 py-4 overflow-y-auto [&_a]:!cursor-pointer [&_button]:!cursor-pointer", isCollapsed ? "px-2" : "px-4")} style={{ pointerEvents: 'auto' }}>
          <div className="relative h-full overflow-x-hidden">
            {/* Main Navigation */}
            <div className={cn(
              "absolute inset-0 space-y-6 transition-transform duration-300 ease-in-out overflow-y-auto",
              showSettingsMenu ? "-translate-x-full" : "translate-x-0"
            )}>
          {/* Main */}
          <div>
            <ul className="space-y-1">
              {navigation.main.map((item) => (
                <li key={item.name} className="cursor-pointer">
                  <NavItem
                    item={item}
                    isActive={location.pathname === item.href}
                    isCollapsed={isCollapsed}
                  />
                </li>
              ))}
            </ul>
          </div>

          {/* CyberArk Management */}
          <div>
            <h3 className={cn(
              "h-5 flex items-center px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 whitespace-nowrap overflow-hidden",
              isCollapsed && "opacity-0 select-none pointer-events-none"
            )}>
              CyberArk Management
            </h3>
            <ul className="space-y-1">
              {navigation.cyberark.map((item) => (
                <li key={item.name}>
                  <NavItem
                    item={item}
                    isActive={location.pathname === item.href}
                    isCollapsed={isCollapsed}
                  />
                </li>
              ))}
            </ul>
          </div>

          {/* Operations & Processing */}
          <div>
            <h3 className={cn(
              "h-5 flex items-center px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 whitespace-nowrap overflow-hidden",
              isCollapsed && "opacity-0 select-none pointer-events-none"
            )}>
              Operations & Processing
            </h3>
            <ul className="space-y-1">
              {navigation.operations.map((item) => (
                <li key={item.name}>
                  <NavItem
                    item={item}
                    isActive={location.pathname === item.href}
                    isCollapsed={isCollapsed}
                  />
                </li>
              ))}
            </ul>
          </div>

          {/* Administration */}
          <div>
            <h3 className={cn(
              "h-5 flex items-center px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 whitespace-nowrap overflow-hidden",
              isCollapsed && "opacity-0 select-none pointer-events-none"
            )}>
              Administration
            </h3>
            <ul className="space-y-1">
              {navigation.administration.map((item) => (
                <li key={item.name}>
                  <NavItem
                    item={item}
                    isActive={item.isSettings ? location.pathname.startsWith('/settings') : location.pathname === item.href}
                    isCollapsed={isCollapsed}
                    onClick={item.isSettings ? () => {
                      navigate('/settings');
                      setShowSettingsMenu(true);
                    } : undefined}
                  />
                </li>
              ))}
            </ul>
          </div>
            </div>

            {/* Settings Navigation */}
            <div className={cn(
              "absolute inset-0 space-y-6 transition-transform duration-300 ease-in-out overflow-y-auto",
              showSettingsMenu ? "translate-x-0" : "translate-x-full"
            )}>
              {/* Back Button */}
              <div>
                <button
                  onClick={() => {
                    setShowSettingsMenu(false);
                    navigate('/');
                  }}
                  style={{ cursor: 'pointer' }}
                  title={isCollapsed ? "Back to Main Menu" : undefined}
                  className={cn(
                    "flex items-center px-3 py-2 text-sm font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded whitespace-nowrap w-full",
                    isCollapsed && "justify-center"
                  )}
                >
                  <ArrowLeft className={cn("h-4 w-4 flex-shrink-0", !isCollapsed && "mr-3")} />
                  {!isCollapsed && "Back to Main Menu"}
                </button>
              </div>

              {/* Settings Items */}
              <div>
                <h3 className={cn(
                  "px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 whitespace-nowrap overflow-hidden",
                  isCollapsed && "opacity-0 select-none pointer-events-none"
                )}>
                  Settings
                </h3>
                <ul className="space-y-1">
                  {settingsNavigation.map((item) => (
                    <li key={item.name}>
                      <NavItem
                        item={item}
                        isActive={location.pathname === item.href}
                        isCollapsed={isCollapsed}
                      />
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </nav>

        {/* User section at bottom */}
        <div className="border-t border-gray-200">
          <div className={cn("px-4 py-4", isCollapsed && "px-2")}>
            {isCollapsed ? (
              <div className="flex flex-col items-center gap-2">
                <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
                  <span className="text-sm font-medium text-gray-700">
                    {user?.username?.[0]?.toUpperCase()}
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
            ) : (
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
            )}
          </div>
        </div>
      </div>

      {/* Main content */}
      <main className={cn("min-h-screen transition-all duration-300", isCollapsed ? "pl-16" : "pl-64")}>
        <Outlet />
      </main>
    </div>
  );
}