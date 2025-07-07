import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { format } from 'date-fns';
import { 
  Operation, 
  Status, 
  operationsApi, 
  getOperationTypeLabel,
  getPriorityColor,
  getStatusColor 
} from '../api/operations';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card';
import { PageContainer } from '../components/PageContainer';
import { 
  
  RefreshCw, 
  ChevronRight,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2
} from 'lucide-react';

export default function Operations() {
  const navigate = useNavigate();
  const [operations, setOperations] = useState<Operation[]>([]);
  const [loading, setLoading] = useState(true);
  const [_refreshInterval, _setRefreshInterval] = useState<ReturnType<typeof setInterval> | null>(null);

  const fetchOperations = async () => {
    try {
      const response = await operationsApi.list();
      setOperations(response.operations);
    } catch (error) {
      console.error('Failed to fetch operations:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOperations();
    
    // Set up auto-refresh every 5 seconds
    const interval = setInterval(fetchOperations, 5000);
    _setRefreshInterval(interval);
    
    return () => {
      if (interval) clearInterval(interval);
    };
  }, []);

  const getStatusIcon = (status: Status) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-4 w-4" />;
      case 'processing':
        return <Loader2 className="h-4 w-4 animate-spin" />;
      case 'completed':
        return <CheckCircle className="h-4 w-4" />;
      case 'failed':
        return <XCircle className="h-4 w-4" />;
      case 'cancelled':
        return <AlertCircle className="h-4 w-4" />;
    }
  };

  const handleCancelOperation = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await operationsApi.cancel(id);
      fetchOperations();
    } catch (error) {
      console.error('Failed to cancel operation:', error);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <PageContainer>
      <div className="flex justify-end gap-2">
        <Button
          variant="outline"
          onClick={fetchOperations}
        >
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      {/* Operations List */}
      <Card>
        <CardHeader>
          <CardTitle>Operations Queue</CardTitle>
        </CardHeader>
        <CardContent>
          {operations.length === 0 ? (
            <p className="text-center text-gray-500 py-8">No operations found</p>
          ) : (
            <div className="space-y-2">
              {operations.map((operation) => (
                <div
                  key={operation.id}
                  className="flex items-center justify-between p-4 border rounded hover:bg-gray-50 cursor-pointer transition-colors"
                  onClick={() => navigate(`/operations/${operation.id}`)}
                >
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(operation.status)}
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(operation.status)}`}>
                        {operation.status}
                      </span>
                    </div>
                    <div>
                      <p className="font-medium">{getOperationTypeLabel(operation.type)}</p>
                      <p className="text-sm text-gray-500">
                        ID: {operation.id} • Created: {format(new Date(operation.created_at), 'MMM d, HH:mm:ss')}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${getPriorityColor(operation.priority)}`}>
                      {operation.priority}
                    </span>
                    {(operation.status === 'pending' || operation.status === 'processing') && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => handleCancelOperation(operation.id, e)}
                      >
                        Cancel
                      </Button>
                    )}
                    <ChevronRight className="h-4 w-4 text-gray-400" />
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

    </PageContainer>
  );
}