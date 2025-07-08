import { PageContainer } from '@/components/PageContainer';
import { PageHeader } from '@/components/PageHeader';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { LayoutDashboard, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

export function Dashboard() {
  const navigate = useNavigate();

  return (
    <PageContainer>
      <PageHeader
        title="Dashboard"
        description="Overview of your ORCA environment and CyberArk instances"
      />
      
      {/* Empty State - Left Aligned */}
      <div className="flex justify-start">
        <Card className="border-gray-200 max-w-md w-full">
          <CardContent className="p-8 text-center">
            <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
              <LayoutDashboard className="h-8 w-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Dashboard Under Construction</h3>
            <p className="text-sm text-gray-600 mb-6">
              We're building a comprehensive dashboard to give you insights into your CyberArk environment. 
              In the meantime, you can explore other areas of the application.
            </p>
            <div className="space-y-3">
              <Button 
                onClick={() => navigate('/instances')}
                className="w-full"
              >
                Configure CyberArk Instances
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
              <Button 
                variant="outline"
                onClick={() => navigate('/operations')}
                className="w-full"
              >
                View Operations Queue
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  );
}