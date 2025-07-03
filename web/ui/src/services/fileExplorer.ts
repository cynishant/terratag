
export interface DirectoryItem {
  name: string;
  path: string;
  is_directory: boolean;
  size?: number;
  mod_time?: string;
  has_terraform?: boolean;
}

export interface DirectoryListing {
  current_path: string;
  parent_path?: string;
  items: DirectoryItem[];
  is_root: boolean;
}

export interface DirectoryInfo {
  path: string;
  exists: boolean;
  is_directory: boolean;
  terraform_files: number;
  subdirectories: number;
  is_initialized: boolean;
  has_terraform: boolean;
}

class FileExplorerService {
  private baseUrl = '/api/v1/files';

  async browseDirectory(path: string = '/'): Promise<DirectoryListing> {
    const response = await fetch(`${this.baseUrl}/browse?path=${encodeURIComponent(path)}`);
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to browse directory');
    }
    return response.json();
  }

  async getDirectoryInfo(path: string): Promise<DirectoryInfo> {
    const response = await fetch(`${this.baseUrl}/info?path=${encodeURIComponent(path)}`);
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to get directory info');
    }
    return response.json();
  }

  // Helper method to format file sizes
  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  // Helper method to get path breadcrumbs
  getPathBreadcrumbs(path: string): Array<{ name: string; path: string }> {
    if (path === '/') {
      return [{ name: 'Root', path: '/' }];
    }

    const parts = path.split('/').filter(part => part !== '');
    const breadcrumbs = [{ name: 'Root', path: '/' }];
    
    let currentPath = '';
    for (const part of parts) {
      currentPath += '/' + part;
      breadcrumbs.push({
        name: part,
        path: currentPath
      });
    }

    return breadcrumbs;
  }

  // Helper method to validate path for different environments
  isValidPath(path: string): boolean {
    // Basic validation
    if (!path || path.includes('..')) {
      return false;
    }

    // In Docker environment, restrict to certain paths
    const allowedPaths = ['/workspace', '/demo-deployment', '/standards', '/tmp'];
    
    if (path === '/') {
      return true;
    }

    return allowedPaths.some(allowed => path.startsWith(allowed));
  }
}

export const fileExplorerService = new FileExplorerService();