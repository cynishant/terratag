import React, { useState, useEffect } from 'react';
import { DirectoryListing, DirectoryItem, fileExplorerService } from '../../services/fileExplorer';
import LoadingSpinner from './LoadingSpinner';
import Modal from './Modal';
import { Folder, FolderOpen, File, FileText, Settings, ArrowUp, X } from 'lucide-react';

interface FileExplorerProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectPath: (path: string) => void;
  initialPath?: string;
  title?: string;
  showFiles?: boolean; // Whether to show files or only directories
}

const FileExplorer: React.FC<FileExplorerProps> = ({
  isOpen,
  onClose,
  onSelectPath,
  initialPath = '/',
  title = 'Select Directory',
  showFiles = false
}) => {
  const [currentListing, setCurrentListing] = useState<DirectoryListing | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedPath, setSelectedPath] = useState<string>(initialPath);

  useEffect(() => {
    if (isOpen) {
      navigateToPath(initialPath);
    }
  }, [isOpen, initialPath]);

  const navigateToPath = async (path: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const listing = await fileExplorerService.browseDirectory(path);
      setCurrentListing(listing);
      setSelectedPath(path);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to browse directory');
    } finally {
      setLoading(false);
    }
  };

  const handleItemClick = (item: DirectoryItem) => {
    if (item.is_directory) {
      navigateToPath(item.path);
    } else if (showFiles) {
      setSelectedPath(item.path);
    }
  };

  const handleBreadcrumbClick = (path: string) => {
    navigateToPath(path);
  };

  const handleSelectPath = () => {
    onSelectPath(selectedPath);
    onClose();
  };

  const getItemIcon = (item: DirectoryItem) => {
    if (item.is_directory) {
      if (item.has_terraform) {
        return <FolderOpen className="w-5 h-5 text-blue-600" />; // Folder with terraform files
      }
      return <Folder className="w-5 h-5 text-gray-600" />;
    }
    
    if (item.name.endsWith('.tf')) {
      return <FileText className="w-5 h-5 text-purple-600" />;
    }
    if (item.name.endsWith('.yaml') || item.name.endsWith('.yml')) {
      return <Settings className="w-5 h-5 text-orange-600" />;
    }
    if (item.name.endsWith('.json')) {
      return <File className="w-5 h-5 text-yellow-600" />;
    }
    return <File className="w-5 h-5 text-gray-600" />;
  };

  const breadcrumbs = currentListing ? fileExplorerService.getPathBreadcrumbs(currentListing.current_path) : [];

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={title}>
      <div className="w-full max-w-4xl max-h-[70vh] flex flex-col">
        {/* Breadcrumb Navigation */}
        <div className="flex items-center space-x-1 p-3 bg-gray-50 border-b text-sm">
          {breadcrumbs.map((crumb, index) => (
            <React.Fragment key={crumb.path}>
              {index > 0 && <span className="text-gray-400">/</span>}
              <button
                onClick={() => handleBreadcrumbClick(crumb.path)}
                className={`px-2 py-1 rounded hover:bg-gray-200 transition-colors ${
                  crumb.path === selectedPath ? 'bg-blue-100 text-blue-700 font-medium' : 'text-gray-700'
                }`}
              >
                {crumb.name}
              </button>
            </React.Fragment>
          ))}
        </div>

        {/* Current Path Display */}
        <div className="p-3 bg-blue-50 border-b">
          <div className="flex items-center justify-between">
            <div className="text-sm text-blue-700">
              <strong>Current path:</strong> <span className="font-mono">{selectedPath}</span>
            </div>
            {currentListing && !currentListing.is_root && (
              <button
                onClick={() => navigateToPath(currentListing.parent_path || '/')}
                className="px-2 py-1 text-sm bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition-colors flex items-center"
              >
                <ArrowUp className="w-3 h-3 mr-1" />
                Up
              </button>
            )}
          </div>
        </div>

        {/* Directory Contents */}
        <div className="flex-1 overflow-y-auto">
          {loading && (
            <div className="flex items-center justify-center p-8">
              <LoadingSpinner />
            </div>
          )}
          
          {error && (
            <div className="p-4 bg-red-50 border border-red-200 m-4 rounded">
              <div className="text-red-700 flex items-center">
                <X className="w-4 h-4 mr-2" />
                Error: {error}
              </div>
              <button
                onClick={() => navigateToPath('/')}
                className="mt-2 px-3 py-1 bg-red-600 text-white rounded text-sm hover:bg-red-700"
              >
                Go to Root
              </button>
            </div>
          )}

          {currentListing && !loading && !error && (
            <div className="space-y-1">
              {currentListing.items.length === 0 ? (
                <div className="text-center py-8 text-gray-500 flex flex-col items-center">
                  <Folder className="w-12 h-12 text-gray-400 mb-2" />
                  Empty directory
                </div>
              ) : (
                currentListing.items
                  .filter(item => showFiles || item.is_directory)
                  .map((item) => (
                    <div
                      key={item.path}
                      onClick={() => handleItemClick(item)}
                      className={`flex items-center space-x-3 p-3 hover:bg-gray-50 cursor-pointer transition-colors border-l-4 ${
                        item.path === selectedPath 
                          ? 'bg-blue-50 border-blue-500' 
                          : 'border-transparent'
                      }`}
                    >
                      <div className="flex-shrink-0">{getItemIcon(item)}</div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center space-x-2">
                          <span className={`font-medium ${
                            item.is_directory ? 'text-blue-700' : 'text-gray-700'
                          }`}>
                            {item.name}
                          </span>
                          {item.has_terraform && (
                            <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                              <Settings className="w-3 h-3 mr-1" />
                              Terraform
                            </span>
                          )}
                        </div>
                        <div className="text-xs text-gray-500 mt-1">
                          {item.is_directory ? 'Directory' : `File • ${fileExplorerService.formatFileSize(item.size || 0)}`}
                          {item.mod_time && ` • Modified: ${item.mod_time}`}
                        </div>
                      </div>
                      {item.is_directory && (
                        <span className="text-gray-400">→</span>
                      )}
                    </div>
                  ))
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 bg-gray-50 border-t flex justify-between items-center">
          <div className="text-sm text-gray-600">
            {currentListing && (
              <span className="flex items-center space-x-4">
                <span className="flex items-center">
                  <Folder className="w-3 h-3 mr-1" />
                  {currentListing.items.filter(i => i.is_directory).length} directories
                </span>
                <span className="flex items-center">
                  <File className="w-3 h-3 mr-1" />
                  {currentListing.items.filter(i => !i.is_directory).length} files
                </span>
              </span>
            )}
          </div>
          <div className="flex space-x-2">
            <button
              onClick={onClose}
              className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
            >
              Cancel
            </button>
            <button
              onClick={handleSelectPath}
              disabled={!selectedPath || selectedPath === '/'}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Select Path
            </button>
          </div>
        </div>
      </div>
    </Modal>
  );
};

export default FileExplorer;