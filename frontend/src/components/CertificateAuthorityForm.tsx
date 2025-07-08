import { useState, useEffect, useRef } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Textarea } from './ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogFooter,
} from './ui/dialog';
import { DialogHeaderStyled } from './ui/dialog-header-styled';
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
import { Loader2, Upload, AlertCircle, FileText, Check, X } from 'lucide-react';
import { certificateAuthoritiesApi, CertificateAuthorityInfo } from '@/api/certificateAuthorities';
import { Alert, AlertDescription, AlertTitle } from './ui/alert';
import { Badge } from './ui/badge';
import { format } from 'date-fns';

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
    }, 'Invalid certificate format. Must be in PEM format'),
  is_active: z.boolean().default(true),
});

const updateSchema = z.object({
  name: z.string()
    .min(1, 'Certificate name is required')
    .max(255, 'Certificate name must be less than 255 characters'),
  description: z.string()
    .max(1000, 'Description must be less than 1000 characters')
    .optional(),
  is_active: z.boolean().default(true),
});

type CreateFormData = z.infer<typeof createSchema>;
type UpdateFormData = z.infer<typeof updateSchema>;

interface CertificateAuthorityFormProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
  certificateAuthority?: CertificateAuthorityInfo | null;
  onDelete?: (ca: CertificateAuthorityInfo) => void;
}

export function CertificateAuthorityForm({ 
  open, 
  onClose, 
  onSuccess, 
  certificateAuthority,
  onDelete
}: CertificateAuthorityFormProps) {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [dragActive, setDragActive] = useState(false);
  
  const isEditMode = !!certificateAuthority;

  // Fetch full certificate details if editing
  const { data: fullCertificate } = useQuery({
    queryKey: ['certificate-authorities', certificateAuthority?.id],
    queryFn: () => certificateAuthority ? certificateAuthoritiesApi.get(certificateAuthority.id) : null,
    enabled: isEditMode && open,
  });

  const createMutation = useMutation({
    mutationFn: certificateAuthoritiesApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificate-authorities'] });
      alert('Certificate authority added successfully');
      onSuccess();
    },
    onError: (error: any) => {
      form.setError('root', {
        message: error.response?.data?.error || 'Failed to add certificate authority'
      });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      certificateAuthoritiesApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificate-authorities'] });
      alert('Certificate authority updated successfully');
      onSuccess();
    },
    onError: (error: any) => {
      form.setError('root', {
        message: error.response?.data?.error || 'Failed to update certificate authority'
      });
    },
  });


  const form = useForm<any>({
    resolver: zodResolver(isEditMode ? updateSchema : createSchema),
    defaultValues: {
      name: '',
      description: '',
      certificate: '',
      is_active: true,
    },
  });

  // Reset form when dialog opens
  useEffect(() => {
    if (open) {
      if (isEditMode && certificateAuthority) {
        form.reset({
          name: certificateAuthority.name,
          description: certificateAuthority.description || '',
          is_active: certificateAuthority.is_active,
        });
      } else {
        form.reset({
          name: '',
          description: '',
          certificate: '',
          is_active: true,
        });
      }
    }
  }, [open, isEditMode, certificateAuthority, form]);

  const handleFileUpload = (file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
      form.setValue('certificate' as any, content.trim());
      form.trigger('certificate' as any);
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

  const onSubmit = async (values: CreateFormData | UpdateFormData) => {
    try {
      if (isEditMode && certificateAuthority) {
        // Update existing certificate authority
        await updateMutation.mutateAsync({
          id: certificateAuthority.id,
          data: values
        });
      } else {
        // Create new certificate authority
        await createMutation.mutateAsync(values as CreateFormData);
      }
    } catch (error) {
      // Error is handled by mutation onError
    }
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <Dialog open={open} onOpenChange={() => !isPending && onClose()}>
      <DialogContent className="max-w-2xl p-0 overflow-hidden max-h-[90vh]">
        <DialogHeaderStyled 
          title={isEditMode ? 'Edit Certificate Authority' : 'Add Certificate Authority'}
          description={isEditMode 
            ? 'Update the certificate authority details'
            : 'Add a trusted certificate authority to verify secure connections'
          }
        />

        <div className="px-6 pb-6 pt-2 overflow-y-auto">
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

              {!isEditMode && (
                <FormField
                  control={form.control}
                  name={"certificate" as any}
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
                            >
                              <FileText className="mr-2 h-4 w-4" />
                              Select File
                            </Button>
                          </div>
                          
                          <Textarea
                            placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                            className="font-mono text-xs min-h-[200px]"
                            {...field}
                          />
                        </div>
                      </FormControl>
                      <FormDescription>
                        Upload or paste a CA certificate in PEM format. Only CA certificates with certificate signing permissions are accepted.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              )}

              {/* Certificate Info Display for Edit Mode */}
              {isEditMode && fullCertificate && (
                <div className="rounded-lg border bg-gray-50 p-4 space-y-3">
                  <h4 className="text-sm font-medium text-gray-900">Certificate Information</h4>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="text-gray-500">Subject:</span>
                      <p className="font-mono text-xs mt-1">{fullCertificate.subject}</p>
                    </div>
                    <div>
                      <span className="text-gray-500">Issuer:</span>
                      <p className="font-mono text-xs mt-1">{fullCertificate.issuer}</p>
                    </div>
                    <div>
                      <span className="text-gray-500">Valid From:</span>
                      <p className="mt-1">{format(new Date(fullCertificate.not_before), 'PPP')}</p>
                    </div>
                    <div>
                      <span className="text-gray-500">Valid Until:</span>
                      <p className="mt-1">{format(new Date(fullCertificate.not_after), 'PPP')}</p>
                    </div>
                    <div className="col-span-2">
                      <span className="text-gray-500">SHA256 Fingerprint:</span>
                      <p className="font-mono text-xs mt-1 break-all">{fullCertificate.fingerprint}</p>
                    </div>
                  </div>
                  {certificateAuthority && (
                    <div className="pt-2">
                      {certificateAuthority.is_expired ? (
                        <Badge variant="destructive" className="gap-1">
                          <X className="h-3 w-3" />
                          Expired
                        </Badge>
                      ) : certificateAuthority.expires_in_days <= 30 ? (
                        <Badge variant="destructive">Expires in {certificateAuthority.expires_in_days} days</Badge>
                      ) : certificateAuthority.expires_in_days <= 90 ? (
                        <Badge variant="secondary">Expires in {certificateAuthority.expires_in_days} days</Badge>
                      ) : (
                        <Badge variant="success" className="gap-1">
                          <Check className="h-3 w-3" />
                          Valid for {certificateAuthority.expires_in_days} days
                        </Badge>
                      )}
                    </div>
                  )}
                </div>
              )}

              <FormCheckbox
                form={form}
                name="is_active"
                label="Active"
                description="Enable this certificate authority for validating connections"
              />

              <DialogFooter className="gap-2 pt-6 mt-6 border-t">
                {isEditMode && onDelete && (
                  <Button
                    type="button"
                    variant="destructive"
                    onClick={() => {
                      if (certificateAuthority) {
                        onClose(); // Close the form modal first
                        onDelete(certificateAuthority); // Then trigger the delete confirmation
                      }
                    }}
                    disabled={isPending}
                  >
                    Delete
                  </Button>
                )}
                
                <div className="flex-1" />
                
                <Button
                  type="button"
                  variant="outline"
                  onClick={onClose}
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
                    isEditMode ? 'Update Certificate' : 'Add Certificate'
                  )}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </div>
      </DialogContent>
    </Dialog>
  );
}