import axios, { AxiosResponse } from 'axios';
import {
  TagStandard,
  CreateTagStandardRequest,
  Operation,
  CreateOperationRequest,
  OperationSummary,
  ApiResponse,
  PaginationParams,
} from '../types';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response interceptor to handle API responses
apiClient.interceptors.response.use(
  (response: AxiosResponse<ApiResponse<any>>) => {
    return response;
  },
  (error) => {
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

// Tag Standards API
export const tagStandardsApi = {
  async list(provider?: string): Promise<TagStandard[]> {
    const params = provider ? { provider } : {};
    const response = await apiClient.get<ApiResponse<TagStandard[]>>('/standards', { params });
    return response.data.data || [];
  },

  async get(id: number): Promise<TagStandard> {
    const response = await apiClient.get<ApiResponse<TagStandard>>(`/standards/${id}`);
    if (!response.data.data) {
      throw new Error('Tag standard not found');
    }
    return response.data.data;
  },

  async create(data: CreateTagStandardRequest): Promise<TagStandard> {
    const response = await apiClient.post<ApiResponse<TagStandard>>('/standards', data);
    if (!response.data.data) {
      throw new Error('Failed to create tag standard');
    }
    return response.data.data;
  },

  async update(id: number, data: CreateTagStandardRequest): Promise<TagStandard> {
    const response = await apiClient.put<ApiResponse<TagStandard>>(`/standards/${id}`, data);
    if (!response.data.data) {
      throw new Error('Failed to update tag standard');
    }
    return response.data.data;
  },

  async delete(id: number): Promise<void> {
    await apiClient.delete(`/standards/${id}`);
  },

  async validateContent(content: string, cloud_provider: string): Promise<{ valid: boolean }> {
    const response = await apiClient.post<ApiResponse<{ valid: boolean }>>('/standards/validate', {
      content,
      cloud_provider,
    });
    return response.data.data || { valid: false };
  },
};

// Operations API
export const operationsApi = {
  async list(params?: PaginationParams): Promise<Operation[]> {
    const response = await apiClient.get<ApiResponse<Operation[]>>('/operations', { params });
    return response.data.data || [];
  },

  async get(id: number): Promise<Operation> {
    const response = await apiClient.get<ApiResponse<Operation>>(`/operations/${id}`);
    if (!response.data.data) {
      throw new Error('Operation not found');
    }
    return response.data.data;
  },

  async create(data: CreateOperationRequest): Promise<Operation> {
    const response = await apiClient.post<ApiResponse<Operation>>('/operations', data);
    if (!response.data.data) {
      throw new Error('Failed to create operation');
    }
    return response.data.data;
  },

  async execute(id: number): Promise<void> {
    await apiClient.post(`/operations/${id}/execute`);
  },

  async getSummary(id: number): Promise<OperationSummary> {
    const response = await apiClient.get<ApiResponse<OperationSummary>>(`/operations/${id}/summary`);
    if (!response.data.data) {
      throw new Error('Operation summary not found');
    }
    return response.data.data;
  },

  async delete(id: number): Promise<void> {
    await apiClient.delete(`/operations/${id}`);
  },
};

// Health check API
export const healthApi = {
  async check(): Promise<{ status: string; version: string }> {
    const response = await apiClient.get<ApiResponse<{ status: string; version: string }>>('/health');
    return response.data.data || { status: 'unknown', version: 'unknown' };
  },
};

export default apiClient;