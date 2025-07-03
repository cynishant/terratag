import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import Card from '../common/Card';
import ValidationResultCard from './ValidationResultCard';
import { useOperationsStore } from '../../store';
import { OperationSummary as OperationSummaryType, OperationResult, OperationLog, ValidationDetails } from '../../types';
import { 
  FileText, 
  Settings, 
  Tag, 
  XCircle, 
  AlertTriangle, 
  CheckCircle, 
  Shield, 
  Minus, 
  TrendingUp,
  BarChart3,
  Folder,
  Clock,
  Code,
  Variable,
  Hash,
  ChevronDown,
  ChevronRight
} from 'lucide-react';

const OperationSummary: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { getOperationSummary, loading, error } = useOperationsStore();
  const [summary, setSummary] = useState<OperationSummaryType | null>(null);
  const [isAllVariablesExpanded, setIsAllVariablesExpanded] = useState(false);

  useEffect(() => {
    if (id) {
      loadSummary(parseInt(id));
    }
  }, [id]);

  const loadSummary = async (operationId: number) => {
    try {
      const summaryData = await getOperationSummary(operationId);
      setSummary(summaryData);
    } catch (err) {
      console.error('Failed to load operation summary:', err);
    }
  };

  const getLogLevelColor = (level: OperationLog['level']) => {
    switch (level) {
      case 'error': return 'text-red-600 bg-red-50';
      case 'warn': return 'text-yellow-600 bg-yellow-50';
      case 'info': return 'text-blue-600 bg-blue-50';
      case 'debug': return 'text-gray-600 bg-gray-50';
      default: return 'text-gray-600 bg-gray-50';
    }
  };


  if (loading) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-300 rounded w-1/4 mb-4"></div>
          <div className="h-32 bg-gray-300 rounded"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <Card>
        <div className="p-6">
          <div className="text-center">
            <AlertTriangle className="w-8 h-8 text-red-600 mb-2 mx-auto" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">Error Loading Operation</h3>
            <p className="text-gray-600">{error}</p>
          </div>
        </div>
      </Card>
    );
  }

  if (!summary) {
    return (
      <Card>
        <div className="p-6">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900 mb-2">Operation Not Found</h3>
            <p className="text-gray-600">The requested operation could not be found.</p>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">
          Operation #{summary.operation.id} Summary
        </h1>
        <p className="mt-2 text-gray-600">
          {summary.operation.type.charAt(0).toUpperCase() + summary.operation.type.slice(1)} operation results
        </p>
      </div>

      {/* Operation Details */}
      <Card>
        <div className="p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Operation Details</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Type</label>
              <p className="mt-1 text-sm text-gray-900">{summary.operation.type}</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Status</label>
              <p className="mt-1 text-sm text-gray-900">{summary.operation.status}</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Directory</label>
              <p className="mt-1 text-sm text-gray-900">{summary.operation.directory_path}</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Created</label>
              <p className="mt-1 text-sm text-gray-900">
                {new Date(summary.operation.created_at).toLocaleString()}
              </p>
            </div>
            {summary.operation.filter_pattern && (
              <div>
                <label className="block text-sm font-medium text-gray-700">Filter Pattern</label>
                <p className="mt-1 text-sm text-gray-900">{summary.operation.filter_pattern}</p>
              </div>
            )}
            {summary.operation.skip_pattern && (
              <div>
                <label className="block text-sm font-medium text-gray-700">Skip Pattern</label>
                <p className="mt-1 text-sm text-gray-900">{summary.operation.skip_pattern}</p>
              </div>
            )}
          </div>
        </div>
      </Card>

      {/* Statistics */}
      <Card>
        <div className="p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
            <BarChart3 className="w-5 h-5 mr-2" />
            Operation Statistics
          </h3>
          
          {/* Primary Metrics */}
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">{summary.summary.total_files}</div>
              <div className="text-sm text-gray-600 flex items-center justify-center">
                <FileText className="w-3 h-3 mr-1" />
                Total Files
              </div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">{summary.summary.processed_files}</div>
              <div className="text-sm text-gray-600 flex items-center justify-center">
                <Settings className="w-3 h-3 mr-1" />
                Processed
              </div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600">{summary.summary.tagged_resources}</div>
              <div className="text-sm text-gray-600 flex items-center justify-center">
                <Tag className="w-3 h-3 mr-1" />
                Tagged
              </div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-red-600">{summary.summary.violations}</div>
              <div className="text-sm text-gray-600 flex items-center justify-center">
                <XCircle className="w-3 h-3 mr-1" />
                Violations
              </div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-orange-600">{summary.summary.errors}</div>
              <div className="text-sm text-gray-600 flex items-center justify-center">
                <AlertTriangle className="w-3 h-3 mr-1" />
                Errors
              </div>
            </div>
          </div>

          {/* Validation-specific metrics */}
          {summary.operation.type === 'validation' && summary.results.length > 0 && (
            <>
              <div className="border-t pt-4">
                <h4 className="text-sm font-semibold text-gray-700 mb-3 flex items-center">
                  <Shield className="w-4 h-4 mr-2" />
                  Validation Breakdown
                </h4>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <div className="text-center bg-green-50 rounded-lg p-3">
                    <div className="text-lg font-bold text-green-600">
                      {summary.results.filter(r => {
                        if (r.action === 'compliant') return true;
                        // Also count non-taggable resources as compliant
                        try {
                          const details = JSON.parse(r.details || '{}') as ValidationDetails;
                          return !details.supports_tagging;
                        } catch { return false; }
                      }).length}
                    </div>
                    <div className="text-xs text-green-700 flex items-center justify-center">
                      <CheckCircle className="w-3 h-3 mr-1" />
                      Compliant Resources
                    </div>
                  </div>
                  <div className="text-center bg-red-50 rounded-lg p-3">
                    <div className="text-lg font-bold text-red-600">
                      {summary.results.filter(r => r.action === 'violation').length}
                    </div>
                    <div className="text-xs text-red-700 flex items-center justify-center">
                      <XCircle className="w-3 h-3 mr-1" />
                      Non-Compliant Resources
                    </div>
                  </div>
                  <div className="text-center bg-yellow-50 rounded-lg p-3">
                    <div className="text-lg font-bold text-yellow-600">
                      {summary.results.filter(r => {
                        try {
                          const details = JSON.parse(r.details || '{}') as ValidationDetails;
                          return details.missing_tags?.length > 0;
                        } catch { return false; }
                      }).length}
                    </div>
                    <div className="text-xs text-yellow-700 flex items-center justify-center">
                      <Minus className="w-3 h-3 mr-1" />
                      Missing Tags
                    </div>
                  </div>
                  <div className="text-center bg-gray-50 rounded-lg p-3">
                    <div className="text-lg font-bold text-gray-600">
                      {summary.results.filter(r => {
                        try {
                          const details = JSON.parse(r.details || '{}') as ValidationDetails;
                          return !details.supports_tagging;
                        } catch { return false; }
                      }).length}
                    </div>
                    <div className="text-xs text-gray-700 flex items-center justify-center">
                      <Shield className="w-3 h-3 mr-1" />
                      Not Taggable
                    </div>
                  </div>
                </div>
              </div>

              {/* Compliance Rate */}
              <div className="border-t pt-4 mt-4">
                <div className="flex justify-between items-center">
                  <h4 className="text-sm font-semibold text-gray-700 flex items-center">
                    <TrendingUp className="w-4 h-4 mr-2" />
                    Compliance Rate
                  </h4>
                  <div className="text-right">
                    {(() => {
                      const compliant = summary.results.filter(r => {
                        if (r.action === 'compliant') return true;
                        // Also count non-taggable resources as compliant
                        try {
                          const details = JSON.parse(r.details || '{}') as ValidationDetails;
                          return !details.supports_tagging;
                        } catch { return false; }
                      }).length;
                      const total = summary.results.length;
                      const rate = total > 0 ? Math.round((compliant / total) * 100) : 0;
                      return (
                        <div>
                          <span className={`text-lg font-bold ${rate >= 80 ? 'text-green-600' : rate >= 60 ? 'text-yellow-600' : 'text-red-600'}`}>
                            {rate}%
                          </span>
                          <div className="text-xs text-gray-500">({compliant}/{total})</div>
                        </div>
                      );
                    })()}
                  </div>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                  {(() => {
                    const compliant = summary.results.filter(r => {
                      if (r.action === 'compliant') return true;
                      // Also count non-taggable resources as compliant
                      try {
                        const details = JSON.parse(r.details || '{}') as ValidationDetails;
                        return !details.supports_tagging;
                      } catch { return false; }
                    }).length;
                    const total = summary.results.length;
                    const rate = total > 0 ? (compliant / total) * 100 : 0;
                    return (
                      <div 
                        className={`h-2 rounded-full transition-all duration-300 ${
                          rate >= 80 ? 'bg-green-500' : rate >= 60 ? 'bg-yellow-500' : 'bg-red-500'
                        }`}
                        style={{ width: `${rate}%` }}
                      />
                    );
                  })()}
                </div>
              </div>
            </>
          )}
        </div>
      </Card>

      {/* All Variables & Locals Section */}
      {(() => {
        // Debug: Check if we have any operation_summary entries
        const operationSummaryEntries = summary.results.filter(r => r.resource_type === 'operation_summary');
        console.log('Operation summary entries:', operationSummaryEntries);
        
        // Find the operation_summary entry with variables_summary action
        const variablesSummaryEntry = summary.results.find(
          r => r.resource_type === 'operation_summary' && r.action === 'variables_summary'
        );
        
        console.log('Variables summary entry found:', variablesSummaryEntry);
        
        if (!variablesSummaryEntry?.details) {
          console.log('No variables summary entry found or no details');
          return null;
        }
        
        let allVariablesInfo;
        try {
          const details = JSON.parse(variablesSummaryEntry.details);
          console.log('Parsed details:', details);
          allVariablesInfo = details.all_variables;
          console.log('All variables info:', allVariablesInfo);
        } catch (error) {
          console.error('Failed to parse all variables info:', error);
          return null;
        }
        
        if (!allVariablesInfo) {
          console.log('No all_variables data found');
          return null;
        }
        
        const variablesCount = Object.keys(allVariablesInfo.variables || {}).length;
        const localsCount = Object.keys(allVariablesInfo.locals || {}).length;
        
        console.log('Variables count:', variablesCount, 'Locals count:', localsCount);
        
        if (variablesCount === 0 && localsCount === 0) {
          console.log('No variables or locals found, not showing section');
          return null;
        }
        
        return (
          <Card>
            <div className="p-6">
              <div 
                className="flex items-center justify-between cursor-pointer hover:bg-purple-50 -m-6 p-6 rounded-lg transition-colors"
                onClick={() => setIsAllVariablesExpanded(!isAllVariablesExpanded)}
              >
                <h3 className="text-lg font-medium text-purple-900 flex items-center">
                  <Code className="w-5 h-5 mr-2" />
                  All Variables & Locals
                  <span className="ml-3 text-sm text-purple-600 bg-purple-100 px-3 py-1 rounded-full">
                    {variablesCount} variables, {localsCount} locals
                  </span>
                </h3>
                <button className="text-purple-600 hover:text-purple-800 transition-colors">
                  {isAllVariablesExpanded ? (
                    <ChevronDown className="w-5 h-5" />
                  ) : (
                    <ChevronRight className="w-5 h-5" />
                  )}
                </button>
              </div>
              
              {isAllVariablesExpanded && (
                <div className="mt-6 space-y-4">
                  
                  {/* Variables Table */}
                  {variablesCount > 0 && (
                    <div>
                      <h4 className="text-sm font-semibold text-purple-700 mb-3 flex items-center">
                        <Variable className="w-4 h-4 mr-2" />
                        Variables ({variablesCount})
                      </h4>
                      <div className="bg-purple-50 rounded-lg border border-purple-200 overflow-hidden">
                        <div className="grid grid-cols-5 gap-2 px-4 py-3 bg-purple-100 text-sm font-semibold text-purple-800">
                          <div>Name</div>
                          <div>Type</div>
                          <div>Value</div>
                          <div>Status</div>
                          <div>Location</div>
                        </div>
                        {Object.entries(allVariablesInfo.variables || {}).map(([name, varInfo]: [string, any], index) => (
                          <div key={index} className="grid grid-cols-5 gap-2 px-4 py-3 border-t border-purple-100 text-sm">
                            <div className="flex items-center">
                              <span className="font-medium text-purple-800">{name}</span>
                            </div>
                            <div className="text-gray-600">{varInfo.type || 'string'}</div>
                            <div className="font-mono text-gray-700 truncate" title={JSON.stringify(varInfo.value || varInfo.default)}>
                              {varInfo.resolved ? (
                                <span className="bg-green-50 px-2 py-1 rounded text-green-800">
                                  {JSON.stringify(varInfo.value || varInfo.default)}
                                </span>
                              ) : (
                                <span className="text-gray-400 italic">unresolved</span>
                              )}
                            </div>
                            <div>
                              {varInfo.resolved ? (
                                <span className="inline-flex items-center text-green-600">
                                  <CheckCircle className="w-3 h-3 mr-1" />
                                  Resolved
                                </span>
                              ) : (
                                <span className="inline-flex items-center text-yellow-600">
                                  <AlertTriangle className="w-3 h-3 mr-1" />
                                  Unresolved
                                </span>
                              )}
                            </div>
                            <div className="text-gray-600 text-sm">
                              {varInfo.file_path && (
                                <span title={varInfo.file_path}>
                                  {varInfo.file_path.split('/').pop()}:{varInfo.line_number || '?'}
                                </span>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Locals Table */}
                  {localsCount > 0 && (
                    <div>
                      <h4 className="text-sm font-semibold text-purple-700 mb-3 flex items-center">
                        <Hash className="w-4 h-4 mr-2" />
                        Locals ({localsCount})
                      </h4>
                      <div className="bg-purple-50 rounded-lg border border-purple-200 overflow-hidden">
                        <div className="grid grid-cols-5 gap-2 px-4 py-3 bg-purple-100 text-sm font-semibold text-purple-800">
                          <div>Name</div>
                          <div>Expression</div>
                          <div>Value</div>
                          <div>Status</div>
                          <div>Location</div>
                        </div>
                        {Object.entries(allVariablesInfo.locals || {}).map(([name, localInfo]: [string, any], index) => (
                          <div key={index} className="grid grid-cols-5 gap-2 px-4 py-3 border-t border-purple-100 text-sm">
                            <div className="flex items-center">
                              <span className="font-medium text-purple-800">{name}</span>
                            </div>
                            <div className="font-mono text-gray-600 truncate" title={localInfo.expression}>
                              {localInfo.expression || '-'}
                            </div>
                            <div className="font-mono text-gray-700 truncate" title={JSON.stringify(localInfo.value)}>
                              {localInfo.resolved ? (
                                <span className="bg-green-50 px-2 py-1 rounded text-green-800">
                                  {JSON.stringify(localInfo.value)}
                                </span>
                              ) : (
                                <span className="text-gray-400 italic">unresolved</span>
                              )}
                            </div>
                            <div>
                              {localInfo.resolved ? (
                                <span className="inline-flex items-center text-green-600">
                                  <CheckCircle className="w-3 h-3 mr-1" />
                                  Resolved
                                </span>
                              ) : (
                                <span className="inline-flex items-center text-yellow-600">
                                  <AlertTriangle className="w-3 h-3 mr-1" />
                                  Unresolved
                                </span>
                              )}
                            </div>
                            <div className="text-gray-600 text-sm">
                              {localInfo.file_path && (
                                <span title={localInfo.file_path}>
                                  {localInfo.file_path.split('/').pop()}:{localInfo.line_number || '?'}
                                </span>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                </div>
              )}
            </div>
          </Card>
        );
      })()}

      {/* Results */}
      {summary.results.filter(r => r.resource_type !== 'operation_summary').length > 0 && (
        <Card>
          <div className="p-6">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium text-gray-900">Validation Results</h3>
              <div className="flex items-center space-x-2 text-sm text-gray-600">
                <span className="flex items-center">
                  <BarChart3 className="w-3 h-3 mr-1" />
                  {summary.results.filter(r => r.resource_type !== 'operation_summary').length} result{summary.results.filter(r => r.resource_type !== 'operation_summary').length > 1 ? 's' : ''}
                </span>
                {summary.operation.type === 'validation' && (
                  <>
                    <span>|</span>
                    <span className="text-green-600 flex items-center">
                      <CheckCircle className="w-3 h-3 mr-1" />
                      {summary.results.filter(r => {
                        if (r.action === 'compliant') return true;
                        // Also count non-taggable resources as compliant
                        try {
                          const details = JSON.parse(r.details || '{}') as ValidationDetails;
                          return !details.supports_tagging;
                        } catch { return false; }
                      }).length} compliant
                    </span>
                    <span>|</span>
                    <span className="text-red-600 flex items-center">
                      <XCircle className="w-3 h-3 mr-1" />
                      {summary.results.filter(r => r.action === 'violation').length} violations
                    </span>
                  </>
                )}
              </div>
            </div>
            
            {/* Results organized by file (highest to lowest violations) */}
            <div className="space-y-6">
              {(() => {
                // Filter out operation_summary entries from main results
                const actualResults = summary.results.filter(r => r.resource_type !== 'operation_summary');
                
                // Group results by file path
                const resultsByFile = actualResults.reduce((acc, result) => {
                  const filePath = result.file_path;
                  if (!acc[filePath]) {
                    acc[filePath] = [];
                  }
                  acc[filePath].push(result);
                  return acc;
                }, {} as Record<string, OperationResult[]>);

                // Calculate violation count per file and sort files by violation count (highest first)
                const sortedFiles = Object.entries(resultsByFile).sort(([, resultsA], [, resultsB]) => {
                  const getViolationCount = (results: OperationResult[]) => {
                    return results.reduce((count, result) => {
                      if (result.action === 'violation') {
                        try {
                          const details = JSON.parse(result.details || '{}') as ValidationDetails;
                          return count + (details.violations?.length || 0) + (details.missing_tags?.length || 0);
                        } catch {
                          return count + 1; // Count as 1 violation if can't parse
                        }
                      }
                      return count;
                    }, 0);
                  };

                  const violationsA = getViolationCount(resultsA);
                  const violationsB = getViolationCount(resultsB);
                  
                  // Sort by violation count (highest first), then by file path
                  if (violationsA !== violationsB) {
                    return violationsB - violationsA;
                  }
                  return resultsA[0].file_path.localeCompare(resultsB[0].file_path);
                });

                return sortedFiles.map(([filePath, fileResults]) => {
                  // Calculate file-level statistics
                  const totalViolations = fileResults.reduce((count, result) => {
                    if (result.action === 'violation') {
                      try {
                        const details = JSON.parse(result.details || '{}') as ValidationDetails;
                        return count + (details.violations?.length || 0) + (details.missing_tags?.length || 0);
                      } catch {
                        return count + 1;
                      }
                    }
                    return count;
                  }, 0);

                  const compliantResources = fileResults.filter(r => {
                    if (r.action === 'compliant') return true;
                    // Also count non-taggable resources as compliant
                    try {
                      const details = JSON.parse(r.details || '{}') as ValidationDetails;
                      return !details.supports_tagging;
                    } catch { return false; }
                  }).length;
                  const violationResources = fileResults.filter(r => r.action === 'violation').length;
                  const nonTaggableResources = fileResults.filter(r => {
                    try {
                      const details = JSON.parse(r.details || '{}') as ValidationDetails;
                      return !details.supports_tagging;
                    } catch {
                      return false;
                    }
                  }).length;

                  return (
                    <div key={filePath} className="border rounded-lg bg-white">
                      {/* File Header */}
                      <div className={`p-4 border-b ${totalViolations > 0 ? 'bg-red-50 border-red-200' : 'bg-green-50 border-green-200'}`}>
                        <div className="flex justify-between items-start">
                          <div>
                            <h4 className="text-lg font-semibold text-gray-900 flex items-center space-x-2">
                              <Folder className="w-5 h-5 text-blue-600" />
                              <span>{filePath}</span>
                            </h4>
                            <div className="flex items-center space-x-4 mt-2 text-sm">
                              <span className="text-gray-600 flex items-center">
                                <BarChart3 className="w-3 h-3 mr-1" />
                                {fileResults.length} resource{fileResults.length > 1 ? 's' : ''}
                              </span>
                              {compliantResources > 0 && (
                                <span className="text-green-600 flex items-center">
                                  <CheckCircle className="w-3 h-3 mr-1" />
                                  {compliantResources} compliant
                                </span>
                              )}
                              {violationResources > 0 && (
                                <span className="text-red-600 flex items-center">
                                  <XCircle className="w-3 h-3 mr-1" />
                                  {violationResources} with violations
                                </span>
                              )}
                              {nonTaggableResources > 0 && (
                                <span className="text-gray-600 flex items-center">
                                  <Shield className="w-3 h-3 mr-1" />
                                  {nonTaggableResources} non-taggable
                                </span>
                              )}
                            </div>
                          </div>
                          <div className="text-right">
                            <div className={`text-2xl font-bold flex items-center ${totalViolations > 0 ? 'text-red-600' : 'text-green-600'}`}>
                              {totalViolations > 0 ? (
                                <>
                                  <XCircle className="w-6 h-6 mr-1" />
                                  {totalViolations}
                                </>
                              ) : (
                                <CheckCircle className="w-6 h-6" />
                              )}
                            </div>
                            <div className="text-xs text-gray-500">
                              {totalViolations > 0 ? 'Total Violations' : 'All Compliant'}
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* File Resources */}
                      <div className="divide-y">
                        {fileResults
                          .sort((a, b) => {
                            // Within a file, sort by violation count, then by resource name
                            const getResourceViolationCount = (result: OperationResult) => {
                              if (result.action === 'violation') {
                                try {
                                  const details = JSON.parse(result.details || '{}') as ValidationDetails;
                                  return (details.violations?.length || 0) + (details.missing_tags?.length || 0);
                                } catch {
                                  return 1;
                                }
                              }
                              return 0;
                            };

                            const violationsA = getResourceViolationCount(a);
                            const violationsB = getResourceViolationCount(b);
                            
                            if (violationsA !== violationsB) {
                              return violationsB - violationsA;
                            }
                            
                            // Then sort by resource name
                            const resourceA = `${a.resource_type || ''}.${a.resource_name || ''}`;
                            const resourceB = `${b.resource_type || ''}.${b.resource_name || ''}`;
                            return resourceA.localeCompare(resourceB);
                          })
                          .map((result: OperationResult) => (
                            <div key={result.id} className="p-4">
                              <ValidationResultCard result={result} />
                            </div>
                          ))}
                      </div>
                    </div>
                  );
                });
              })()}
            </div>

            {/* No results message */}
            {summary.results.filter(r => r.resource_type !== 'operation_summary').length === 0 && (
              <div className="text-center py-8">
                <FileText className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                <p className="text-gray-500">No validation results found.</p>
                <p className="text-sm text-gray-400 mt-1">
                  This might indicate that no resources were processed or no violations were found.
                </p>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* Logs */}
      {summary.logs.length > 0 && (
        <Card>
          <div className="p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
              <FileText className="w-5 h-5 mr-2" />
              Logs
            </h3>
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {summary.logs.map((log: OperationLog) => (
                <div key={log.id} className={`p-3 rounded-lg ${getLogLevelColor(log.level)}`}>
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <span className="text-xs font-medium uppercase">{log.level}</span>
                        <Clock className="w-3 h-3 text-gray-400" />
                        <span className="text-xs text-gray-500">
                          {new Date(log.created_at).toLocaleTimeString()}
                        </span>
                      </div>
                      <p className="text-sm mt-1">{log.message}</p>
                      {log.details && (
                        <p className="text-xs mt-1 opacity-75">{log.details}</p>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </Card>
      )}
    </div>
  );
};

export default OperationSummary;