import React, { useEffect, useState } from 'react';
import { Plus, Edit, Trash2, Eye, Download, FileText, Cloud, Zap, Globe, Settings } from 'lucide-react';
import { TagStandard, CLOUD_PROVIDERS } from '../../types';
import { tagStandardsApi } from '../../api/client';
import { useStore } from '../../store';
import Button from '../common/Button';
import Card from '../common/Card';
import LoadingSpinner from '../common/LoadingSpinner';

interface StandardsListProps {
  onCreateNew: () => void;
  onEdit: (standard: TagStandard) => void;
  onView: (standard: TagStandard) => void;
}

const StandardsList: React.FC<StandardsListProps> = ({ 
  onCreateNew, 
  onEdit, 
  onView 
}) => {
  const { tagStandards, setTagStandards, removeTagStandard, loading, setLoading, setError } = useStore();
  const [filter, setFilter] = useState<string>('all');

  useEffect(() => {
    loadStandards();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filter]);

  const loadStandards = async () => {
    try {
      setLoading(true);
      const provider = filter === 'all' ? undefined : filter;
      const standards = await tagStandardsApi.list(provider);
      setTagStandards(standards);
    } catch (error) {
      setError('Failed to load tag standards');
      console.error('Error loading standards:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this tag standard?')) {
      return;
    }

    try {
      await tagStandardsApi.delete(id);
      removeTagStandard(id);
    } catch (error) {
      setError('Failed to delete tag standard');
      console.error('Error deleting standard:', error);
    }
  };

  const handleDownload = (standard: TagStandard) => {
    const blob = new Blob([standard.content], { type: 'application/yaml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${standard.name}.yaml`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const filteredStandards = tagStandards.filter(standard => 
    filter === 'all' || standard.cloud_provider === filter
  );

  const getProviderIcon = (providerId: string) => {
    switch (providerId) {
      case 'aws': return <Cloud className="w-6 h-6 text-orange-500" />;
      case 'azure': return <Zap className="w-6 h-6 text-blue-500" />;
      case 'gcp': return <Globe className="w-6 h-6 text-green-500" />;
      case 'generic': return <Settings className="w-6 h-6 text-gray-500" />;
      default: return <Settings className="w-6 h-6 text-gray-500" />;
    }
  };

  const getProviderName = (providerId: string) => {
    const provider = CLOUD_PROVIDERS.find(p => p.id === providerId);
    return provider?.name || providerId;
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Tag Standards</h1>
          <p className="mt-2 text-gray-600">
            Manage and configure tag standardization rules for your cloud resources.
          </p>
        </div>
        <Button icon={Plus} onClick={onCreateNew}>
          Create Standard
        </Button>
      </div>

      {/* Filters */}
      <div className="flex items-center space-x-4">
        <label className="text-sm font-medium text-gray-700">Filter by provider:</label>
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="block w-48 rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
        >
          <option value="all">All Providers</option>
          {CLOUD_PROVIDERS.map((provider) => (
            <option key={provider.id} value={provider.id}>
              {provider.name}
            </option>
          ))}
        </select>
      </div>

      {/* Standards Grid */}
      {filteredStandards.length === 0 ? (
        <Card>
          <div className="text-center py-12">
            <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No standards found</h3>
            <p className="text-gray-500 mb-6">
              {filter === 'all' 
                ? "Get started by creating your first tag standard."
                : `No standards found for ${getProviderName(filter)}.`
              }
            </p>
            <Button icon={Plus} onClick={onCreateNew}>
              Create Standard
            </Button>
          </div>
        </Card>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredStandards.map((standard) => (
            <Card key={standard.id} className="hover:shadow-lg transition-shadow">
              <div className="space-y-4">
                <div className="flex items-start justify-between">
                  <div className="flex items-center space-x-3">
                    <div className="flex-shrink-0">{getProviderIcon(standard.cloud_provider)}</div>
                    <div>
                      <h3 className="text-lg font-medium text-gray-900">{standard.name}</h3>
                      <p className="text-sm text-gray-500">{getProviderName(standard.cloud_provider)}</p>
                    </div>
                  </div>
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 text-primary-800">
                    v{standard.version}
                  </span>
                </div>

                {standard.description && (
                  <p className="text-sm text-gray-600 line-clamp-2">{standard.description}</p>
                )}

                <div className="text-xs text-gray-500">
                  <p>Created: {new Date(standard.created_at).toLocaleDateString()}</p>
                  <p>Updated: {new Date(standard.updated_at).toLocaleDateString()}</p>
                </div>

                <div className="flex items-center justify-between pt-4 border-t border-gray-200">
                  <div className="flex items-center space-x-2">
                    <Button 
                      variant="outline" 
                      size="sm" 
                      icon={Eye}
                      onClick={() => onView(standard)}
                    >
                      View
                    </Button>
                    <Button 
                      variant="outline" 
                      size="sm" 
                      icon={Download}
                      onClick={() => handleDownload(standard)}
                    >
                      Download
                    </Button>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Button 
                      variant="secondary" 
                      size="sm" 
                      icon={Edit}
                      onClick={() => onEdit(standard)}
                    >
                      Edit
                    </Button>
                    <Button 
                      variant="danger" 
                      size="sm" 
                      icon={Trash2}
                      onClick={() => handleDelete(standard.id)}
                    >
                      Delete
                    </Button>
                  </div>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};

export default StandardsList;