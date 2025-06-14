import { apiRequest, authHeaders } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import { Account, CreateAccountInput } from "@/lib/models/account";

export async function listAccounts(signal?: AbortSignal): Promise<Account[]> {
  return apiRequest<Account[]>(
    `${API_BASE_URL}/account`,
    {
      headers: authHeaders(),
      credentials: "include",
      signal,
    },
    "account",
    [],
    "Failed to fetch accounts"
  );
}

export async function getAccount(id: number): Promise<Account> {
  return apiRequest<Account>(
    `${API_BASE_URL}/account/${id}`,
    {
      headers: authHeaders(),
      credentials: "include",
    },
    "account",
    [],
    "Failed to fetch account"
  );
}

export async function createAccount(
  input: CreateAccountInput
): Promise<Account> {
  return apiRequest<Account>(
    `${API_BASE_URL}/account`,
    {
      method: "POST",
      headers: authHeaders(),
      credentials: "include",
      body: JSON.stringify(input),
    },
    "account",
    [],
    "Failed to create account"
  );
}

export async function updateAccount(
  id: number,
  input: Partial<CreateAccountInput>
): Promise<Account> {
  return apiRequest<Account>(
    `${API_BASE_URL}/account/${id}`,
    {
      method: "PATCH",
      headers: authHeaders(),
      credentials: "include",
      body: JSON.stringify(input),
    },
    "account",
    [],
    "Failed to update account"
  );
}

export async function deleteAccount(id: number): Promise<void> {
  return apiRequest<void>(
    `${API_BASE_URL}/account/${id}`,
    {
      method: "DELETE",
      headers: authHeaders(),
      credentials: "include",
    },
    "account",
    [],
    "Failed to delete account"
  );
}
