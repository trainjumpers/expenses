import { API_BASE_URL } from "@/lib/constants/api";

export interface StatementPreview {
  headers: string[];
  sample_data: string[][];
  total_rows: number;
  file_type: string;
}

export interface ColumnMapping {
  date_column: number;
  description_column: number;
  amount_column: number;
  reference_column: number;
  date_format?: string;
}

export interface ParsedTransaction {
  date: string;
  description: string;
  amount: number;
  reference?: string;
}

export interface ParsedStatementResult {
  transactions: ParsedTransaction[];
  total_rows: number;
  successful_rows: number;
  failed_rows: number;
  errors?: string[];
}

export interface PreviewStatementRequest {
  account_id: number;
  file: File;
}

export interface PreviewStatementResponse {
  statement_id: number;
  preview: StatementPreview;
  account_id: number;
  filename: string;
}

export interface ParseStatementRequest {
  account_id: number;
  file: File;
  has_headers: boolean;
  column_mapping: ColumnMapping;
}

export interface ParseStatementResponse {
  statement: Record<string, unknown>; // Statement object
  parse_result: ParsedStatementResult;
}

export async function previewStatement(data: PreviewStatementRequest): Promise<PreviewStatementResponse> {
  const formData = new FormData();
  formData.append("account_id", data.account_id.toString());
  formData.append("file", data.file);

  const response = await fetch(`${API_BASE_URL}/statements/preview`, {
    method: "POST",
    credentials: "include",
    body: formData,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || "Failed to preview statement");
  }

  const result = await response.json();
  return result.data;
}

export async function parseStatementDirect(data: ParseStatementRequest): Promise<ParseStatementResponse> {
  const formData = new FormData();
  formData.append("account_id", data.account_id.toString());
  formData.append("file", data.file);
  formData.append("has_headers", data.has_headers.toString());
  
  // Add column mapping fields
  formData.append("date_column", data.column_mapping.date_column.toString());
  formData.append("description_column", data.column_mapping.description_column.toString());
  formData.append("amount_column", data.column_mapping.amount_column.toString());
  formData.append("reference_column", data.column_mapping.reference_column.toString());
  
  if (data.column_mapping.date_format) {
    formData.append("date_format", data.column_mapping.date_format);
  }

  const response = await fetch(`${API_BASE_URL}/statements/parse-direct`, {
    method: "POST",
    credentials: "include",
    body: formData,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || "Failed to parse statement");
  }

  const result = await response.json();
  return result.data;
}
