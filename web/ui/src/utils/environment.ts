// Environment detection and path utilities for Terratag UI

export interface EnvironmentInfo {
  type: 'docker' | 'native';
  pathPrefix: string;
  pathSeparator: string;
  examples: {
    absolute: string;
    relative: string;
  };
  validation: {
    requiresAbsolute: boolean;
    recommendedPrefix?: string;
  };
}

/**
 * Detect the current environment (Docker vs Native)
 */
export const detectEnvironment = async (): Promise<EnvironmentInfo> => {
  try {
    // Check multiple indicators for Docker environment
    const isDocker = 
      // Check if we're running on Docker's typical port mapping
      (window.location.hostname === 'localhost' && window.location.port === '8080') ||
      // Check if running in container via user agent
      navigator.userAgent.includes('Docker') ||
      // Check if the URL suggests Docker environment
      window.location.hostname === '0.0.0.0' ||
      // Check via health API if available
      await checkDockerViaAPI();

    if (isDocker) {
      return {
        type: 'docker',
        pathPrefix: '/workspace',
        pathSeparator: '/',
        examples: {
          absolute: '/workspace/terraform',
          relative: '/workspace/my-project'
        },
        validation: {
          requiresAbsolute: true,
          recommendedPrefix: '/workspace'
        }
      };
    } else {
      // Detect OS for native environment
      const isWindows = navigator.platform.toLowerCase().includes('win');
      
      return {
        type: 'native',
        pathPrefix: isWindows ? 'C:\\\\' : '/home',
        pathSeparator: isWindows ? '\\\\' : '/',
        examples: {
          absolute: isWindows ? 'C:\\\\projects\\\\terraform' : '/home/user/terraform',
          relative: isWindows ? '.\\\\terraform' : './terraform'
        },
        validation: {
          requiresAbsolute: false
        }
      };
    }
  } catch (error) {
    // Fallback to native environment
    console.warn('Failed to detect environment, defaulting to native:', error);
    return {
      type: 'native',
      pathPrefix: '',
      pathSeparator: '/',
      examples: {
        absolute: '/home/user/terraform',
        relative: './terraform'
      },
      validation: {
        requiresAbsolute: false
      }
    };
  }
};

/**
 * Check if running in Docker via API call
 */
const checkDockerViaAPI = async (): Promise<boolean> => {
  try {
    const response = await fetch('/api/v1/health', { 
      method: 'GET',
      timeout: 2000 
    } as any);
    const data = await response.json();
    return data?.environment === 'docker' || data?.container === true;
  } catch (error) {
    return false;
  }
};

/**
 * Validate a path based on the environment
 */
export interface PathValidation {
  valid: boolean;
  message: string;
  severity: 'error' | 'warning' | 'info';
}

export const validatePath = (path: string, envInfo: EnvironmentInfo): PathValidation => {
  if (!path.trim()) {
    return {
      valid: false,
      message: 'Path is required',
      severity: 'error'
    };
  }

  if (envInfo.type === 'docker') {
    // Docker environment validation
    if (!path.startsWith('/')) {
      return {
        valid: false,
        message: 'Docker environment requires absolute paths (e.g., /workspace/terraform)',
        severity: 'error'
      };
    }
    
    if (!path.startsWith('/workspace')) {
      return {
        valid: true,
        message: 'Note: In Docker, your code should typically be mounted to /workspace. Example: /workspace/terraform',
        severity: 'warning'
      };
    }
    
    return {
      valid: true,
      message: 'Valid Docker path',
      severity: 'info'
    };
  } else {
    // Native environment validation
    if (path.includes('\\\\') && !path.includes('/')) {
      // Windows path detected
      return {
        valid: true,
        message: 'Windows path detected. Example: C:\\\\projects\\\\terraform or .\\\\terraform',
        severity: 'info'
      };
    }
    
    if (path.startsWith('~')) {
      return {
        valid: true,
        message: 'Home directory path detected (e.g., ~/projects/terraform)',
        severity: 'info'
      };
    }
    
    if (path.startsWith('./') || path.startsWith('../')) {
      return {
        valid: true,
        message: 'Relative path detected',
        severity: 'info'
      };
    }
    
    if (path.startsWith('/')) {
      return {
        valid: true,
        message: 'Absolute path detected',
        severity: 'info'
      };
    }
    
    return {
      valid: true,
      message: 'Path appears valid',
      severity: 'info'
    };
  }
};

/**
 * Get path suggestions based on environment
 */
export const getPathSuggestions = (envInfo: EnvironmentInfo): string[] => {
  if (envInfo.type === 'docker') {
    return [
      '/workspace',
      '/workspace/terraform',
      '/workspace/infrastructure',
      '/workspace/tf',
      '/workspace/iac'
    ];
  } else {
    const isWindows = envInfo.pathSeparator === '\\\\';
    
    if (isWindows) {
      return [
        '.\\\\terraform',
        '.\\\\infrastructure',
        'C:\\\\projects\\\\terraform',
        'C:\\\\code\\\\terraform',
        'D:\\\\projects\\\\terraform'
      ];
    } else {
      return [
        './terraform',
        './infrastructure',
        '~/projects/terraform',
        '/home/user/terraform',
        '/opt/terraform',
        '/var/terraform'
      ];
    }
  }
};

/**
 * Format path for display
 */
export const formatPathForDisplay = (path: string, envInfo: EnvironmentInfo): string => {
  if (!path) return '';
  
  // Handle Windows paths in display
  if (envInfo.type === 'native' && envInfo.pathSeparator === '\\\\') {
    return path.replace(/\\\\/g, '\\\\');
  }
  
  return path;
};

/**
 * Get environment-specific help text
 */
export const getEnvironmentHelpText = (envInfo: EnvironmentInfo): string => {
  if (envInfo.type === 'docker') {
    return `
Running in Docker container. Your Terraform files should be mounted as volumes to the container.
Typically, you would run:
  docker run -v /host/path/to/terraform:/workspace/terraform terratag
Then use paths like: /workspace/terraform
    `.trim();
  } else {
    return `
Running natively on your local machine. You can use:
- Relative paths: ./terraform, ../infrastructure
- Absolute paths: /home/user/terraform, C:\\\\projects\\\\terraform
- Home directory: ~/projects/terraform
    `.trim();
  }
};