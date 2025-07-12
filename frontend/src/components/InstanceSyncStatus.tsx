import { useInstanceSyncConfigs } from '@/hooks/useSyncJobs';
import { Badge } from '@/components/ui/badge';
import { Clock, CheckCircle, XCircle, AlertCircle, Users, Shield, FileText } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { cn } from '@/lib/utils';

interface InstanceSyncStatusProps {
  instanceId: string;
  className?: string;
}

const syncTypeConfig = {
  users: { icon: Users, label: 'Users' },
  safes: { icon: Shield, label: 'Safes' },
  groups: { icon: FileText, label: 'Groups' },
};

export function InstanceSyncStatus({ instanceId, className }: InstanceSyncStatusProps) {
  const { data: syncConfigs, isLoading } = useInstanceSyncConfigs(instanceId);

  if (isLoading || !syncConfigs?.configs) {
    return null;
  }

  const configs = syncConfigs.configs;
  const syncTypes = ['users', 'safes', 'groups'] as const;

  return (
    <div className={cn("flex flex-wrap gap-2", className)}>
      {syncTypes.map((syncType) => {
        const config = configs[syncType];
        if (!config) return null;

        const typeConfig = syncTypeConfig[syncType];
        const Icon = typeConfig.icon;

        // Determine status
        let statusIcon = AlertCircle;
        let statusColor = 'text-gray-400';
        let statusText = 'Not synced';
        
        if (!config.enabled) {
          statusIcon = XCircle;
          statusColor = 'text-gray-400';
          statusText = 'Disabled';
        } else if (config.last_run_at) {
          const lastRunDate = new Date(config.last_run_at);
          const hoursSinceRun = (Date.now() - lastRunDate.getTime()) / (1000 * 60 * 60);
          
          if (config.last_run_status === 'completed') {
            if (hoursSinceRun < 1) {
              statusIcon = CheckCircle;
              statusColor = 'text-green-500';
              statusText = 'Synced';
            } else if (hoursSinceRun < 24) {
              statusIcon = Clock;
              statusColor = 'text-yellow-500';
              statusText = formatDistanceToNow(lastRunDate, { addSuffix: true });
            } else {
              statusIcon = AlertCircle;
              statusColor = 'text-orange-500';
              statusText = 'Stale';
            }
          } else {
            statusIcon = XCircle;
            statusColor = 'text-red-500';
            statusText = 'Failed';
          }
        }

        const StatusIcon = statusIcon;

        return (
          <div
            key={syncType}
            className="flex items-center gap-1.5 text-xs"
            title={`${typeConfig.label}: ${statusText}${config.last_run_at ? ` (${formatDistanceToNow(new Date(config.last_run_at), { addSuffix: true })})` : ''}`}
          >
            <Icon className="h-3.5 w-3.5 text-gray-500" />
            <StatusIcon className={cn("h-3.5 w-3.5", statusColor)} />
          </div>
        );
      })}
    </div>
  );
}