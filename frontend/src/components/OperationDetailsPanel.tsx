import { useState } from 'react';
import { format, formatDistanceToNow, formatDistance } from 'date-fns';
import { Operation, getOperationTypeLabel, operationsApi } from '../api/operations';
import { Button } from './ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogBody,
  DialogFooter,
} from './ui/dialog';
import { Clock, AlertCircle, RotateCcw, Calendar, Play, CheckCircle2, XCircle, Ban } from 'lucide-react';
import { cn } from '@/lib/utils';

interface OperationDetailsPanelProps {
  operation: Operation | null;
  onClose: () => void;
  onUpdate?: () => void;
  onCancelRequest?: (operation: Operation) => void;
}

export function OperationDetailsPanel({ operation, onClose, onUpdate, onCancelRequest }: OperationDetailsPanelProps) {
  if (!operation) return null;

  const scheduledTime = new Date(operation.scheduled_at);
  const now = new Date();
  const isScheduledForFuture = operation.status === 'pending' && scheduledTime > now;
  const showFooter = operation.status === 'pending' || operation.status === 'processing' || operation.status === 'failed';
  
  const handleRetry = async () => {
    // TODO: Implement retry functionality
    console.log('Retry operation:', operation.id);
  };

  return (
    <Dialog open={!!operation} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader 
          title="Operation Details"
          description={operation.id}
        />

        <DialogBody className={!showFooter ? 'pb-6' : ''}>
          <div className="space-y-6">
            {/* Status, Type and Priority */}
            <div className="grid grid-cols-3 gap-4">
              <div>
                <label className="text-sm font-medium text-gray-500">Status</label>
                <div className="mt-1">
                  <span className={`inline-block px-3 py-1 rounded-full text-[11px] font-semibold uppercase tracking-wider ${
                    operation.status === 'completed' ? 'bg-green-100 text-green-700 border border-green-200' :
                    operation.status === 'failed' ? 'bg-red-100 text-red-700 border border-red-200' :
                    operation.status === 'processing' ? 'bg-blue-100 text-blue-700 border border-blue-200' :
                    operation.status === 'cancelled' ? 'bg-amber-100 text-amber-700 border border-amber-200' :
                    'bg-gray-100 text-gray-600 border border-gray-200'
                  }`}>
                    {operation.status}
                  </span>
                </div>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-500">Type</label>
                <p className="mt-1 text-sm text-gray-900">{getOperationTypeLabel(operation.type)}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-500">Priority</label>
                <div className="mt-1">
                  {(operation.status === 'pending' || operation.status === 'processing') ? (
                    <select
                      value={operation.priority}
                      onChange={async (e) => {
                        try {
                          await operationsApi.updatePriority(operation.id, e.target.value as any);
                          onUpdate?.();
                        } catch (error) {
                          console.error('Failed to update priority:', error);
                          // TODO: Show error toast or alert
                        }
                      }}
                      className="text-[11px] font-semibold uppercase tracking-wider px-3 py-1 rounded-full border bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="low">Low</option>
                      <option value="normal">Normal</option>
                      <option value="high">High</option>
                    </select>
                  ) : (
                    <span className={`inline-block px-3 py-1 rounded-full text-[11px] font-semibold uppercase tracking-wider ${
                      operation.priority === 'high' ? 'bg-red-100 text-red-700 border border-red-200' :
                      operation.priority === 'normal' ? 'bg-blue-100 text-blue-700 border border-blue-200' :
                      'bg-gray-100 text-gray-600 border border-gray-200'
                    }`}>
                      {operation.priority}
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Timeline */}
            <div>
              <label className="text-sm font-medium text-gray-500 block mb-3">Timeline</label>
              <div className="relative">
                {/* Timeline line */}
                <div className="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200" style={{
                  height: operation.completed_at && ['completed', 'failed', 'cancelled'].includes(operation.status) 
                    ? 'calc(100% - 2rem)' 
                    : '100%'
                }}></div>
                
                {/* Timeline events */}
                <div className="space-y-6">
                {/* Created */}
                  {/* Created Event */}
                  <div className="relative flex items-start gap-4">
                    <div className="relative z-10 flex items-center justify-center w-8 h-8 bg-white border-2 border-gray-300 rounded-full">
                      <Calendar className="h-4 w-4 text-gray-600" />
                    </div>
                    <div className="flex-1 pb-2">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-sm font-semibold text-gray-900">Created</span>
                        <span className="text-xs text-gray-500">
                          {format(new Date(operation.created_at), 'MMM d, yyyy')} at {format(new Date(operation.created_at), 'h:mm a')}
                        </span>
                      </div>
                      <div className="text-xs text-gray-500">
                        {formatDistanceToNow(new Date(operation.created_at), { addSuffix: true })}
                      </div>
                      {operation.created_by_user && (
                        <div className="text-xs text-gray-500 mt-1">
                          by {operation.created_by_user.username} <span className="text-gray-400">({operation.created_by_user.id})</span>
                        </div>
                      )}
                    </div>
                  </div>

                  
                  {/* Scheduled Event (if different from created) */}
                  {operation.status === 'pending' && (
                    <div className="relative flex items-start gap-4">
                      <div className="relative z-10 flex items-center justify-center w-8 h-8 bg-white border-2 border-blue-400 rounded-full">
                        <Clock className="h-4 w-4 text-blue-600" />
                      </div>
                      <div className="flex-1 pb-2">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-sm font-semibold text-gray-900">Scheduled</span>
                          <span className="text-xs text-gray-500">
                            {format(scheduledTime, 'MMM d, yyyy')} at {format(scheduledTime, 'h:mm a')}
                          </span>
                        </div>
                        <div className="text-xs text-blue-600 font-medium">
                          {isScheduledForFuture ? 
                            formatDistanceToNow(scheduledTime, { addSuffix: true }) : 
                            'Waiting to start...'
                          }
                        </div>
                      </div>
                    </div>
                  )}

                  
                  {/* Started Event */}
                  {operation.started_at && (
                    <div className="relative flex items-start gap-4">
                      <div className="relative z-10 flex items-center justify-center w-8 h-8 bg-white border-2 border-indigo-400 rounded-full">
                        <Play className="h-4 w-4 text-indigo-600" />
                      </div>
                      <div className="flex-1 pb-2">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-sm font-semibold text-gray-900">Started</span>
                          <span className="text-xs text-gray-500">
                            {format(new Date(operation.started_at), 'MMM d, yyyy')} at {format(new Date(operation.started_at), 'h:mm a')}
                          </span>
                        </div>
                        <div className="text-xs text-gray-500">
                          {formatDistanceToNow(new Date(operation.started_at), { addSuffix: true })}
                        </div>
                      </div>
                    </div>
                  )}

                  
                  {/* Completed/Failed/Cancelled Event */}
                  {operation.completed_at && (
                    <div className="relative flex items-start gap-4">
                      <div className={cn(
                        "relative z-10 flex items-center justify-center w-8 h-8 bg-white rounded-full border-2",
                        operation.status === 'completed' ? 'border-green-400' : 
                        operation.status === 'failed' ? 'border-red-400' : 
                        operation.status === 'cancelled' ? 'border-amber-400' :
                        'border-gray-400'
                      )}>
                        {operation.status === 'completed' ? (
                          <CheckCircle2 className="h-4 w-4 text-green-600" />
                        ) : operation.status === 'failed' ? (
                          <XCircle className="h-4 w-4 text-red-600" />
                        ) : operation.status === 'cancelled' ? (
                          <Ban className="h-4 w-4 text-amber-600" />
                        ) : (
                          <Clock className="h-4 w-4 text-gray-600" />
                        )}
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <span className={cn(
                            "text-sm font-semibold",
                            operation.status === 'completed' ? 'text-green-700' : 
                            operation.status === 'failed' ? 'text-red-700' : 
                            operation.status === 'cancelled' ? 'text-amber-700' :
                            'text-gray-700'
                          )}>
                            {operation.status === 'completed' ? 'Completed' : 
                             operation.status === 'failed' ? 'Failed' : 
                             operation.status === 'cancelled' ? 'Cancelled' :
                             'Ended'
                            }
                          </span>
                          <span className="text-xs text-gray-500">
                            {format(new Date(operation.completed_at), 'MMM d, yyyy')} at {format(new Date(operation.completed_at), 'h:mm a')}
                          </span>
                        </div>
                        <div className="text-xs text-gray-500">
                          {formatDistanceToNow(new Date(operation.completed_at), { addSuffix: true })}
                        </div>
                        {operation.started_at && (
                          <div className="text-xs text-gray-600 mt-1 font-medium">
                            Duration: {formatDistance(new Date(operation.started_at), new Date(operation.completed_at))}
                          </div>
                        )}
                        {operation.status === 'cancelled' && (
                          <div className="text-xs text-gray-500 mt-1">
                            {/* TODO: Add cancelled_by_user when backend supports it */}
                            {/* by {operation.cancelled_by_user?.username} ({operation.cancelled_by_user?.id}) */}
                            <span className="italic">Cancellation user not tracked</span>
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>

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
        </DialogBody>

        {showFooter && (
          <DialogFooter className="gap-2">
            {(operation.status === 'pending' || operation.status === 'processing') && (
              <Button 
                variant="destructive" 
                size="sm" 
                onClick={() => {
                  onClose();
                  onCancelRequest?.(operation);
                }}
                disabled={false}
              >
                Cancel Operation
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
        )}
      </DialogContent>
    </Dialog>
  );
}