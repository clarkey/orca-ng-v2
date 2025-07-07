import { PageHeader } from '@/components/PageHeader';
import { PageContainer } from '@/components/PageContainer';
import { Button } from '@/components/ui/button';
import { Key } from 'lucide-react';

export function SettingsAPI() {
  return (
    <PageContainer>
      <PageHeader
        title="API Settings"
        description="Configure API access, rate limits, and integration settings"
        actions={
          <Button>
            <Key className="h-4 w-4 mr-2" />
            Generate API Key
          </Button>
        }
      />
      <div className="bg-white border border-gray-200 rounded p-6">
        <p className="text-gray-600">API configuration and access management will be displayed here.</p>
      </div>
    </PageContainer>
  );
}

export default SettingsAPI;