import { useState, useCallback } from 'react';

export interface Toast {
  id: string;
  title: string;
  description?: string;
  variant?: 'default' | 'destructive';
}

interface ToastOptions {
  title: string;
  description?: string;
  variant?: 'default' | 'destructive';
}

// Simple in-memory toast state for now
// In a real app, this would be managed by a context or global state
let toastListeners: ((toasts: Toast[]) => void)[] = [];
let toasts: Toast[] = [];

const notifyListeners = () => {
  toastListeners.forEach(listener => listener([...toasts]));
};

export function useToast() {
  const [, forceUpdate] = useState({});
  // Use forceUpdate to trigger re-renders when needed (currently not used)
  void forceUpdate;

  const toast = useCallback((options: ToastOptions) => {
    const id = Date.now().toString();
    const newToast: Toast = {
      id,
      ...options,
    };
    
    toasts = [...toasts, newToast];
    notifyListeners();
    
    // Auto-remove toast after 5 seconds
    setTimeout(() => {
      toasts = toasts.filter(t => t.id !== id);
      notifyListeners();
    }, 5000);
  }, []);

  const dismiss = useCallback((id: string) => {
    toasts = toasts.filter(t => t.id !== id);
    notifyListeners();
  }, []);

  return {
    toast,
    dismiss,
  };
}