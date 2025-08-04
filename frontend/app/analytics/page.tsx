import { AnalyticsDashboard } from '../../components/analytics/AnalyticsDashboard';

export default function AnalyticsPage() {
  return (
    <div className="container mx-auto py-6">
      <AnalyticsDashboard />
    </div>
  );
}

export const metadata = {
  title: 'Analytics - NeuroSpend',
  description: 'Comprehensive analytics and insights for your expenses',
};
