import { getOperationTypeLabel, OperationType } from '../api/operations';
import { usePipelineData } from '@/hooks/usePipelineMetrics';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card';
import { Progress } from '../components/ui/progress';
import { PageContainer } from '../components/PageContainer';
import { PageHeader } from '../components/PageHeader';
import { 
  Activity,
  Zap,
  Clock,
  CheckCircle,
  XCircle,
  RefreshCw,
  Settings,
  Loader2
} from 'lucide-react';
import { Button } from '../components/ui/button';

export default function PipelineDashboard() {
  const { metrics, config, isLoading, refetch } = usePipelineData(5000); // 5 second refresh

  if (isLoading || !metrics || !config) {
    return (
      <PageContainer>
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin" />
        </div>
      </PageContainer>
    );
  }

  const totalQueued = Object.values(metrics.queue_depth || {}).reduce((a, b) => a + b, 0);
  const totalProcessing = Object.values(metrics.processing_count || {}).reduce((a, b) => a + b, 0);
  const totalCompleted = Object.values(metrics.completed_count || {}).reduce((a, b) => a + b, 0);
  const totalFailed = Object.values(metrics.failed_count || {}).reduce((a, b) => a + b, 0);

  const priorityOrder = ['high', 'normal', 'low'] as const;

  return (
    <PageContainer>
      <PageHeader
        title="Queue Monitoring"
        description="Monitor and manage operation processing queues"
        actions={
          <>
            <Button variant="outline" onClick={() => refetch()}>
              <RefreshCw className="h-4 w-4" />
            </Button>
            <Button variant="outline">
              <Settings className="h-4 w-4 mr-2" />
              Configure
            </Button>
          </>
        }
      />

      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Capacity Utilization</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(metrics.worker_utilization * 100).toFixed(1)}%
            </div>
            <Progress 
              value={metrics.worker_utilization * 100} 
              className="mt-2"
            />
            <p className="text-xs text-muted-foreground mt-2">
              {totalProcessing} of {config.total_capacity} workers active
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Queued Operations</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalQueued}</div>
            <p className="text-xs text-muted-foreground mt-2">
              Waiting to be processed
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Completed</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{totalCompleted}</div>
            <p className="text-xs text-muted-foreground mt-2">
              Successfully processed
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Failed</CardTitle>
            <XCircle className="h-4 w-4 text-red-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{totalFailed}</div>
            <p className="text-xs text-muted-foreground mt-2">
              Failed operations
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Priority Lanes */}
      <Card>
        <CardHeader>
          <CardTitle>Priority Lanes</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {priorityOrder.map((priority) => {
            const queueDepth = metrics.queue_depth?.[priority] || 0;
            const processing = metrics.processing_count?.[priority] || 0;
            const capacity = Math.round(config.total_capacity * (config.priority_allocation?.[priority] || 0));
            const utilization = capacity > 0 ? (processing / capacity) * 100 : 0;

            return (
              <div key={priority} className="space-y-2">
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-2">
                    <Zap className={`h-4 w-4 ${
                      priority === 'high' ? 'text-red-600' :
                      priority === 'normal' ? 'text-blue-600' :
                      'text-gray-600'
                    }`} />
                    <span className="font-medium capitalize">{priority} Priority</span>
                  </div>
                  <div className="text-sm text-gray-500">
                    {processing}/{capacity} workers • {queueDepth} queued
                  </div>
                </div>
                <Progress value={utilization} className="h-2" />
              </div>
            );
          })}
        </CardContent>
      </Card>

      {/* Operation Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Operations by Type</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {Object.entries(metrics.completed_count || {}).map(([type, count]) => {
                const failed = metrics.failed_count?.[type as OperationType] || 0;
                const total = count + failed;
                const successRate = total > 0 ? (count / total) * 100 : 0;
                const avgTime = metrics.avg_processing_time?.[type as OperationType] || 0;

                return (
                  <div key={type} className="space-y-1">
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">
                        {getOperationTypeLabel(type as OperationType)}
                      </span>
                      <span className="text-sm text-gray-500">
                        {count} completed
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Progress value={successRate} className="flex-1 h-2" />
                      <span className="text-xs text-gray-500 w-16 text-right">
                        {successRate.toFixed(0)}%
                      </span>
                    </div>
                    {avgTime > 0 && (
                      <p className="text-xs text-gray-500">
                        Avg time: {avgTime.toFixed(1)}s
                      </p>
                    )}
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-gray-500">Total Capacity</p>
              <p className="text-lg font-semibold">{config.total_capacity} concurrent operations</p>
            </div>
            
            <div>
              <p className="text-sm font-medium text-gray-500 mb-2">Priority Allocation</p>
              <div className="space-y-1">
                {priorityOrder.map((priority) => (
                  <div key={priority} className="flex justify-between text-sm">
                    <span className="capitalize">{priority}:</span>
                    <span>{(config.priority_allocation[priority] * 100).toFixed(0)}%</span>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <p className="text-sm font-medium text-gray-500">Retry Policy</p>
              <div className="text-sm space-y-1 mt-1">
                <p>Max attempts: {config.retry_policy.max_attempts}</p>
                <p>Backoff: {config.retry_policy.backoff_base_seconds}s × {config.retry_policy.backoff_multiplier}</p>
                <p>Jitter: {config.retry_policy.backoff_jitter ? 'Enabled' : 'Disabled'}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  );
}