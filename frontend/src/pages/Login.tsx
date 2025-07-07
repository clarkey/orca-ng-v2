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
    // Select a random wallpaper
    const wallpaperNumber = Math.floor(Math.random() * 5) + 1;
    const extension = wallpaperNumber === 5 ? 'JPG' : 'jpg';
    const imagePath = `/wallpapers/${wallpaperNumber}.${extension}`;
    
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
      {/* Background Image with Overlay */}
      <div 
        className={`absolute inset-0 z-0 transition-opacity duration-1000 ${imageLoaded ? 'opacity-100' : 'opacity-0'}`}
        style={{
          backgroundImage: backgroundImage ? `url(${backgroundImage})` : '',
          backgroundSize: 'cover',
          backgroundPosition: 'center',
          backgroundRepeat: 'no-repeat',
        }}
      >
        {/* Dark overlay for better readability */}
        <div className="absolute inset-0 bg-black/40" />
      </div>
      
      {/* Fallback gradient while image loads */}
      <div className="absolute inset-0 z-0 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900" />

      {/* Header */}
      <header className="relative z-10">
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
      <main className="relative z-10 flex flex-col items-center justify-center px-4 py-12 sm:px-6 lg:px-8" style={{ minHeight: 'calc(100vh - 4rem)' }}>
        <div className="w-full max-w-sm">
          {/* Glass card effect */}
          <div className="backdrop-blur-md bg-white/10 border border-white/20 rounded-2xl shadow-2xl p-8">
            <div className="flex justify-center mb-8">
              <LogoWithStroke className="h-32 w-auto drop-shadow-lg" />
            </div>

            <div className="space-y-6">
              <Button
                type="button"
                onClick={() => window.location.href = '/auth/entra'}
                className="w-full h-11 bg-white/90 hover:bg-white text-gray-900 font-medium shadow-lg transition-all hover:shadow-xl"
                disabled={isLoading}
              >
                Login with Entra ID
              </Button>
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-white/30" />
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-2 bg-transparent text-white/70">or</span>
                </div>
              </div>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
                  {error && (
                    <Alert variant="destructive" className="bg-red-500/20 border-red-500/30 text-white">
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
                            className="h-11 bg-white/10 border-white/20 text-white placeholder:text-white/50 focus:bg-white/20 focus:border-white/40 focus-visible:ring-white/30"
                            disabled={isLoading}
                            {...field}
                          />
                        </FormControl>
                        <FormMessage className="text-white/90" />
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
                            className="h-11 bg-white/10 border-white/20 text-white placeholder:text-white/50 focus:bg-white/20 focus:border-white/40 focus-visible:ring-white/30"
                            disabled={isLoading}
                            {...field}
                          />
                        </FormControl>
                        <FormMessage className="text-white/90" />
                      </FormItem>
                    )}
                  />
                  
                  <Button
                    type="submit"
                    className="w-full h-11 bg-white/10 hover:bg-white/20 text-white font-medium border border-white/20 shadow-lg transition-all hover:shadow-xl"
                    disabled={isLoading}
                  >
                    {isLoading ? 'Logging in...' : 'Log in with ORCA'}
                  </Button>
                </form>
              </Form>
            </div>
          </div>

          <p className="text-center text-xs text-white/60 mt-8">
            © 2025 Privilent. All rights reserved.
          </p>
        </div>
      </main>
    </div>
  );
}