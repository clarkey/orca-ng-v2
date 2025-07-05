import { createContext, useContext, useState, ReactNode } from 'react';

interface CyberArkInstance {
  id: string;
  name: string;
  type: 'overview' | 'instance';
  url?: string;
  status?: 'connected' | 'disconnected' | 'error';
}

interface InstanceContextType {
  instances: CyberArkInstance[];
  currentInstanceId: string;
  currentInstance: CyberArkInstance | null;
  setCurrentInstanceId: (id: string) => void;
  isOverviewMode: boolean;
}

const InstanceContext = createContext<InstanceContextType | undefined>(undefined);

export function useInstance() {
  const context = useContext(InstanceContext);
  if (!context) {
    throw new Error('useInstance must be used within InstanceProvider');
  }
  return context;
}

interface InstanceProviderProps {
  children: ReactNode;
}

export function InstanceProvider({ children }: InstanceProviderProps) {
  // Mock instances - in real app, these would come from API
  const instances: CyberArkInstance[] = [
    { id: 'all', name: 'All Instances', type: 'overview' },
    { id: 'prod-us', name: 'Production US', type: 'instance', url: 'https://cyberark-us.example.com', status: 'connected' },
    { id: 'prod-eu', name: 'Production EU', type: 'instance', url: 'https://cyberark-eu.example.com', status: 'connected' },
    { id: 'dev', name: 'Development', type: 'instance', url: 'https://cyberark-dev.example.com', status: 'disconnected' },
  ];

  const [currentInstanceId, setCurrentInstanceId] = useState<string>('all');
  
  const currentInstance = instances.find(inst => inst.id === currentInstanceId) || null;
  const isOverviewMode = currentInstance?.type === 'overview';

  return (
    <InstanceContext.Provider 
      value={{
        instances,
        currentInstanceId,
        currentInstance,
        setCurrentInstanceId,
        isOverviewMode,
      }}
    >
      {children}
    </InstanceContext.Provider>
  );
}