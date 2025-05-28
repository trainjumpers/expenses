import { API_BASE_URL } from "@/lib/constants/api";
import { AuthResponse } from "@/lib/models/auth";
import { handleApiError } from "@/lib/utils/toast";
import { toast } from "sonner";

export async function login(
  email: string,
  password: string
): Promise<AuthResponse> {
  const response = await fetch(`${API_BASE_URL}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  const data = await response.json();
  if (!response.ok) {
    if (response.status === 401) {
      toast.error("The email or password is incorrect");
    } else {
      handleApiError(response.status, "user");
    }
    throw new Error(data.error || "Login failed");
  }
  return data.data;
}

export async function signup(
  name: string,
  email: string,
  password: string
): Promise<AuthResponse> {
  const response = await fetch(`${API_BASE_URL}/signup`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, email, password }),
  });
  const data = await response.json();
  if (!response.ok) {
    if (response.status === 409) {
      toast.info("Account already exists. Please login", {
        action: {
          label: "Login",
          onClick: () => {
            window.location.href = "/login";
          },
        },
      });
    } else {
      handleApiError(response.status, "user");
    }
    throw new Error(data.error || "Signup failed");
  }
  return data.data;
}
