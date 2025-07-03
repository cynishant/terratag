export interface TagStandard {
  id: number;
  name: string;
  description: string;
  cloud_provider: string;
  version: number;
  content: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTagStandardRequest {
  name: string;
  description: string;
  cloud_provider: string;
  version: number;
  content: string;
}

export interface Operation {
  id: number;
  type: 'validation' | 'tagging';
  status: 'pending' | 'running' | 'completed' | 'failed';
  standard_id?: number;
  directory_path: string;
  filter_pattern?: string;
  skip_pattern?: string;
  settings?: string;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface CreateOperationRequest {
  type: 'validation' | 'tagging';
  standard_id?: number;
  directory_path: string;
  filter_pattern?: string;
  skip_pattern?: string;
  settings?: string;
}

export interface OperationResult {
  id: number;
  operation_id: number;
  file_path: string;
  resource_type?: string;
  resource_name?: string;
  line_number?: number;
  snippet?: string;
  action: string;
  violation_type?: string;
  details?: string;
  created_at: string;
}

// Detailed validation result structure from the details JSON
export interface ValidationDetails {
  violations: TagViolation[];
  compliance_status: boolean;
  supports_tagging: boolean;
  missing_tags: string[];
  extra_tags: string[];
  message?: string;
  variable_resolution?: VariableResolution;
}

// Variable resolution information
export interface VariableResolution {
  literal_values: Record<string, string>;
  resolved_variables: Record<string, ResolvedVariable>;
  unresolved_references: string[];
  all_variables?: AllVariablesInfo; // New field for all variables in the codebase
}

// Information about all variables in the codebase
export interface AllVariablesInfo {
  variables: Record<string, VariableInfo>;
  locals: Record<string, LocalInfo>;
}

export interface VariableInfo {
  name: string;
  type?: string;
  description?: string;
  default?: any;
  value?: any; // Resolved value if available
  resolved: boolean;
  file_path?: string;
  line_number?: number;
}

export interface LocalInfo {
  name: string;
  expression?: string;
  value?: any; // Resolved value if available
  resolved: boolean;
  file_path?: string;
  line_number?: number;
}

export interface ResolvedVariable {
  reference: string;
  resolved: boolean;
  source: 'variable' | 'local' | 'interpolation';
  tag_location: string;
  value: string;
}

export interface TagViolation {
  tag_key: string;
  tag_value: string;
  violation_type: string;
  expected?: string;
  message: string;
}

export interface OperationLog {
  id: number;
  operation_id: number;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  details?: string;
  created_at: string;
}

export interface OperationStats {
  total_files: number;
  processed_files: number;
  tagged_resources: number;
  violations: number;
  errors: number;
}

export interface OperationSummary {
  operation: Operation;
  results: OperationResult[];
  logs: OperationLog[];
  summary: OperationStats;
}

export interface ApiResponse<T> {
  message: string;
  data?: T;
}

export interface ApiError {
  error: string;
  message: string;
  code: number;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface CloudProvider {
  id: string;
  name: string;
  icon: string;
}

export const CLOUD_PROVIDERS: CloudProvider[] = [
  { id: 'aws', name: 'Amazon Web Services', icon: 'aws' },
  { id: 'azure', name: 'Microsoft Azure', icon: 'azure' },
  { id: 'gcp', name: 'Google Cloud Platform', icon: 'gcp' },
  { id: 'generic', name: 'Generic Cloud', icon: 'generic' },
];

export interface TagRequirement {
  key: string;
  description: string;
  required: boolean;
  data_type: 'string' | 'number' | 'boolean';
  default_value?: string;
  allowed_values?: string[];
  pattern?: string;
  examples?: string[];
}

export interface TagStandardSchema {
  version: number;
  cloud_provider: string;
  required_tags: TagRequirement[];
  optional_tags: TagRequirement[];
  validation_rules?: {
    [key: string]: any;
  };
}