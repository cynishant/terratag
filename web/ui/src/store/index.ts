import { create } from 'zustand';
import { TagStandard, Operation, OperationSummary, CreateOperationRequest, CreateTagStandardRequest } from '../types';
import { tagStandardsApi, operationsApi } from '../api/client';

// Tag Standards Store
interface TagStandardsState {
  standards: TagStandard[];
  selectedStandard: TagStandard | null;
  loading: boolean;
  error: string | null;
  
  // Actions
  fetchStandards: (provider?: string) => Promise<void>;
  createStandard: (data: CreateTagStandardRequest) => Promise<TagStandard>;
  updateStandard: (id: number, data: CreateTagStandardRequest) => Promise<TagStandard>;
  deleteStandard: (id: number) => Promise<void>;
  setSelectedStandard: (standard: TagStandard | null) => void;
  setError: (error: string | null) => void;
}

export const useTagStandardsStore = create<TagStandardsState>((set, get) => ({
  standards: [],
  selectedStandard: null,
  loading: false,
  error: null,
  
  fetchStandards: async (provider?: string) => {
    set({ loading: true, error: null });
    try {
      const standards = await tagStandardsApi.list(provider);
      set({ standards, loading: false });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch standards', loading: false });
    }
  },
  
  createStandard: async (data: CreateTagStandardRequest) => {
    set({ loading: true, error: null });
    try {
      const standard = await tagStandardsApi.create(data);
      set((state) => ({ 
        standards: [...state.standards, standard], 
        loading: false 
      }));
      return standard;
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to create standard', loading: false });
      throw error;
    }
  },
  
  updateStandard: async (id: number, data: CreateTagStandardRequest) => {
    set({ loading: true, error: null });
    try {
      const standard = await tagStandardsApi.update(id, data);
      set((state) => ({
        standards: state.standards.map(s => s.id === id ? standard : s),
        loading: false
      }));
      return standard;
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to update standard', loading: false });
      throw error;
    }
  },
  
  deleteStandard: async (id: number) => {
    set({ loading: true, error: null });
    try {
      await tagStandardsApi.delete(id);
      set((state) => ({
        standards: state.standards.filter(s => s.id !== id),
        loading: false
      }));
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to delete standard', loading: false });
      throw error;
    }
  },
  
  setSelectedStandard: (standard) => set({ selectedStandard: standard }),
  setError: (error) => set({ error }),
}));

// Operations Store
interface OperationsState {
  operations: Operation[];
  selectedOperation: Operation | null;
  operationSummary: OperationSummary | null;
  loading: boolean;
  error: string | null;
  
  // Actions
  fetchOperations: () => Promise<void>;
  createOperation: (data: CreateOperationRequest) => Promise<Operation>;
  executeOperation: (id: number) => Promise<void>;
  getOperationSummary: (id: number) => Promise<OperationSummary>;
  deleteOperation: (id: number) => Promise<void>;
  setSelectedOperation: (operation: Operation | null) => void;
  setError: (error: string | null) => void;
}

export const useOperationsStore = create<OperationsState>((set, get) => ({
  operations: [],
  selectedOperation: null,
  operationSummary: null,
  loading: false,
  error: null,
  
  fetchOperations: async () => {
    set({ loading: true, error: null });
    try {
      const operations = await operationsApi.list();
      set({ operations, loading: false });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch operations', loading: false });
    }
  },
  
  createOperation: async (data: CreateOperationRequest) => {
    set({ loading: true, error: null });
    try {
      const operation = await operationsApi.create(data);
      set((state) => ({ 
        operations: [...state.operations, operation], 
        loading: false 
      }));
      return operation;
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to create operation', loading: false });
      throw error;
    }
  },
  
  executeOperation: async (id: number) => {
    set({ loading: true, error: null });
    try {
      await operationsApi.execute(id);
      // Update the operation status to 'running'
      set((state) => ({
        operations: state.operations.map(o => 
          o.id === id ? { ...o, status: 'running' as const } : o
        ),
        loading: false
      }));
      
      // Poll for status updates
      const pollStatus = async () => {
        try {
          const operation = await operationsApi.get(id);
          set((state) => ({
            operations: state.operations.map(o => o.id === id ? operation : o)
          }));
          
          if (operation.status === 'running') {
            setTimeout(pollStatus, 2000); // Poll every 2 seconds
          }
        } catch (error) {
          console.error('Failed to poll operation status:', error);
        }
      };
      
      setTimeout(pollStatus, 1000); // Start polling after 1 second
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to execute operation', loading: false });
      throw error;
    }
  },
  
  getOperationSummary: async (id: number) => {
    set({ loading: true, error: null });
    try {
      const summary = await operationsApi.getSummary(id);
      set({ operationSummary: summary, loading: false });
      return summary;
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to get operation summary', loading: false });
      throw error;
    }
  },
  
  deleteOperation: async (id: number) => {
    set({ loading: true, error: null });
    try {
      await operationsApi.delete(id);
      set((state) => ({
        operations: state.operations.filter(o => o.id !== id),
        loading: false
      }));
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to delete operation', loading: false });
      throw error;
    }
  },
  
  setSelectedOperation: (operation) => set({ selectedOperation: operation }),
  setError: (error) => set({ error }),
}));

// Legacy combined store for backward compatibility
interface AppState {
  // Tag Standards
  tagStandards: TagStandard[];
  selectedStandard: TagStandard | null;
  
  // Operations
  operations: Operation[];
  selectedOperation: Operation | null;
  operationSummary: OperationSummary | null;
  
  // UI State
  loading: boolean;
  error: string | null;
  
  // Actions
  setTagStandards: (standards: TagStandard[]) => void;
  setSelectedStandard: (standard: TagStandard | null) => void;
  addTagStandard: (standard: TagStandard) => void;
  updateTagStandard: (id: number, standard: TagStandard) => void;
  removeTagStandard: (id: number) => void;
  
  setOperations: (operations: Operation[]) => void;
  setSelectedOperation: (operation: Operation | null) => void;
  setOperationSummary: (summary: OperationSummary | null) => void;
  addOperation: (operation: Operation) => void;
  updateOperation: (id: number, operation: Operation) => void;
  removeOperation: (id: number) => void;
  
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useStore = create<AppState>((set) => ({
  // Initial state
  tagStandards: [],
  selectedStandard: null,
  operations: [],
  selectedOperation: null,
  operationSummary: null,
  loading: false,
  error: null,
  
  // Tag Standards actions
  setTagStandards: (standards) => set({ tagStandards: standards }),
  setSelectedStandard: (standard) => set({ selectedStandard: standard }),
  addTagStandard: (standard) => set((state) => ({ 
    tagStandards: [...state.tagStandards, standard] 
  })),
  updateTagStandard: (id, standard) => set((state) => ({
    tagStandards: state.tagStandards.map(s => s.id === id ? standard : s)
  })),
  removeTagStandard: (id) => set((state) => ({
    tagStandards: state.tagStandards.filter(s => s.id !== id)
  })),
  
  // Operations actions
  setOperations: (operations) => set({ operations }),
  setSelectedOperation: (operation) => set({ selectedOperation: operation }),
  setOperationSummary: (summary) => set({ operationSummary: summary }),
  addOperation: (operation) => set((state) => ({ 
    operations: [...state.operations, operation] 
  })),
  updateOperation: (id, operation) => set((state) => ({
    operations: state.operations.map(o => o.id === id ? operation : o)
  })),
  removeOperation: (id) => set((state) => ({
    operations: state.operations.filter(o => o.id !== id)
  })),
  
  // UI actions
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
}));