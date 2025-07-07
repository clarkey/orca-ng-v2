import React, { useEffect, useState } from 'react';
import { PageContainer } from '@/components/PageContainer';
import { PageHeader } from '@/components/PageHeader';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Switch } from '@/components/ui/switch';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { CyberArkInstanceForm } from '@/components/CyberArkInstanceForm';
import { format } from 'date-fns';
import {
  Plus,
  MoreHorizontal,
  Edit,
  Trash2,
  RefreshCw,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Server,
  Globe,
  User,
  Clock,
  Loader2,
} from 'lucide-react';
import { cyberarkApi, CyberArkInstance } from '@/api/cyberark';

export function Instances() {
  const [instances, setInstances] = useState<CyberArkInstance[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingInstance, setEditingInstance] = useState<CyberArkInstance | null>(null);
  const [testingInstanceId, setTestingInstanceId] = useState<string | null>(null);

  useEffect(() => {
    fetchInstances();
  }, []);

  const fetchInstances = async () => {
    try {
      const response = await cyberarkApi.listInstances();
      setInstances(response.instances);
    } catch (error) {
      console.error('Failed to fetch instances:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleTestConnection = async (instance: CyberArkInstance) => {
    setTestingInstanceId(instance.id);
    try {
      const result = await cyberarkApi.testInstanceConnection(instance.id);
      // Refresh to show updated test results
      await fetchInstances();
      
      // Show success/failure message
      if (!result.success) {
        alert(result.message);
      }
    } catch (error) {
      console.error('Failed to test connection:', error);
      alert('Failed to test connection');
    } finally {
      setTestingInstanceId(null);
    }
  };

  const handleToggleActive = async (instance: CyberArkInstance) => {
    try {
      await cyberarkApi.updateInstance(instance.id, {
        is_active: !instance.is_active,
      });
      await fetchInstances();
    } catch (error) {
      console.error('Failed to update instance:', error);
      alert('Failed to update instance');
    }
  };

  const handleDelete = async (instance: CyberArkInstance) => {
    if (!confirm(`Are you sure you want to delete "${instance.name}"?`)) {
      return;
    }

    try {
      await cyberarkApi.deleteInstance(instance.id);
      await fetchInstances();
    } catch (error) {
      console.error('Failed to delete instance:', error);
      alert('Failed to delete instance');
    }
  };

  const handleEdit = (instance: CyberArkInstance) => {
    setEditingInstance(instance);
    setShowForm(true);
  };

  const handleFormClose = () => {
    setShowForm(false);
    setEditingInstance(null);
  };

  const handleFormSuccess = () => {
    fetchInstances();
    handleFormClose();
  };

  const getStatusIcon = (instance: CyberArkInstance) => {
    if (testingInstanceId === instance.id) {
      return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
    }

    if (!instance.last_test_at) {
      return <AlertCircle className="h-4 w-4 text-gray-400" />;
    }

    if (instance.last_test_success) {
      return <CheckCircle2 className="h-4 w-4 text-green-500" />;
    }

    return <XCircle className="h-4 w-4 text-red-500" />;
  };

  const getStatusText = (instance: CyberArkInstance) => {
    if (testingInstanceId === instance.id) {
      return 'Testing...';
    }

    if (!instance.last_test_at) {
      return 'Not tested';
    }

    if (instance.last_test_success) {
      return 'Connected';
    }

    return 'Failed';
  };

  if (loading) {
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

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Instance</TableHead>
                <TableHead>Connection</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Last Tested</TableHead>
                <TableHead className="text-center">Active</TableHead>
                <TableHead className="w-[50px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {instances.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                    No instances configured. Click "Add Instance" to get started.
                  </TableCell>
                </TableRow>
              ) : (
                instances.map((instance) => (
                  <TableRow key={instance.id}>
                    <TableCell>
                      <div className="flex items-start gap-3">
                        <Server className="h-5 w-5 text-gray-400 mt-0.5" />
                        <div>
                          <div className="font-medium">{instance.name}</div>
                          <div className="text-xs text-gray-500 font-mono">{instance.id}</div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="space-y-1">
                        <div className="flex items-center gap-2 text-sm">
                          <Globe className="h-3.5 w-3.5 text-gray-400" />
                          <span className="font-mono text-xs">{instance.base_url}</span>
                        </div>
                        <div className="flex items-center gap-2 text-sm">
                          <User className="h-3.5 w-3.5 text-gray-400" />
                          <span className="text-gray-600">{instance.username}</span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {getStatusIcon(instance)}
                        <span className="text-sm">{getStatusText(instance)}</span>
                      </div>
                      {instance.last_test_error && !instance.last_test_success && (
                        <div className="text-xs text-red-600 mt-1 max-w-xs truncate" title={instance.last_test_error}>
                          {instance.last_test_error}
                        </div>
                      )}
                    </TableCell>
                    <TableCell>
                      {instance.last_test_at ? (
                        <div className="flex items-center gap-2 text-sm text-gray-600">
                          <Clock className="h-3.5 w-3.5" />
                          {format(new Date(instance.last_test_at), 'MMM d, HH:mm')}
                        </div>
                      ) : (
                        <span className="text-sm text-gray-400">Never</span>
                      )}
                    </TableCell>
                    <TableCell className="text-center">
                      <Switch
                        checked={instance.is_active}
                        onCheckedChange={() => handleToggleActive(instance)}
                        aria-label={`Toggle ${instance.name} active state`}
                      />
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuLabel>Actions</DropdownMenuLabel>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem 
                            onClick={() => handleTestConnection(instance)}
                            disabled={testingInstanceId === instance.id}
                          >
                            <RefreshCw className="h-4 w-4 mr-2" />
                            Test Connection
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleEdit(instance)}>
                            <Edit className="h-4 w-4 mr-2" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => handleDelete(instance)}
                            className="text-red-600"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <CyberArkInstanceForm
        open={showForm}
        onClose={handleFormClose}
        onSuccess={handleFormSuccess}
        instance={editingInstance}
      />
    </PageContainer>
  );
}

export default Instances;