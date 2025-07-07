import { createContext, useContext, useState, ReactNode, useMemo } from 'react';
import { CyberArkInstance as ApiCyberArkInstance } from '@/api/cyberark';
import { useCyberArkInstances } from '@/hooks/useCyberArkInstances';

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
  isLoading: boolean;
  refetch: () => void;
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
  const [currentInstanceId, setCurrentInstanceId] = useState<string>('all');
  const { data: response, isLoading, refetch } = useCyberArkInstances();

  const getInstanceStatus = (instance: ApiCyberArkInstance): 'connected' | 'disconnected' | 'error' => {
    if (!instance.last_test_at) {
      return 'disconnected';
    }
    return instance.last_test_success ? 'connected' : 'error';
  };

  const instances = useMemo(() => {
    const apiInstances = response?.instances || [];
    
    // Convert API instances to context format
    const contextInstances: CyberArkInstance[] = [
      { id: 'all', name: 'All Instances', type: 'overview' },
      ...apiInstances
        .filter(inst => inst.is_active) // Only include active instances
        .map(inst => ({
          id: inst.id,
          name: inst.name,
          type: 'instance' as const,
          url: inst.base_url,
          status: getInstanceStatus(inst),
        }))
    ];
    
    // If current instance is no longer in the list, reset to overview
    if (currentInstanceId !== 'all' && !contextInstances.find(i => i.id === currentInstanceId)) {
      setCurrentInstanceId('all');
    }
    
    return contextInstances;
  }, [response, currentInstanceId]);
  
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
        isLoading,
        refetch,
      }}
    >
      {children}
    </InstanceContext.Provider>
  );
}