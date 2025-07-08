import { apiClient } from '@/api/client';

export interface CertificateAuthority {
  id: string;
  name: string;
  description?: string;
  certificate: string;
  fingerprint: string;
  subject: string;
  issuer: string;
  not_before: string;
  not_after: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CertificateAuthorityInfo {
  id: string;
  name: string;
  description?: string;
  fingerprint: string;
  subject: string;
  issuer: string;
  not_before: string;
  not_after: string;
  is_active: boolean;
  is_expired: boolean;
  expires_in_days: number;
  created_at: string;
  updated_at: string;
}

export interface CreateCertificateAuthorityRequest {
  name: string;
  description?: string;
  certificate: string;
  is_active?: boolean;
}

export interface UpdateCertificateAuthorityRequest {
  name?: string;
  description?: string;
  is_active?: boolean;
}

export const certificateAuthoritiesApi = {
  list: () =>
    apiClient.get<{ certificate_authorities: CertificateAuthorityInfo[] }>('/certificate-authorities'),

  get: (id: string) =>
    apiClient.get<CertificateAuthority>(`/certificate-authorities/${id}`),

  create: (data: CreateCertificateAuthorityRequest) =>
    apiClient.post<CertificateAuthority>('/certificate-authorities', data),

  update: (id: string, data: UpdateCertificateAuthorityRequest) =>
    apiClient.put<CertificateAuthority>(`/certificate-authorities/${id}`, data),

  delete: (id: string) =>
    apiClient.delete<void>(`/certificate-authorities/${id}`),
};