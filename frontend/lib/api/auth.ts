import { API_BASE_URL } from "@/lib/constants/api";
import { AuthResponse } from "@/lib/models/auth";
import { toast } from "sonner";

import { handleApiError } from "../utils/toast";

/**
 * Authenticates a user with the provided email and password.
 *
 * Sends a POST request to the `/login` endpoint and returns authentication data on success.
 *
 * @param email - The user's email address.
 * @param password - The user's password.
 * @returns The authentication response data.
 *
 * @throws {Error} If authentication fails, with the error message from the API response or "Login failed".
 */
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
  return data;
}

/**
 * Registers a new user account with the provided name, email, and password.
 *
 * Sends a POST request to the `/signup` endpoint and returns the authentication response data on success.
 *
 * @param name - The user's display name.
 * @param email - The user's email address.
 * @param password - The user's chosen password.
 * @returns The authentication response data for the newly created account.
 *
 * @throws {Error} If the signup fails, including when the account already exists or other API errors occur.
 * @remark If the account already exists (HTTP 409), an informational toast is shown with an option to redirect to the login page.
 */
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
  return data;
}
