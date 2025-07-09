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
import { format } from 'date-fns';
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
  X,
} from 'lucide-react';
import { CyberArkInstance } from '@/api/cyberark';
import { 
  useCyberArkInstances, 
  useDeleteCyberArkInstance, 
  useUpdateCyberArkInstance 
} from '@/hooks/useCyberArkInstances';

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
          <div>
            <div className="font-medium">{row.original.name}</div>
            <div className="text-xs text-gray-500 font-mono">{row.original.id}</div>
          </div>
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
        header: 'Status',
        size: 150,
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
        cell: ({ row }) => (
          row.original.last_test_at ? (
            <div className="flex items-center gap-2 text-sm text-gray-600">
              <Clock className="h-3.5 w-3.5" />
              {format(new Date(row.original.last_test_at), 'MMM d, HH:mm')}
            </div>
          ) : (
            <span className="text-sm text-gray-400">Never</span>
          )
        ),
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
        size: 40,
        cell: () => (
          <div className="flex justify-end">
            <ChevronRight className="h-4 w-4 text-gray-400" />
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
        title="CyberArk Instances"
        description="Manage your CyberArk PVWA instance configurations"
        actions={
          <Button onClick={() => setShowForm(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Instance
          </Button>
        }
      />


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
                No CyberArk instances configured
              </h3>
              <p className="text-sm text-gray-500 mb-6">
                Add your first CyberArk PVWA instance to get started
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
                        className="cursor-pointer hover:bg-gray-50"
                        onClick={() => handleEdit(row.original)}
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