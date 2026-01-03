import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import type { Account, CreateAccountInput } from "@/lib/models/account";
import { toast } from "sonner";

export async function listAccounts(signal?: AbortSignal): Promise<Account[]> {
  return apiRequest<Account[]>(
    `${API_BASE_URL}/account`,
    {
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
      credentials: "include",
      body: JSON.stringify(input),
    },
    "account",
    [],
    "Failed to update account"
  );
}

export async function deleteAccount(id: number): Promise<void> {
  const customErrorHandlers = [
    (response: Response, data: unknown) => {
      if (response.status === 409) {
        const errorData = data as { message?: string };
        if (errorData.message?.includes("existing transactions")) {
          toast.error("Cannot delete account with existing transactions", {
            duration: 2000,
          });
          return true;
        }
      }
      return false;
    },
  ];

  return apiRequest<void>(
    `${API_BASE_URL}/account/${id}`,
    {
      method: "DELETE",
      credentials: "include",
    },
    "account",
    customErrorHandlers,
    "Failed to delete account"
  );
}
