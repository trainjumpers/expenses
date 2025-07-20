export type StatementStatus = "pending" | "processing" | "done" | "error";

export interface Statement {
  id: number;
  account_id: number;
  created_by: number;
  original_filename: string;
  file_type: string;
  status: StatementStatus;
  message?: string;
  created_at: string;
}

export interface CreateStatementRequest {
  account_id: number;
  file: File;
  metadata?: StatementImportMetadata;
}

export interface StatementImportMetadata {
  skip_rows?: number;
  mappings?: ColumnMapping[];
}

export interface StatementUploadResponse {
  statement: Statement;
  message: string;
}
// Column mapping interface for unified import
export interface ColumnMapping {
  source_column: string;
  target_field: 'name' | 'amount' | 'description' | 'date' | 'credit' | 'debit';
}
