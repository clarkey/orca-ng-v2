import { PageHeader } from '@/components/PageHeader';
import { PageContainer } from '@/components/PageContainer';

export function SettingsGeneral() {
  return (
    <PageContainer>
      <PageHeader
        title="General Settings"
        description="Configure general application settings and preferences"
      />
      <div className="bg-white border border-gray-200 rounded p-6">
        <p className="text-gray-600">General application settings will be displayed here.</p>
      </div>
    </PageContainer>
  );
}

export default SettingsGeneral;