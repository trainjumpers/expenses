import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import type { AuthResponse } from "@/lib/models/auth";
import { toast } from "sonner";

export async function login(
  email: string,
  password: string
): Promise<AuthResponse> {
  return apiRequest<AuthResponse>(
    `${API_BASE_URL}/login`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    },
    "user",
    [
      (response) => {
        if (response.status === 401) {
          toast.error("The email or password is incorrect", {
            id: "login-error",
          });
          return true;
        }
        return false;
      },
    ]
  );
}

export async function signup(
  name: string,
  email: string,
  password: string
): Promise<AuthResponse> {
  return apiRequest<AuthResponse>(
    `${API_BASE_URL}/signup`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, email, password }),
    },
    "user",
    [
      (response) => {
        if (response.status === 409) {
          toast.info("Account already exists.", {
            action: {
              label: "Login",
              onClick: () => {
                window.location.href = "/login";
              },
            },
          });
          return true;
        }
        return false;
      },
    ]
  );
}

export async function logout(): Promise<void> {
  try {
    await fetch(`${API_BASE_URL}/logout`, {
      method: "POST",
      credentials: "include",
    });
  } catch (e) {
    console.error("Logout API call failed:", e);
  }
}

export async function refresh(): Promise<Response> {
  return fetch(`${API_BASE_URL}/refresh`, {
    method: "POST",
    credentials: "include",
  });
}
