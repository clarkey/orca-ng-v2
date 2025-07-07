import React from 'react';
import { format, formatDistanceToNow, formatDistance } from 'date-fns';
import { Operation, getOperationTypeLabel, operationsApi } from '../api/operations';
import { Button } from './ui/button';
import {
  Dialog,
  DialogContent,
  DialogFooter,
} from './ui/dialog';
import { DialogHeaderStyled } from './ui/dialog-header-styled';
import { Clock, Calendar, AlertCircle, Ban, RotateCcw } from 'lucide-react';

interface OperationDetailsPanelProps {
  operation: Operation | null;
  onClose: () => void;
  onUpdate?: () => void;
}

export function OperationDetailsPanel({ operation, onClose, onUpdate }: OperationDetailsPanelProps) {
  const [isCancelling, setIsCancelling] = React.useState(false);
  
  if (!operation) return null;

  const scheduledTime = new Date(operation.scheduled_at);
  const now = new Date();
  const isScheduledForFuture = operation.status === 'pending' && scheduledTime > now;
  
  const handleCancel = async () => {
    if (!confirm('Are you sure you want to cancel this operation?')) {
      return;
    }
    
    setIsCancelling(true);
    try {
      await operationsApi.cancel(operation.id);
      onUpdate?.();
      onClose();
    } catch (error) {
      console.error('Failed to cancel operation:', error);
      alert('Failed to cancel operation. Please try again.');
    } finally {
      setIsCancelling(false);
    }
  };
  
  const handleRetry = async () => {
    // TODO: Implement retry functionality
    console.log('Retry operation:', operation.id);
  };

  return (
    <Dialog open={!!operation} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-2xl p-0 overflow-hidden">
        <DialogHeaderStyled 
          title="Operation Details"
          description={operation.id}
        />

        <div className="px-6 pb-6 pt-2 max-h-[70vh] overflow-y-auto">
          <div className="space-y-6">
            {/* Scheduled Time Alert for Future Operations */}
            {isScheduledForFuture && (
              <div className="flex items-start gap-3 p-3 bg-blue-50 border border-blue-200 rounded">
                <Clock className="h-5 w-5 text-blue-600 mt-0.5" />
                <div className="flex-1">
                  <p className="text-sm font-medium text-blue-900">Scheduled Operation</p>
                  <p className="text-sm text-blue-700">
                    This operation will start {formatDistanceToNow(scheduledTime, { addSuffix: true })}
                  </p>
                  <p className="text-xs text-blue-600 mt-1">
                    {format(scheduledTime, 'PPpp')}
                  </p>
                </div>
              </div>
            )}

            {/* Status and Priority */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium text-gray-500">Status</label>
                <div className="mt-1">
                  <span className={`inline-block px-2 py-0.5 rounded text-[10px] font-medium uppercase ${
                    operation.status === 'completed' ? 'bg-green-100 text-green-700 border border-green-200' :
                    operation.status === 'failed' ? 'bg-red-100 text-red-700 border border-red-200' :
                    operation.status === 'processing' ? 'bg-blue-100 text-blue-700 border border-blue-200' :
                    operation.status === 'cancelled' ? 'bg-orange-100 text-orange-700 border border-orange-200' :
                    'bg-gray-100 text-gray-600 border border-gray-200'
                  }`}>
                    {operation.status}
                  </span>
                </div>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-500">Priority</label>
                <div className="mt-1">
                  <span className={`inline-block px-2 py-0.5 rounded text-[10px] font-medium uppercase ${
                    operation.priority === 'high' ? 'bg-red-100 text-red-700 border border-red-200' :
                    operation.priority === 'normal' ? 'bg-blue-100 text-blue-700 border border-blue-200' :
                    'bg-gray-100 text-gray-600 border border-gray-200'
                  }`}>
                    {operation.priority}
                  </span>
                </div>
              </div>
            </div>

            {/* Type */}
            <div>
              <label className="text-sm font-medium text-gray-500">Operation Type</label>
              <p className="mt-1 text-sm text-gray-900">{getOperationTypeLabel(operation.type)}</p>
            </div>

            {/* Timestamps */}
            <div className="space-y-3">
              <div>
                <label className="text-sm font-medium text-gray-500">Created</label>
                <p className="mt-1 text-sm text-gray-900">
                  {format(new Date(operation.created_at), 'PPpp')}
                </p>
              </div>
              
              {/* Show scheduled time for pending operations or if scheduled_at differs from created_at */}
              {operation.status === 'pending' && (
                <div>
                  <label className="text-sm font-medium text-gray-500">Scheduled For</label>
                  <p className="mt-1 text-sm text-gray-900">
                    {format(scheduledTime, 'PPpp')}
                  </p>
                  {!isScheduledForFuture && (
                    <p className="text-xs text-gray-500">
                      Waiting to start...
                    </p>
                  )}
                </div>
              )}
              
              {/* Show when operation actually started processing */}
              {operation.started_at && (
                <div>
                  <label className="text-sm font-medium text-gray-500">Started Processing</label>
                  <p className="mt-1 text-sm text-gray-900">
                    {format(new Date(operation.started_at), 'PPpp')}
                  </p>
                  {/* Show delay if started later than scheduled */}
                  {operation.status !== 'pending' && Math.abs(new Date(operation.started_at).getTime() - scheduledTime.getTime()) > 60000 && (
                    <p className="text-xs text-gray-500">
                      Started {formatDistance(scheduledTime, new Date(operation.started_at), { addSuffix: true })} scheduled time
                    </p>
                  )}
                </div>
              )}
              
              {operation.completed_at && (
                <div>
                  <label className="text-sm font-medium text-gray-500">Completed</label>
                  <p className="mt-1 text-sm text-gray-900">
                    {format(new Date(operation.completed_at), 'PPpp')}
                  </p>
                  {operation.started_at && (
                    <p className="text-xs text-gray-500">
                      Duration: {formatDistance(new Date(operation.started_at), new Date(operation.completed_at), { includeSeconds: true })}
                    </p>
                  )}
                </div>
              )}
            </div>

            {/* Created By */}
            {operation.created_by_user && (
              <div>
                <label className="text-sm font-medium text-gray-500">Created By</label>
                <p className="mt-1 text-sm text-gray-900">{operation.created_by_user.username}</p>
                <p className="text-xs text-gray-500 font-mono">{operation.created_by_user.id}</p>
              </div>
            )}

            {/* Error Message */}
            {operation.error_message && (
              <div>
                <label className="text-sm font-medium text-gray-500">Error</label>
                <div className="mt-1 p-3 bg-red-50 border border-red-200 rounded">
                  <div className="flex items-start gap-2">
                    <AlertCircle className="h-4 w-4 text-red-600 mt-0.5 flex-shrink-0" />
                    <p className="text-sm text-red-800">{operation.error_message}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Payload */}
            {operation.payload && (
              <div>
                <label className="text-sm font-medium text-gray-500">Payload</label>
                <div className="mt-1 p-3 bg-gray-50 border border-gray-200 rounded">
                  <pre className="text-xs text-gray-700 whitespace-pre-wrap font-mono">
                    {JSON.stringify(operation.payload, null, 2)}
                  </pre>
                </div>
              </div>
            )}

            {/* Result */}
            {operation.result && (
              <div>
                <label className="text-sm font-medium text-gray-500">Result</label>
                <div className="mt-1 p-3 bg-green-50 border border-green-200 rounded">
                  <pre className="text-xs text-green-800 whitespace-pre-wrap font-mono">
                    {JSON.stringify(operation.result, null, 2)}
                  </pre>
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="px-6 py-4 border-t border-gray-200">
          <DialogFooter className="gap-2">
          {(operation.status === 'pending' || operation.status === 'processing') && (
            <Button 
              variant="outline" 
              size="sm" 
              className="text-red-600 hover:text-red-700 hover:bg-red-50"
              onClick={handleCancel}
              disabled={isCancelling}
            >
              {isCancelling ? (
                <>
                  <Ban className="h-4 w-4 mr-2 animate-pulse" />
                  Cancelling...
                </>
              ) : (
                <>
                  <Ban className="h-4 w-4 mr-2" />
                  Cancel Operation
                </>
              )}
            </Button>
          )}
          {operation.status === 'failed' && (
            <Button 
              variant="outline" 
              size="sm"
              onClick={handleRetry}
            >
              <RotateCcw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          )}
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  );
}