-- Tag Standards table to store tag standardization configurations
CREATE TABLE tag_standards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    cloud_provider TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    content TEXT NOT NULL, -- YAML content
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Operations table to track validation and tagging operations
CREATE TABLE operations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('validation', 'tagging')),
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    standard_id INTEGER REFERENCES tag_standards(id),
    directory_path TEXT NOT NULL,
    filter_pattern TEXT,
    skip_pattern TEXT,
    settings TEXT, -- JSON settings
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME
);

-- Operation Results table to store detailed results
CREATE TABLE operation_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operation_id INTEGER NOT NULL REFERENCES operations(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    resource_type TEXT,
    resource_name TEXT,
    action TEXT NOT NULL, -- 'tagged', 'validated', 'skipped', 'error'
    violation_type TEXT, -- For validation: 'missing_tag', 'invalid_value', etc.
    details TEXT, -- JSON with detailed information
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Operation Logs table for storing operation logs
CREATE TABLE operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operation_id INTEGER NOT NULL REFERENCES operations(id) ON DELETE CASCADE,
    level TEXT NOT NULL CHECK (level IN ('info', 'warn', 'error', 'debug')),
    message TEXT NOT NULL,
    details TEXT, -- JSON with additional details
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_tag_standards_name ON tag_standards(name);
CREATE INDEX idx_tag_standards_cloud_provider ON tag_standards(cloud_provider);
CREATE INDEX idx_operations_type ON operations(type);
CREATE INDEX idx_operations_status ON operations(status);
CREATE INDEX idx_operations_created_at ON operations(created_at);
CREATE INDEX idx_operation_results_operation_id ON operation_results(operation_id);
CREATE INDEX idx_operation_results_file_path ON operation_results(file_path);
CREATE INDEX idx_operation_logs_operation_id ON operation_logs(operation_id);
CREATE INDEX idx_operation_logs_level ON operation_logs(level);

-- Create triggers to update the updated_at timestamp
CREATE TRIGGER update_tag_standards_updated_at
    AFTER UPDATE ON tag_standards
    FOR EACH ROW
BEGIN
    UPDATE tag_standards SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_operations_updated_at
    AFTER UPDATE ON operations
    FOR EACH ROW
BEGIN
    UPDATE operations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;