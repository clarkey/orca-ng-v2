import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { format } from 'date-fns';
import { 
  Operation, 
  operationsApi, 
  getOperationTypeLabel,
  getPriorityColor,
  getStatusColor 
} from '../api/operations';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card';
import { PageContainer } from '../components/PageContainer';
import { 
  ArrowLeft,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2,
  RefreshCw
} from 'lucide-react';

export default function OperationDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [operation, setOperation] = useState<Operation | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshInterval, setRefreshInterval] = useState<ReturnType<typeof setInterval> | null>(null);

  const fetchOperation = async () => {
    if (!id) return;
    
    try {
      const data = await operationsApi.get(id);
      setOperation(data);
      
      // Stop auto-refresh if operation is in terminal state
      if (data.status === 'completed' || data.status === 'failed' || data.status === 'cancelled') {
        if (refreshInterval) {
          clearInterval(refreshInterval);
          setRefreshInterval(null);
        }
      }
    } catch (error) {
      console.error('Failed to fetch operation:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOperation();
    
    // Set up auto-refresh every 2 seconds for active operations
    const interval = setInterval(fetchOperation, 2000);
    setRefreshInterval(interval);
    
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [id]);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-6 w-6" />;
      case 'processing':
        return <Loader2 className="h-6 w-6 animate-spin" />;
      case 'completed':
        return <CheckCircle className="h-6 w-6 text-green-600" />;
      case 'failed':
        return <XCircle className="h-6 w-6 text-red-600" />;
      case 'cancelled':
        return <AlertCircle className="h-6 w-6 text-orange-600" />;
      default:
        return null;
    }
  };

  const handleCancel = async () => {
    if (!operation) return;
    
    try {
      await operationsApi.cancel(operation.id);
      fetchOperation();
    } catch (error) {
      console.error('Failed to cancel operation:', error);
    }
  };

  if (loading || !operation) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  const duration = operation.completed_at 
    ? new Date(operation.completed_at).getTime() - new Date(operation.created_at).getTime()
    : null;

  return (
    <PageContainer maxWidth="2xl">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          onClick={() => navigate('/operations')}
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Operations
        </Button>
        <Button
          variant="ghost"
          onClick={fetchOperation}
        >
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              {getStatusIcon(operation.status)}
              <div>
                <CardTitle className="text-2xl">
                  {getOperationTypeLabel(operation.type)}
                </CardTitle>
                <p className="text-sm text-gray-500 mt-1">{operation.id}</p>
              </div>
            </div>
            {(operation.status === 'pending' || operation.status === 'processing') && (
              <Button
                variant="destructive"
                onClick={handleCancel}
              >
                Cancel Operation
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Status and Priority */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm font-medium text-gray-500 mb-1">Status</p>
              <span className={`inline-flex px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(operation.status)}`}>
                {operation.status}
              </span>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500 mb-1">Priority</p>
              <span className={`inline-flex px-3 py-1 rounded-full text-sm font-medium ${getPriorityColor(operation.priority)}`}>
                {operation.priority}
              </span>
            </div>
          </div>

          {/* Instance Info */}
          {operation.cyberark_instance_info && (
            <div>
              <p className="text-sm font-medium text-gray-500">CyberArk Instance</p>
              <p className="text-sm">{operation.cyberark_instance_info.name}</p>
            </div>
          )}

          {/* Created By */}
          {operation.created_by_user && (
            <div>
              <p className="text-sm font-medium text-gray-500">Created By</p>
              <p className="text-sm">{operation.created_by_user.username}</p>
            </div>
          )}

          {/* Timestamps */}
          <div className="space-y-2">
            <div>
              <p className="text-sm font-medium text-gray-500">Created</p>
              <p className="text-sm">{format(new Date(operation.created_at), 'PPpp')}</p>
            </div>
            {operation.started_at && (
              <div>
                <p className="text-sm font-medium text-gray-500">Started</p>
                <p className="text-sm">{format(new Date(operation.started_at), 'PPpp')}</p>
              </div>
            )}
            {operation.completed_at && (
              <div>
                <p className="text-sm font-medium text-gray-500">Completed</p>
                <p className="text-sm">
                  {format(new Date(operation.completed_at), 'PPpp')}
                  {duration && (
                    <span className="text-gray-500 ml-2">
                      ({Math.round(duration / 1000)}s duration)
                    </span>
                  )}
                </p>
              </div>
            )}
          </div>

          {/* Payload */}
          {operation.payload && (
            <div>
              <p className="text-sm font-medium text-gray-500 mb-2">Operation Details</p>
              <div className="bg-gray-50 rounded p-4">
                {operation.type === 'user_sync' && operation.payload.sync_mode && (
                  <div className="space-y-1 text-sm">
                    <p><span className="font-medium">Sync Mode:</span> {operation.payload.sync_mode}</p>
                    {operation.payload.page_size && (
                      <p><span className="font-medium">Page Size:</span> {operation.payload.page_size}</p>
                    )}
                  </div>
                )}
                {operation.type !== 'user_sync' && (
                  <pre className="text-sm overflow-auto">
                    {JSON.stringify(operation.payload, null, 2)}
                  </pre>
                )}
              </div>
            </div>
          )}

          {/* Error Message */}
          {operation.error_message && (
            <div className="bg-red-50 border border-red-200 rounded p-4">
              <p className="text-sm font-medium text-red-800 mb-1">Error</p>
              <p className="text-sm text-red-700">{operation.error_message}</p>
            </div>
          )}

          {/* Result */}
          {operation.result && (
            <div>
              <p className="text-sm font-medium text-gray-500 mb-2">Result</p>
              <div className="bg-gray-50 rounded p-4">
                {operation.type === 'user_sync' && operation.result.total_users !== undefined && (
                  <div className="space-y-2 text-sm">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <p className="font-medium text-gray-600">Total Users</p>
                        <p className="text-2xl font-semibold">{operation.result.total_users}</p>
                      </div>
                      <div>
                        <p className="font-medium text-gray-600">Processed</p>
                        <p className="text-2xl font-semibold">{operation.result.processed_users || 0}</p>
                      </div>
                    </div>
                    <div className="grid grid-cols-3 gap-4 mt-4">
                      <div>
                        <p className="font-medium text-gray-600">New Users</p>
                        <p className="text-lg font-semibold text-green-600">{operation.result.new_users || 0}</p>
                      </div>
                      <div>
                        <p className="font-medium text-gray-600">Updated</p>
                        <p className="text-lg font-semibold text-blue-600">{operation.result.updated_users || 0}</p>
                      </div>
                      <div>
                        <p className="font-medium text-gray-600">Deleted</p>
                        <p className="text-lg font-semibold text-red-600">{operation.result.deleted_users || 0}</p>
                      </div>
                    </div>
                    {operation.result.errors && operation.result.errors.length > 0 && (
                      <div className="mt-4">
                        <p className="font-medium text-gray-600 mb-2">Errors</p>
                        <ul className="list-disc list-inside space-y-1">
                          {operation.result.errors.map((error: string, index: number) => (
                            <li key={index} className="text-red-600">{error}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                )}
                {(operation.type !== 'user_sync' || operation.result.total_users === undefined) && (
                  <pre className="text-sm overflow-auto">
                    {JSON.stringify(operation.result, null, 2)}
                  </pre>
                )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  );
}