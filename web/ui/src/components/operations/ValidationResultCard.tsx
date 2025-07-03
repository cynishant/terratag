import React, { useState } from 'react';
import { OperationResult, ValidationDetails, TagViolation } from '../../types';
import { 
  CheckCircle, 
  XCircle, 
  AlertTriangle, 
  FileText, 
  Settings, 
  Tag, 
  ChevronDown, 
  ChevronRight,
  AlertCircle,
  Minus,
  Plus,
  Shield,
  Hash,
  Code
} from 'lucide-react';

interface ValidationResultCardProps {
  result: OperationResult;
}

const ValidationResultCard: React.FC<ValidationResultCardProps> = ({ result }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  // Parse validation details from JSON
  const getValidationDetails = (): ValidationDetails | null => {
    if (!result.details) return null;
    try {
      const parsed = JSON.parse(result.details);
      return parsed as ValidationDetails;
    } catch (error) {
      console.error('Failed to parse validation details:', error);
      return null;
    }
  };

  const validationDetails = getValidationDetails();

  const getActionIcon = (action: string) => {
    switch (action.toLowerCase()) {
      case 'compliant': return <CheckCircle className="w-5 h-5 text-green-600" />;
      case 'violation': return <XCircle className="w-5 h-5 text-red-600" />;
      case 'error': return <AlertTriangle className="w-5 h-5 text-yellow-600" />;
      case 'processed': return <FileText className="w-5 h-5 text-blue-600" />;
      default: return <FileText className="w-5 h-5 text-gray-600" />;
    }
  };

  const getActionColor = (action: string) => {
    switch (action.toLowerCase()) {
      case 'compliant': return 'text-green-700 bg-green-50 border-green-200';
      case 'violation': return 'text-red-700 bg-red-50 border-red-200';
      case 'error': return 'text-yellow-700 bg-yellow-50 border-yellow-200';
      case 'processed': return 'text-blue-700 bg-blue-50 border-blue-200';
      default: return 'text-gray-700 bg-gray-50 border-gray-200';
    }
  };

  const getViolationTypeIcon = (violationType: string) => {
    switch (violationType.toLowerCase()) {
      case 'missing_required': return <Minus className="w-4 h-4 text-red-600" />;
      case 'invalid_value': return <AlertTriangle className="w-4 h-4 text-orange-600" />;
      case 'invalid_format': return <FileText className="w-4 h-4 text-orange-600" />;
      case 'length_exceeded': return <AlertCircle className="w-4 h-4 text-orange-600" />;
      case 'length_too_short': return <AlertCircle className="w-4 h-4 text-orange-600" />;
      case 'invalid_data_type': return <Settings className="w-4 h-4 text-orange-600" />;
      default: return <AlertCircle className="w-4 h-4 text-orange-600" />;
    }
  };

  const formatViolationType = (violationType: string) => {
    return violationType
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  };



  return (
    <div className={`border rounded-lg transition-all duration-200 ${getActionColor(result.action)}`}>
      {/* Header */}
      <div 
        className="p-4 cursor-pointer hover:bg-opacity-80 transition-colors"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex justify-between items-start">
          <div className="flex-1">
            <div className="flex items-center space-x-3 mb-2">
              <div className="flex-shrink-0">{getActionIcon(result.action)}</div>
              <div className="flex-1">
                <div className="flex items-center space-x-2 mb-1">
                  {result.resource_type && result.resource_name ? (
                    <div className="flex items-center space-x-2">
                      <Settings className="w-4 h-4 text-blue-600" />
                      <span className="text-sm font-bold text-blue-700">
                        {result.resource_type}.{result.resource_name}
                      </span>
                      {result.line_number && (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-700">
                          <Hash className="w-3 h-3 mr-1" />
                          Line {result.line_number}
                        </span>
                      )}
                      {validationDetails && !validationDetails.supports_tagging && (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-700">
                          <Shield className="w-3 h-3 mr-1" />
                          Non-Taggable (Compliant)
                        </span>
                      )}
                    </div>
                  ) : (
                    <div className="flex items-center space-x-2">
                      <FileText className="w-4 h-4 text-gray-600" />
                      <span className="text-sm font-bold text-gray-800">
                        Resource in {result.file_path}
                      </span>
                      {result.line_number && (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-700">
                          <Hash className="w-3 h-3 mr-1" />
                          Line {result.line_number}
                        </span>
                      )}
                    </div>
                  )}
                </div>
                {result.resource_type && (
                  <div className="text-xs text-gray-500">
                    Resource Type: {result.resource_type}
                    {result.resource_name && ` | Name: ${result.resource_name}`}
                  </div>
                )}
              </div>
            </div>

            {/* Quick Summary */}
            {validationDetails && (
              <div className="flex flex-wrap gap-2 mt-2">
                {validationDetails.supports_tagging ? (
                  <>
                    {validationDetails.missing_tags.length > 0 && (
                      <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800">
                        <Minus className="w-3 h-3 mr-1" />
                        {validationDetails.missing_tags.length} Missing Tag{validationDetails.missing_tags.length > 1 ? 's' : ''}
                      </span>
                    )}
                    {validationDetails.violations.length > 0 && (
                      <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-orange-100 text-orange-800">
                        <AlertTriangle className="w-3 h-3 mr-1" />
                        {validationDetails.violations.length} Violation{validationDetails.violations.length > 1 ? 's' : ''}
                      </span>
                    )}
                    {validationDetails.extra_tags.length > 0 && (
                      <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                        <Plus className="w-3 h-3 mr-1" />
                        {validationDetails.extra_tags.length} Extra Tag{validationDetails.extra_tags.length > 1 ? 's' : ''}
                      </span>
                    )}
                    {validationDetails.compliance_status && (
                      <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                        <CheckCircle className="w-3 h-3 mr-1" />
                        Compliant
                      </span>
                    )}
                  </>
                ) : (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                    <Shield className="w-3 h-3 mr-1" />
                    Compliant (Non-Taggable Resource)
                  </span>
                )}
              </div>
            )}
          </div>

          <div className="ml-4 flex items-center space-x-2">
            <span className={`px-3 py-1 rounded-full text-xs font-medium border ${getActionColor(result.action)}`}>
              {result.action.charAt(0).toUpperCase() + result.action.slice(1)}
            </span>
            <button className="text-gray-400 hover:text-gray-600 transition-colors">
              {isExpanded ? (
                <ChevronDown className="w-5 h-5" />
              ) : (
                <ChevronRight className="w-5 h-5" />
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Expanded Details */}
      {isExpanded && validationDetails && (
        <div className="border-t bg-white bg-opacity-50 p-4 space-y-4">
          
          {/* Resource Information */}
          {result.resource_type && (
            <div className="bg-white rounded-lg p-3 border">
              <h4 className="text-sm font-semibold text-gray-800 mb-2 flex items-center">
                <Settings className="w-4 h-4 mr-2" />
                Resource Information
              </h4>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-2 text-sm">
                <div>
                  <span className="font-medium text-gray-600">Type:</span>
                  <span className="ml-2 text-gray-800">{result.resource_type}</span>
                </div>
                <div>
                  <span className="font-medium text-gray-600">Name:</span>
                  <span className="ml-2 text-gray-800">{result.resource_name || 'N/A'}</span>
                </div>
                <div>
                  <span className="font-medium text-gray-600">Supports Tagging:</span>
                  <span className={`ml-2 flex items-center ${validationDetails.supports_tagging ? 'text-green-600' : 'text-red-600'}`}>
                    {validationDetails.supports_tagging ? (
                      <>
                        <CheckCircle className="w-3 h-3 mr-1" />
                        Yes
                      </>
                    ) : (
                      <>
                        <XCircle className="w-3 h-3 mr-1" />
                        No
                      </>
                    )}
                  </span>
                </div>
              </div>
            </div>
          )}


          {/* Missing Tags */}
          {validationDetails.missing_tags.length > 0 && (
            <div className="bg-red-50 rounded-lg p-3 border border-red-200">
              <h4 className="text-sm font-semibold text-red-800 mb-2 flex items-center">
                <Minus className="w-4 h-4 mr-2" />
                Missing Required Tags
              </h4>
              <div className="flex flex-wrap gap-2">
                {validationDetails.missing_tags.map((tag, index) => (
                  <span 
                    key={index}
                    className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-red-100 text-red-800 border border-red-300"
                  >
                    <Tag className="w-3 h-3 mr-1" />
                    {tag}
                  </span>
                ))}
              </div>
              <p className="text-xs text-red-600 mt-2">
                These tags are required by your tag standard and must be added to the resource.
              </p>
            </div>
          )}

          {/* Tag Violations */}
          {validationDetails.violations.length > 0 && (
            <div className="bg-orange-50 rounded-lg p-3 border border-orange-200">
              <h4 className="text-sm font-semibold text-orange-800 mb-2 flex items-center">
                <AlertTriangle className="w-4 h-4 mr-2" />
                Tag Value Violations
              </h4>
              <div className="space-y-2">
                {validationDetails.violations.map((violation: TagViolation, index) => (
                  <div key={index} className="bg-white rounded border border-orange-200 p-3">
                    <div className="flex items-start space-x-2">
                      <div className="flex-shrink-0 mt-0.5">{getViolationTypeIcon(violation.violation_type)}</div>
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-1">
                          <Tag className="w-3 h-3 text-orange-600" />
                          <span className="font-medium text-orange-800">{violation.tag_key}</span>
                          <span className="px-2 py-1 rounded text-xs bg-orange-100 text-orange-700">
                            {formatViolationType(violation.violation_type)}
                          </span>
                        </div>
                        <div className="text-sm text-gray-700 space-y-1">
                          <div>
                            <span className="font-medium">Current Value:</span>
                            <span className="ml-2 px-2 py-1 bg-red-100 text-red-800 rounded text-xs font-mono">
                              "{violation.tag_value}"
                            </span>
                          </div>
                          {violation.expected && (
                            <div>
                              <span className="font-medium">Expected:</span>
                              <span className="ml-2 px-2 py-1 bg-green-100 text-green-800 rounded text-xs font-mono">
                                {violation.expected}
                              </span>
                            </div>
                          )}
                          <div className="text-xs text-gray-600 mt-1">
                          {violation.message}
                        </div>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Extra Tags */}
          {validationDetails.extra_tags.length > 0 && (
            <div className="bg-yellow-50 rounded-lg p-3 border border-yellow-200">
              <h4 className="text-sm font-semibold text-yellow-800 mb-2 flex items-center">
                <Plus className="w-4 h-4 mr-2" />
                Extra Tags (Not in Standard)
              </h4>
              <div className="flex flex-wrap gap-2">
                {validationDetails.extra_tags.map((tag, index) => (
                  <span 
                    key={index}
                    className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-yellow-100 text-yellow-800 border border-yellow-300"
                  >
                    <Tag className="w-3 h-3 mr-1" />
                    {tag}
                  </span>
                ))}
              </div>
              <p className="text-xs text-yellow-600 mt-2">
                These tags are not defined in your tag standard. Consider adding them to the standard or removing them from the resource.
              </p>
            </div>
          )}

          {/* Code Snippet */}
          {result.snippet && (
            <div className="bg-gray-50 rounded-lg p-3 border border-gray-200">
              <h4 className="text-sm font-semibold text-gray-800 mb-2 flex items-center justify-between">
                <div className="flex items-center">
                  <Code className="w-4 h-4 mr-2" />
                  Resource Definition
                  {result.line_number && (
                    <span className="ml-2 text-xs text-gray-500">
                      (Starting at Line {result.line_number})
                    </span>
                  )}
                </div>
                <span className="text-xs text-gray-500 bg-gray-200 px-2 py-1 rounded">
                  {result.snippet?.split('\n').length || 0} lines
                </span>
              </h4>
              <div className="relative">
                <pre className="text-xs text-gray-700 bg-white p-4 rounded border overflow-x-auto max-h-96 overflow-y-auto font-mono leading-relaxed shadow-inner">
                  <code className="language-hcl">{result.snippet}</code>
                </pre>
                <div className="absolute top-2 right-2">
                  <button 
                    onClick={() => navigator.clipboard.writeText(result.snippet || '')}
                    className="text-gray-400 hover:text-gray-600 transition-colors p-1 rounded bg-white/80 hover:bg-white"
                    title="Copy to clipboard"
                  >
                    <Code className="w-3 h-3" />
                  </button>
                </div>
              </div>
              <div className="mt-2 text-xs text-gray-500 flex items-center justify-between">
                <span>Complete resource definition extracted from Terraform file</span>
                <span>{result.snippet?.length || 0} characters</span>
              </div>
            </div>
          )}

          {/* Success State */}
          {!validationDetails.supports_tagging ? (
            <div className="bg-green-50 rounded-lg p-3 border border-green-200">
              <div className="flex items-center space-x-2">
                <Shield className="w-5 h-5 text-green-600" />
                <div>
                  <h4 className="text-sm font-semibold text-green-800">Non-Taggable Resource - Compliant</h4>
                  <p className="text-xs text-green-600">
                    This resource type does not support tagging and is considered compliant by default.
                  </p>
                </div>
              </div>
            </div>
          ) : (
            validationDetails.compliance_status && 
            validationDetails.missing_tags.length === 0 && 
            validationDetails.violations.length === 0 && 
            validationDetails.extra_tags.length === 0 && (
              <div className="bg-green-50 rounded-lg p-3 border border-green-200">
                <div className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5 text-green-600" />
                  <div>
                    <h4 className="text-sm font-semibold text-green-800">Perfect Compliance</h4>
                    <p className="text-xs text-green-600">
                      This resource meets all tag requirements defined in your standard.
                    </p>
                  </div>
                </div>
              </div>
            )
          )}

        </div>
      )}
    </div>
  );
};

export default ValidationResultCard;