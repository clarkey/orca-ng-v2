import { useState } from 'react';
import { useTriggerSync } from '@/hooks/useSyncJobs';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { RefreshCw, Users, Shield, FileText, Loader2 } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { syncJobsKeys } from '@/hooks/useSyncJobs';

interface InstanceSyncActionsProps {
  instanceId: string;
  instanceName: string;
  size?: 'sm' | 'default';
}

export function InstanceSyncActions({ instanceId, instanceName, size = 'sm' }: InstanceSyncActionsProps) {
  const [syncing, setSyncing] = useState<string | null>(null);
  const triggerSync = useTriggerSync();
  const queryClient = useQueryClient();

  const handleSync = async (syncType: 'users' | 'safes' | 'groups') => {
    setSyncing(syncType);
    try {
      await triggerSync.mutateAsync({
        instance_id: instanceId,
        sync_type: syncType,
      });
      
      // Invalidate sync configs to refresh status
      queryClient.invalidateQueries({ queryKey: syncJobsKeys.config(instanceId) });
      
      // Show success message (you might want to use a toast library)
      alert(`${syncType} sync triggered for ${instanceName}`);
    } catch (error) {
      console.error('Failed to trigger sync:', error);
      alert('Failed to trigger sync');
    } finally {
      setSyncing(null);
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size={size} disabled={syncing !== null}>
          {syncing ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <RefreshCw className="h-4 w-4" />
          )}
          <span className="ml-2">Sync</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>Trigger Sync</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem 
          onClick={() => handleSync('users')}
          disabled={syncing !== null}
        >
          <Users className="h-4 w-4 mr-2" />
          Sync Users
        </DropdownMenuItem>
        <DropdownMenuItem 
          onClick={() => handleSync('safes')}
          disabled={syncing !== null}
        >
          <Shield className="h-4 w-4 mr-2" />
          Sync Safes
        </DropdownMenuItem>
        <DropdownMenuItem 
          onClick={() => handleSync('groups')}
          disabled={syncing !== null}
        >
          <FileText className="h-4 w-4 mr-2" />
          Sync Groups
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}