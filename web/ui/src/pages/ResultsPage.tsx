import React from 'react';
import Card from '../components/common/Card';

const ResultsPage: React.FC = () => {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Results</h1>
        <p className="mt-2 text-gray-600">
          View results and analytics from your tag validation and application operations.
        </p>
      </div>

      <Card>
        <div className="text-center py-12">
          <h3 className="text-lg font-medium text-gray-900 mb-2">Results Dashboard Coming Soon</h3>
          <p className="text-gray-500">
            This section will display detailed results, statistics, and insights from your operations.
          </p>
        </div>
      </Card>
    </div>
  );
};

export default ResultsPage;