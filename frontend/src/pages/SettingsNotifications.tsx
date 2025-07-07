import { PageHeader } from '@/components/PageHeader';
import { PageContainer } from '@/components/PageContainer';

export function SettingsNotifications() {
  return (
    <PageContainer>
      <PageHeader
        title="Notifications"
        description="Configure notification preferences and alert settings"
      />
      <div className="bg-white border border-gray-200 rounded p-6">
        <p className="text-gray-600">Notification settings will be displayed here.</p>
      </div>
    </PageContainer>
  );
}

export default SettingsNotifications;