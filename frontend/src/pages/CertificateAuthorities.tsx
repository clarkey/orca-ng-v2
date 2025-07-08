import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
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
import { Plus, Shield, Check, X, Loader2, ChevronUp, ChevronDown, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { certificateAuthoritiesApi, CertificateAuthorityInfo } from '@/api/certificateAuthorities';
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
import { PageContainer } from '@/components/PageContainer';
import { PageHeader } from '@/components/PageHeader';
import { CertificateAuthorityForm } from '@/components/CertificateAuthorityForm';
import { ConfirmationDialog } from '@/components/ui/confirmation-dialog';

const STORAGE_KEY_PAGE_SIZE = 'orca-certificate-authorities-page-size';

export function CertificateAuthorities() {
  const queryClient = useQueryClient();
  const [selectedCA, setSelectedCA] = useState<CertificateAuthorityInfo | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [caToDelete, setCaToDelete] = useState<CertificateAuthorityInfo | null>(null);
  
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

  // Fetch data
  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ['certificate-authorities'],
    queryFn: certificateAuthoritiesApi.list,
  });

  const deleteMutation = useMutation({
    mutationFn: certificateAuthoritiesApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificate-authorities'] });
      alert('Certificate authority deleted successfully');
      setShowDeleteDialog(false);
      setCaToDelete(null);
      refetch();
    },
    onError: (error: any) => {
      alert(error.response?.data?.error || 'Failed to delete certificate authority');
    },
  });


  // Define columns
  const columns = useMemo<ColumnDef<CertificateAuthorityInfo>[]>(
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
              Name
              {column.getIsSorted() === "asc" && <ChevronUp className="ml-2 h-4 w-4" />}
              {column.getIsSorted() === "desc" && <ChevronDown className="ml-2 h-4 w-4" />}
            </Button>
          )
        },
        size: 250,
        cell: ({ row }) => (
          <div>
            <div className="font-medium">{row.original.name}</div>
            {row.original.description && (
              <div className="text-sm text-gray-500 mt-1">{row.original.description}</div>
            )}
          </div>
        ),
      },
      {
        id: 'subject',
        accessorKey: 'subject',
        header: 'Subject',
        size: 350,
        cell: ({ row }) => (
          <div className="text-sm">
            <div className="truncate max-w-[350px]" title={row.original.subject}>
              {row.original.subject}
            </div>
            <div className="text-xs text-gray-500 font-mono mt-1">
              SHA256: {row.original.fingerprint.substring(0, 16)}...
            </div>
          </div>
        ),
      },
      {
        id: 'issuer',
        accessorKey: 'issuer',
        header: 'Issuer',
        size: 250,
        cell: ({ row }) => (
          <div className="text-sm truncate max-w-[250px]" title={row.original.issuer}>
            {row.original.issuer}
          </div>
        ),
      },
      {
        id: 'validity',
        accessorKey: 'expires_in_days',
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
              Validity
              {column.getIsSorted() === "asc" && <ChevronUp className="ml-2 h-4 w-4" />}
              {column.getIsSorted() === "desc" && <ChevronDown className="ml-2 h-4 w-4" />}
            </Button>
          )
        },
        size: 180,
        cell: ({ row }) => {
          const ca = row.original;
          if (ca.is_expired) {
            return <Badge variant="destructive">Expired</Badge>;
          }
          if (ca.expires_in_days <= 30) {
            return <Badge variant="destructive">Expires in {ca.expires_in_days} days</Badge>;
          }
          if (ca.expires_in_days <= 90) {
            return <Badge variant="secondary">Expires in {ca.expires_in_days} days</Badge>;
          }
          return <Badge variant="outline">Valid for {ca.expires_in_days} days</Badge>;
        },
      },
      {
        id: 'status_badge',
        accessorKey: 'is_active',
        header: 'Status',
        size: 100,
        cell: ({ row }) => {
          const ca = row.original;
          if (ca.is_expired) {
            return (
              <Badge variant="destructive" className="gap-1">
                <X className="h-3 w-3" />
                Expired
              </Badge>
            );
          }
          return ca.is_active ? (
            <Badge variant="success" className="gap-1">
              <Check className="h-3 w-3" />
              Active
            </Badge>
          ) : (
            <Badge variant="secondary" className="gap-1">
              <X className="h-3 w-3" />
              Inactive
            </Badge>
          );
        },
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
    () => data?.certificate_authorities || [],
    [data?.certificate_authorities]
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

  const handleEdit = (ca: CertificateAuthorityInfo) => {
    setSelectedCA(ca);
    setShowForm(true);
  };

  const handleFormClose = () => {
    setShowForm(false);
    setSelectedCA(null);
  };

  const handleFormSuccess = () => {
    refetch();
    handleFormClose();
  };

  const handleDeleteRequest = (ca: CertificateAuthorityInfo) => {
    setCaToDelete(ca);
    setShowDeleteDialog(true);
  };

  if (isLoading && tableData.length === 0) {
    return (
      <PageContainer>
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin" />
        </div>
      </PageContainer>
    );
  }

  if (error) {
    return (
      <PageContainer>
        <div className="text-center py-8 text-red-500">
          Failed to load certificate authorities
        </div>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <PageHeader
        title="Certificate Authorities"
        description="Manage trusted certificate authorities for secure connections"
        actions={
          <Button onClick={() => setShowForm(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Add Certificate
          </Button>
        }
      />


      {/* Data Table */}
      <Card className="overflow-hidden">
        <CardContent className="p-0 relative">
          {/* Loading overlay for refetch */}
          {isFetching && !isLoading && (
            <div className="absolute inset-0 bg-white/50 backdrop-blur-sm z-10 flex items-center justify-center">
              <Loader2 className="h-6 w-6 animate-spin text-gray-600" />
            </div>
          )}
          
          {tableData.length === 0 ? (
            <div className="text-center py-16">
              <Shield className="h-12 w-12 mx-auto mb-4 text-gray-400" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                No certificate authorities configured
              </h3>
              <p className="text-sm text-gray-500 mb-6">
                Add certificate authorities to establish secure connections
              </p>
              <Button onClick={() => setShowForm(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Add Certificate
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
                        No certificate authorities found.
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
                  of {table.getFilteredRowModel().rows.length} certificate{table.getFilteredRowModel().rows.length !== 1 ? 's' : ''}
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
                      value={table.getState().pagination.pageIndex + 1}
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

      {/* Certificate Authority Form Dialog */}
      <CertificateAuthorityForm
        open={showForm}
        onClose={handleFormClose}
        onSuccess={handleFormSuccess}
        certificateAuthority={selectedCA}
        onDelete={handleDeleteRequest}
      />

      {/* Delete Confirmation Dialog */}
      <ConfirmationDialog
        open={showDeleteDialog}
        onOpenChange={setShowDeleteDialog}
        title="Delete Certificate Authority"
        description={`Are you sure you want to delete "${caToDelete?.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={() => {
          if (caToDelete) {
            deleteMutation.mutate(caToDelete.id);
          }
        }}
      />

    </PageContainer>
  );
}