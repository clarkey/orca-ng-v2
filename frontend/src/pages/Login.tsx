import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { OrcaIcon } from '@/components/OrcaIcon';
import { LogoWithStroke } from '@/components/LogoWithStroke';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertCircle } from 'lucide-react';

const loginSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export function Login() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const form = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  });

  const handleSubmit = async (values: LoginFormData) => {
    setError('');
    setIsLoading(true);

    try {
      await login({ username: values.username, password: values.password });
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
            <OrcaIcon className="h-10 w-10 text-gray-900" />
            
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
          <div className="flex justify-center mb-8">
            <LogoWithStroke className="h-32 w-auto" />
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
            <Form {...form}>
              <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
                {error && (
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}
                
                <FormField
                  control={form.control}
                  name="username"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <Input
                          placeholder="Username"
                          autoComplete="username"
                          className="h-11 focus-visible:ring-slate-700"
                          disabled={isLoading}
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                
                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <Input
                          type="password"
                          placeholder="Password"
                          autoComplete="current-password"
                          className="h-11 focus-visible:ring-slate-700"
                          disabled={isLoading}
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                
                <Button
                  type="submit"
                  className="w-full h-11 bg-white hover:bg-gray-50 text-gray-900 font-medium border border-gray-300"
                  disabled={isLoading}
                >
                  {isLoading ? 'Logging in...' : 'Log in with ORCA'}
                </Button>
              </form>
            </Form>
          </div>

          <p className="text-center text-xs text-gray-500 mt-8">
            © 2025 Privilent. All rights reserved.
          </p>
        </div>
      </main>
    </div>
  );
}