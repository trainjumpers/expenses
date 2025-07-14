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
}

export interface StatementUploadResponse {
  statement: Statement;
  message: string;
}
