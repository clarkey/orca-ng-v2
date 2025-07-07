import React, { useEffect, useState, useCallback, useMemo } from 'react';
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
  RowSelectionState,
} from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';
import { 
  Operation, 
  OperationType,
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
import { Input } from '../components/ui/input';
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../components/ui/table';
import { Checkbox } from '../components/ui/checkbox';
import { 
  Plus, 
  RefreshCw, 
  ChevronRight,
  Clock,
  CheckCircle,
  XCircle,
  Loader2,
  Search,
  Calendar,
  Download,
  BarChart3,
  Filter,
  ChevronLeft,
  ChevronsLeft,
  ChevronsRight,
  Ban,
  ChevronUp,
  ChevronDown
} from 'lucide-react';
import { OperationDetailsPanel } from '../components/OperationDetailsPanel';

export default function OperationsTable() {
  const searchInputRef = React.useRef<HTMLInputElement>(null);
  const [data, setData] = useState<Operation[]>([]);
  const [loading, setLoading] = useState(true);
  const [totalCount, setTotalCount] = useState(0);
  const [showNewOperationDialog, setShowNewOperationDialog] = useState(false);
  const [showStatsDialog, setShowStatsDialog] = useState(false);
  const [stats, setStats] = useState<any>(null);
  const [selectedOperation, setSelectedOperation] = useState<Operation | null>(null);
  
  // Table state
  const [sorting, setSorting] = useState<SortingState>([]);
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 50,
  });
  
  // Filters
  const [globalFilter, setGlobalFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState<Status | 'all'>('all');
  const [priorityFilter, setPriorityFilter] = useState<Priority | 'all'>('all');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [dateRange, setDateRange] = useState<{ start: string; end: string }>({ start: '', end: '' });
  
  // Debounced search
  const [searchValue, setSearchValue] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchValue);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchValue]);

  // Fetch data
  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: any = {
        page: pagination.pageIndex + 1,
        page_size: pagination.pageSize,
      };
      
      if (debouncedSearch) params.search = debouncedSearch;
      if (statusFilter !== 'all') params.status = statusFilter;
      if (priorityFilter !== 'all') params.priority = priorityFilter;
      if (typeFilter !== 'all') params.type = typeFilter;
      if (dateRange.start && dateRange.start !== '') params.start_date = dateRange.start;
      if (dateRange.end && dateRange.end !== '') params.end_date = dateRange.end;
      
      // Add sorting parameters
      if (sorting.length > 0) {
        params.sort_by = sorting[0].id;
        params.sort_order = sorting[0].desc ? 'desc' : 'asc';
      }
      
      const response = await operationsApi.list(params);
      setData(response.operations);
      setTotalCount(response.pagination.total_count);
    } catch (error) {
      console.error('Failed to fetch operations:', error);
    } finally {
      setLoading(false);
    }
  }, [pagination, debouncedSearch, statusFilter, priorityFilter, typeFilter, dateRange, sorting]);

  // Fetch data when filters change
  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Track if we're currently searching to maintain focus
  const [isSearching, setIsSearching] = useState(false);

  // Maintain focus on search input while typing
  useEffect(() => {
    if (isSearching && searchInputRef.current) {
      const cursorPosition = searchInputRef.current.selectionStart;
      searchInputRef.current.focus();
      // Restore cursor position
      if (cursorPosition !== null) {
        searchInputRef.current.setSelectionRange(cursorPosition, cursorPosition);
      }
    }
  }, [data, isSearching]);

  // Fetch stats
  const fetchStats = async () => {
    try {
      const params: any = {};
      if (dateRange.start) params.start_date = dateRange.start;
      if (dateRange.end) params.end_date = dateRange.end;
      
      const response = await operationsApi.getStats(params);
      setStats(response);
      setShowStatsDialog(true);
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  const getStatusIcon = (status: Status) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-4 w-4" />;
      case 'processing':
        return <Loader2 className="h-4 w-4 animate-spin" />;
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-600" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-600" />;
      case 'cancelled':
        return <Ban className="h-4 w-4 text-orange-600" />;
    }
  };

  // Define columns
  const columns = useMemo<ColumnDef<Operation>[]>(
    () => [
      {
        id: 'select',
        header: ({ table }) => (
          <div className="flex items-center h-full">
            <Checkbox
              checked={
                table.getIsAllPageRowsSelected() ||
                (table.getIsSomePageRowsSelected() && "indeterminate")
              }
              onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
              aria-label="Select all"
            />
          </div>
        ),
        cell: ({ row }) => (
          <div className="flex items-center h-full">
            <Checkbox
              checked={row.getIsSelected()}
              onCheckedChange={(value) => row.toggleSelected(!!value)}
              aria-label="Select row"
              onClick={(e) => e.stopPropagation()}
            />
          </div>
        ),
        size: 40,
      },
      {
        accessorKey: 'status',
        header: '',
        size: 20,
        cell: ({ row }) => {
          const statusIcons: Record<Status, { icon: React.ReactNode; color: string }> = {
            pending: { 
              icon: <Clock className="h-3.5 w-3.5" />, 
              color: 'text-gray-400' 
            },
            processing: { 
              icon: <Loader2 className="h-3.5 w-3.5 animate-spin" />, 
              color: 'text-blue-500' 
            },
            completed: { 
              icon: <CheckCircle className="h-3.5 w-3.5" />, 
              color: 'text-green-600' 
            },
            failed: { 
              icon: <XCircle className="h-3.5 w-3.5" />, 
              color: 'text-red-500' 
            },
            cancelled: { 
              icon: <Ban className="h-3.5 w-3.5" />, 
              color: 'text-orange-500' 
            },
          };
          
          const status = statusIcons[row.original.status];
          
          return (
            <div className="flex items-center justify-center" title={row.original.status}>
              <span className={status.color}>
                {status.icon}
              </span>
            </div>
          );
        },
      },
      {
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
        cell: ({ row }) => (
          <div className="space-y-1">
            <div className="font-medium text-sm">{getOperationTypeLabel(row.original.type)}</div>
            <div className="font-mono text-xs text-gray-500">{row.original.id}</div>
          </div>
        ),
      },
      {
        accessorKey: 'subject',
        header: 'Subject',
        size: 200,
        cell: ({ row }) => {
          // Configuration for extracting subject fields by operation type
          const subjectConfig: Record<OperationType, { primary: string[], secondary: string[] }> = {
            safe_provision: {
              primary: ['safe_name', 'safeName'],
              secondary: ['safe_id', 'safeId', 'id']
            },
            safe_modify: {
              primary: ['safe_name', 'safeName'],
              secondary: ['safe_id', 'safeId', 'id']
            },
            safe_delete: {
              primary: ['safe_name', 'safeName'],
              secondary: ['safe_id', 'safeId', 'id']
            },
            access_grant: {
              primary: ['username', 'user', 'user_id'],
              secondary: ['safe_name', 'safeName', 'role', 'role_name', 'target']
            },
            access_revoke: {
              primary: ['username', 'user', 'user_id'],
              secondary: ['safe_name', 'safeName', 'role', 'role_name', 'target']
            },
            user_sync: {
              primary: ['username', 'user_id', 'email'],
              secondary: ['source', 'directory', 'domain']
            },
            safe_sync: {
              primary: ['safe_name', 'safeName'],
              secondary: ['safe_id', 'safeId', 'id']
            },
            group_sync: {
              primary: ['group_name', 'groupName', 'group_id'],
              secondary: ['source', 'directory', 'domain']
            }
          };
          
          // Helper function to extract value from payload using field paths
          const extractValue = (payload: any, fields: string[]): string => {
            for (const field of fields) {
              const value = field.split('.').reduce((obj, key) => obj?.[key], payload);
              if (value !== undefined && value !== null && value !== '') {
                return String(value);
              }
            }
            return '-';
          };
          
          const payload = row.original.payload as any;
          const config = subjectConfig[row.original.type];
          
          let primarySubject = '-';
          let secondarySubject = '-';
          
          if (config) {
            primarySubject = extractValue(payload, config.primary);
            secondarySubject = extractValue(payload, config.secondary);
          } else {
            // Fallback for unknown operation types
            primarySubject = extractValue(payload, ['name', 'target', 'id']);
            secondarySubject = extractValue(payload, ['description', 'type']);
          }
          
          return (
            <div className="space-y-1">
              <div className="text-sm font-medium text-gray-900 truncate" title={primarySubject}>
                {primarySubject}
              </div>
              <div className="text-xs text-gray-500 truncate" title={secondarySubject}>
                {secondarySubject}
              </div>
            </div>
          );
        },
      },
      {
        accessorKey: 'priority',
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
              Priority
              {column.getIsSorted() === "asc" && <ChevronUp className="ml-2 h-4 w-4" />}
              {column.getIsSorted() === "desc" && <ChevronDown className="ml-2 h-4 w-4" />}
            </Button>
          )
        },
        size: 100,
        cell: ({ row }) => {
          const priorityBadgeColors: Record<Priority, string> = {
            high: 'bg-red-100 text-red-700 border border-red-200',
            medium: 'bg-amber-100 text-amber-700 border border-amber-200', 
            normal: 'bg-blue-100 text-blue-700 border border-blue-200',
            low: 'bg-gray-100 text-gray-600 border border-gray-200',
          };
          
          return (
            <span className={`inline-block px-2 py-0.5 rounded text-[10px] font-medium uppercase ${priorityBadgeColors[row.original.priority]}`}>
              {row.original.priority}
            </span>
          );
        },
      },
      {
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
              <div className="text-sm">{format(createdDate, 'MMM d, HH:mm:ss')}</div>
              <div className="text-xs text-gray-500">{relativeTime}</div>
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
    pageCount: Math.max(1, Math.ceil(totalCount / pagination.pageSize)),
    state: {
      sorting,
      pagination,
      rowSelection,
    },
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    manualPagination: true,
    manualSorting: true,
    enableSorting: true,
  });

  const handleCancelOperation = async (id: string) => {
    try {
      await operationsApi.cancel(id);
      fetchData();
    } catch (error) {
      console.error('Failed to cancel operation:', error);
    }
  };

  const handleRowClick = (operation: Operation) => {
    setSelectedOperation(operation);
  };

  // Handle bulk cancel
  const handleBulkCancel = async () => {
    const selectedRows = table.getFilteredSelectedRowModel().rows;
    const selectedIds = selectedRows.map(row => row.original.id);
    
    if (selectedIds.length === 0) return;
    
    if (!confirm(`Are you sure you want to cancel ${selectedIds.length} operations?`)) {
      return;
    }
    
    try {
      // Cancel operations in parallel
      await Promise.all(selectedIds.map(id => operationsApi.cancel(id)));
      setRowSelection({});
      fetchData();
    } catch (error) {
      console.error('Failed to cancel operations:', error);
    }
  };

  // Export operations
  const handleExport = async () => {
    try {
      const params: any = {
        page_size: 1000, // Export more records
      };
      
      if (debouncedSearch) params.search = debouncedSearch;
      if (statusFilter !== 'all') params.status = statusFilter;
      if (priorityFilter !== 'all') params.priority = priorityFilter;
      if (typeFilter !== 'all') params.type = typeFilter;
      if (dateRange.start && dateRange.start !== '') params.start_date = dateRange.start;
      if (dateRange.end && dateRange.end !== '') params.end_date = dateRange.end;
      
      const response = await operationsApi.list(params);
      
      // Convert to CSV
      const headers = ['ID', 'Type', 'Status', 'Priority', 'Created At', 'Error Message'];
      const rows = response.operations.map(op => [
        op.id,
        op.type,
        op.status,
        op.priority,
        op.created_at,
        op.error_message || ''
      ]);
      
      const csv = [
        headers.join(','),
        ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
      ].join('\n');
      
      // Download
      const blob = new Blob([csv], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `operations-${format(new Date(), 'yyyy-MM-dd-HHmmss')}.csv`;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export operations:', error);
    }
  };

  if (loading && data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <PageContainer>
      {/* Search and Actions Bar */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              ref={searchInputRef}
              id="operations-search"
              placeholder="Search operations by ID, type, or error..."
              value={searchValue}
              onChange={(e) => {
                setSearchValue(e.target.value);
                setIsSearching(true);
              }}
              onBlur={() => setIsSearching(false)}
              className="pl-10 h-11"
            />
          </div>
        </div>
        
        <div className="flex gap-2">
          {Object.keys(rowSelection).length > 0 ? (
            <>
              <div className="flex items-center text-sm text-gray-600 px-3">
                <CheckCircle className="h-4 w-4 mr-2 text-blue-600" />
                {Object.keys(rowSelection).length} selected
              </div>
              <Button 
                variant="destructive" 
                size="default"
                onClick={handleBulkCancel}
              >
                <Ban className="h-4 w-4 mr-2" />
                Cancel Selected
              </Button>
              <Button 
                variant="ghost" 
                size="default"
                onClick={() => setRowSelection({})}
              >
                Clear
              </Button>
            </>
          ) : (
            <>
              <Button variant="outline" size="icon" onClick={fetchData} title="Refresh">
                <RefreshCw className="h-4 w-4" />
              </Button>
              <Button variant="outline" onClick={fetchStats}>
                <BarChart3 className="h-4 w-4 mr-2" />
                Stats
              </Button>
              <Button variant="outline" onClick={handleExport}>
                <Download className="h-4 w-4 mr-2" />
                Export
              </Button>
              <Button onClick={() => setShowNewOperationDialog(true)}>
                <Plus className="h-4 w-4 mr-2" />
                New Operation
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Advanced Filters */}
      <div className="space-y-4">
        {/* Main filters row */}
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
          {/* Status Filter */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-gray-600">Status</label>
            <Select value={statusFilter} onValueChange={(value: any) => setStatusFilter(value)}>
              <SelectTrigger className="h-9">
                <SelectValue>
                  {statusFilter === 'all' ? (
                    <span className="text-gray-500">All Statuses</span>
                  ) : (
                    <div className="flex items-center gap-2">
                      {getStatusIcon(statusFilter as Status)}
                      <span className="capitalize">{statusFilter}</span>
                    </div>
                  )}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Statuses</SelectItem>
                <SelectItem value="pending">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4" />
                    Pending
                  </div>
                </SelectItem>
                <SelectItem value="processing">
                  <div className="flex items-center gap-2">
                    <Loader2 className="h-4 w-4" />
                    Processing
                  </div>
                </SelectItem>
                <SelectItem value="completed">
                  <div className="flex items-center gap-2">
                    <CheckCircle className="h-4 w-4 text-green-600" />
                    Completed
                  </div>
                </SelectItem>
                <SelectItem value="failed">
                  <div className="flex items-center gap-2">
                    <XCircle className="h-4 w-4 text-red-600" />
                    Failed
                  </div>
                </SelectItem>
                <SelectItem value="cancelled">
                  <div className="flex items-center gap-2">
                    <Ban className="h-4 w-4 text-orange-600" />
                    Cancelled
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Priority Filter */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-gray-600">Priority</label>
            <Select value={priorityFilter} onValueChange={(value: any) => setPriorityFilter(value)}>
              <SelectTrigger className="h-9">
                <SelectValue>
                  {priorityFilter === 'all' ? (
                    <span className="text-gray-500">All Priorities</span>
                  ) : (
                    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${getPriorityColor(priorityFilter as Priority)}`}>
                      {priorityFilter}
                    </span>
                  )}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Priorities</SelectItem>
                <SelectItem value="high">
                  <span className="px-2 py-0.5 rounded-full text-xs font-medium text-red-600 bg-red-100">
                    High
                  </span>
                </SelectItem>
                <SelectItem value="medium">
                  <span className="px-2 py-0.5 rounded-full text-xs font-medium text-yellow-600 bg-yellow-100">
                    Medium
                  </span>
                </SelectItem>
                <SelectItem value="normal">
                  <span className="px-2 py-0.5 rounded-full text-xs font-medium text-blue-600 bg-blue-100">
                    Normal
                  </span>
                </SelectItem>
                <SelectItem value="low">
                  <span className="px-2 py-0.5 rounded-full text-xs font-medium text-gray-600 bg-gray-100">
                    Low
                  </span>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Type Filter */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-gray-600">Type</label>
            <Select value={typeFilter} onValueChange={setTypeFilter}>
              <SelectTrigger className="h-9">
                <SelectValue>
                  {typeFilter === 'all' ? (
                    <span className="text-gray-500">All Types</span>
                  ) : (
                    <span>{getOperationTypeLabel(typeFilter as any)}</span>
                  )}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value="safe_provision">Safe Provision</SelectItem>
                <SelectItem value="user_sync">User Sync</SelectItem>
                <SelectItem value="safe_sync">Safe Sync</SelectItem>
                <SelectItem value="group_sync">Group Sync</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Date filters row */}
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3">
          {/* Start Date */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-gray-600">Start Date</label>
            <div className="relative">
              <Calendar className="absolute left-2.5 top-1/2 transform -translate-y-1/2 h-3.5 w-3.5 text-gray-400" />
              <Input
                type="datetime-local"
                value={(() => {
                  if (!dateRange.start) return '';
                  const date = new Date(dateRange.start);
                  const year = date.getFullYear();
                  const month = String(date.getMonth() + 1).padStart(2, '0');
                  const day = String(date.getDate()).padStart(2, '0');
                  const hours = String(date.getHours()).padStart(2, '0');
                  const minutes = String(date.getMinutes()).padStart(2, '0');
                  return `${year}-${month}-${day}T${hours}:${minutes}`;
                })()}
                onChange={(e) => {
                  if (e.target.value) {
                    const date = new Date(e.target.value);
                    setDateRange({ ...dateRange, start: date.toISOString() });
                  } else {
                    setDateRange({ ...dateRange, start: '' });
                  }
                }}
                className="pl-8 h-9 text-xs w-full"
                title="Start date"
              />
            </div>
          </div>

          {/* End Date */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-gray-600">End Date</label>
            <div className="relative">
              <Calendar className="absolute left-2.5 top-1/2 transform -translate-y-1/2 h-3.5 w-3.5 text-gray-400" />
              <Input
                type="datetime-local"
                value={(() => {
                  if (!dateRange.end) return '';
                  const date = new Date(dateRange.end);
                  const year = date.getFullYear();
                  const month = String(date.getMonth() + 1).padStart(2, '0');
                  const day = String(date.getDate()).padStart(2, '0');
                  const hours = String(date.getHours()).padStart(2, '0');
                  const minutes = String(date.getMinutes()).padStart(2, '0');
                  return `${year}-${month}-${day}T${hours}:${minutes}`;
                })()}
                onChange={(e) => {
                  if (e.target.value) {
                    const date = new Date(e.target.value);
                    setDateRange({ ...dateRange, end: date.toISOString() });
                  } else {
                    setDateRange({ ...dateRange, end: '' });
                  }
                }}
                className="pl-8 h-9 text-xs w-full"
                title="End date"
              />
            </div>
          </div>
        </div>

        {/* Quick Filters */}
        <div className="flex items-center justify-between gap-2 pt-2">
          <div className="flex items-center gap-2">
            <span className="text-xs text-gray-500">Quick filters:</span>
            <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-xs"
              onClick={() => {
                const now = new Date();
                const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
                setDateRange({ 
                  start: today.toISOString(), 
                  end: now.toISOString() 
                });
              }}
            >
              Today
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-xs"
              onClick={() => {
                const now = new Date();
                const yesterday = new Date(now);
                yesterday.setDate(yesterday.getDate() - 1);
                yesterday.setHours(0, 0, 0, 0);
                const endOfYesterday = new Date(yesterday);
                endOfYesterday.setHours(23, 59, 59, 999);
                setDateRange({ 
                  start: yesterday.toISOString(), 
                  end: endOfYesterday.toISOString() 
                });
              }}
            >
              Yesterday
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-xs"
              onClick={() => {
                const now = new Date();
                const weekAgo = new Date(now);
                weekAgo.setDate(weekAgo.getDate() - 7);
                setDateRange({ 
                  start: weekAgo.toISOString(), 
                  end: now.toISOString() 
                });
              }}
            >
              Last 7 days
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-xs"
              onClick={() => {
                setStatusFilter('failed');
                setPriorityFilter('all');
                setTypeFilter('all');
              }}
            >
              Failed only
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-xs"
              onClick={() => {
                setStatusFilter('processing');
                setPriorityFilter('all');
                setTypeFilter('all');
              }}
            >
              In progress
            </Button>
          </div>
          </div>
          {(statusFilter !== 'all' || priorityFilter !== 'all' || typeFilter !== 'all' || dateRange.start || dateRange.end) && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setStatusFilter('all');
                setPriorityFilter('all');
                setTypeFilter('all');
                setDateRange({ start: '', end: '' });
              }}
              className="h-7 text-xs"
            >
              <XCircle className="h-3 w-3 mr-1" />
              Clear All Filters
            </Button>
          )}
        </div>
      </div>

      {/* Data Table */}
      <Card className="overflow-hidden">
        <CardContent className="p-0">
          <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead 
                        key={header.id} 
                        style={{ width: header.getSize() }}
                        className={header.column.id === 'status' ? 'px-0' : ''}
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
                          className={cell.column.id === 'status' ? 'px-0' : ''}
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
                />
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
              
              <Select
                value={pagination.pageSize.toString()}
                onValueChange={(value) => table.setPageSize(Number(value))}
              >
                <SelectTrigger className="w-24">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
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

      {/* Statistics Dialog */}
      <Dialog open={showStatsDialog} onOpenChange={setShowStatsDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Operation Statistics</DialogTitle>
            <DialogDescription>
              Aggregated statistics for the selected time period
            </DialogDescription>
          </DialogHeader>
          
          {stats && (
            <div className="grid grid-cols-3 gap-4 py-4">
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">By Status</CardTitle>
                </CardHeader>
                <CardContent>
                  {Object.entries(stats.by_status).map(([status, count]) => (
                    <div key={status} className="flex justify-between text-sm py-1">
                      <span className="capitalize">{status}:</span>
                      <span className="font-medium">{count as number}</span>
                    </div>
                  ))}
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">By Type</CardTitle>
                </CardHeader>
                <CardContent>
                  {Object.entries(stats.by_type).map(([type, count]) => (
                    <div key={type} className="flex justify-between text-sm py-1">
                      <span>{getOperationTypeLabel(type as any)}:</span>
                      <span className="font-medium">{count as number}</span>
                    </div>
                  ))}
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">Performance</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2 text-sm">
                    <div>
                      <span className="text-gray-500">Total Operations:</span>
                      <p className="font-medium">{stats.total_count.toLocaleString()}</p>
                    </div>
                    <div>
                      <span className="text-gray-500">Avg Wait Time:</span>
                      <p className="font-medium">{stats.avg_wait_time_seconds.toFixed(1)}s</p>
                    </div>
                    <div>
                      <span className="text-gray-500">Avg Process Time:</span>
                      <p className="font-medium">{stats.avg_process_time_seconds.toFixed(1)}s</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* New Operation Dialog */}
      <Dialog open={showNewOperationDialog} onOpenChange={setShowNewOperationDialog}>
        <DialogContent className="bg-white">
          <DialogHeader>
            <DialogTitle>Create New Operation</DialogTitle>
            <DialogDescription>
              Create a new operation to be processed by the pipeline.
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Operation Type</label>
              <Select defaultValue="user_sync">
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
              <Select defaultValue="normal">
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
            <Button variant="outline" onClick={() => setShowNewOperationDialog(false)}>
              Cancel
            </Button>
            <Button onClick={() => {
              // TODO: Implement operation creation
              setShowNewOperationDialog(false);
              fetchData();
            }}>
              Create Operation
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Operation Details Panel */}
      <OperationDetailsPanel 
        operation={selectedOperation} 
        onClose={() => setSelectedOperation(null)}
        onUpdate={fetchData}
      />
    </PageContainer>
  );
}