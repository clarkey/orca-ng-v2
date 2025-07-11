import { useState, useMemo } from 'react';
import { format } from 'date-fns';
import {
  useReactTable,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  ColumnDef,
  flexRender,
  SortingState,
  PaginationState,
} from '@tanstack/react-table';
import { 
  Operation, 
  OperationType,
  Status, 
  getOperationTypeLabel
} from '../api/operations';
import { useOperations, useCancelOperation } from '@/hooks/useOperations';
import { Button } from '../components/ui/button';
import { Card, CardContent } from '../components/ui/card';
import { PageContainer } from '../components/PageContainer';
import { PageHeader } from '../components/PageHeader';
import { Input } from '../components/ui/input';
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../components/ui/table';
import { 
  RefreshCw, 
  ChevronRight,
  Clock,
  CheckCircle,
  XCircle,
  Loader2,
  ChevronLeft,
  ChevronsLeft,
  ChevronsRight,
  Ban,
  ChevronUp,
  ChevronDown,
  CheckCircle2
} from 'lucide-react';
import { OperationDetailsPanel } from '../components/OperationDetailsPanel';
import { ConfirmationDialog } from '../components/ui/confirmation-dialog';
import { operationsApi } from '../api/operations';

const STORAGE_KEY_PAGE_SIZE = 'orca-operations-page-size';

