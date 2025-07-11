import { useState, useMemo } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  ColumnDef,
  flexRender,
  SortingState,
  PaginationState,
} from '@tanstack/react-table';
import { PageContainer } from '@/components/PageContainer';
import { PageHeader } from '@/components/PageHeader';
import { Button } from '@/components/ui/button';
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
import { Switch } from '@/components/ui/switch';
import { CyberArkInstanceForm } from '@/components/CyberArkInstanceForm';
import { ConfirmationDialog } from '@/components/ui/confirmation-dialog';
import { format, formatDistanceToNow } from 'date-fns';
import {
  Plus,
  AlertCircle,
  Server,
  Globe,
  User,
  Clock,
  Loader2,
  ShieldOff,
  ChevronUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Check,
  CheckCircle,
  X,
  RefreshCw,
  Users as UsersIcon,
  Shield as ShieldIcon,
  FileText,
  MoreVertical,
} from 'lucide-react';
import { CyberArkInstance } from '@/api/cyberark';
import { 
  useCyberArkInstances, 
  useDeleteCyberArkInstance, 
  useUpdateCyberArkInstance 
} from '@/hooks/useCyberArkInstances';
import { useInstanceSyncConfigs, useTriggerSync } from '@/hooks/useSyncJobs';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { InstanceSyncStatus } from '@/components/InstanceSyncStatus';
import { InstanceSyncActions } from '@/components/InstanceSyncActions';

const STORAGE_KEY_PAGE_SIZE = 'cyberark-instances-page-size';

