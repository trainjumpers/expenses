import {
  getStatement,
  listStatements,
  previewStatement,
  uploadStatement,
} from "@/lib/api/statement";
import type { PaginatedStatementResponse } from "@/lib/api/statement";
import type {
  CreateStatementRequest,
  StatementPreviewResponse,
} from "@/lib/models/statement";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

const STATEMENT_KEYS = {
  all: ["statements"] as const,
  lists: () => [...STATEMENT_KEYS.all, "list"] as const,
  list: (filters: Record<string, unknown>) =>
    [...STATEMENT_KEYS.lists(), filters] as const,
  details: () => [...STATEMENT_KEYS.all, "detail"] as const,
  detail: (id: number) => [...STATEMENT_KEYS.details(), id] as const,
};

export const useStatements = (page: number = 1, pageSize: number = 10) => {
  return useQuery<PaginatedStatementResponse>({
    queryKey: ["statements", page, pageSize],
    queryFn: ({ signal }) =>
      listStatements(signal, { page, page_size: pageSize }),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useStatement = (id: number) => {
  return useQuery({
    queryKey: STATEMENT_KEYS.detail(id),
    queryFn: () => getStatement(id),
    enabled: !!id,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useUploadStatement = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateStatementRequest) => uploadStatement(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: STATEMENT_KEYS.all });
      queryClient.invalidateQueries({ queryKey: ["transactions"] });

      toast.success(
        "Statement uploaded successfully! Processing will begin shortly."
      );
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to upload statement");
    },
  });
};

export const usePreviewStatement = () => {
  return useMutation<
    StatementPreviewResponse,
    Error,
    { file: File; skipRows: number; rowSize: number }
  >({
    mutationFn: ({ file, skipRows, rowSize }) =>
      previewStatement(file, skipRows, rowSize),
    onError: (error: Error) => {
      toast.error(error.message || "Failed to preview statement");
    },
  });
};
