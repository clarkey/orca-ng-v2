import { apiClient } from './client';

export interface CyberArkInstance {
  id: string;
  name: string;
  base_url: string;
  username: string;
  concurrent_sessions: boolean;
  is_active: boolean;
  last_test_at?: string;
  last_test_success?: boolean;
  last_test_error?: string;
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
}

export interface CreateCyberArkInstanceRequest {
  name: string;
  base_url: string;
  username: string;
  password: string;
  concurrent_sessions?: boolean;
}

export interface UpdateCyberArkInstanceRequest {
  name?: string;
  base_url?: string;
  username?: string;
  password?: string;
  concurrent_sessions?: boolean;
  is_active?: boolean;
}

export interface TestConnectionRequest {
  base_url: string;
  username: string;
  password: string;
}

export interface TestConnectionResponse {
  success: boolean;
  message: string;
  response_time_ms: number;
  version?: string;
}

export interface CyberArkInstancesResponse {
  instances: CyberArkInstance[];
  count: number;
}

export const cyberarkApi = {
  // List all instances
  listInstances: async (onlyActive?: boolean): Promise<CyberArkInstancesResponse> => {
    const params = onlyActive ? { active: 'true' } : {};
    return await apiClient.get<CyberArkInstancesResponse>('/cyberark/instances', { params });
  },

  // Get a single instance
  getInstance: async (id: string): Promise<CyberArkInstance> => {
    return await apiClient.get<CyberArkInstance>(`/cyberark/instances/${id}`);
  },

  // Create a new instance
  createInstance: async (data: CreateCyberArkInstanceRequest): Promise<CyberArkInstance> => {
    return await apiClient.post<CyberArkInstance>('/cyberark/instances', data);
  },

  // Update an instance
  updateInstance: async (id: string, data: UpdateCyberArkInstanceRequest): Promise<{ message: string }> => {
    return await apiClient.put<{ message: string }>(`/cyberark/instances/${id}`, data);
  },

  // Delete an instance
  deleteInstance: async (id: string): Promise<{ message: string }> => {
    return await apiClient.delete<{ message: string }>(`/cyberark/instances/${id}`);
  },

  // Test a connection (without saving)
  testConnection: async (data: TestConnectionRequest): Promise<TestConnectionResponse> => {
    return await apiClient.post<TestConnectionResponse>('/cyberark/test-connection', data);
  },

  // Test an existing instance's connection
  testInstanceConnection: async (id: string): Promise<TestConnectionResponse> => {
    return await apiClient.post<TestConnectionResponse>(`/cyberark/instances/${id}/test`);
  },
};