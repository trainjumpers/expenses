import { API_BASE_URL } from "@/lib/constants/api";
import { User } from "@/lib/models/user";
import { getCookie } from "@/lib/utils/cookies";
import { handleApiError } from "@/lib/utils/toast";

export async function getUser(): Promise<User> {
  const response = await fetch(`${API_BASE_URL}/user`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getCookie("access_token")}`,
    },
  });
  const data = await response.json();
  if (!response.ok) {
    handleApiError(response.status, "user");
    throw new Error(data.error || "Failed to get user");
  }
  return data.data;
}

export async function updateUser(user: Partial<User>): Promise<User> {
  const response = await fetch(`${API_BASE_URL}/user`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getCookie("access_token")}`,
    },
    body: JSON.stringify(user),
  });
  const data = await response.json();
  if (!response.ok) {
    handleApiError(response.status, "user");
    throw new Error(data.error || "Failed to update user");
  }
  return data.data;
}

export async function updatePassword(
  currentPassword: string,
  newPassword: string
): Promise<User> {
  const response = await fetch(`${API_BASE_URL}/user/password`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getCookie("access_token")}`,
    },
    body: JSON.stringify({
      old_password: currentPassword,
      new_password: newPassword,
    }),
  });
  const data = await response.json();
  if (!response.ok) {
    handleApiError(response.status, "password");
    throw new Error(data.error || "Change password failed");
  }
  return data.data;
}
