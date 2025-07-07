import { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { cyberarkApi, CyberArkInstance as ApiCyberArkInstance } from '@/api/cyberark';

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
  const [instances, setInstances] = useState<CyberArkInstance[]>([
    { id: 'all', name: 'All Instances', type: 'overview' }
  ]);
  const [currentInstanceId, setCurrentInstanceId] = useState<string>('all');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchInstances();
  }, []);

  const fetchInstances = async () => {
    try {
      const response = await cyberarkApi.listInstances(true); // Only get active instances
      const apiInstances = response.instances || [];
      
      // Convert API instances to context format
      const contextInstances: CyberArkInstance[] = [
        { id: 'all', name: 'All Instances', type: 'overview' },
        ...apiInstances.map(inst => ({
          id: inst.id,
          name: inst.name,
          type: 'instance' as const,
          url: inst.base_url,
          status: getInstanceStatus(inst),
        }))
      ];
      
      setInstances(contextInstances);
      
      // If current instance is no longer in the list, reset to overview
      if (currentInstanceId !== 'all' && !contextInstances.find(i => i.id === currentInstanceId)) {
        setCurrentInstanceId('all');
      }
    } catch (error) {
      console.error('Failed to fetch CyberArk instances:', error);
    } finally {
      setLoading(false);
    }
  };

  const getInstanceStatus = (instance: ApiCyberArkInstance): 'connected' | 'disconnected' | 'error' => {
    if (!instance.last_test_at) {
      return 'disconnected';
    }
    return instance.last_test_success ? 'connected' : 'error';
  };
  
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