import React from 'react';
import { Operation } from '../../types';
import { 
  CheckCircle, 
  XCircle, 
  Clock, 
  Pause, 
  HelpCircle, 
  Search, 
  Tag, 
  Clipboard, 
  Folder, 
  Filter, 
  SkipForward, 
  Calendar,
  PartyPopper,
  AlertTriangle
} from 'lucide-react';

interface OperationResultsPreviewProps {
  operation: Operation;
  onClick: () => void;
}

const OperationResultsPreview: React.FC<OperationResultsPreviewProps> = ({ operation, onClick }) => {
  const getStatusIcon = (status: Operation['status']) => {
    switch (status) {
      case 'completed': return <CheckCircle className="w-3 h-3" />;
      case 'running': return <Clock className="w-3 h-3" />;
      case 'failed': return <XCircle className="w-3 h-3" />;
      case 'pending': return <Pause className="w-3 h-3" />;
      default: return <HelpCircle className="w-3 h-3" />;
    }
  };

  const getStatusColor = (status: Operation['status']) => {
    switch (status) {
      case 'completed': return 'text-green-600 bg-green-100 border-green-200';
      case 'running': return 'text-blue-600 bg-blue-100 border-blue-200';
      case 'failed': return 'text-red-600 bg-red-100 border-red-200';
      case 'pending': return 'text-yellow-600 bg-yellow-100 border-yellow-200';
      default: return 'text-gray-600 bg-gray-100 border-gray-200';
    }
  };

  const getTypeIcon = (type: Operation['type']) => {
    switch (type) {
      case 'validation': return <Search className="w-5 h-5" />;
      case 'tagging': return <Tag className="w-5 h-5" />;
      default: return <Clipboard className="w-5 h-5" />;
    }
  };

  return (
    <div 
      className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer"
      onClick={onClick}
    >
      <div className="flex justify-between items-start">
        <div className="flex-1">
          <div className="flex items-center space-x-3 mb-2">
            <div className="text-blue-600">{getTypeIcon(operation.type)}</div>
            <div>
              <span className="text-sm font-medium text-gray-900">
                {operation.type.charAt(0).toUpperCase() + operation.type.slice(1)} Operation
              </span>
              <div className="flex items-center space-x-2 mt-1">
                <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getStatusColor(operation.status)} flex items-center space-x-1`}>
                  {getStatusIcon(operation.status)} 
                  <span>{operation.status.charAt(0).toUpperCase() + operation.status.slice(1)}</span>
                </span>
                <span className="text-xs text-gray-500">
                  #{operation.id}
                </span>
              </div>
            </div>
          </div>

          <div className="space-y-1 text-sm text-gray-600">
            <div className="flex items-center space-x-1">
              <Folder className="w-3 h-3 text-gray-500" />
              <span className="font-medium">Path:</span>
              <span className="font-mono text-xs bg-gray-100 px-1 rounded">{operation.directory_path}</span>
            </div>
            
            {operation.filter_pattern && (
              <div className="flex items-center space-x-1">
                <Filter className="w-3 h-3 text-gray-500" />
                <span className="font-medium">Filter:</span>
                <span className="font-mono text-xs bg-blue-100 px-1 rounded text-blue-800">{operation.filter_pattern}</span>
              </div>
            )}
            
            {operation.skip_pattern && (
              <div className="flex items-center space-x-1">
                <SkipForward className="w-3 h-3 text-gray-500" />
                <span className="font-medium">Skip:</span>
                <span className="font-mono text-xs bg-yellow-100 px-1 rounded text-yellow-800">{operation.skip_pattern}</span>
              </div>
            )}
            
            <div className="flex items-center space-x-1 text-xs text-gray-500 pt-1">
              <Calendar className="w-3 h-3" />
              <span>Created: {new Date(operation.created_at).toLocaleString()}</span>
              {operation.completed_at && (
                <>
                  <span>•</span>
                  <span>Completed: {new Date(operation.completed_at).toLocaleString()}</span>
                </>
              )}
            </div>
          </div>
        </div>

        <div className="ml-4 flex flex-col items-end space-y-2">
          <button className="px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors">
            View Details
          </button>
          
          {operation.status === 'completed' && operation.type === 'validation' && (
            <div className="text-xs text-gray-500 text-right">
              <div>Click to see</div>
              <div>validation results</div>
            </div>
          )}
        </div>
      </div>

      {/* Quick status indicators */}
      {(operation.status === 'completed' || operation.status === 'failed') && (
        <div className="mt-3 pt-3 border-t border-gray-200">
          <div className="flex justify-between items-center text-xs">
            <div className="flex items-center space-x-1 text-gray-500">
              {operation.status === 'completed' ? (
                <>
                  <PartyPopper className="w-3 h-3" />
                  <span>Ready to view results</span>
                </>
              ) : (
                <>
                  <AlertTriangle className="w-3 h-3" />
                  <span>Check logs for details</span>
                </>
              )}
            </div>
            <span className="text-blue-600 hover:text-blue-800 font-medium">
              View Summary →
            </span>
          </div>
        </div>
      )}
    </div>
  );
};

export default OperationResultsPreview;