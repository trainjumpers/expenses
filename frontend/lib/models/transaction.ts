export interface Transaction {
  id: number;
  date: string;
  name: string;
  description: string | null;
  amount: number;
  category_ids: number[];
  account_id: number;
}

export interface CreateTransaction {
  name: string;
  description?: string;
  amount: number;
  date: string;
  category_ids: number[];
  account_id: number;
}

export interface PaginatedTransactionsResponse {
  transactions: Transaction[];
  total: number;
  page: number;
  page_size: number;
}

export interface TransactionQueryParams {
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: "asc" | "desc";
  account_id?: number;
  category_id?: number;
  uncategorized?: boolean;
  min_amount?: number;
  max_amount?: number;
  date_from?: string; // YYYY-MM-DD
  date_to?: string; // YYYY-MM-DD
  search?: string;
}
