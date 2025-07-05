import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { format } from 'date-fns';
import { 
  Operation, 
  Priority, 
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../components/ui/dialog';
import { 
  Plus, 
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
  const [statusFilter, setStatusFilter] = useState<Status | 'all'>('all');
  const [priorityFilter, setPriorityFilter] = useState<Priority | 'all'>('all');
  const [refreshInterval, setRefreshInterval] = useState<NodeJS.Timer | null>(null);
  const [showNewOperationDialog, setShowNewOperationDialog] = useState(false);
  const [newOperationType, setNewOperationType] = useState<OperationType>('user_sync');
  const [newOperationPriority, setNewOperationPriority] = useState<Priority>('normal');
  const [creatingOperation, setCreatingOperation] = useState(false);

  const fetchOperations = async () => {
    try {
      const params: any = {};
      if (statusFilter !== 'all') params.status = statusFilter;
      if (priorityFilter !== 'all') params.priority = priorityFilter;
      
      const response = await operationsApi.list(params);
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
    setRefreshInterval(interval);
    
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [statusFilter, priorityFilter]);

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

  const handleCreateOperation = async () => {
    setCreatingOperation(true);
    try {
      // Create a sample payload based on operation type
      let payload: any = {};
      
      switch (newOperationType) {
        case 'user_sync':
        case 'safe_sync':
        case 'group_sync':
          payload = {
            cyberark_instance_id: 'inst_01ABC',
            sync_type: 'full'
          };
          break;
        case 'safe_provision':
          payload = {
            safe_name: `TestSafe_${Date.now()}`,
            description: 'Test safe created from UI',
            cyberark_instance_id: 'inst_01ABC',
            permissions: []
          };
          break;
        default:
          payload = { test: true };
      }

      await operationsApi.create({
        type: newOperationType,
        priority: newOperationPriority,
        payload
      });

      setShowNewOperationDialog(false);
      fetchOperations();
    } catch (error) {
      console.error('Failed to create operation:', error);
      // In a real app, show an error toast
    } finally {
      setCreatingOperation(false);
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
          onClick={() => setShowNewOperationDialog(true)}
        >
          <Plus className="h-4 w-4 mr-2" />
          New Operation
        </Button>
        <Button
          variant="outline"
          onClick={fetchOperations}
        >
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <div className="flex-1">
              <label className="text-sm font-medium mb-2 block">Status</label>
              <Select value={statusFilter} onValueChange={(value: any) => setStatusFilter(value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                  <SelectItem value="processing">Processing</SelectItem>
                  <SelectItem value="completed">Completed</SelectItem>
                  <SelectItem value="failed">Failed</SelectItem>
                  <SelectItem value="cancelled">Cancelled</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex-1">
              <label className="text-sm font-medium mb-2 block">Priority</label>
              <Select value={priorityFilter} onValueChange={(value: any) => setPriorityFilter(value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Priorities</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="normal">Normal</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

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
                  className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 cursor-pointer transition-colors"
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

      {/* New Operation Dialog */}
      <Dialog open={showNewOperationDialog} onOpenChange={setShowNewOperationDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Operation</DialogTitle>
            <DialogDescription>
              Create a new operation to be processed by the pipeline.
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Operation Type</label>
              <Select 
                value={newOperationType} 
                onValueChange={(value) => setNewOperationType(value as OperationType)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="safe_provision">Safe Provision</SelectItem>
                  <SelectItem value="user_sync">User Sync</SelectItem>
                  <SelectItem value="safe_sync">Safe Sync</SelectItem>
                  <SelectItem value="group_sync">Group Sync</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="space-y-2">
              <label className="text-sm font-medium">Priority</label>
              <Select 
                value={newOperationPriority}
                onValueChange={(value) => setNewOperationPriority(value as Priority)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="normal">Normal</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="text-sm text-gray-500">
              Note: This is a demo dialog. In a real implementation, you would have 
              specific fields based on the operation type selected.
            </div>
          </div>
          
          <DialogFooter>
            <Button 
              variant="outline" 
              onClick={() => setShowNewOperationDialog(false)}
              disabled={creatingOperation}
            >
              Cancel
            </Button>
            <Button 
              onClick={handleCreateOperation}
              disabled={creatingOperation}
            >
              {creatingOperation ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Creating...
                </>
              ) : (
                'Create Operation'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageContainer>
  );
}