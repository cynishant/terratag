-- Add line number and snippet fields to operation_results table
ALTER TABLE operation_results ADD COLUMN line_number INTEGER;
ALTER TABLE operation_results ADD COLUMN snippet TEXT;

-- Create index for line number for better performance when filtering by location
CREATE INDEX idx_operation_results_line_number ON operation_results(line_number);