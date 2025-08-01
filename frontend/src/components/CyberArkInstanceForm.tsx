import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { 
  Dialog,
  DialogContent,
  DialogHeader,
  DialogBody,
  DialogFooter,
} from './ui/dialog';
import { FormCheckbox } from './ui/form-fields';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from './ui/form';
import { Loader2, CheckCircle, XCircle } from 'lucide-react';
import { cyberarkApi, CyberArkInstance, TestConnectionResponse } from '../api/cyberark';
import { cn } from '@/lib/utils';
import { Alert, AlertDescription, AlertTitle } from './ui/alert';
import { useCreateCyberArkInstance, useUpdateCyberArkInstance } from '@/hooks/useCyberArkInstances';
import { SyncConfiguration } from '@/components/SyncConfiguration';

const baseSchema = z.object({
  name: z.string()
    .min(1, 'Instance name is required')
    .max(255, 'Instance name must be less than 255 characters')
    .regex(/^[a-zA-Z0-9\s\-_]+$/, 'Instance name can only contain letters, numbers, spaces, hyphens, and underscores'),
  base_url: z.string()
    .min(1, 'Base URL is required')
    .url('Please enter a valid URL (e.g., https://cyberark.company.com/PasswordVault)')
    .refine((url) => {
      try {
        const u = new URL(url);
        return u.protocol === 'http:' || u.protocol === 'https:';
      } catch {
        return false;
      }
    }, 'URL must use HTTP or HTTPS protocol'),
  username: z.string()
    .min(1, 'Username is required')
    .min(3, 'Username must be at least 3 characters'),
  concurrent_sessions: z.boolean().default(true),
});

const createSchema = baseSchema.extend({
  password: z.string()
    .min(1, 'Password is required'),
  skip_tls_verify: z.boolean().default(false),
});

const editSchema = baseSchema.extend({
  password: z.string().optional(),
  skip_tls_verify: z.boolean().default(false),
});

type CreateFormData = z.infer<typeof createSchema>;
type EditFormData = z.infer<typeof editSchema>;
type FormData = CreateFormData | EditFormData;

interface CyberArkInstanceFormProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
  instance?: CyberArkInstance | null;
  onDelete?: (instance: CyberArkInstance) => void;
}

