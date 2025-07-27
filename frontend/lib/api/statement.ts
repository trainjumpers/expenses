import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import {
  CreateStatementRequest,
  Statement,
  StatementPreviewResponse,
  StatementUploadResponse,
} from "@/lib/models/statement";

export async function uploadStatement(
  data: CreateStatementRequest
): Promise<StatementUploadResponse> {
  const formData = new FormData();
  formData.append("account_id", data.account_id.toString());
  formData.append("file", data.file);

  if (data.bank_type) {
    formData.append("bank_type", data.bank_type);
  }
  if (data.metadata) {
    formData.append("metadata", data.metadata);
  }

  const response = await fetch(`${API_BASE_URL}/statement`, {
    method: "POST",
    credentials: "include",
    body: formData,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || "Failed to upload statement");
  }

  const result = await response.json();
  return {
    statement: result.data,
    message: result.message || "Statement uploaded successfully",
  };
}

export interface PaginatedStatementResponse {
  statements: Statement[];
  total: number;
  page: number;
  page_size: number;
}

export async function listStatements(
  signal?: AbortSignal,
  params?: { page?: number; page_size?: number }
): Promise<PaginatedStatementResponse> {
  const query = [];
  if (params?.page) query.push(`page=${params.page}`);
  if (params?.page_size) query.push(`page_size=${params.page_size}`);
  const queryString = query.length > 0 ? `?${query.join("&")}` : "";
  return apiRequest<PaginatedStatementResponse>(
    `${API_BASE_URL}/statement${queryString}`,
    {
      credentials: "include",
      signal,
    },
    "statement",
    [],
    "Failed to fetch statements"
  );
}

export async function getStatement(id: number): Promise<Statement> {
  return apiRequest<Statement>(
    `${API_BASE_URL}/statement/${id}`,
    {
      credentials: "include",
    },
    "statement",
    [],
    "Failed to fetch statement"
  );
}

export async function previewStatement(
  file: File,
  skipRows: number,
  rowSize: number
): Promise<StatementPreviewResponse> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append(
    "metadata",
    JSON.stringify({ skip_rows: skipRows, row_size: rowSize })
  );

  const response = await fetch(`${API_BASE_URL}/statement/preview`, {
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
