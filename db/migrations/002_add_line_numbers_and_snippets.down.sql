-- Remove the added columns for line number and snippet
DROP INDEX IF EXISTS idx_operation_results_line_number;
ALTER TABLE operation_results DROP COLUMN snippet;
ALTER TABLE operation_results DROP COLUMN line_number;