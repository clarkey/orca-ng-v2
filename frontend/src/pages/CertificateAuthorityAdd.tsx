import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
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
import { Loader2, Upload, AlertCircle, FileText, ArrowLeft } from 'lucide-react';
import { certificateAuthoritiesApi } from '@/api/certificateAuthorities';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { PageContainer } from '@/components/PageContainer';
import { PageHeader } from '@/components/PageHeader';

const createSchema = z.object({
  name: z.string()
    .min(1, 'Certificate name is required')
    .max(255, 'Certificate name must be less than 255 characters'),
  description: z.string()
    .max(1000, 'Description must be less than 1000 characters')
    .optional(),
  certificate: z.string()
    .min(1, 'Certificate is required')
    .refine((cert) => {
      // Basic PEM format validation
      return cert.includes('-----BEGIN CERTIFICATE-----') && 
             cert.includes('-----END CERTIFICATE-----');
    }, 'Invalid certificate format. Must be in PEM format')
    .transform((cert) => {
      // Count certificates in the input
      const certCount = (cert.match(/-----BEGIN CERTIFICATE-----/g) || []).length;
      if (certCount > 1) {
        console.log(`Detected certificate chain with ${certCount} certificates`);
      }
      return cert;
    }),
  is_active: z.boolean().default(true),
});

type CreateFormData = z.infer<typeof createSchema>;

export function CertificateAuthorityAdd() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [dragActive, setDragActive] = useState(false);
  const [detectedCertCount, setDetectedCertCount] = useState<number>(0);

  const createMutation = useMutation({
    mutationFn: certificateAuthoritiesApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificate-authorities'] });
      navigate('/settings/certificates');
    },
    onError: (error: any) => {
      form.setError('root', {
        message: error.response?.data?.error || 'Failed to add certificate authority'
      });
    },
  });

  const form = useForm<CreateFormData>({
    resolver: zodResolver(createSchema),
    defaultValues: {
      name: '',
      description: '',
      certificate: '',
      is_active: true,
    },
  });

  const detectCertificateCount = (content: string) => {
    const certCount = (content.match(/-----BEGIN CERTIFICATE-----/g) || []).length;
    setDetectedCertCount(certCount);
    return certCount;
  };

  const handleFileUpload = (file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
      const trimmedContent = content.trim();
      detectCertificateCount(trimmedContent);
      form.setValue('certificate', trimmedContent);
      form.trigger('certificate');
    };
    reader.readAsText(file);
  };

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFileUpload(e.dataTransfer.files[0]);
    }
  };

  const onSubmit = async (values: CreateFormData) => {
    await createMutation.mutateAsync(values);
  };

  const isPending = createMutation.isPending;

  return (
    <PageContainer>
      <PageHeader
        title="Add Certificate Authority"
        description="Add a trusted certificate authority to verify secure connections"
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

              <FormField
                control={form.control}
                name="certificate"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Certificate (PEM format)</FormLabel>
                    <FormControl>
                      <div className="space-y-4">
                        <div
                          className={`border-2 border-dashed rounded-lg p-6 text-center transition-colors ${
                            dragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'
                          }`}
                          onDragEnter={handleDrag}
                          onDragLeave={handleDrag}
                          onDragOver={handleDrag}
                          onDrop={handleDrop}
                        >
                          <Upload className="mx-auto h-12 w-12 text-gray-400 mb-3" />
                          <p className="text-sm text-gray-600 mb-2">
                            Drag and drop a certificate file here, or click to select
                          </p>
                          <p className="text-xs text-gray-500">
                            Supports single certificates or certificate chains
                          </p>
                          <input
                            ref={fileInputRef}
                            type="file"
                            className="hidden"
                            accept=".pem,.crt,.cer"
                            onChange={(e) => {
                              if (e.target.files?.[0]) {
                                handleFileUpload(e.target.files[0]);
                              }
                            }}
                          />
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => fileInputRef.current?.click()}
                            className="mt-2"
                          >
                            <FileText className="mr-2 h-4 w-4" />
                            Select File
                          </Button>
                        </div>
                        
                        <div className="relative">
                          <Textarea
                            placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                            className="font-mono text-xs min-h-[200px]"
                            {...field}
                            onChange={(e) => {
                              field.onChange(e);
                              detectCertificateCount(e.target.value);
                            }}
                          />
                          {detectedCertCount > 0 && (
                            <div className="absolute top-2 right-2 pointer-events-none">
                              <Badge variant={detectedCertCount > 1 ? "default" : "secondary"} className="text-xs">
                                {detectedCertCount} certificate{detectedCertCount !== 1 ? 's' : ''} detected
                              </Badge>
                            </div>
                          )}
                        </div>
                      </div>
                    </FormControl>
                    <div className="space-y-2">
                      <FormDescription>
                        Upload or paste CA certificates in PEM format. Certificate chains are supported.
                      </FormDescription>
                      <Alert className="mt-2">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle className="text-sm">Certificate Chain Order</AlertTitle>
                        <AlertDescription className="text-xs">
                          When uploading a certificate chain, include certificates in this order:
                          <ol className="list-decimal list-inside mt-2 space-y-1">
                            <li>Intermediate CA certificate (the one that signed your service certificates)</li>
                            <li>Higher-level intermediate CAs (if any)</li>
                            <li>Root CA certificate (self-signed) last</li>
                          </ol>
                          <div className="mt-2 p-2 bg-gray-50 rounded text-xs">
                            <div className="font-medium mb-1">Example for a service cert signed by an intermediate:</div>
                            <div className="font-mono text-gray-600">
                              [Intermediate CA] ← signs service certs<br/>
                              [Root CA] ← signs intermediate
                            </div>
                          </div>
                        </AlertDescription>
                      </Alert>
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormCheckbox
                form={form}
                name="is_active"
                label="Active"
                description="Enable this certificate authority for validating connections"
              />

              <div className="flex justify-end gap-4 pt-6 border-t">
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
                    'Add Certificate'
                  )}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </PageContainer>
  );
}