import { PageContainer } from '../components/PageContainer';
import { PageHeader } from '../components/PageHeader';
import { Card, CardContent } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Activity, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

export default function PipelineDashboard() {
  const navigate = useNavigate();

  return (
    <PageContainer>
      <PageHeader
        title="Queue Monitoring"
        description="Monitor and manage operation processing queues"
      />
      
      {/* Empty State - Left Aligned */}
      <div className="flex justify-start">
        <Card className="border-gray-200 max-w-md w-full">
          <CardContent className="p-8 text-center">
            <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
              <Activity className="h-8 w-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Queue Monitoring Under Construction</h3>
            <p className="text-sm text-gray-600 mb-6">
              We're building a comprehensive queue monitoring dashboard to give you real-time insights into your operations pipeline. 
              In the meantime, you can view the operations queue.
            </p>
            <div className="space-y-3">
              <Button 
                onClick={() => navigate('/operations')}
                className="w-full"
              >
                View Operations Queue
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
              <Button 
                variant="outline"
                onClick={() => navigate('/instances')}
                className="w-full"
              >
                Configure Instances
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  );
}