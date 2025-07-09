import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { FormCheckbox } from '@/components/ui/form-fields';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Loader2, AlertCircle, Check, X, ArrowLeft } from 'lucide-react';
import { certificateAuthoritiesApi } from '@/api/certificateAuthorities';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { PageContainer } from '@/components/PageContainer';
import { PageHeader } from '@/components/PageHeader';
import { format } from 'date-fns';
import { Skeleton } from '@/components/ui/skeleton';

const updateSchema = z.object({
  name: z.string()
    .min(1, 'Certificate name is required')
    .max(255, 'Certificate name must be less than 255 characters'),
  description: z.string()
    .max(1000, 'Description must be less than 1000 characters')
    .optional(),
  is_active: z.boolean().default(true),
});

type UpdateFormData = z.infer<typeof updateSchema>;

export function CertificateAuthorityEdit() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();

  // Fetch certificate details
  const { data: certificate, isLoading } = useQuery({
    queryKey: ['certificate-authorities', id],
    queryFn: () => certificateAuthoritiesApi.get(id!),
    enabled: !!id,
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      certificateAuthoritiesApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificate-authorities'] });
      navigate('/settings/certificates');
    },
    onError: (error: any) => {
      form.setError('root', {
        message: error.response?.data?.error || 'Failed to update certificate authority'
      });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: certificateAuthoritiesApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificate-authorities'] });
      navigate('/settings/certificates');
    },
    onError: (error: any) => {
      alert(error.response?.data?.error || 'Failed to delete certificate authority');
    },
  });

  const form = useForm<UpdateFormData>({
    resolver: zodResolver(updateSchema),
    defaultValues: {
      name: '',
      description: '',
      is_active: true,
    },
  });

  // Update form when certificate data is loaded
  useEffect(() => {
    if (certificate) {
      form.reset({
        name: certificate.name,
        description: certificate.description || '',
        is_active: certificate.is_active,
      });
    }
  }, [certificate, form]);

  const onSubmit = async (values: UpdateFormData) => {
    if (!id) return;
    await updateMutation.mutateAsync({ id, data: values });
  };

  const handleDelete = () => {
    if (!id || !certificate) return;
    
    if (confirm(`Are you sure you want to delete the certificate authority "${certificate.name}"? This action cannot be undone.`)) {
      deleteMutation.mutate(id);
    }
  };

  const isPending = updateMutation.isPending || deleteMutation.isPending;

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader 
          title={<Skeleton className="h-8 w-64 inline-block" />}
          description={<Skeleton className="h-4 w-96 inline-block mt-2" />}
        />
        <Card>
          <CardContent className="pt-6 space-y-6">
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-48 w-full" />
          </CardContent>
        </Card>
      </PageContainer>
    );
  }

  if (!certificate) {
    return (
      <PageContainer>
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>Certificate authority not found</AlertDescription>
        </Alert>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <PageHeader
        title="Edit Certificate Authority"
        description="Update the certificate authority details"
        actions={
          <Button 
            variant="outline" 
            onClick={() => navigate('/settings/certificates')}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back
          </Button>
        }
      />

      <Card>
        <CardContent className="pt-6">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              {form.formState.errors.root && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>Error</AlertTitle>
                  <AlertDescription>{form.formState.errors.root.message}</AlertDescription>
                </Alert>
              )}

              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Certificate Name</FormLabel>
                    <FormControl>
                      <Input placeholder="Company Root CA" {...field} />
                    </FormControl>
                    <FormDescription>
                      A unique name to identify this certificate authority
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Description <span className="font-normal text-gray-500">(optional)</span></FormLabel>
                    <FormControl>
                      <Input 
                        placeholder="Internal root certificate for company services" 
                        {...field} 
                        value={field.value || ''}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Certificate Info Display */}
              <div className="rounded-lg border bg-gray-50 p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-medium text-gray-900">Certificate Information</h4>
                  <div className="flex items-center gap-2">
                    {certificate.certificate_count > 1 && (
                      <Badge variant="outline" className="text-xs">
                        Chain ({certificate.certificate_count} certificates)
                      </Badge>
                    )}
                    {certificate.is_root_ca && (
                      <Badge variant="secondary" className="text-xs">Root CA</Badge>
                    )}
                    {certificate.is_intermediate && (
                      <Badge variant="secondary" className="text-xs">Intermediate CA</Badge>
                    )}
                  </div>
                </div>
                
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-gray-500">Subject:</span>
                    <p className="font-mono text-xs mt-1">{certificate.subject}</p>
                  </div>
                  <div>
                    <span className="text-gray-500">Issuer:</span>
                    <p className="font-mono text-xs mt-1">{certificate.issuer}</p>
                  </div>
                  <div>
                    <span className="text-gray-500">Valid From:</span>
                    <p className="mt-1">{format(new Date(certificate.not_before), 'PPP')}</p>
                  </div>
                  <div>
                    <span className="text-gray-500">Valid Until:</span>
                    <p className="mt-1">{format(new Date(certificate.not_after), 'PPP')}</p>
                  </div>
                  <div className="col-span-2">
                    <span className="text-gray-500">SHA256 Fingerprint:</span>
                    <p className="font-mono text-xs mt-1 break-all">{certificate.fingerprint}</p>
                  </div>
                </div>

                {/* Show certificate chain details if available */}
                {certificate.chain_info && (() => {
                  try {
                    const chainInfo = typeof certificate.chain_info === 'string' 
                      ? JSON.parse(certificate.chain_info) 
                      : certificate.chain_info;
                    
                    if (Array.isArray(chainInfo) && chainInfo.length > 1) {
                      return (
                        <div className="pt-3 border-t">
                          <h5 className="text-xs font-medium text-gray-700 mb-2">Certificate Chain</h5>
                          <div className="space-y-2">
                            {chainInfo.map((cert: any, index: number) => (
                              <div key={index} className="text-xs p-2 bg-white rounded border">
                                <div className="flex items-center justify-between mb-1">
                                  <span className="font-medium">
                                    {index === 0 ? 'Primary Certificate' : 
                                     cert.is_self_signed ? 'Root CA' : 'Intermediate CA'}
                                  </span>
                                  {cert.is_self_signed && (
                                    <Badge variant="outline" className="text-xs">Self-signed</Badge>
                                  )}
                                </div>
                                <div className="font-mono text-gray-600 truncate" title={cert.subject}>
                                  {cert.subject}
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      );
                    }
                    return null;
                  } catch (e) {
                    return null;
                  }
                })()}

                <div className="pt-2">
                  {certificate.is_expired ? (
                    <Badge variant="destructive" className="gap-1">
                      <X className="h-3 w-3" />
                      Expired
                    </Badge>
                  ) : certificate.expires_in_days <= 30 ? (
                    <Badge variant="destructive">Expires in {certificate.expires_in_days} days</Badge>
                  ) : certificate.expires_in_days <= 90 ? (
                    <Badge variant="secondary">Expires in {certificate.expires_in_days} days</Badge>
                  ) : (
                    <Badge variant="success" className="gap-1">
                      <Check className="h-3 w-3" />
                      Valid for {certificate.expires_in_days} days
                    </Badge>
                  )}
                </div>
              </div>

              <FormCheckbox
                form={form}
                name="is_active"
                label="Active"
                description="Enable this certificate authority for validating connections"
              />

              <div className="flex justify-between pt-6 border-t">
                <Button
                  type="button"
                  variant="destructive"
                  onClick={handleDelete}
                  disabled={isPending}
                >
                  Delete Certificate
                </Button>

                <div className="flex gap-4">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => navigate('/settings/certificates')}
                    disabled={isPending}
                  >
                    Cancel
                  </Button>
                  
                  <Button 
                    type="submit" 
                    disabled={isPending}
                  >
                    {isPending ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        Saving...
                      </>
                    ) : (
                      'Update Certificate'
                    )}
                  </Button>
                </div>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </PageContainer>
  );
}