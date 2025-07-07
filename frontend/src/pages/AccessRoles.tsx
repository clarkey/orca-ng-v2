import { PageHeader } from '@/components/PageHeader';
import { PageContainer } from '@/components/PageContainer';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';

export function AccessRoles() {
  return (
    <PageContainer>
      <PageHeader
        title="Access Roles"
        description="Manage logical safe access roles that abstract and aggregate permissions on safes within CyberArk"
        actions={
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            Create Role
          </Button>
        }
      />
      <div className="bg-white border border-gray-200 rounded p-6">
        <div className="space-y-4">
          <div className="border rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">Role Templates</h4>
            <p className="text-sm text-gray-600">Configure role templates for common access patterns.</p>
          </div>
          <div className="border rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">Permission Mappings</h4>
            <p className="text-sm text-gray-600">Define how permissions map to CyberArk safe access controls.</p>
          </div>
        </div>
      </div>
    </PageContainer>
  );
}

export default AccessRoles;