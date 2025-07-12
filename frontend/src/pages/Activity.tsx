import { useState, useMemo } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  ColumnDef,
  flexRender,
  SortingState,
  ColumnFiltersState,
} from '@tanstack/react-table';
import { PageContainer } from '@/components/PageContainer';
import { PageHeader } from '@/components/PageHeader';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import {
  CheckCircle,
  XCircle,
  Clock,
  Loader2,
  Activity as ActivityIcon,
  ChevronUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Search,
  Filter,
  RefreshCw,
  Settings,
} from 'lucide-react';
import { format } from 'date-fns';
import { useActivity } from '@/hooks/useActivity';
import { ActivityItem } from '@/api/activity';
import { cn } from '@/lib/utils';
import { useActivityStream } from '@/hooks/useSSE';

export function Activity() {
  const [instanceFilter, setInstanceFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'created_at', desc: true }
  ]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [globalFilter, setGlobalFilter] = useState('');

  const { data: response, isLoading } = useActivity({
    instance_id: instanceFilter || undefined,
    type: typeFilter as 'operation' | 'sync' | undefined,
    status: statusFilter || undefined,
    limit: 100,
  });
  
  // Enable real-time updates
  useActivityStream();

  const activities = response?.activities || [];

  // Define columns
  const columns = useMemo<ColumnDef<ActivityItem>[]>(
    () => [
      {
        id: 'status',
        accessorKey: 'status',
        header: 'Status',
        size: 40,
        minSize: 40,
        maxSize: 40,
        cell: ({ row }) => {
          const status = row.original.status;
          const iconClass = 'h-5 w-5';
          
          const statusConfig = {
            pending: { icon: Clock, className: 'text-gray-500' },
            running: { icon: Loader2, className: 'text-blue-500 animate-spin' },
            processing: { icon: Loader2, className: 'text-blue-500 animate-spin' },
            completed: { icon: CheckCircle, className: 'text-green-500' },
            failed: { icon: XCircle, className: 'text-red-500' },
            cancelled: { icon: XCircle, className: 'text-gray-500' },
          };
          
          const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.pending;
          const Icon = config.icon;
          
          return (
            <div className="flex justify-center">
              <Icon className={cn(iconClass, config.className)} />
            </div>
          );
        },
      },
      {
        id: 'title',
        accessorKey: 'title',
        header: 'Activity',
        size: 300,
        cell: ({ row }) => {
          const item = row.original;
          return (
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="text-xs">
                  {item.type === 'sync' ? (
                    <>
                      <RefreshCw className="h-3 w-3 mr-1" />
                      Sync
                    </>
                  ) : (
                    <>
                      <Settings className="h-3 w-3 mr-1" />
                      Operation
                    </>
                  )}
                </Badge>
                <span className="font-medium">{item.title}</span>
              </div>
              {item.subtitle && (
                <div className="text-xs text-gray-500">{item.subtitle}</div>
              )}
              {item.error && (
                <div className="text-xs text-red-600 truncate" title={item.error}>
                  Error: {item.error}
                </div>
              )}
            </div>
          );
        },
      },
      {
        id: 'instance',
        accessorKey: 'instance.name',
        header: 'Instance',
        size: 150,
        cell: ({ row }) => row.original.instance?.name || '-',
      },
      {
        id: 'created_by',
        accessorKey: 'created_by.username',
        header: 'Triggered By',
        size: 120,
        cell: ({ row }) => row.original.created_by?.username || 'System',
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
                  column.toggleSorting(false);
                } else if (isSorted === "asc") {
                  column.toggleSorting(true);
                } else {
                  column.clearSorting();
                }
              }}
              className="h-auto p-0 font-medium hover:bg-transparent"
            >
              Started
              {column.getIsSorted() === "asc" && <ChevronUp className="ml-2 h-4 w-4" />}
              {column.getIsSorted() === "desc" && <ChevronDown className="ml-2 h-4 w-4" />}
            </Button>
          )
        },
        size: 140,
        cell: ({ row }) => (
          <div className="text-sm text-gray-600">
            {format(new Date(row.original.created_at), 'MMM d, HH:mm:ss')}
          </div>
        ),
      },
      {
        id: 'duration',
        accessorKey: 'duration_seconds',
        header: 'Duration',
        size: 100,
        cell: ({ row }) => {
          const duration = row.original.duration_seconds;
          if (duration === undefined || duration === null) {
            if (row.original.status === 'running' || row.original.status === 'processing') {
              return <span className="text-sm text-gray-500">In progress...</span>;
            }
            return <span className="text-sm text-gray-400">-</span>;
          }
          
          if (duration < 60) {
            return <span className="text-sm">{duration.toFixed(1)}s</span>;
          } else {
            const minutes = Math.floor(duration / 60);
            const seconds = duration % 60;
            return <span className="text-sm">{minutes}m {seconds.toFixed(0)}s</span>;
          }
        },
      },
      {
        id: 'details',
        header: '',
        size: 100,
        cell: ({ row }) => {
          const item = row.original;
          let href = '';
          
          if (item.type === 'operation' && item.operation) {
            href = `/operations/${item.operation.id}`;
          } else if (item.type === 'sync' && item.sync_job) {
            href = `/sync-jobs/${item.sync_job.id}`;
          }
          
          if (!href) return null;
          
          return (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => window.location.href = href}
            >
              View Details
            </Button>
          );
        },
      },
    ],
    []
  );

  const table = useReactTable({
    data: activities,
    columns,
    state: {
      sorting,
      columnFilters,
      globalFilter,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    initialState: {
      pagination: {
        pageSize: 20,
      },
    },
  });

  // Get unique instances for filter
  const uniqueInstances = useMemo(() => {
    const instances = new Map();
    activities.forEach(item => {
      if (item.instance) {
        instances.set(item.instance.id, item.instance.name);
      }
    });
    return Array.from(instances, ([id, name]) => ({ id, name }));
  }, [activities]);

  if (isLoading && activities.length === 0) {
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
        title="Activity"
        description="View all operations and synchronization activities"
        icon={<ActivityIcon className="h-8 w-8" />}
      />

      {/* Filters */}
      <Card className="mb-6">
        <CardContent className="pt-6">
          <div className="flex flex-wrap gap-4">
            <div className="flex-1 min-w-[200px]">
              <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-gray-500" />
                <Input
                  placeholder="Search activities..."
                  value={globalFilter}
                  onChange={(e) => setGlobalFilter(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
            
            <Select value={instanceFilter} onValueChange={setInstanceFilter}>
              <SelectTrigger className="w-[180px]">
                <Filter className="h-4 w-4 mr-2" />
                <SelectValue placeholder="All instances" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All instances</SelectItem>
                {uniqueInstances.map(instance => (
                  <SelectItem key={instance.id} value={instance.id}>
                    {instance.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            
            <Select value={typeFilter} onValueChange={setTypeFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="All types" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All types</SelectItem>
                <SelectItem value="operation">Operations</SelectItem>
                <SelectItem value="sync">Sync Jobs</SelectItem>
              </SelectContent>
            </Select>
            
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="All statuses" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All statuses</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="running">Running</SelectItem>
                <SelectItem value="completed">Completed</SelectItem>
                <SelectItem value="failed">Failed</SelectItem>
                <SelectItem value="cancelled">Cancelled</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Data Table */}
      <Card>
        <CardContent className="p-0">
          {activities.length === 0 ? (
            <div className="text-center py-16">
              <ActivityIcon className="h-12 w-12 mx-auto mb-4 text-gray-400" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                No activity yet
              </h3>
              <p className="text-sm text-gray-500">
                Operations and sync jobs will appear here
              </p>
            </div>
          ) : (
            <>
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
                      <TableRow key={row.id}>
                        {row.getVisibleCells().map((cell) => (
                          <TableCell 
                            key={cell.id}
                            className="align-middle"
                          >
                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={columns.length} className="h-24 text-center">
                        No activities found.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>

              {/* Pagination */}
              <div className="flex items-center justify-between px-4 py-4 border-t">
                <div className="text-sm text-gray-500">
                  Showing {table.getState().pagination.pageIndex * table.getState().pagination.pageSize + 1} to{' '}
                  {Math.min(
                    (table.getState().pagination.pageIndex + 1) * table.getState().pagination.pageSize,
                    table.getFilteredRowModel().rows.length
                  )}{' '}
                  of {table.getFilteredRowModel().rows.length} activities
                </div>
                
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.setPageIndex(0)}
                    disabled={!table.getCanPreviousPage()}
                  >
                    <ChevronsLeft className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.previousPage()}
                    disabled={!table.getCanPreviousPage()}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  
                  <div className="flex items-center gap-1">
                    <span className="text-sm">Page</span>
                    <span className="text-sm font-medium">
                      {table.getState().pagination.pageIndex + 1}
                    </span>
                    <span className="text-sm">of {table.getPageCount()}</span>
                  </div>
                  
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.nextPage()}
                    disabled={!table.getCanNextPage()}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.setPageIndex(table.getPageCount() - 1)}
                    disabled={!table.getCanNextPage()}
                  >
                    <ChevronsRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  );
}

export default Activity;