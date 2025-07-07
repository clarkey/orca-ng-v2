const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

export interface User {
  id: string;
  username: string;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
  is_active: boolean;
  is_admin: boolean;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  message: string;
}

export interface Safe {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateSafeRequest {
  name: string;
  description?: string;
}

export interface UpdateSafeRequest {
  name?: string;
  description?: string;
}

export interface Group {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  username: string;
  password: string;
  is_admin?: boolean;
  is_active?: boolean;
}

export interface UpdateUserRequest {
  username?: string;
  password?: string;
  is_admin?: boolean;
  is_active?: boolean;
}

class ApiClient {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    const config: RequestInit = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      credentials: 'include',
    };

    const response = await fetch(url, config);

    if (!response.ok) {
      const error = await response.json().catch(() => ({
        error: 'An error occurred',
      }));
      throw new Error(error.error || 'An error occurred');
    }

    return response.json();
  }

  async login(data: LoginRequest): Promise<LoginResponse> {
    return this.request<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async logout(): Promise<void> {
    await this.request('/auth/logout', {
      method: 'POST',
    });
  }

  async getCurrentUser(): Promise<User> {
    return this.request<User>('/auth/me');
  }

  // Safe methods
  async getSafes(filters?: any): Promise<{ safes: Safe[] }> {
    return this.get('/safes', { params: filters });
  }

  async getSafe(id: string): Promise<Safe> {
    return this.get(`/safes/${id}`);
  }

  async createSafe(data: CreateSafeRequest): Promise<Safe> {
    return this.post('/safes', data);
  }

  async updateSafe(id: string, data: UpdateSafeRequest): Promise<Safe> {
    return this.put(`/safes/${id}`, data);
  }

  async deleteSafe(id: string): Promise<void> {
    return this.delete(`/safes/${id}`);
  }

  // User methods
  async getUsers(filters?: any): Promise<{ users: User[] }> {
    return this.get('/users', { params: filters });
  }

  async getUser(id: string): Promise<User> {
    return this.get(`/users/${id}`);
  }

  async createUser(data: CreateUserRequest): Promise<User> {
    return this.post('/users', data);
  }

  async updateUser(id: string, data: UpdateUserRequest): Promise<User> {
    return this.put(`/users/${id}`, data);
  }

  async deleteUser(id: string): Promise<void> {
    return this.delete(`/users/${id}`);
  }

  // Group methods
  async getGroups(filters?: any): Promise<{ groups: Group[] }> {
    return this.get('/groups', { params: filters });
  }

  async getGroup(id: string): Promise<Group> {
    return this.get(`/groups/${id}`);
  }

  // Generic HTTP methods for other API calls
  async get<T>(endpoint: string, options?: { params?: Record<string, any> }): Promise<T> {
    let url = endpoint;
    if (options?.params) {
      const searchParams = new URLSearchParams();
      Object.entries(options.params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          searchParams.append(key, String(value));
        }
      });
      const queryString = searchParams.toString();
      if (queryString) {
        url += `?${queryString}`;
      }
    }
    return this.request<T>(url);
  }

  async post<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'DELETE',
    });
  }
}

export const apiClient = new ApiClient();