import React, { useState, useEffect } from 'react';
import { Save, Plus, Trash2, CheckCircle, XCircle, Edit3, FileText, Lightbulb, AlertTriangle } from 'lucide-react';
import { TagStandard, CreateTagStandardRequest, CLOUD_PROVIDERS, TagRequirement } from '../../types';
import { tagStandardsApi } from '../../api/client';
import { useStore } from '../../store';
import Button from '../common/Button';
import Modal from '../common/Modal';
import * as yaml from 'js-yaml';

interface StandardEditorProps {
  isOpen: boolean;
  onClose: () => void;
  standard?: TagStandard | null;
}

const StandardEditor: React.FC<StandardEditorProps> = ({ 
  isOpen, 
  onClose, 
  standard 
}) => {
  const { addTagStandard, updateTagStandard, setError } = useStore();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<CreateTagStandardRequest>({
    name: '',
    description: '',
    cloud_provider: 'aws',
    version: 1,
    content: '',
  });

  const [yamlMode, setYamlMode] = useState(false);
  const [requiredTags, setRequiredTags] = useState<TagRequirement[]>([]);
  const [optionalTags, setOptionalTags] = useState<TagRequirement[]>([]);
  const [validationStatus, setValidationStatus] = useState<'idle' | 'validating' | 'valid' | 'invalid'>('idle');
  const [validationError, setValidationError] = useState<string>('');
  const [hasComplexContent, setHasComplexContent] = useState(false);

  useEffect(() => {
    if (standard) {
      setFormData({
        name: standard.name,
        description: standard.description,
        cloud_provider: standard.cloud_provider,
        version: standard.version,
        content: standard.content,
      });
      
      // Parse YAML content to extract tags
      if (standard.content) {
        const parsed = parseYamlContent(standard.content);
        setRequiredTags(parsed.required_tags);
        setOptionalTags(parsed.optional_tags);
        setHasComplexContent(parsed.hasComplexContent);
        
        // Use YAML mode if we have complex content or no tags
        if (parsed.hasComplexContent || (parsed.required_tags.length === 0 && parsed.optional_tags.length === 0)) {
          setYamlMode(true);
        } else {
          setYamlMode(false);
        }
      }
    } else {
      resetForm();
    }
  }, [standard, isOpen]);

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      cloud_provider: 'aws',
      version: 1,
      content: '',
    });
    setRequiredTags([]);
    setOptionalTags([]);
    setYamlMode(false);
    setHasComplexContent(false);
  };

  const parseYamlContent = (content: string) => {
    try {
      const parsed = yaml.load(content) as any;
      if (!parsed) {
        return { required_tags: [], optional_tags: [], hasComplexContent: false };
      }
      
      // Check if the YAML has complex content that can't be represented in form
      const complexFields = ['global_excludes', 'resource_rules', 'case_sensitive', 'extends', 'mixins'];
      const hasComplex = complexFields.some(field => parsed[field] !== undefined) ||
        (parsed.validation_rules && Object.keys(parsed.validation_rules).length > 3);
      
      // Convert parsed YAML to our TagRequirement format
      const convertTags = (tags: any[]): TagRequirement[] => {
        if (!Array.isArray(tags)) return [];
        
        return tags.map(tag => ({
          key: tag.key || '',
          description: tag.description || '',
          required: true,
          data_type: tag.data_type || 'string',
          default_value: tag.default_value,
          allowed_values: tag.allowed_values,
          pattern: tag.format || tag.pattern,
          examples: tag.examples,
        }));
      };
      
      return {
        required_tags: convertTags(parsed.required_tags || []),
        optional_tags: convertTags(parsed.optional_tags || []).map(tag => ({ ...tag, required: false })),
        hasComplexContent: hasComplex,
      };
    } catch (error) {
      console.error('Error parsing YAML content:', error);
      // Return empty structure on parse error
      return { required_tags: [], optional_tags: [], hasComplexContent: false };
    }
  };

  const generateYamlContent = () => {
    // If we're editing an existing standard, try to preserve its structure
    if (standard && standard.content) {
      try {
        // Parse the existing content to preserve any extra fields
        const existingParsed = yaml.load(standard.content) as any;
        
        // Update with our form values
        const updatedObject = {
          ...existingParsed, // Preserve any existing fields
          version: formData.version || existingParsed.version || 1,
          metadata: {
            ...(existingParsed.metadata || {}),
            description: formData.description || existingParsed.metadata?.description || 'Updated via Terratag UI',
            author: existingParsed.metadata?.author || 'Terratag UI',
            updated_date: new Date().toISOString().split('T')[0],
          },
          cloud_provider: formData.cloud_provider,
          required_tags: requiredTags.filter(tag => tag.key).map(tag => {
            const spec: any = {
              key: tag.key,
              description: tag.description || '',
              data_type: tag.data_type || 'string',
            };
            
            if (tag.default_value) spec.default_value = tag.default_value;
            if (tag.pattern) spec.format = tag.pattern;
            if (tag.allowed_values && tag.allowed_values.length > 0) {
              spec.allowed_values = tag.allowed_values;
            }
            if (tag.examples && tag.examples.length > 0) {
              spec.examples = tag.examples;
            }
            
            return spec;
          }),
          optional_tags: optionalTags.filter(tag => tag.key).map(tag => {
            const spec: any = {
              key: tag.key,
              description: tag.description || '',
              data_type: tag.data_type || 'string',
            };
            
            if (tag.default_value) spec.default_value = tag.default_value;
            if (tag.pattern) spec.format = tag.pattern;
            if (tag.allowed_values && tag.allowed_values.length > 0) {
              spec.allowed_values = tag.allowed_values;
            }
            if (tag.examples && tag.examples.length > 0) {
              spec.examples = tag.examples;
            }
            
            return spec;
          }),
          // Preserve validation_rules if they exist, otherwise use defaults
          validation_rules: existingParsed.validation_rules || {
            case_sensitive_keys: false,
            allow_extra_tags: true,
            strict_mode: false,
          },
        };
        
        return yaml.dump(updatedObject, {
          indent: 2,
          lineWidth: 120,
          noRefs: true,
          sortKeys: false,
        });
      } catch (error) {
        console.error('Error preserving existing YAML structure:', error);
        // Fall back to generating new structure
      }
    }
    
    // Generate new structure for new standards or if parsing failed
    const standardObject = {
      version: formData.version || 1,
      metadata: {
        description: formData.description || 'Generated by Terratag UI',
        author: 'Terratag UI',
        created_date: new Date().toISOString().split('T')[0],
      },
      cloud_provider: formData.cloud_provider,
      required_tags: requiredTags.filter(tag => tag.key).map(tag => {
        const spec: any = {
          key: tag.key,
          description: tag.description || '',
          data_type: tag.data_type || 'string',
        };
        
        if (tag.default_value) spec.default_value = tag.default_value;
        if (tag.pattern) spec.format = tag.pattern;
        if (tag.allowed_values && tag.allowed_values.length > 0) {
          spec.allowed_values = tag.allowed_values;
        }
        if (tag.examples && tag.examples.length > 0) {
          spec.examples = tag.examples;
        }
        
        return spec;
      }),
      optional_tags: optionalTags.filter(tag => tag.key).map(tag => {
        const spec: any = {
          key: tag.key,
          description: tag.description || '',
          data_type: tag.data_type || 'string',
        };
        
        if (tag.default_value) spec.default_value = tag.default_value;
        if (tag.pattern) spec.format = tag.pattern;
        if (tag.allowed_values && tag.allowed_values.length > 0) {
          spec.allowed_values = tag.allowed_values;
        }
        if (tag.examples && tag.examples.length > 0) {
          spec.examples = tag.examples;
        }
        
        return spec;
      }),
      validation_rules: {
        case_sensitive_keys: false,
        allow_extra_tags: true,
        strict_mode: false,
      },
    };
    
    try {
      return yaml.dump(standardObject, {
        indent: 2,
        lineWidth: 120,
        noRefs: true,
        sortKeys: false,
      });
    } catch (error) {
      console.error('Error generating YAML:', error);
      return '# Error generating YAML content';
    }
  };

  // Real-time validation function
  const validateContent = async (content: string) => {
    if (!content.trim()) {
      setValidationStatus('idle');
      setValidationError('');
      return;
    }

    setValidationStatus('validating');
    setValidationError('');

    try {
      await tagStandardsApi.validateContent(content, formData.cloud_provider);
      setValidationStatus('valid');
      setValidationError('');
    } catch (error: any) {
      setValidationStatus('invalid');
      setValidationError(error.response?.data?.message || error.message || 'Invalid YAML content');
    }
  };

  // Debounced validation for YAML mode
  const debouncedValidate = React.useCallback(
    (() => {
      let timeoutId: NodeJS.Timeout;
      return (content: string) => {
        clearTimeout(timeoutId);
        timeoutId = setTimeout(() => validateContent(content), 1000);
      };
    })(),
    [formData.cloud_provider]
  );

  // Effect to validate content when it changes in YAML mode
  useEffect(() => {
    if (yamlMode && formData.content) {
      debouncedValidate(formData.content);
    }
  }, [formData.content, yamlMode, debouncedValidate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const content = yamlMode ? formData.content : generateYamlContent();
      const dataToSubmit = { ...formData, content };

      if (standard) {
        const updated = await tagStandardsApi.update(standard.id, dataToSubmit);
        updateTagStandard(standard.id, updated);
      } else {
        const created = await tagStandardsApi.create(dataToSubmit);
        addTagStandard(created);
      }

      onClose();
    } catch (error) {
      setError(`Failed to ${standard ? 'update' : 'create'} tag standard`);
      console.error('Error saving standard:', error);
    } finally {
      setLoading(false);
    }
  };

  const addTag = (type: 'required' | 'optional') => {
    const newTag: TagRequirement = {
      key: '',
      description: '',
      required: type === 'required',
      data_type: 'string',
    };

    if (type === 'required') {
      setRequiredTags([...requiredTags, newTag]);
    } else {
      setOptionalTags([...optionalTags, newTag]);
    }
  };

  const removeTag = (type: 'required' | 'optional', index: number) => {
    if (type === 'required') {
      setRequiredTags(requiredTags.filter((_, i) => i !== index));
    } else {
      setOptionalTags(optionalTags.filter((_, i) => i !== index));
    }
  };

  const updateTag = (type: 'required' | 'optional', index: number, field: keyof TagRequirement, value: any) => {
    if (type === 'required') {
      const updated = [...requiredTags];
      (updated[index] as any)[field] = value;
      setRequiredTags(updated);
    } else {
      const updated = [...optionalTags];
      (updated[index] as any)[field] = value;
      setOptionalTags(updated);
    }
  };

  const renderTagEditor = (tags: TagRequirement[], type: 'required' | 'optional') => (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-gray-900 capitalize">{type} Tags</h4>
        <Button 
          type="button"
          variant="outline" 
          size="sm" 
          icon={Plus}
          onClick={() => addTag(type)}
        >
          Add Tag
        </Button>
      </div>
      
      {tags.map((tag, index) => (
        <div key={index} className="border border-gray-200 rounded-lg p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h5 className="text-sm font-medium text-gray-700">Tag {index + 1}</h5>
            <Button 
              type="button"
              variant="danger" 
              size="sm" 
              icon={Trash2}
              onClick={() => removeTag(type, index)}
            >
              Remove
            </Button>
          </div>
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Key</label>
              <input
                type="text"
                value={tag.key}
                onChange={(e) => updateTag(type, index, 'key', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                placeholder="e.g., Environment"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Data Type</label>
              <select
                value={tag.data_type}
                onChange={(e) => updateTag(type, index, 'data_type', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              >
                <option value="string">String</option>
                <option value="number">Number</option>
                <option value="boolean">Boolean</option>
              </select>
            </div>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700">Description</label>
            <input
              type="text"
              value={tag.description}
              onChange={(e) => updateTag(type, index, 'description', e.target.value)}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              placeholder="Describe the purpose of this tag"
            />
          </div>
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Default Value</label>
              <input
                type="text"
                value={tag.default_value || ''}
                onChange={(e) => updateTag(type, index, 'default_value', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                placeholder="Optional default value"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Validation Pattern (Regex)</label>
              <input
                type="text"
                value={tag.pattern || ''}
                onChange={(e) => updateTag(type, index, 'pattern', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                placeholder="e.g., ^[A-Z][a-z]+$"
              />
            </div>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700">Allowed Values (comma-separated)</label>
            <input
              type="text"
              value={tag.allowed_values?.join(', ') || ''}
              onChange={(e) => {
                const values = e.target.value
                  .split(',')
                  .map(v => v.trim())
                  .filter(v => v.length > 0);
                updateTag(type, index, 'allowed_values', values.length > 0 ? values : undefined);
              }}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              placeholder="e.g., Production, Staging, Development"
            />
            {tag.allowed_values && tag.allowed_values.length > 0 && (
              <p className="mt-1 text-xs text-gray-500">
                Values: {tag.allowed_values.map(v => `"${v}"`).join(', ')}
              </p>
            )}
          </div>
        </div>
      ))}
    </div>
  );

  return (
    <Modal 
      isOpen={isOpen} 
      onClose={onClose} 
      title={standard ? 'Edit Tag Standard' : 'Create Tag Standard'}
      size="xl"
    >
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Information */}
        <div className="grid grid-cols-2 gap-6">
          <div>
            <label className="block text-sm font-medium text-gray-700">Name</label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              placeholder="Production Environment Tags"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">Cloud Provider</label>
            <select
              value={formData.cloud_provider}
              onChange={(e) => setFormData({ ...formData, cloud_provider: e.target.value })}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
            >
              {CLOUD_PROVIDERS.map((provider) => (
                <option key={provider.id} value={provider.id}>
                  {provider.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Description</label>
          <textarea
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            rows={2}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
            placeholder="Describe the purpose and scope of this tag standard"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Version</label>
          <input
            type="number"
            min="1"
            value={formData.version}
            onChange={(e) => setFormData({ ...formData, version: parseInt(e.target.value) })}
            className="mt-1 block w-20 rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          />
        </div>

        {/* Content Editor Mode Toggle */}
        <div className="flex items-center space-x-4">
          <label className="block text-sm font-medium text-gray-700">Content Mode:</label>
          <div className="flex space-x-2">
            <Button
              type="button"
              variant={!yamlMode ? 'primary' : 'outline'}
              size="sm"
              onClick={() => {
                if (yamlMode && formData.content) {
                  // Switching from YAML to Form - parse YAML content
                  const parsed = parseYamlContent(formData.content);
                  setRequiredTags(parsed.required_tags);
                  setOptionalTags(parsed.optional_tags);
                  setHasComplexContent(parsed.hasComplexContent);
                }
                setYamlMode(false);
                setValidationStatus('idle');
                setValidationError('');
              }}
              icon={Edit3}
            >
              Form Editor
            </Button>
            <Button
              type="button"
              variant={yamlMode ? 'primary' : 'outline'}
              size="sm"
              onClick={() => {
                if (!yamlMode) {
                  // Switching from Form to YAML - generate YAML content
                  const generatedYaml = generateYamlContent();
                  setFormData({ ...formData, content: generatedYaml });
                }
                setYamlMode(true);
              }}
              icon={FileText}
            >
              YAML Editor
            </Button>
          </div>
        </div>

        {/* Content */}
        {yamlMode ? (
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="block text-sm font-medium text-gray-700">YAML Content</label>
              <div className="flex items-center space-x-2">
                {validationStatus === 'validating' && (
                  <div className="flex items-center text-yellow-600">
                    <div className="animate-spin h-4 w-4 border-2 border-yellow-600 border-t-transparent rounded-full mr-1"></div>
                    <span className="text-xs">Validating...</span>
                  </div>
                )}
                {validationStatus === 'valid' && (
                  <div className="flex items-center text-green-600">
                    <CheckCircle className="h-4 w-4 mr-1" />
                    <span className="text-xs">Valid YAML</span>
                  </div>
                )}
                {validationStatus === 'invalid' && (
                  <div className="flex items-center text-red-600">
                    <XCircle className="h-4 w-4 mr-1" />
                    <span className="text-xs">Invalid YAML</span>
                  </div>
                )}
              </div>
            </div>
            <div className="relative">
              <textarea
                value={formData.content}
                onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                rows={20}
                className={`mt-1 block w-full rounded-md shadow-sm sm:text-sm font-mono focus:ring-2 focus:ring-offset-2 ${
                  validationStatus === 'invalid' 
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500' 
                    : validationStatus === 'valid'
                    ? 'border-green-300 focus:border-green-500 focus:ring-green-500'
                    : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                }`}
                placeholder="Enter YAML content or switch to Form Editor to build it..."
              />
              {validationStatus === 'valid' && (
                <div className="absolute top-2 right-2">
                  <CheckCircle className="h-5 w-5 text-green-500" />
                </div>
              )}
              {validationStatus === 'invalid' && (
                <div className="absolute top-2 right-2">
                  <XCircle className="h-5 w-5 text-red-500" />
                </div>
              )}
            </div>
            {validationError && (
              <div className="mt-2 p-3 bg-red-50 border border-red-200 rounded-md">
                <div className="flex items-start">
                  <XCircle className="h-5 w-5 text-red-400 mt-0.5 mr-2 flex-shrink-0" />
                  <div>
                    <h4 className="text-sm font-medium text-red-800">Validation Error</h4>
                    <p className="mt-1 text-sm text-red-700 whitespace-pre-wrap">{validationError}</p>
                  </div>
                </div>
              </div>
            )}
            <div className="mt-2 text-xs text-gray-500 flex items-start space-x-1">
              <Lightbulb className="w-3 h-3 flex-shrink-0 mt-0.5" />
              <span>Tip: Switch to Form Editor to build the standard visually, then switch back to see the generated YAML.</span>
            </div>
          </div>
        ) : (
          <div className="space-y-8">
            {hasComplexContent && (
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <div className="flex items-start space-x-2">
                  <AlertTriangle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <h4 className="text-sm font-medium text-yellow-900">Complex YAML Content Detected</h4>
                    <p className="mt-1 text-sm text-yellow-700">
                      This standard contains advanced features (like global_excludes or resource_rules) that can't be fully edited in form mode. 
                      Switch to YAML Editor to modify these features, or continue with form editing to update basic tag requirements only.
                    </p>
                  </div>
                </div>
              </div>
            )}
            {renderTagEditor(requiredTags, 'required')}
            {renderTagEditor(optionalTags, 'optional')}
          </div>
        )}

        {/* Actions */}
        <div className="flex items-center justify-end space-x-4 pt-6 border-t border-gray-200">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" loading={loading} icon={Save}>
            {standard ? 'Update' : 'Create'} Standard
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default StandardEditor;