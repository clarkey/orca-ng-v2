import { useState, useEffect } from 'react';
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
  const [backgroundImage, setBackgroundImage] = useState('');
  const [imageLoaded, setImageLoaded] = useState(false);

  const form = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  });

  useEffect(() => {
    // Use wallpaper 2.jpg
    const imagePath = `/wallpapers/2.jpg`;
    
    // Preload the image
    const img = new Image();
    img.onload = () => {
      setBackgroundImage(imagePath);
      setImageLoaded(true);
    };
    img.src = imagePath;
  }, []);

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
    <div className="relative min-h-screen overflow-hidden">
      {/* Fallback gradient while image loads */}
      <div className="absolute inset-0 z-0 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900" />
      
      {/* Background Image with Overlay */}
      <div 
        className={`absolute inset-0 z-10 transition-opacity duration-[1500ms] ${imageLoaded ? 'opacity-100' : 'opacity-0'}`}
        style={{
          backgroundImage: backgroundImage ? `url(${backgroundImage})` : '',
          backgroundSize: 'cover',
          backgroundPosition: 'center',
          backgroundRepeat: 'no-repeat',
        }}
      >
        {/* Dark overlay for better readability */}
        <div className="absolute inset-0 bg-black/30" />
      </div>

      {/* Header */}
      <header className="relative z-20">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <OrcaIcon className="h-10 w-10 text-white drop-shadow-lg" />
            
            {/* Support Button */}
            <Button
              variant="ghost"
              size="sm"
              className="text-white/90 hover:text-white hover:bg-white/10 font-normal"
              onClick={() => window.location.href = '/support'}
            >
              Support
            </Button>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="relative z-20 flex flex-col items-center justify-center px-4 py-12 sm:px-6 lg:px-8" style={{ minHeight: 'calc(100vh - 4rem)' }}>
        <div className="w-full max-w-md">
          {/* Form card */}
          <div className="bg-white rounded shadow-2xl ring-1 ring-black/5 p-8" style={{boxShadow: '0 25px 60px -15px rgba(0, 0, 0, 0.4), 0 0 25px rgba(0, 0, 0, 0.15)'}}>
              {/* Logo on the left */}
              <div className="mb-8">
                <LogoWithStroke className="h-16 w-auto text-gray-700" />
              </div>
              
              {/* Divider line */}
              <div className="w-full border-t border-gray-200 mb-8" />
              
              <div className="space-y-6">
                <Form {...form}>
                  <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
                    {error && (
                      <Alert variant="destructive" className="bg-red-50 border-red-200">
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
                              className="h-12 bg-gray-50 border-gray-200 text-gray-900 placeholder:text-gray-400 focus:bg-white focus:border-gray-400 rounded"
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
                              className="h-12 bg-gray-50 border-gray-200 text-gray-900 placeholder:text-gray-400 focus:bg-white focus:border-gray-400 rounded"
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
                      className="w-full h-12 bg-gray-700 hover:bg-gray-800 text-white font-medium rounded transition-all transform hover:scale-[1.02]"
                      disabled={isLoading}
                    >
                      {isLoading ? 'Logging in...' : 'Log in with ORCA'}
                    </Button>
                  </form>
                </Form>
                
                <div className="relative">
                  <div className="absolute inset-0 flex items-center">
                    <div className="w-full border-t border-gray-200" />
                  </div>
                  <div className="relative flex justify-center text-sm">
                    <span className="px-3 bg-white text-gray-500">or continue with</span>
                  </div>
                </div>
                
                <Button
                  type="button"
                  onClick={() => window.location.href = '/auth/entra'}
                  className="w-full h-12 bg-white hover:bg-gray-50 text-gray-700 font-medium border-2 border-gray-700 rounded transition-all transform hover:scale-[1.02]"
                  disabled={isLoading}
                >
                  Log in with Entra ID
                </Button>
              </div>
          </div>

          {/* Footer info */}
          <div className="flex justify-between items-center mt-6 px-4 text-xs text-white/80">
            <p className="drop-shadow-lg">© 2025 Privilent. All rights reserved.</p>
            <p className="drop-shadow-lg">ORCA v2.0.0 • Build 2025.01</p>
          </div>
        </div>
      </main>
    </div>
  );
}