export default function OperationsTable() {
  const [selectedOperation, setSelectedOperation] = useState<Operation | null>(null);
  const [operationToCancel, setOperationToCancel] = useState<Operation | null>(null);
  const [isCancelling, setIsCancelling] = useState(false);
  
  // Table state
  const [sorting, setSorting] = useState<SortingState>([]);
  
  const handleConfirmCancel = async () => {
    if (!operationToCancel) return;
    
    setIsCancelling(true);
    try {
      await operationsApi.cancel(operationToCancel.id);
      refetch();
    } catch (error) {
      console.error('Failed to cancel operation:', error);
      alert('Failed to cancel operation. Please try again.');
    } finally {
      setIsCancelling(false);
      setOperationToCancel(null);
    }
  };
  
  // Initialize pagination with saved page size
  const [pagination, setPagination] = useState<PaginationState>(() => {
    const savedPageSize = localStorage.getItem(STORAGE_KEY_PAGE_SIZE);
    return {
      pageIndex: 0,
      pageSize: savedPageSize ? parseInt(savedPageSize, 10) : 50,
    };
  });

  // Build query params
  const queryParams = useMemo(() => {
    const params: any = {
      page: pagination.pageIndex + 1,
      page_size: pagination.pageSize,
    };
    
    // Add sorting parameters
    if (sorting.length > 0) {
      params.sort_by = sorting[0].id;
      params.sort_order = sorting[0].desc ? 'desc' : 'asc';
    }
    
    return params;
  }, [pagination, sorting]);

  // Fetch operations using React Query with next page prefetching
  const { data: response, isLoading, refetch, isFetching } = useOperations(queryParams, { prefetchNext: true });
  const cancelMutation = useCancelOperation();
  
  // Use real data from the API
  const data = response?.operations || [];
  const totalCount = response?.pagination?.total_count || 0;
  const pageCount = response?.pagination?.total_pages || 0;



  // Define columns
  const columns = useMemo<ColumnDef<Operation>[]>(
    () => [
      {
        id: 'status',
        accessorKey: 'status',
        header: () => <span></span>,
        size: 40,
        minSize: 40,
        maxSize: 40,
        cell: ({ row }) => {
          const operation = row.original;
          const scheduledTime = new Date(operation.scheduled_at);
          const now = new Date();
          const isScheduledForFuture = operation.status === 'pending' && scheduledTime > now;
          
          if (operation.status === 'pending' && isScheduledForFuture) {
            // Scheduled state (blue)
            return (
              <div className="flex items-center justify-center h-full" title="Scheduled">
                <Clock className="h-5 w-5 text-blue-600" />
              </div>
            );
          } else if (operation.status === 'pending') {
            // Pending/Created state (gray)
            return (
              <div className="flex items-center justify-center h-full" title="Pending">
                <Clock className="h-5 w-5 text-gray-600" />
              </div>
            );
          } else if (operation.status === 'processing') {
            // Processing/Started state (indigo)
            return (
              <div className="flex items-center justify-center h-full" title="Processing">
                <Loader2 className="h-5 w-5 text-indigo-600 animate-spin" />
              </div>
            );
          } else if (operation.status === 'completed') {
            // Completed state (green)
            return (
              <div className="flex items-center justify-center h-full" title="Completed">
                <CheckCircle2 className="h-5 w-5 text-green-600" />
              </div>
            );
          } else if (operation.status === 'failed') {
            // Failed state (red)
            return (
              <div className="flex items-center justify-center h-full" title="Failed">
                <XCircle className="h-5 w-5 text-red-600" />
              </div>
            );
          } else if (operation.status === 'cancelled') {
            // Cancelled state (amber)
            return (
              <div className="flex items-center justify-center h-full" title="Cancelled">
                <Ban className="h-5 w-5 text-amber-600" />
              </div>
            );
          }
          
          // Fallback
          return (
            <div className="flex items-center justify-center h-full" title={operation.status}>
              <Clock className="h-5 w-5 text-gray-600" />
            </div>
          );
        },
      },
      {
        id: 'type',
        accessorKey: 'type',
        header: ({ column }) => {
          return (
            <Button
              variant="ghost"
              onClick={() => {
                const isSorted = column.getIsSorted();
                if (isSorted === false) {
                  column.toggleSorting(false); // Set to ascending
                } else if (isSorted === "asc") {
                  column.toggleSorting(true); // Set to descending
                } else {
                  column.clearSorting(); // Clear sorting
                }
              }}
              className="h-auto p-0 font-medium hover:bg-transparent"
            >
              Type
              {column.getIsSorted() === "asc" && <ChevronUp className="ml-2 h-4 w-4" />}
              {column.getIsSorted() === "desc" && <ChevronDown className="ml-2 h-4 w-4" />}
            </Button>
          )
        },
        size: 220,
        enableSorting: true,
        cell: ({ row }) => {
          const instanceInfo = row.original.cyberark_instance_info;
          const instanceId = row.original.cyberark_instance_id;
          const instanceName = instanceInfo?.name || (instanceId ? 'Unknown Instance' : null);
          
          return (
            <div className="space-y-1">
              <div className="font-medium text-sm">{getOperationTypeLabel(row.original.type)}</div>
              {instanceName && (
                <div className="text-xs text-gray-500">{instanceName}</div>
              )}
            </div>
          );
        },
      },
      {
        id: 'subject',
        accessorKey: 'subject',
        header: 'Subject',
        size: 200,
        cell: ({ row }) => {
          const payload = row.original.payload || {};
          
          // Define display configuration with labels for each operation type
          const displayConfig: Record<OperationType, Array<{key: string, label: string, primary?: boolean}>> = {
            safe_provision: [
              { key: 'safe_name', label: 'Safe', primary: true },
              { key: 'description', label: 'Description' }
            ],
            safe_modify: [
              { key: 'safe_name', label: 'Safe', primary: true },
              { key: 'changes', label: 'Changes' }
            ],
            safe_delete: [
              { key: 'safe_name', label: 'Safe', primary: true },
              { key: 'reason', label: 'Reason' }
            ],
            access_grant: [
              { key: 'username', label: 'User', primary: true },
              { key: 'safe_name', label: 'Safe' },
              { key: 'permission_set', label: 'Permissions' }
            ],
            access_revoke: [
              { key: 'username', label: 'User', primary: true },
              { key: 'safe_name', label: 'Safe' }
            ],
            user_sync: [
              { key: 'sync_mode', label: 'Mode', primary: true },
              { key: 'page_size', label: 'Page Size' }
            ],
            safe_sync: [
              { key: 'sync_type', label: 'Type', primary: true },
              { key: 'affected_safes', label: 'Safes' }
            ],
            group_sync: [
              { key: 'sync_type', label: 'Type', primary: true },
              { key: 'affected_groups', label: 'Groups' }
            ],
          };
          
          const config = displayConfig[row.original.type];
          if (!config) {
            return <span className="text-gray-400 text-xs">No details</span>;
          }
          
          const allFields = config.filter(f => payload[f.key]);
          
          return (
            <div className="space-y-0.5">
              {allFields.length > 0 ? (
                allFields.map((field, idx) => (
                  <div key={idx} className="text-sm">
                    <span className="text-gray-500">{field.label}:</span>{' '}
                    <span className="text-gray-900">{payload[field.key]}</span>
                  </div>
                ))
              ) : (
                <span className="text-gray-400 text-sm">No details</span>
              )}
            </div>
          );
        },
      },
      {
        id: 'created_at',
        accessorKey: 'created_at',
        header: ({ column }) => {
          return (
            <Button
              variant="ghost"
              onClick={() => {
                const isSorted = column.getIsSorted();
                if (isSorted === false) {
                  column.toggleSorting(false); // Set to ascending
                } else if (isSorted === "asc") {
                  column.toggleSorting(true); // Set to descending
                } else {
                  column.clearSorting(); // Clear sorting
                }
              }}
              className="h-auto p-0 font-medium hover:bg-transparent"
            >
              Added
              {column.getIsSorted() === "asc" && <ChevronUp className="ml-2 h-4 w-4" />}
              {column.getIsSorted() === "desc" && <ChevronDown className="ml-2 h-4 w-4" />}
            </Button>
          )
        },
        size: 180,
        cell: ({ row }) => {
          const createdDate = new Date(row.original.created_at);
          const now = new Date();
          const diffInSeconds = Math.floor((now.getTime() - createdDate.getTime()) / 1000);
          
          let relativeTime: string;
          if (diffInSeconds < 60) {
            relativeTime = 'just now';
          } else if (diffInSeconds < 3600) {
            const minutes = Math.floor(diffInSeconds / 60);
            relativeTime = `${minutes}m ago`;
          } else if (diffInSeconds < 86400) {
            const hours = Math.floor(diffInSeconds / 3600);
            relativeTime = `${hours}h ago`;
          } else if (diffInSeconds < 604800) {
            const days = Math.floor(diffInSeconds / 86400);
            relativeTime = `${days}d ago`;
          } else {
            relativeTime = format(createdDate, 'MMM d');
          }
          
          // Get username
          const userInfo = row.original.created_by_user;
          const userId = row.original.created_by;
          const username = userInfo?.username || (userId ? 'Unknown' : 'System');
          
          return (
            <div className="space-y-0.5">
              <div className="text-sm">
                {relativeTime}
                <span className="text-gray-500 ml-1">by {username}</span>
              </div>
              <div className="text-xs text-gray-500">{format(createdDate, 'MMM d, HH:mm:ss')}</div>
            </div>
          );
        },
      },
      {
        id: 'actions',
        header: '',
        size: 40,
        cell: ({ row }) => (
          <div className="flex justify-end pt-1">
            <ChevronRight className="h-4 w-4 text-gray-400" />
          </div>
        ),
      },
    ],
    []
  );

  const table = useReactTable({
    data,
    columns,
    pageCount: pageCount || Math.max(1, Math.ceil(totalCount / pagination.pageSize)),
    state: {
      sorting,
      pagination,
    },
    onSortingChange: setSorting,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    manualPagination: true,
    manualSorting: true,
  });

  const handleCancelOperation = async (id: string) => {
    try {
      await cancelMutation.mutateAsync(id);
      refetch();
    } catch (error) {
      console.error('Failed to cancel operation:', error);
    }
  };

  const handleRowClick = (operation: Operation) => {
    setSelectedOperation(operation);
  };

  if (isLoading && data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <PageContainer>
      <PageHeader
        title="Operations Queue"
        description="Monitor and manage CyberArk operations and tasks"
        actions={
          <Button 
            onClick={() => refetch()} 
            disabled={isLoading}
            variant="outline"
          >
            {isLoading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4" />
            )}
            <span className="ml-2">Refresh</span>
          </Button>
        }
      />

      {/* Data Table */}
      <Card className="overflow-hidden">
        <CardContent className="p-0 relative">
          {/* Loading overlay for pagination changes */}
          {isFetching && !isLoading && (
            <div className="absolute inset-0 bg-white/50 backdrop-blur-sm z-10 flex items-center justify-center">
              <Loader2 className="h-6 w-6 animate-spin text-gray-600" />
            </div>
          )}
          <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead 
                        key={header.id} 
                        style={{ width: header.getSize(), minWidth: header.getSize(), maxWidth: header.getSize() }}
                        className={header.id === 'status' ? 'w-[60px] min-w-[60px] max-w-[60px]' : ''}
                      >
                        {header.isPlaceholder
                          ? null
                          : flexRender(header.column.columnDef.header, header.getContext())}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      className="cursor-pointer hover:bg-gray-50"
                      onClick={() => handleRowClick(row.original)}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <TableCell 
                          key={cell.id}
                          className={`align-top ${cell.column.id === 'status' ? 'p-2' : ''}`}
                        >
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={columns.length} className="h-24 text-center">
                      No operations found.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>

          {/* Pagination */}
          <div className="flex items-center justify-between px-4 py-4 border-t">
            <div className="text-sm text-gray-500">
              Showing {pagination.pageIndex * pagination.pageSize + 1} to{' '}
              {Math.min((pagination.pageIndex + 1) * pagination.pageSize, totalCount)} of{' '}
              {totalCount.toLocaleString()} operations
            </div>
            
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.setPageIndex(0)}
                disabled={!table.getCanPreviousPage() || isFetching}
              >
                <ChevronsLeft className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage() || isFetching}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              
              <div className="flex items-center gap-1">
                <span className="text-sm">Page</span>
                <Input
                  type="number"
                  value={pagination.pageIndex + 1}
                  onChange={(e) => {
                    const page = e.target.value ? Number(e.target.value) - 1 : 0;
                    table.setPageIndex(page);
                  }}
                  className="w-16 text-center"
                  min={1}
                  max={table.getPageCount()}
                  disabled={isFetching}
                />
                <span className="text-sm">of {table.getPageCount()}</span>
              </div>
              
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage() || isFetching}
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.setPageIndex(table.getPageCount() - 1)}
                disabled={!table.getCanNextPage() || isFetching}
              >
                <ChevronsRight className="h-4 w-4" />
              </Button>
              
              <Select
                value={pagination.pageSize.toString()}
                onValueChange={(value) => {
                  const pageSize = Number(value);
                  table.setPageSize(pageSize);
                  localStorage.setItem(STORAGE_KEY_PAGE_SIZE, value);
                }}
              >
                <SelectTrigger className="w-24">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="10">10</SelectItem>
                  <SelectItem value="20">20</SelectItem>
                  <SelectItem value="50">50</SelectItem>
                  <SelectItem value="100">100</SelectItem>
                  <SelectItem value="200">200</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Operation Details Panel */}
      <OperationDetailsPanel 
        operation={selectedOperation} 
        onClose={() => setSelectedOperation(null)}
        onUpdate={refetch}
        onCancelRequest={(operation) => {
          setSelectedOperation(null);
          setOperationToCancel(operation);
        }}
      />

      <ConfirmationDialog
        open={!!operationToCancel}
        onOpenChange={(open) => !open && setOperationToCancel(null)}
        title="Cancel Operation"
        description={operationToCancel ? 
          `Are you sure you want to cancel this ${getOperationTypeLabel(operationToCancel.type).toLowerCase()}? This action cannot be undone.` : 
          ''
        }
        confirmLabel={isCancelling ? "Cancelling..." : "Cancel Operation"}
        variant="destructive"
        onConfirm={handleConfirmCancel}
        confirmDisabled={isCancelling}
      />
    </PageContainer>
  );
}