import React, { useState, useEffect } from 'react';
import Card from '../components/common/Card';
import StandardEditor from '../components/standards/StandardEditor';
import OperationResultsPreview from '../components/operations/OperationResultsPreview';
import FileExplorer from '../components/common/FileExplorer';
import { useOperationsStore, useTagStandardsStore } from '../store';
import { Operation, CreateOperationRequest, TagStandard } from '../types';
import { detectEnvironment, validatePath, getPathSuggestions, EnvironmentInfo, PathValidation } from '../utils/environment';
import { 
  Container, 
  Monitor, 
  Clipboard, 
  Plus, 
  Edit, 
  RefreshCw, 
  CheckCircle, 
  AlertTriangle, 
  FileText, 
  Rocket 
} from 'lucide-react';

const OperationsPage: React.FC = () => {
  const { operations, createOperation, executeOperation, fetchOperations, loading, error } = useOperationsStore();
  const { standards, fetchStandards } = useTagStandardsStore();
  
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [showStandardEditor, setShowStandardEditor] = useState(false);
  const [selectedStandardForEdit, setSelectedStandardForEdit] = useState<TagStandard | null>(null);
  const [showFileExplorer, setShowFileExplorer] = useState(false);
  const [formData, setFormData] = useState<CreateOperationRequest>({
    type: 'validation',
    directory_path: '',
    filter_pattern: '',
    skip_pattern: '',
    settings: ''
  });
  const [environmentInfo, setEnvironmentInfo] = useState<EnvironmentInfo | null>(null);
  const [pathValidation, setPathValidation] = useState<PathValidation>({ valid: true, message: '', severity: 'info' });

  useEffect(() => {
    fetchOperations();
    fetchStandards();
    // Detect environment type
    initializeEnvironment();
  }, []);

  const initializeEnvironment = async () => {
    try {
      const envInfo = await detectEnvironment();
      setEnvironmentInfo(envInfo);
    } catch (error) {
      console.error('Failed to detect environment:', error);
      // Fallback to native environment
      setEnvironmentInfo({
        type: 'native',
        pathPrefix: '',
        pathSeparator: '/',
        examples: { absolute: '/home/user/terraform', relative: './terraform' },
        validation: { requiresAbsolute: false }
      });
    }
  };

  const handlePathValidation = (path: string) => {
    if (!environmentInfo) return;
    
    const validation = validatePath(path, environmentInfo);
    setPathValidation(validation);
  };

  const handleInputChange = (field: keyof CreateOperationRequest, value: string | number) => {
    const newFormData = { ...formData, [field]: value };
    setFormData(newFormData);
    
    if (field === 'directory_path') {
      handlePathValidation(value as string);
    }
  };

  const handlePathSelection = (selectedPath: string) => {
    handleInputChange('directory_path', selectedPath);
    setShowFileExplorer(false);
  };

  const handleCreateOperation = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!pathValidation.valid) {
      return;
    }
    
    try {
      const operation = await createOperation(formData);
      setShowCreateForm(false);
      setFormData({
        type: 'validation',
        directory_path: '',
        filter_pattern: '',
        skip_pattern: '',
        settings: ''
      });
      // Auto-execute the operation
      await executeOperation(operation.id);
    } catch (err) {
      console.error('Failed to create operation:', err);
    }
  };


  if (!environmentInfo) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-300 rounded w-1/4 mb-4"></div>
          <div className="h-32 bg-gray-300 rounded"></div>
        </div>
      </div>
    );
  }

  const getEnvironmentDisplayInfo = () => {
    if (environmentInfo.type === 'docker') {
      return {
        title: 'Docker Environment Detected',
        icon: Container,
        description: 'Your Terraform files should be mounted to /workspace in the container',
        pathExample: environmentInfo.examples.absolute,
        pathNote: 'Use absolute paths starting with /workspace'
      };
    } else {
      return {
        title: 'Native Environment',
        icon: Monitor,
        description: 'Running directly on your local machine',
        pathExample: `${environmentInfo.examples.relative} or ${environmentInfo.examples.absolute}`,
        pathNote: 'Use relative or absolute paths to your Terraform files'
      };
    }
  };

  const envDisplayInfo = getEnvironmentDisplayInfo();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Operations</h1>
        <p className="mt-2 text-gray-600">
          Run validation and tagging operations on your Terraform files.
        </p>
      </div>

      {/* Environment Info */}
      <Card>
        <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="flex items-center space-x-2">
            <envDisplayInfo.icon className="w-4 h-4 text-blue-700" />
            <h3 className="text-sm font-medium text-blue-900">{envDisplayInfo.title}</h3>
          </div>
          <p className="text-sm text-blue-700 mt-1">{envDisplayInfo.description}</p>
          <p className="text-xs text-blue-600 mt-2">
            <strong>Path example:</strong> {envDisplayInfo.pathExample}<br/>
            <strong>Note:</strong> {envDisplayInfo.pathNote}
          </p>
        </div>
      </Card>

      {/* Create Operation Form */}
      {showCreateForm ? (
        <Card>
          <div className="p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Create New Operation</h3>
            <form onSubmit={handleCreateOperation} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Operation Type
                  </label>
                  <select
                    value={formData.type}
                    onChange={(e) => handleInputChange('type', e.target.value as 'validation' | 'tagging')}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="validation">Validation</option>
                    <option value="tagging">Tagging</option>
                  </select>
                </div>
                
                {(formData.type === 'validation' || formData.type === 'tagging') && (
                  <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <Clipboard className="w-4 h-4 text-blue-700" />
                        <label className="block text-sm font-medium text-blue-900">
                          Tag Standard {formData.type === 'validation' ? '(Required for validation)' : '(Optional for tagging)'}
                        </label>
                      </div>
                      <div className="flex space-x-2">
                        <button
                          type="button"
                          onClick={() => {
                            setSelectedStandardForEdit(null);
                            setShowStandardEditor(true);
                          }}
                          className="px-3 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700 transition-colors flex items-center space-x-1"
                        >
                          <Plus className="w-3 h-3" />
                          <span>Create New</span>
                        </button>
                        {formData.standard_id && (
                          <button
                            type="button"
                            onClick={() => {
                              const standard = standards.find(s => s.id === formData.standard_id);
                              if (standard) {
                                setSelectedStandardForEdit(standard);
                                setShowStandardEditor(true);
                              }
                            }}
                            className="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors flex items-center space-x-1"
                          >
                            <Edit className="w-3 h-3" />
                            <span>Edit</span>
                          </button>
                        )}
                        <button
                          type="button"
                          onClick={() => fetchStandards()}
                          className="px-3 py-1 text-xs bg-gray-600 text-white rounded hover:bg-gray-700 transition-colors flex items-center space-x-1"
                        >
                          <RefreshCw className="w-3 h-3" />
                          <span>Refresh</span>
                        </button>
                      </div>
                    </div>
                    
                    <select
                      value={formData.standard_id || ''}
                      onChange={(e) => {
                        const value = e.target.value;
                        if (value) {
                          handleInputChange('standard_id', parseInt(value));
                        } else {
                          setFormData({ ...formData, standard_id: undefined });
                        }
                      }}
                      className="w-full px-3 py-2 border border-blue-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
                      required={formData.type === 'validation'}
                    >
                      <option value="">
                        {formData.type === 'validation' ? 'Select a standard (required)' : 'Select a standard (optional)'}
                      </option>
                      {standards.length === 0 && (
                        <option value="" disabled>No standards available - create one first</option>
                      )}
                      {standards.map((standard: TagStandard) => (
                        <option key={standard.id} value={standard.id}>
                          {standard.name} ({standard.cloud_provider.toUpperCase()}) 
                          {standard.description && ` - ${standard.description.substring(0, 50)}${standard.description.length > 50 ? '...' : ''}`}
                        </option>
                      ))}
                    </select>
                    
                    {formData.standard_id && (
                      <div className="mt-3 p-3 bg-white rounded border border-blue-200">
                        <div className="flex items-center space-x-2 text-sm text-blue-700 font-medium mb-1">
                          <CheckCircle className="w-4 h-4" />
                          <span>Selected Standard: {standards.find(s => s.id === formData.standard_id)?.name}</span>
                        </div>
                        <div className="text-xs text-gray-600">
                          Provider: {standards.find(s => s.id === formData.standard_id)?.cloud_provider.toUpperCase()} | 
                          Version: {standards.find(s => s.id === formData.standard_id)?.version} | 
                          Description: {standards.find(s => s.id === formData.standard_id)?.description || 'No description'}
                        </div>
                      </div>
                    )}
                    
                    {!formData.standard_id && formData.type === 'validation' && (
                      <div className="mt-2 text-sm text-red-600 bg-red-50 p-2 rounded border border-red-200 flex items-start space-x-2">
                        <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                        <span>A tag standard is required for validation operations.</span> 
                        <button
                          type="button"
                          onClick={() => {
                            setSelectedStandardForEdit(null);
                            setShowStandardEditor(true);
                          }}
                          className="ml-1 text-red-700 underline hover:text-red-800"
                        >
                          Create one now
                        </button>
                      </div>
                    )}
                    
                    {standards.length === 0 && (
                      <div className="mt-2 text-sm text-yellow-700 bg-yellow-50 p-2 rounded border border-yellow-200 flex items-start space-x-2">
                        <FileText className="w-4 h-4 flex-shrink-0 mt-0.5" />
                        <span>No tag standards found.</span> 
                        <button
                          type="button"
                          onClick={() => {
                            setSelectedStandardForEdit(null);
                            setShowStandardEditor(true);
                          }}
                          className="ml-1 text-yellow-800 underline hover:text-yellow-900"
                        >
                          Create your first standard
                        </button> to get started.
                      </div>
                    )}
                  </div>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Directory Path *
                </label>
                <div className="relative">
                  <input
                    type="text"
                    value={formData.directory_path}
                    onChange={(e) => handleInputChange('directory_path', e.target.value)}
                    placeholder={envDisplayInfo.pathExample}
                    className={`w-full px-3 py-2 pr-20 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                      pathValidation.valid ? 'border-gray-300' : 'border-red-300'
                    }`}
                    required
                    list="path-suggestions"
                  />
                  <button
                    type="button"
                    onClick={() => setShowFileExplorer(true)}
                    className="absolute right-2 top-1/2 transform -translate-y-1/2 px-2 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 flex items-center"
                  >
                    Browse
                  </button>
                  <datalist id="path-suggestions">
                    {getPathSuggestions(environmentInfo).map((suggestion, index) => (
                      <option key={index} value={suggestion} />
                    ))}
                  </datalist>
                </div>
                {pathValidation.message && (
                  <p className={`text-xs mt-1 ${
                    pathValidation.severity === 'error' ? 'text-red-600' : 
                    pathValidation.severity === 'warning' ? 'text-yellow-600' : 'text-blue-600'
                  }`}>
                    {pathValidation.message}
                  </p>
                )}
                <div className="mt-1 text-xs text-gray-500">
                  Common paths: {getPathSuggestions(environmentInfo).slice(0, 3).join(', ')}
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Filter Pattern (Optional)
                  </label>
                  <input
                    type="text"
                    value={formData.filter_pattern}
                    onChange={(e) => handleInputChange('filter_pattern', e.target.value)}
                    placeholder="*.tf or aws_*"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Include only files/resources matching this pattern
                  </p>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Skip Pattern (Optional)
                  </label>
                  <input
                    type="text"
                    value={formData.skip_pattern}
                    onChange={(e) => handleInputChange('skip_pattern', e.target.value)}
                    placeholder="test_* or *_backup.tf"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Exclude files/resources matching this pattern
                  </p>
                </div>
              </div>

              <div className="flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => setShowCreateForm(false)}
                  className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={!pathValidation.valid || loading}
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? 'Creating...' : 'Create & Execute'}
                </button>
              </div>
            </form>
          </div>
        </Card>
      ) : (
        <Card>
          <div className="p-6">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-medium text-gray-900">Operations</h3>
              <button
                onClick={() => setShowCreateForm(true)}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                New Operation
              </button>
            </div>
          </div>
        </Card>
      )}

      {/* Operations List */}
      <Card>
        <div className="p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Recent Operations</h3>
          
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}
          
          {operations.length === 0 ? (
            <div className="text-center py-12">
              <Clipboard className="w-16 h-16 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">No Operations Yet</h3>
              <p className="text-gray-500 mb-4">Create your first operation to validate or tag your Terraform resources.</p>
              <button
                onClick={() => setShowCreateForm(true)}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 flex items-center space-x-2"
              >
                <Rocket className="w-4 h-4" />
                <span>Create First Operation</span>
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {operations
                .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
                .map((operation: Operation) => (
                  <OperationResultsPreview
                    key={operation.id}
                    operation={operation}
                    onClick={() => window.location.href = `/operations/${operation.id}/summary`}
                  />
                ))}
            </div>
          )}
        </div>
      </Card>
      
      {/* Standard Editor Modal */}
      <StandardEditor
        isOpen={showStandardEditor}
        onClose={() => {
          setShowStandardEditor(false);
          setSelectedStandardForEdit(null);
          // Refresh standards after creating/editing
          fetchStandards();
        }}
        standard={selectedStandardForEdit}
      />

      {/* File Explorer Modal */}
      <FileExplorer
        isOpen={showFileExplorer}
        onClose={() => setShowFileExplorer(false)}
        onSelectPath={handlePathSelection}
        initialPath={formData.directory_path || '/workspace'}
        title="Select Terraform Directory"
        showFiles={false}
      />
    </div>
  );
};

export default OperationsPage;