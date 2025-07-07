import { PageHeader } from '@/components/PageHeader';
import { PageContainer } from '@/components/PageContainer';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';

export function SettingsSSO() {
  return (
    <PageContainer>
      <PageHeader
        title="SSO Configuration"
        description="Configure Single Sign-On providers for authentication"
        actions={
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            Add Provider
          </Button>
        }
      />
      <div className="bg-white border border-gray-200 rounded p-6">
        <div className="space-y-4">
          <div className="border rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">SAML Configuration</h4>
            <p className="text-sm text-gray-600">Set up SAML-based SSO providers.</p>
          </div>
          <div className="border rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">OAuth/OIDC Configuration</h4>
            <p className="text-sm text-gray-600">Configure OAuth 2.0 or OpenID Connect providers.</p>
          </div>
        </div>
      </div>
    </PageContainer>
  );
}

export default SettingsSSO;