export function Instances() {
  const [showForm, setShowForm] = useState(false);
  const [selectedInstance, setSelectedInstance] = useState<CyberArkInstance | null>(null);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [instanceToDelete, setInstanceToDelete] = useState<CyberArkInstance | null>(null);
  
  // Table state
  const [sorting, setSorting] = useState<SortingState>([]);
  
  // Initialize pagination with saved page size
  const [pagination, setPagination] = useState<PaginationState>(() => {
    const savedPageSize = localStorage.getItem(STORAGE_KEY_PAGE_SIZE);
    return {
      pageIndex: 0,
      pageSize: savedPageSize ? parseInt(savedPageSize, 10) : 20,
    };
  });

  const { data: response, isLoading, refetch } = useCyberArkInstances();
  const deleteMutation = useDeleteCyberArkInstance();
  const updateMutation = useUpdateCyberArkInstance();

  const instances = response?.instances || [];


  const handleToggleActive = async (instance: CyberArkInstance) => {
    try {
      await updateMutation.mutateAsync({
        id: instance.id,
        data: { is_active: !instance.is_active }
      });
    } catch (error) {
      console.error('Failed to update instance:', error);
      alert('Failed to update instance');
    }
  };

  const handleDelete = async () => {
    if (!instanceToDelete) return;
    
    try {
      await deleteMutation.mutateAsync(instanceToDelete.id);
      alert('CyberArk instance deleted successfully');
      setShowDeleteDialog(false);
      setInstanceToDelete(null);
      refetch();
    } catch (error) {
      console.error('Failed to delete instance:', error);
      alert('Failed to delete instance');
    }
  };

  const handleDeleteRequest = (instance: CyberArkInstance) => {
    setInstanceToDelete(instance);
    setShowDeleteDialog(true);
  };

  const handleEdit = (instance: CyberArkInstance) => {
    setSelectedInstance(instance);
    setShowForm(true);
  };

  const handleFormClose = () => {
    setShowForm(false);
    setSelectedInstance(null);
  };

  const handleFormSuccess = () => {
    refetch();
    handleFormClose();
  };

  // Define columns
  const columns = useMemo<ColumnDef<CyberArkInstance>[]>(
    () => [
      {
        id: 'name',
        accessorKey: 'name',
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
              Instance
              {column.getIsSorted() === "asc" && <ChevronUp className="ml-2 h-4 w-4" />}
              {column.getIsSorted() === "desc" && <ChevronDown className="ml-2 h-4 w-4" />}
            </Button>
          )
        },
        size: 250,
        cell: ({ row }) => (
          <button
            onClick={() => handleEdit(row.original)}
            className="text-left hover:underline focus:outline-none"
          >
            <div className="font-medium">{row.original.name}</div>
            <div className="text-xs text-gray-500 font-mono">{row.original.id}</div>
          </button>
        ),
      },
      {
        id: 'connection',
        accessorKey: 'base_url',
        header: 'Connection',
        size: 350,
        cell: ({ row }) => (
          <div className="space-y-1">
            <div className="flex items-center gap-2 text-sm">
              <Globe className="h-3.5 w-3.5 text-gray-400" />
              <span className="font-mono text-xs truncate max-w-[300px]" title={row.original.base_url}>
                {row.original.base_url}
              </span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <User className="h-3.5 w-3.5 text-gray-400" />
              <span className="text-gray-600">{row.original.username}</span>
            </div>
            {row.original.skip_tls_verify && (
              <div className="flex items-center gap-2 text-sm mt-1">
                <ShieldOff className="h-3.5 w-3.5 text-orange-500" />
                <span className="text-orange-600 text-xs">TLS verification disabled</span>
              </div>
            )}
          </div>
        ),
      },
      {
        id: 'status',
        accessorKey: 'last_test_success',
        header: 'Connection',
        size: 120,
        cell: ({ row }) => {
          const instance = row.original;
          if (!instance.last_test_at) {
            return (
              <Badge variant="secondary" className="gap-1">
                <AlertCircle className="h-3 w-3" />
                Not tested
              </Badge>
            );
          }
          
          // Calculate how long ago the test was performed
          const testDate = new Date(instance.last_test_at);
          const now = new Date();
          const hoursSinceTest = (now.getTime() - testDate.getTime()) / (1000 * 60 * 60);
          
          // If test is older than 24 hours, consider it stale
          if (hoursSinceTest > 24) {
            return (
              <Badge variant="secondary" className="gap-1">
                <AlertCircle className="h-3 w-3" />
                Unknown
              </Badge>
            );
          }
          
          // If test is older than 1 hour but less than 24 hours, show warning
          if (hoursSinceTest > 1) {
            if (instance.last_test_success) {
              return (
                <Badge variant="warning" className="gap-1">
                  <Clock className="h-3 w-3" />
                  Stale
                </Badge>
              );
            }
          }
          
          // Recent test results
          if (instance.last_test_success) {
            return (
              <Badge variant="success" className="gap-1">
                <Check className="h-3 w-3" />
                Connected
              </Badge>
            );
          }
          return (
            <Badge variant="destructive" className="gap-1">
              <X className="h-3 w-3" />
              Failed
            </Badge>
          );
        },
      },
      {
        id: 'sync_status',
        header: 'Sync Status',
        size: 180,
        cell: ({ row }) => (
          <InstanceSyncStatus instanceId={row.original.id} />
        ),
      },
      {
        id: 'last_test_at',
        accessorKey: 'last_test_at',
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
              Last Tested
              {column.getIsSorted() === "asc" && <ChevronUp className="ml-2 h-4 w-4" />}
              {column.getIsSorted() === "desc" && <ChevronDown className="ml-2 h-4 w-4" />}
            </Button>
          )
        },
        size: 180,
        cell: ({ row }) => {
          if (!row.original.last_test_at) {
            return <span className="text-sm text-gray-400">Never</span>;
          }
          
          const testDate = new Date(row.original.last_test_at);
          const now = new Date();
          const hoursSinceTest = (now.getTime() - testDate.getTime()) / (1000 * 60 * 60);
          
          let textColor = "text-gray-600";
          let iconColor = "";
          
          if (hoursSinceTest > 24) {
            textColor = "text-gray-400";
            iconColor = "text-gray-400";
          } else if (hoursSinceTest > 1) {
            textColor = "text-yellow-600";
            iconColor = "text-yellow-600";
          }
          
          return (
            <div className={`flex items-center gap-2 text-sm ${textColor}`}>
              <Clock className={`h-3.5 w-3.5 ${iconColor}`} />
              {format(testDate, 'MMM d, HH:mm')}
            </div>
          );
        },
      },
      {
        id: 'is_active',
        accessorKey: 'is_active',
        header: 'Active',
        size: 100,
        cell: ({ row }) => (
          <div className="flex justify-center">
            <Switch
              checked={row.original.is_active}
              onCheckedChange={() => handleToggleActive(row.original)}
              aria-label={`Toggle ${row.original.name} active state`}
              onClick={(e) => e.stopPropagation()}
            />
          </div>
        ),
      },
      {
        id: 'actions',
        header: '',
        size: 140,
        cell: ({ row }) => (
          <div className="flex justify-end gap-2">
            <InstanceSyncActions 
              instanceId={row.original.id} 
              instanceName={row.original.name}
              size="sm"
            />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => handleEdit(row.original)}>
                  Edit Instance
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={() => handleDeleteRequest(row.original)}
                  className="text-red-600"
                >
                  Delete Instance
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        ),
      },
    ],
    []
  );

  const tableData = useMemo(
    () => instances || [],
    [instances]
  );

  const table = useReactTable({
    data: tableData,
    columns,
    state: {
      sorting,
      pagination,
    },
    onSortingChange: setSorting,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  if (isLoading && tableData.length === 0) {
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
        title="Instance Administration"
        description="Manage CyberArk PVWA instances and their synchronization settings"
        actions={
          <Button onClick={() => setShowForm(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Instance
          </Button>
        }
      />

      {/* Summary Stats */}
      {instances.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <Card className="px-6 py-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Instances</p>
                <p className="text-2xl font-semibold">{instances.length}</p>
              </div>
              <Server className="h-8 w-8 text-gray-400" />
            </div>
          </Card>
          <Card className="px-6 py-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Active</p>
                <p className="text-2xl font-semibold">
                  {instances.filter(i => i.is_active).length}
                </p>
              </div>
              <CheckCircle className="h-8 w-8 text-green-500" />
            </div>
          </Card>
          <Card className="px-6 py-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Connected</p>
                <p className="text-2xl font-semibold">
                  {instances.filter(i => i.last_test_success).length}
                </p>
              </div>
              <Globe className="h-8 w-8 text-blue-500" />
            </div>
          </Card>
          <Card className="px-6 py-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Sync Active</p>
                <p className="text-2xl font-semibold">
                  {instances.filter(i => i.is_active).length}
                </p>
              </div>
              <RefreshCw className="h-8 w-8 text-purple-500" />
            </div>
          </Card>
        </div>
      )}


      {/* Data Table */}
      <Card className="overflow-hidden">
        <CardContent className="p-0 relative">
          {/* Loading overlay for refetch */}
          {isLoading && !isLoading && (
            <div className="absolute inset-0 bg-white/50 backdrop-blur-sm z-10 flex items-center justify-center">
              <Loader2 className="h-6 w-6 animate-spin text-gray-600" />
            </div>
          )}
          
          {tableData.length === 0 ? (
            <div className="text-center py-16">
              <Server className="h-12 w-12 mx-auto mb-4 text-gray-400" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                No instances configured
              </h3>
              <p className="text-sm text-gray-500 mb-6">
                Add your first CyberArk PVWA instance to start synchronizing data
              </p>
              <Button onClick={() => setShowForm(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Add Instance
              </Button>
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
                      <TableRow
                        key={row.id}
                        className="hover:bg-gray-50"
                      >
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
                        No instances found.
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
                  of {table.getFilteredRowModel().rows.length} instance{table.getFilteredRowModel().rows.length !== 1 ? 's' : ''}
                </div>
                
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.setPageIndex(0)}
                    disabled={!table.getCanPreviousPage() || isLoading}
                  >
                    <ChevronsLeft className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.previousPage()}
                    disabled={!table.getCanPreviousPage() || isLoading}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  
                  <div className="flex items-center gap-1">
                    <span className="text-sm">Page</span>
                    <Input
                      type="number"
                      value={table.getState().pagination.pageIndex + 1}
                      onChange={(e) => {
                        const page = e.target.value ? Number(e.target.value) - 1 : 0;
                        table.setPageIndex(page);
                      }}
                      className="w-16 text-center"
                      min={1}
                      max={table.getPageCount()}
                      disabled={isLoading}
                    />
                    <span className="text-sm">of {table.getPageCount()}</span>
                  </div>
                  
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.nextPage()}
                    disabled={!table.getCanNextPage() || isLoading}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => table.setPageIndex(table.getPageCount() - 1)}
                    disabled={!table.getCanNextPage() || isLoading}
                  >
                    <ChevronsRight className="h-4 w-4" />
                  </Button>
                  
                  <Select
                    value={table.getState().pagination.pageSize.toString()}
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
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* CyberArk Instance Form Dialog */}
      <CyberArkInstanceForm
        open={showForm}
        onClose={handleFormClose}
        onSuccess={handleFormSuccess}
        instance={selectedInstance}
        onDelete={handleDeleteRequest}
      />

      {/* Delete Confirmation Dialog */}
      <ConfirmationDialog
        open={showDeleteDialog}
        onOpenChange={setShowDeleteDialog}
        title="Delete CyberArk Instance"
        description={`Are you sure you want to delete "${instanceToDelete?.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={handleDelete}
      />
    </PageContainer>
  );
}

export default Instances;