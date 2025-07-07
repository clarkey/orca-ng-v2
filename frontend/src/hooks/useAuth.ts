import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, LoginRequest } from '@/api/client';
import { useNavigate } from 'react-router-dom';

export const authKeys = {
  all: ['auth'] as const,
  user: () => [...authKeys.all, 'user'] as const,
};

// Hook to get current user
export function useCurrentUser() {
  return useQuery({
    queryKey: authKeys.user(),
    queryFn: () => apiClient.getCurrentUser(),
    retry: false,
    staleTime: Infinity, // User data doesn't change often
  });
}

// Hook for login mutation
export function useLogin() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  
  return useMutation({
    mutationFn: (data: LoginRequest) => apiClient.login(data),
    onSuccess: (response) => {
      // Set user data in cache
      queryClient.setQueryData(authKeys.user(), response.user);
      navigate('/');
    },
  });
}

// Hook for logout mutation
export function useLogout() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  
  return useMutation({
    mutationFn: () => apiClient.logout(),
    onSuccess: () => {
      // Clear all queries
      queryClient.clear();
      navigate('/login');
    },
  });
}