import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import { ACCESS_TOKEN_NAME } from "@/lib/constants/cookie";
import { User } from "@/lib/models/user";
import { getCookie } from "@/lib/utils/cookies";
import { toast } from "sonner";

export async function getUser(): Promise<User> {
  return apiRequest<User>(
    `${API_BASE_URL}/user`,
    {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${getCookie(ACCESS_TOKEN_NAME)}`,
      },
    },
    "user"
  );
}

export async function updateUser(user: Partial<User>): Promise<User> {
  return apiRequest<User>(
    `${API_BASE_URL}/user`,
    {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${getCookie(ACCESS_TOKEN_NAME)}`,
      },
      body: JSON.stringify(user),
    },
    "user"
  );
}

export async function updatePassword(
  currentPassword: string,
  newPassword: string
): Promise<User> {
  return apiRequest<User>(
    `${API_BASE_URL}/user/password`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${getCookie(ACCESS_TOKEN_NAME)}`,
      },
      body: JSON.stringify({
        old_password: currentPassword,
        new_password: newPassword,
      }),
    },
    "password",
    [
      (response) => {
        if (response.status === 401) {
          toast.error("Current password is incorrect");
          return true;
        }
        return false;
      },
    ]
  );
}
