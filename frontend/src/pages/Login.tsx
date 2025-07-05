import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { OrcaIcon } from '@/components/OrcaIcon';

export function Login() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await login({ username, password });
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-white">
      {/* Header */}
      <header className="bg-white">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <OrcaIcon className="h-8 w-8 text-gray-900" />
            
            {/* Support Button */}
            <Button
              variant="ghost"
              size="sm"
              className="text-gray-600 hover:text-gray-900 font-normal"
              onClick={() => window.location.href = '/support'}
            >
              Support
            </Button>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="flex flex-col items-center justify-center px-4 py-12 sm:px-6 lg:px-8" style={{ minHeight: 'calc(100vh - 4rem)' }}>
        <div className="w-full max-w-sm">
          <div className="text-center mb-8">
            <h2 className="text-3xl font-normal text-gray-900">Log in to ORCA</h2>
          </div>

          <div className="space-y-6">
            <Button
              type="button"
              onClick={() => window.location.href = '/auth/entra'}
              className="w-full h-11 bg-slate-700 hover:bg-slate-800 text-white font-medium"
              disabled={isLoading}
            >
              Login with Entra ID
            </Button>
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">or</span>
              </div>
            </div>
            <form onSubmit={handleSubmit} className="space-y-6">
              <div>
                <Input
                  id="username"
                  type="text"
                  placeholder="Username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                  disabled={isLoading}
                  autoComplete="username"
                  className="h-11 focus-visible:ring-slate-700"
                />
              </div>
              <div>
                <Input
                  id="password"
                  type="password"
                  placeholder="Password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={isLoading}
                  autoComplete="current-password"
                  className="h-11 focus-visible:ring-slate-700"
                />
              </div>
              {error && (
                <div className="rounded-md bg-red-50 border border-red-200 p-3">
                  <p className="text-sm text-red-800">{error}</p>
                </div>
              )}
              <Button
                type="submit"
                className="w-full h-11 bg-white hover:bg-gray-50 text-gray-900 font-medium border border-gray-300"
                disabled={isLoading}
              >
                {isLoading ? 'Logging in...' : 'Log in with ORCA'}
              </Button>
            </form>
          </div>

          <p className="text-center text-xs text-gray-500 mt-8">
            © 2025 Privilent. All rights reserved.
          </p>
        </div>
      </main>
    </div>
  );
}