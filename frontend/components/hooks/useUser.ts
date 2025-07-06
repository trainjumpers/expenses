"use client";

import {
  login as loginApi,
  logout as logoutApi,
  signup as signupApi,
} from "@/lib/api/auth";
import {
  getUser,
  updatePassword as updatePasswordApi,
  updateUser as updateUserApi,
} from "@/lib/api/user";
import { User } from "@/lib/models/user";
import { queryKeys } from "@/lib/query-client";
import {
  ApiErrorType,
  getErrorMessage,
  getErrorStatus,
  isAuthError,
} from "@/lib/types/errors";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { useSession } from "./useSession";

export function useUser() {
  const { isAuthenticated, isLoading: sessionLoading } = useSession();

  return useQuery({
    queryKey: queryKeys.user,
    queryFn: getUser,
    enabled: isAuthenticated && !sessionLoading,
    staleTime: 10 * 60 * 1000,
    retry: (failureCount, error: ApiErrorType) => {
      if (isAuthError(error)) return false;
      const status = getErrorStatus(error);
      if (status && status >= 400 && status < 500) return false;
      return failureCount < 3;
    },
  });
}

export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      loginApi(email, password),
    onSuccess: (user) => {
      queryClient.setQueryData(queryKeys.user, user);
      queryClient.invalidateQueries({ queryKey: queryKeys.session });
      queryClient.invalidateQueries({ queryKey: queryKeys.user });
    },
    onError: (error: ApiErrorType) => {
      const message = getErrorMessage(error);
      console.error(message || "Login failed. Please try again.");
    },
  });
}

export function useSignup() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      name,
      email,
      password,
    }: {
      name: string;
      email: string;
      password: string;
    }) => signupApi(name, email, password),
    onSuccess: (user) => {
      queryClient.setQueryData(queryKeys.user, user);
      queryClient.invalidateQueries({ queryKey: queryKeys.session });
      queryClient.invalidateQueries({ queryKey: queryKeys.user });
      toast.success("Account created successfully!");
    },
    onError: (error: ApiErrorType) => {
      const message = getErrorMessage(error);
      console.error(message || "Signup failed. Please try again.");
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: logoutApi,
    onSuccess: () => {
      queryClient.clear();
      toast.success("Logged out successfully");
    },
    onError: (error: ApiErrorType) => {
      queryClient.clear();
      const message = getErrorMessage(error);
      console.error("Logout error:", message);
    },
  });
}

export function useUpdateUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userData: Partial<User>) => updateUserApi(userData),
    onSuccess: (updatedUser) => {
      queryClient.setQueryData(queryKeys.user, updatedUser);
      toast.success("Profile updated successfully!");
    },
    onError: (error: ApiErrorType) => {
      const message = getErrorMessage(error);
      console.error(message || "Failed to update profile. Please try again.");
    },
  });
}

export function useUpdatePassword() {
  return useMutation({
    mutationFn: ({
      currentPassword,
      newPassword,
    }: {
      currentPassword: string;
      newPassword: string;
    }) => updatePasswordApi(currentPassword, newPassword),
    onSuccess: () => {
      toast.success("Password updated successfully!");
    },
    onError: (error: ApiErrorType) => {
      const message = getErrorMessage(error);
      console.error(message || "Failed to update password. Please try again.");
    },
  });
}
