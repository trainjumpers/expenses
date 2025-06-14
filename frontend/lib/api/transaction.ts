import { apiRequest, authHeaders } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import {
  CreateTransaction,
  PaginatedTransactionsResponse,
  Transaction,
  TransactionQueryParams,
} from "@/lib/models/transaction";

export async function getAllTransactions(
  params: TransactionQueryParams = {}
): Promise<PaginatedTransactionsResponse> {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      query.append(key, String(value));
    }
  });
  return apiRequest<PaginatedTransactionsResponse>(
    `${API_BASE_URL}/transaction${query.toString() ? `?${query.toString()}` : ""}`,
    {
      method: "GET",
      headers: { "Content-Type": "application/json", ...authHeaders() },
    },
    "transaction",
    [],
    "Failed to get transactions"
  );
}

export async function createTransaction(
  transaction: CreateTransaction
): Promise<Transaction> {
  if (transaction.date.split("T").length === 1) {
    transaction.date = new Date(transaction.date).toISOString();
  }
  return apiRequest<Transaction>(
    `${API_BASE_URL}/transaction`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json", ...authHeaders() },
      body: JSON.stringify(transaction),
    },
    "transaction",
    [],
    "Failed to create transaction"
  );
}

export async function updateTransaction(
  id: number,
  update: Partial<Transaction>
): Promise<Transaction> {
  return apiRequest<Transaction>(
    `${API_BASE_URL}/transaction/${id}`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json", ...authHeaders() },
      body: JSON.stringify(update),
    },
    "transaction",
    [],
    "Failed to update transaction"
  );
}
