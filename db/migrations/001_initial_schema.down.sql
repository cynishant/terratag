-- Drop triggers
DROP TRIGGER IF EXISTS update_operations_updated_at;
DROP TRIGGER IF EXISTS update_tag_standards_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_operation_logs_level;
DROP INDEX IF EXISTS idx_operation_logs_operation_id;
DROP INDEX IF EXISTS idx_operation_results_file_path;
DROP INDEX IF EXISTS idx_operation_results_operation_id;
DROP INDEX IF EXISTS idx_operations_created_at;
DROP INDEX IF EXISTS idx_operations_status;
DROP INDEX IF EXISTS idx_operations_type;
DROP INDEX IF EXISTS idx_tag_standards_cloud_provider;
DROP INDEX IF EXISTS idx_tag_standards_name;

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS operation_logs;
DROP TABLE IF EXISTS operation_results;
DROP TABLE IF EXISTS operations;
DROP TABLE IF EXISTS tag_standards;