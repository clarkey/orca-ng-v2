import { useState } from 'react';
import { PageContainer } from '@/components/PageContainer';
import { PageHeader } from '@/components/PageHeader';
import { Button } from '@/components/ui/button';
import { Search, Users, UserPlus } from 'lucide-react';
import { Input } from '@/components/ui/input';

export default function UsersAndGroups() {
  const [activeTab, setActiveTab] = useState('users');

  return (
    <PageContainer>
      <PageHeader
        title="Users & Groups"
        description="Manage CyberArk users and groups"
        actions={
          <>
            <Button variant="outline" size="sm">
              <UserPlus className="h-4 w-4 mr-2" />
              Add User
            </Button>
            <Button size="sm">
              <Users className="h-4 w-4 mr-2" />
              Create Group
            </Button>
          </>
        }
      />

      <div className="space-y-4">
        {/* Search bar */}
        <div className="relative max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input
            placeholder="Search users or groups..."
            className="pl-10"
          />
        </div>

        {/* Manual tabs implementation */}
        <div className="w-full">
          <div className="inline-flex h-10 items-center justify-center rounded bg-gray-100 p-1 text-gray-500 max-w-md w-full">
            <button
              onClick={() => setActiveTab('users')}
              className={`inline-flex items-center justify-center whitespace-nowrap rounded px-3 py-1.5 text-sm font-medium w-1/2 ${
                activeTab === 'users'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Users
            </button>
            <button
              onClick={() => setActiveTab('groups')}
              className={`inline-flex items-center justify-center whitespace-nowrap rounded px-3 py-1.5 text-sm font-medium w-1/2 ${
                activeTab === 'groups'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Groups
            </button>
          </div>
          
          <div className="mt-4">
            {activeTab === 'users' && (
              <div className="rounded border bg-white">
                <div className="p-8 text-center text-gray-500">
                  <p>Users table will be displayed here</p>
                </div>
              </div>
            )}
            
            {activeTab === 'groups' && (
              <div className="rounded border bg-white">
                <div className="p-8 text-center text-gray-500">
                  <p>Groups table will be displayed here</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </PageContainer>
  );
}