import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Loader2, Save, Clock, RefreshCw, FileText, Users, Shield } from 'lucide-react';
import { useInstanceSyncConfigs, useUpdateSyncConfig } from '@/hooks/useSyncJobs';
import { format } from 'date-fns';
import { Badge } from '@/components/ui/badge';

interface SyncConfigurationProps {
  instanceId: string;
}

const syncTypeConfig = {
  users: {
    label: 'Users',
    icon: Users,
    description: 'Synchronize users and their group memberships',
  },
  safes: {
    label: 'Safes',
    icon: Shield,
    description: 'Synchronize safe configurations and metadata',
  },
  groups: {
    label: 'Groups',
    icon: FileText,
    description: 'Synchronize group definitions and settings',
  },
} as const;

export function SyncConfiguration({ instanceId }: SyncConfigurationProps) {
  const { data: configsData, isLoading } = useInstanceSyncConfigs(instanceId);
  const updateConfig = useUpdateSyncConfig();
  
  const [editedConfigs, setEditedConfigs] = useState<Record<string, any>>({});
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    if (configsData?.configs) {
      setEditedConfigs(configsData.configs);
    }
  }, [configsData]);

  const handleConfigChange = (syncType: string, field: string, value: any) => {
    setEditedConfigs(prev => ({
      ...prev,
      [syncType]: {
        ...prev[syncType],
        [field]: value,
      },
    }));
    setHasChanges(true);
  };

  const handleSave = async (syncType: 'users' | 'safes' | 'groups') => {
    const config = editedConfigs[syncType];
    if (!config) return;

    try {
      await updateConfig.mutateAsync({
        instanceId,
        syncType,
        data: {
          enabled: config.enabled,
          interval_minutes: parseInt(config.interval_minutes),
          page_size: parseInt(config.page_size),
          retry_attempts: parseInt(config.retry_attempts),
          timeout_minutes: parseInt(config.timeout_minutes),
        },
      });
      setHasChanges(false);
    } catch (error) {
      console.error('Failed to update sync config:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold">Synchronization Settings</h3>
      
      {Object.entries(syncTypeConfig).map(([syncType, typeInfo]) => {
        const config = editedConfigs[syncType as keyof typeof editedConfigs];
        if (!config) return null;

        const Icon = typeInfo.icon;
        const isUpdating = updateConfig.isPending && 
          updateConfig.variables?.syncType === syncType;

        return (
          <Card key={syncType}>
            <CardHeader className="pb-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Icon className="h-5 w-5 text-gray-600" />
                  <div>
                    <CardTitle className="text-base">{typeInfo.label} Sync</CardTitle>
                    <p className="text-sm text-gray-500 mt-1">{typeInfo.description}</p>
                  </div>
                </div>
                <Switch
                  checked={config.enabled}
                  onCheckedChange={(checked) => handleConfigChange(syncType, 'enabled', checked)}
                />
              </div>
            </CardHeader>
            
            <CardContent className="space-y-4">
              {/* Last run info */}
              {config.last_run_at && (
                <div className="flex items-center justify-between text-sm bg-gray-50 p-3 rounded-md">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-gray-500" />
                    <span className="text-gray-600">Last run:</span>
                    <span className="font-medium">
                      {format(new Date(config.last_run_at), 'MMM d, HH:mm')}
                    </span>
                  </div>
                  {config.last_run_status && (
                    <Badge 
                      variant={config.last_run_status === 'completed' ? 'success' : 'destructive'}
                      className="text-xs"
                    >
                      {config.last_run_status}
                    </Badge>
                  )}
                </div>
              )}

              {/* Configuration fields */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor={`${syncType}-interval`}>Interval (minutes)</Label>
                  <Input
                    id={`${syncType}-interval`}
                    type="number"
                    min="5"
                    value={config.interval_minutes}
                    onChange={(e) => handleConfigChange(syncType, 'interval_minutes', e.target.value)}
                    disabled={!config.enabled}
                  />
                </div>
                
                <div>
                  <Label htmlFor={`${syncType}-pagesize`}>Page Size</Label>
                  <Input
                    id={`${syncType}-pagesize`}
                    type="number"
                    min="1"
                    max="1000"
                    value={config.page_size}
                    onChange={(e) => handleConfigChange(syncType, 'page_size', e.target.value)}
                    disabled={!config.enabled}
                  />
                </div>
                
                <div>
                  <Label htmlFor={`${syncType}-retries`}>Retry Attempts</Label>
                  <Input
                    id={`${syncType}-retries`}
                    type="number"
                    min="0"
                    max="10"
                    value={config.retry_attempts}
                    onChange={(e) => handleConfigChange(syncType, 'retry_attempts', e.target.value)}
                    disabled={!config.enabled}
                  />
                </div>
                
                <div>
                  <Label htmlFor={`${syncType}-timeout`}>Timeout (minutes)</Label>
                  <Input
                    id={`${syncType}-timeout`}
                    type="number"
                    min="1"
                    max="120"
                    value={config.timeout_minutes}
                    onChange={(e) => handleConfigChange(syncType, 'timeout_minutes', e.target.value)}
                    disabled={!config.enabled}
                  />
                </div>
              </div>

              {/* Next run info */}
              {config.enabled && config.next_run_at && (
                <div className="flex items-center gap-2 text-sm text-gray-600">
                  <RefreshCw className="h-4 w-4" />
                  <span>Next sync: {format(new Date(config.next_run_at), 'MMM d, HH:mm')}</span>
                </div>
              )}

              {/* Save button */}
              <div className="flex justify-end">
                <Button
                  size="sm"
                  disabled={!hasChanges || isUpdating}
                  onClick={() => handleSave(syncType as 'users' | 'safes' | 'groups')}
                >
                  {isUpdating ? (
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  ) : (
                    <Save className="h-4 w-4 mr-2" />
                  )}
                  Save Changes
                </Button>
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}