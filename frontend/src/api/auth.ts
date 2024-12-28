import { BACKEND_URL } from "@/constants/web";
import type { LoginResponse } from "@/types/user";

export async function login(
  email: string,
  password: string
): Promise<LoginResponse> {
  const response = await fetch(`${BACKEND_URL}/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  if (response.status === 401) {
    throw new Error("Incorrect email or password");
  }

  if (response.status === 404) {
    throw new Error("The email address is not registered");
  }

  if (!response.ok) {
    throw new Error("Something went wrong");
  }

  const data = (await response.json()) as LoginResponse;
  return data;
}

export async function register(
  email: string,
  password: string,
  name: string
): Promise<LoginResponse> {
  const response = await fetch(`${BACKEND_URL}/signup`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password, name }),
  });

  if (response.status === 409) {
    throw new Error("Email already exists. Please login instead.");
  }

  if (!response.ok) {
    throw new Error("Something went wrong");
  }

  const data = (await response.json()) as LoginResponse;
  return data;
}
