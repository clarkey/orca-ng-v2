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
  ChevronDown
} from 'lucide-react';
import { OperationDetailsPanel } from '../components/OperationDetailsPanel';

const STORAGE_KEY_PAGE_SIZE = 'orca-operations-page-size';

export default function OperationsTable() {
  const [selectedOperation, setSelectedOperation] = useState<Operation | null>(null);
  
  // Table state
  const [sorting, setSorting] = useState<SortingState>([]);
  
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
  const { data: response, isLoading, refetch, isFetching, isPreviousData } = useOperations(queryParams, { prefetchNext: true });
  const cancelMutation = useCancelOperation();
  
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
        size: 30,
        cell: ({ row }) => {
          const statusIcons: Record<Status, { icon: React.ReactNode; color: string }> = {
            pending: { 
              icon: <Clock className="h-4 w-4" />, 
              color: 'text-gray-400' 
            },
            processing: { 
              icon: <Loader2 className="h-4 w-4 animate-spin" />, 
              color: 'text-blue-600' 
            },
            completed: { 
              icon: <CheckCircle className="h-4 w-4" />, 
              color: 'text-green-600' 
            },
            failed: { 
              icon: <XCircle className="h-4 w-4" />, 
              color: 'text-red-600' 
            },
            cancelled: { 
              icon: <Ban className="h-4 w-4" />, 
              color: 'text-gray-400' 
            },
          };
          
          const status = statusIcons[row.original.status];
          
          return (
            <div className="flex items-center pl-2" title={row.original.status}>
              <span className={status.color}>
                {status.icon}
              </span>
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
              Type / ID
              {column.getIsSorted() === "asc" && <ChevronUp className="ml-2 h-4 w-4" />}
              {column.getIsSorted() === "desc" && <ChevronDown className="ml-2 h-4 w-4" />}
            </Button>
          )
        },
        size: 220,
        enableSorting: true,
        cell: ({ row }) => (
          <div className="space-y-1">
            <div className="font-medium text-sm">{getOperationTypeLabel(row.original.type)}</div>
            <div className="font-mono text-xs text-gray-500">{row.original.id}</div>
          </div>
        ),
      },
      {
        id: 'subject',
        accessorKey: 'subject',
        header: 'Subject',
        size: 200,
        cell: ({ row }) => {
          // Configuration for extracting subject fields by operation type
          const subjectConfig: Record<OperationType, { primary: string[], secondary: string[] }> = {
            safe_provision: { 
              primary: ['safe_name'], 
              secondary: ['cyberark_instance_id'] 
            },
            safe_modify: { 
              primary: ['safe_name'], 
              secondary: ['cyberark_instance_id'] 
            },
            safe_delete: { 
              primary: ['safe_name'], 
              secondary: ['cyberark_instance_id'] 
            },
            access_grant: { 
              primary: ['username', 'user_id'], 
              secondary: ['safe_name', 'permission_set'] 
            },
            access_revoke: { 
              primary: ['username', 'user_id'], 
              secondary: ['safe_name'] 
            },
            user_sync: { 
              primary: ['cyberark_instance_id'], 
              secondary: ['sync_type'] 
            },
            safe_sync: { 
              primary: ['cyberark_instance_id'], 
              secondary: ['sync_type'] 
            },
            group_sync: { 
              primary: ['cyberark_instance_id'], 
              secondary: ['sync_type'] 
            },
          };
          
          const config = subjectConfig[row.original.type];
          if (!config || !row.original.payload) {
            return <span className="text-gray-400 text-xs">-</span>;
          }
          
          // Extract values from payload
          const payload = row.original.payload;
          const primaryValue = config.primary
            .map(field => payload[field])
            .filter(Boolean)
            .join(' / ');
          
          const secondaryValue = config.secondary
            .map(field => payload[field])
            .filter(Boolean)
            .join(' • ');
          
          return (
            <div className="space-y-0.5">
              {primaryValue && (
                <div className="text-sm font-medium text-gray-900 truncate" title={primaryValue}>
                  {primaryValue}
                </div>
              )}
              {secondaryValue && (
                <div className="text-xs text-gray-500 truncate" title={secondaryValue}>
                  {secondaryValue}
                </div>
              )}
              {!primaryValue && !secondaryValue && (
                <span className="text-gray-400 text-xs">-</span>
              )}
            </div>
          );
        },
      },
      {
        id: 'cyberark_instance',
        accessorKey: 'cyberark_instance',
        header: 'Instance',
        size: 150,
        cell: ({ row }) => {
          const instanceInfo = row.original.cyberark_instance_info;
          const instanceId = row.original.cyberark_instance_id;
          
          if (!instanceInfo && !instanceId) {
            return <span className="text-gray-400 text-xs">-</span>;
          }
          
          const instanceName = instanceInfo?.name || instanceId || 'Unknown';
          const displayId = instanceInfo?.id || instanceId || '-';
          
          return (
            <div className="space-y-0.5">
              <div className="text-sm font-medium text-gray-900 truncate" title={instanceName}>
                {instanceName}
              </div>
              <div className="text-xs text-gray-500 font-mono truncate" title={displayId}>
                {displayId}
              </div>
            </div>
          );
        },
      },
      {
        id: 'created_by',
        accessorKey: 'created_by',
        header: 'Created By',
        size: 150,
        cell: ({ row }) => {
          const userInfo = row.original.created_by_user;
          const userId = row.original.created_by;
          
          if (!userId && !userInfo) {
            return <span className="text-gray-400 text-xs">System</span>;
          }
          
          const username = userInfo?.username || 'Unknown User';
          const displayId = userInfo?.id || userId || '-';
          
          return (
            <div className="space-y-1">
              <div className="text-sm font-medium text-gray-900">{username}</div>
              <div className="text-xs text-gray-500 font-mono">{displayId}</div>
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
              Created
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
          
          return (
            <div className="space-y-0.5">
              <div className="text-sm">{relativeTime}</div>
              <div className="text-xs text-gray-500">{format(createdDate, 'MMM d, HH:mm:ss')}</div>
            </div>
          );
        },
      },
      {
        id: 'actions',
        header: '',
        size: 100,
        cell: ({ row }) => (
          <div className="flex items-center justify-end gap-2">
            {(row.original.status === 'pending' || row.original.status === 'processing') && (
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation();
                  handleCancelOperation(row.original.id);
                }}
              >
                Cancel
              </Button>
            )}
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
                        style={{ width: header.getSize() }}
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
                {isPreviousData && (
                  <span className="text-xs text-gray-500 ml-2">(Stale)</span>
                )}
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
      />
    </PageContainer>
  );
}