export function CyberArkInstanceForm({ open, onClose, onSuccess, instance, onDelete }: CyberArkInstanceFormProps) {
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState<TestConnectionResponse | null>(null);
  const [hasTestedSuccessfully, setHasTestedSuccessfully] = useState(false);
  
  const createMutation = useCreateCyberArkInstance();
  const updateMutation = useUpdateCyberArkInstance();

  const form = useForm({
    resolver: zodResolver(instance ? editSchema : createSchema),
    defaultValues: {
      name: '',
      base_url: '',
      username: '',
      password: '',
      concurrent_sessions: true,
      skip_tls_verify: false,
    },
  });

  // Reset form when instance changes or dialog opens
  useEffect(() => {
    if (open) {
      if (instance) {
        form.reset({
          name: instance.name,
          base_url: instance.base_url,
          username: instance.username,
          password: '', // Password is never sent from backend
          concurrent_sessions: instance.concurrent_sessions ?? true,
          skip_tls_verify: instance.skip_tls_verify ?? false,
        });
      } else {
        // Try to load saved values from localStorage for new instances
        const savedValues = localStorage.getItem('cyberark-instance-form');
        if (savedValues) {
          try {
            const parsed = JSON.parse(savedValues);
            form.reset({
              name: parsed.name || '',
              base_url: parsed.base_url || '',
              username: parsed.username || '',
              password: '', // Never restore password
              concurrent_sessions: parsed.concurrent_sessions ?? true,
              skip_tls_verify: parsed.skip_tls_verify ?? false,
            });
          } catch {
            // If parse fails, use defaults
            form.reset({
              name: '',
              base_url: '',
              username: '',
              password: '',
              concurrent_sessions: true,
              skip_tls_verify: false,
            });
          }
        } else {
          form.reset({
            name: '',
            base_url: '',
            username: '',
            password: '',
            concurrent_sessions: true,
            skip_tls_verify: false,
          });
        }
      }
      setTestResult(null);
      setHasTestedSuccessfully(false);
    }
  }, [open, instance, form]);

  // Save form values to localStorage (except password)
  useEffect(() => {
    if (!instance && open) {
      const subscription = form.watch((values, { name }) => {
        const { password, ...valuesToSave } = values;
        localStorage.setItem('cyberark-instance-form', JSON.stringify(valuesToSave));
        
        // Reset test status if connection-related fields change
        if (name && ['base_url', 'username', 'password'].includes(name)) {
          setHasTestedSuccessfully(false);
          setTestResult(null);
        }
      });
      return () => subscription.unsubscribe();
    }
  }, [form, instance, open]);

  const handleTestConnection = async () => {
    const values = form.getValues();
    
    // Validate form first
    const valid = await form.trigger();
    if (!valid) return;

    setIsTesting(true);
    setTestResult(null);

    try {
      const result = await cyberarkApi.testConnection({
        base_url: values.base_url,
        username: values.username,
        password: values.password || '',
        skip_tls_verify: values.skip_tls_verify || false,
      });
      setTestResult(result);
      setHasTestedSuccessfully(result.success);
    } catch (error: any) {
      const errorResult = {
        success: false,
        message: error.response?.data?.error || 'Failed to test connection',
        response_time_ms: 0,
      };
      setTestResult(errorResult);
      setHasTestedSuccessfully(false);
    } finally {
      setIsTesting(false);
    }
  };

  const onSubmit = async (values: FormData) => {
    // For new instances, ensure connection has been tested successfully
    if (!instance && !hasTestedSuccessfully) {
      form.setError('root', {
        message: 'Please test the connection successfully before creating the instance'
      });
      return;
    }

    try {
      if (instance) {
        // Update existing instance
        const updateData: any = {};
        if (values.name !== instance.name) updateData.name = values.name;
        if (values.base_url !== instance.base_url) updateData.base_url = values.base_url;
        if (values.username !== instance.username) updateData.username = values.username;
        if (values.password) updateData.password = values.password;
        if (values.concurrent_sessions !== instance.concurrent_sessions) updateData.concurrent_sessions = values.concurrent_sessions;
        if (values.skip_tls_verify !== instance.skip_tls_verify) updateData.skip_tls_verify = values.skip_tls_verify;

        await updateMutation.mutateAsync({ id: instance.id, data: updateData });
      } else {
        // Create new instance
        await createMutation.mutateAsync({
          ...values,
          password: values.password || ''
        });
        // Clear saved form values on successful creation
        localStorage.removeItem('cyberark-instance-form');
      }
      
      onSuccess();
      onClose();
    } catch (error: any) {
      form.setError('root', {
        message: error.response?.data?.error || 'Failed to save instance'
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={() => !createMutation.isPending && !updateMutation.isPending && onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader
          title={instance ? 'Edit CyberArk Instance' : 'Add CyberArk Instance'}
          description="Configure a CyberArk PVWA instance connection. The connection will be tested before saving."
        />

        <DialogBody>
          <Form {...form}>
            <form id="cyberark-instance-form" onSubmit={form.handleSubmit(onSubmit)} autoComplete="off">
              <div className="space-y-6">
              {form.formState.errors.root && (
                <Alert variant="destructive">
                  <XCircle className="h-4 w-4" />
                  <AlertTitle>Error</AlertTitle>
                  <AlertDescription>{form.formState.errors.root.message}</AlertDescription>
                </Alert>
              )}

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Instance Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Production CyberArk" {...field} />
                  </FormControl>
                  <FormDescription>
                    A unique name to identify this CyberArk instance
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="base_url"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Base URL</FormLabel>
                  <FormControl>
                    <Input 
                      placeholder="https://cyberark.company.com/PasswordVault" 
                      {...field} 
                    />
                  </FormControl>
                  <FormDescription>
                    The full URL of your CyberArk PVWA server including the path (e.g., /PasswordVault)
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-6">
              <FormField
                control={form.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username</FormLabel>
                    <FormControl>
                      <Input 
                        placeholder="orca_service" 
                        autoComplete="off"
                        {...field} 
                      />
                    </FormControl>
                    <FormDescription>
                      API user with access permissions
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password {instance && <span className="font-normal text-gray-500">(optional)</span>}</FormLabel>
                    <FormControl>
                      <Input 
                        type="password" 
                        placeholder="••••••••" 
                        autoComplete="off"
                        {...field} 
                      />
                    </FormControl>
                    <FormDescription>
                      Encrypted storage
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormCheckbox
              form={form}
              name="concurrent_sessions"
              label="Allow concurrent sessions"
              description="Enable up to 300 simultaneous connections to this CyberArk instance (default: on)"
            />

            <FormCheckbox
              form={form}
              name="skip_tls_verify"
              label="Skip TLS certificate verification"
              description="WARNING: This disables SSL/TLS certificate validation. Only use for testing with self-signed certificates."
            />

            {/* Test Result */}
            {testResult && (
              <div className={cn(
                "rounded border p-4",
                testResult.success 
                  ? "border-green-200 bg-green-50/50" 
                  : "border-gray-200 bg-gray-50"
              )}>
                <div className="flex items-start gap-3 mb-3">
                  {testResult.success ? (
                    <CheckCircle className="h-5 w-5 text-green-600 flex-shrink-0" />
                  ) : (
                    <XCircle className="h-5 w-5 text-red-600 flex-shrink-0" />
                  )}
                  <h3 className={cn(
                    "text-sm font-semibold",
                    testResult.success ? "text-green-900" : "text-gray-900"
                  )}>
                    {testResult.success ? 'Connection Successful' : 'Connection Failed'}
                  </h3>
                </div>
                <div className="space-y-2">
                  <pre className="text-xs font-mono bg-white rounded border border-gray-200 p-3 overflow-x-auto whitespace-pre-wrap break-words">
                    <code className="text-gray-700">{testResult.message}</code>
                  </pre>
                  {testResult.response_time_ms > 0 && (
                    <p className="text-xs text-gray-500">Response time: {testResult.response_time_ms}ms</p>
                  )}
                </div>
              </div>
            )}
              </div>
            </form>
          </Form>
          
          {/* Show sync configuration for existing instances */}
          {instance && (
            <div className="mt-6">
              <SyncConfiguration instanceId={instance.id} />
            </div>
          )}
        </DialogBody>
        
        <DialogFooter className="gap-2">
                {instance && onDelete && (
                  <Button
                    type="button"
                    variant="destructive"
                    onClick={() => {
                      onClose(); // Close the form modal first
                      onDelete(instance); // Then trigger the delete confirmation
                    }}
                    disabled={createMutation.isPending || updateMutation.isPending}
                  >
                    Delete
                  </Button>
                )}
                
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleTestConnection}
                  disabled={isTesting || createMutation.isPending || updateMutation.isPending}
                >
                  {isTesting ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Testing...
                    </>
                  ) : (
                    'Test Connection'
                  )}
                </Button>
                
                <div className="flex-1" />
                
                <Button
                  type="button"
                  variant="outline"
                  onClick={onClose}
                  disabled={createMutation.isPending || updateMutation.isPending}
                >
                  Cancel
                </Button>
                
                <Button 
                  type="submit"
                  form="cyberark-instance-form"
                  disabled={createMutation.isPending || updateMutation.isPending || isTesting || (!instance && !hasTestedSuccessfully)}
                >
                  {createMutation.isPending || updateMutation.isPending ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    instance ? 'Update Instance' : 'Create Instance'
                  )}
                </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}