import { PageHeader } from '@/components/PageHeader';
import { PageContainer } from '@/components/PageContainer';
import { Button } from '@/components/ui/button';
import { Database, Download } from 'lucide-react';

export function SettingsDatabase() {
  return (
    <PageContainer>
      <PageHeader
        title="Database Settings"
        description="Database configuration and maintenance options"
        actions={
          <div className="flex gap-2">
            <Button variant="outline">
              <Download className="h-4 w-4 mr-2" />
              Export Data
            </Button>
            <Button variant="outline">
              <Database className="h-4 w-4 mr-2" />
              Backup Now
            </Button>
          </div>
        }
      />
      <div className="bg-white border border-gray-200 rounded p-6">
        <p className="text-gray-600">Database configuration options will be displayed here.</p>
      </div>
    </PageContainer>
  );
}

export default SettingsDatabase;