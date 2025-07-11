import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { 
  Clock, 
  RefreshCw, 
  Users, 
  Shield, 
  FolderLock,
  Pause,
  Play,
  AlertCircle,
  CheckCircle2,
  Loader2,
  Info,
  Settings
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { formatDistanceToNow } from "date-fns";
import { syncApi, type SyncSchedule, type EntitySchedule } from "@/api/sync";
import { cn } from "@/lib/utils";
import { PageContainer } from "@/components/PageContainer";
import { PageHeader } from "@/components/PageHeader";
import { cyberarkApi } from "@/api/cyberark";
import { useUpdateCyberArkInstance } from "@/hooks/useCyberArkInstances";

// Helper components
function SyncStatusBadge({ status, lastSyncAt }: { status?: string; lastSyncAt?: string }) {
  if (!status) return null;
  
  const config = {
    success: { icon: CheckCircle2, class: "text-green-600", label: "Success" },
    failed: { icon: AlertCircle, class: "text-red-600", label: "Failed" },
    running: { icon: Loader2, class: "text-blue-600 animate-spin", label: "Syncing" }
  }[status] || { icon: Info, class: "text-gray-600", label: "Unknown" };
  
  const Icon = config.icon;
  
  return (
    <div className="flex items-center gap-2">
      <Icon className={`h-4 w-4 ${config.class}`} />
      <div className="text-sm">
        <div className="font-medium">{config.label}</div>
        {lastSyncAt && (
          <div className="text-muted-foreground">
            {formatDistanceToNow(new Date(lastSyncAt), { addSuffix: true })}
          </div>
        )}
      </div>
    </div>
  );
}

function EntityIcon({ type }: { type: string }) {
  const icons = {
    users: Users,
    groups: Shield,
    safes: FolderLock
  };
  const Icon = icons[type as keyof typeof icons] || Users;
  return <Icon className="w-5 h-5" />;
}

// Page size configuration component
function PageSizeConfig({ instanceId, currentSize }: { instanceId: string; currentSize?: number }) {
  const [pageSize, setPageSize] = useState(currentSize || 100);
  const [isOpen, setIsOpen] = useState(false);
  const updateInstance = useUpdateCyberArkInstance();
  
  const handleSave = async () => {
    await updateInstance.mutateAsync({
      id: instanceId,
      data: { user_sync_page_size: pageSize }
    });
    setIsOpen(false);
  };
  
  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
          <Settings className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="end">
        <div className="space-y-4">
          <div className="space-y-2">
            <h4 className="font-medium leading-none">User Sync Configuration</h4>
            <p className="text-sm text-muted-foreground">
              Configure pagination settings for user synchronization
            </p>
          </div>
          <div className="space-y-2">
            <Label htmlFor="pageSize">Page Size</Label>
            <Input
              id="pageSize"
              type="number"
              min="1"
              max="1000"
              value={pageSize}
              onChange={(e) => setPageSize(Number(e.target.value))}
              placeholder="100"
            />
            <p className="text-xs text-muted-foreground">
              Number of users to fetch per request (1-1000)
            </p>
          </div>
          <div className="flex justify-end gap-2">
            <Button size="sm" variant="outline" onClick={() => setIsOpen(false)}>
              Cancel
            </Button>
            <Button 
              size="sm" 
              onClick={handleSave}
              disabled={updateInstance.isPending || pageSize < 1 || pageSize > 1000}
            >
              {updateInstance.isPending ? "Saving..." : "Save"}
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

export default function InstanceSync() {
  const queryClient = useQueryClient();
  const [globalPaused, setGlobalPaused] = useState(false);

  // Fetch sync schedules
  const { data: schedules = [], isLoading, isFetching, refetch } = useQuery({
    queryKey: ["sync-schedules"],
    queryFn: syncApi.getSchedules,
    refetchInterval: 30000, // Refresh every 30 seconds
    refetchIntervalInBackground: false, // Don't refetch when tab is not visible
    refetchOnWindowFocus: false, // Don't refetch on window focus
    staleTime: 10000, // Consider data stale after 10 seconds
    gcTime: 5 * 60 * 1000, // Keep in cache for 5 minutes
    retry: 1 // Only retry once on failure
  });

  // Mutations
  const updateSchedule = useMutation({
    mutationFn: ({ instanceId, data }: { instanceId: string; data: Partial<SyncSchedule> }) =>
      syncApi.updateSchedule(instanceId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sync-schedules"] });
    }
  });

  const updateEntitySchedule = useMutation({
    mutationFn: ({ instanceId, entityType, data }: { instanceId: string; entityType: string; data: Partial<EntitySchedule> }) =>
      syncApi.updateEntitySchedule(instanceId, entityType, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sync-schedules"] });
    }
  });

  const triggerSync = useMutation({
    mutationFn: ({ instanceId, entityType }: { instanceId: string; entityType: string }) =>
      syncApi.triggerSync(instanceId, entityType),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sync-schedules"] });
    }
  });

  const pauseAll = useMutation({
    mutationFn: () => globalPaused ? syncApi.resumeAll() : syncApi.pauseAll(),
    onSuccess: () => {
      setGlobalPaused(!globalPaused);
      // Delay invalidation to prevent immediate refetch
      setTimeout(() => {
        queryClient.invalidateQueries({ queryKey: ["sync-schedules"] });
      }, 500);
    }
  });

  // Calculate summary stats
  const stats = schedules?.reduce(
    (acc, schedule) => {
      if (schedule.enabled) {
        acc.activeInstances++;
        schedule.schedules.forEach(s => {
          if (s.lastStatus === "failed") acc.failedSyncs++;
          acc.totalRecords += s.recordCount || 0;
        });
      }
      return acc;
    },
    { activeInstances: 0, failedSyncs: 0, totalRecords: 0 }
  ) || { activeInstances: 0, failedSyncs: 0, totalRecords: 0 };

  const intervalOptions = [
    { value: "15", label: "Every 15 minutes" },
    { value: "30", label: "Every 30 minutes" },
    { value: "60", label: "Every hour" },
    { value: "120", label: "Every 2 hours" },
    { value: "240", label: "Every 4 hours" },
    { value: "0", label: "Disabled" }
  ];

  if (isLoading) {
    return (
      <PageContainer>
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin" />
        </div>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <PageHeader
        title="Instance Synchronization"
        description="Manage automated data synchronization schedules for CyberArk instances"
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="default"
              onClick={() => refetch()}
              disabled={isFetching}
            >
              <RefreshCw className={cn("h-4 w-4 mr-2", isFetching && "animate-spin")} />
              Refresh
            </Button>
            <Button
              variant={globalPaused ? "default" : "outline"}
              size="default"
              onClick={() => pauseAll.mutate()}
            >
              {globalPaused ? (
                <>
                  <Play className="h-4 w-4 mr-2" />
                  Resume All
                </>
              ) : (
                <>
                  <Pause className="h-4 w-4 mr-2" />
                  Pause All
                </>
              )}
            </Button>
          </div>
        }
      />

      {/* Summary Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base font-medium">Active Syncs</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats.activeInstances}/{schedules?.length || 0}
            </div>
            <p className="text-sm text-muted-foreground">instances</p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base font-medium">Failed (24h)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{stats.failedSyncs}</div>
            <p className="text-sm text-muted-foreground">sync failures</p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base font-medium">Records Synced</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalRecords.toLocaleString()}</div>
            <p className="text-sm text-muted-foreground">total records</p>
          </CardContent>
        </Card>
      </div>

      {/* Instance Sync Cards */}
      <div className="space-y-4">
        {schedules?.map((schedule) => (
          <Card key={schedule.instanceId} className={!schedule.enabled ? "opacity-60" : ""}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <CardTitle className="text-xl">{schedule.instanceName}</CardTitle>
                  <Badge variant={schedule.enabled ? "default" : "secondary"}>
                    {schedule.enabled ? "Enabled" : "Disabled"}
                  </Badge>
                </div>
                <Switch
                  checked={schedule.enabled}
                  onCheckedChange={(checked) =>
                    updateSchedule.mutate({
                      instanceId: schedule.instanceId,
                      data: { enabled: checked }
                    })
                  }
                />
              </div>
              <CardDescription>Instance ID: {schedule.instanceId}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {schedule.schedules.map((entitySchedule) => (
                  <div
                    key={entitySchedule.entityType}
                    className="flex items-center justify-between p-4 border rounded-lg"
                  >
                    <div className="flex items-center gap-4">
                      <EntityIcon type={entitySchedule.entityType} />
                      <div>
                        <div className="font-medium capitalize">{entitySchedule.entityType}</div>
                        {entitySchedule.recordCount !== undefined && (
                          <div className="text-sm text-muted-foreground">
                            {entitySchedule.recordCount.toLocaleString()} records
                          </div>
                        )}
                        {entitySchedule.entityType === "users" && schedule.userSyncPageSize && (
                          <div className="text-sm text-muted-foreground">
                            Page size: {schedule.userSyncPageSize}
                          </div>
                        )}
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-6">
                      <div className="flex flex-col gap-1">
                        <SyncStatusBadge 
                          status={entitySchedule.lastStatus} 
                          lastSyncAt={entitySchedule.lastSyncAt}
                        />
                        {entitySchedule.enabled && entitySchedule.nextSyncAt && (
                          <div className="flex items-center gap-1 text-xs text-muted-foreground">
                            <Clock className="h-3 w-3" />
                            <span>Next: {formatDistanceToNow(new Date(entitySchedule.nextSyncAt), { addSuffix: true })}</span>
                          </div>
                        )}
                      </div>
                      
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4 text-muted-foreground" />
                        <Select
                          value={entitySchedule.interval.toString()}
                          onValueChange={(value) => {
                            updateEntitySchedule.mutate({
                              instanceId: schedule.instanceId,
                              entityType: entitySchedule.entityType,
                              data: { interval: parseInt(value) }
                            });
                          }}
                        >
                          <SelectTrigger className="w-[180px]">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {intervalOptions.map((option) => (
                              <SelectItem key={option.value} value={option.value}>
                                {option.label}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() =>
                          triggerSync.mutate({
                            instanceId: schedule.instanceId,
                            entityType: entitySchedule.entityType
                          })
                        }
                        disabled={!schedule.enabled || !entitySchedule.enabled}
                      >
                        <RefreshCw className="h-4 w-4 mr-2" />
                        Sync Now
                      </Button>
                      
                      {entitySchedule.entityType === "users" && (
                        <PageSizeConfig 
                          instanceId={schedule.instanceId} 
                          currentSize={schedule.userSyncPageSize}
                        />
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Empty state */}
      {schedules.length === 0 && !isLoading && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Shield className="w-12 h-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium">No instances configured</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Add CyberArk instances to begin synchronizing data
            </p>
          </CardContent>
        </Card>
      )}
    </PageContainer>
  );
}