import { API_BASE_URL } from "@/lib/constants/api";
import { ACCESS_TOKEN_NAME } from "@/lib/constants/cookie";
import { User } from "@/lib/models/user";
import { getCookie } from "@/lib/utils/cookies";
import { handleApiError } from "@/lib/utils/toast";
import { toast } from "sonner";

export async function getUser(): Promise<User> {
  let toastShown = false;
  try {
    const response = await fetch(`${API_BASE_URL}/user`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${getCookie(ACCESS_TOKEN_NAME)}`,
      },
    });
    const data = await response.json();
    if (!response.ok) {
      toastShown = true;
      handleApiError(response.status, "user");
      throw new Error(data.error || "Failed to get user");
    }
    return data.data;
  } catch (err) {
    if (!toastShown) toast.error("Something went wrong. Please try again.");
    throw err;
  }
}

export async function updateUser(user: Partial<User>): Promise<User> {
  let toastShown = false;
  try {
    const response = await fetch(`${API_BASE_URL}/user`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${getCookie(ACCESS_TOKEN_NAME)}`,
      },
      body: JSON.stringify(user),
    });
    const data = await response.json();
    if (!response.ok) {
      toastShown = true;
      handleApiError(response.status, "user");
      throw new Error(data.error || "Failed to update user");
    }
    return data.data;
  } catch (err) {
    if (!toastShown) toast.error("Something went wrong. Please try again.");
    throw err;
  }
}

export async function updatePassword(
  currentPassword: string,
  newPassword: string
): Promise<User> {
  let toastShown = false;
  try {
    const response = await fetch(`${API_BASE_URL}/user/password`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${getCookie(ACCESS_TOKEN_NAME)}`,
      },
      body: JSON.stringify({
        old_password: currentPassword,
        new_password: newPassword,
      }),
    });
    const data = await response.json();
    if (!response.ok) {
      toastShown = true;
      if (response.status === 401) {
        toast.error("Current password is incorrect");
        throw new Error("Current password is incorrect");
      }
      handleApiError(response.status, "password");
      throw new Error(data.error || "Change password failed");
    }
    return data.data;
  } catch (err) {
    if (!toastShown) toast.error("Something went wrong. Please try again.");
    throw err;
  }
}
