import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import { User } from "@/lib/models/user";
import { toast } from "sonner";

export async function getUser(): Promise<User> {
  return apiRequest<User>(
    `${API_BASE_URL}/user`,
    {
      method: "GET",
      credentials: "include",
    },
    "user"
  );
}

export async function checkUser(): Promise<Response | null> {
  const response = await fetch(`${API_BASE_URL}/user`, {
    method: "GET",
    credentials: "include",
  });
  return response;
}

export async function updateUser(user: Partial<User>): Promise<User> {
  return apiRequest<User>(
    `${API_BASE_URL}/user`,
    {
      method: "PATCH",
      credentials: "include",
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
      credentials: "include",
      body: JSON.stringify({
        old_password: currentPassword,
        new_password: newPassword,
      }),
    },
    "password",
    [
      (response) => {
        if (response.status === 401) {
          toast.error("Current password is incorrect", {
            id: "password-error",
          });
          return true;
        }
        return false;
      },
    ]
  );
}
