"use client";

import { getUser, updateUser, updatePassword as updatePasswordApi } from "@/lib/api/user";
import { User } from "@/lib/models/user";
import { deleteCookie, getCookie } from "@/lib/utils/cookies";
import { createResource } from "@/lib/utils/suspense";
import { useRouter } from "next/navigation";
import React, { ReactNode, createContext, useContext, useState } from "react";
import { ACCESS_TOKEN_NAME, REFRESH_TOKEN_NAME } from "../constants/cookie";

type UserResource = {
  read: () => User;
  logout: () => void;
  update: (user: Partial<User>) => Promise<User>;
  updatePassword: (currentPassword: string, newPassword: string) => Promise<User>;
};

const UserContext = createContext<UserResource | null>(null);

export const UserProvider = ({ children }: { children: ReactNode }) => {
  const router = useRouter();

  const logout = () => {
    deleteCookie(ACCESS_TOKEN_NAME);
    deleteCookie(REFRESH_TOKEN_NAME);
    router.replace("/login");
  };

  const update = async (user: Partial<User>) => {
    const updatedUser = await updateUser(user);
    return updatedUser;
  };

  const updatePassword = async (currentPassword: string, newPassword: string) => {
    const updatedUser = await updatePasswordApi(currentPassword, newPassword);
    return updatedUser;
  };

  const [resource] = useState(() => {
    const token = getCookie("access_token");
    if (!token) {
      // Redirect to login if no token
      if (typeof window !== "undefined") {
        router.replace("/login");
      }
      return {
        read: () => {
          throw new Error("No access token. Redirecting to login.");
        },
        logout,
        update,
        updatePassword,
      };
    }
    const userResource = createResource<User>(getUser);
    return {
      read: () => {
        const user = userResource.read();
        if (!user) throw new Error("User not found");
        return user;
      },
      logout,
      update,
      updatePassword,
    };
  });

  return (
    <UserContext.Provider value={resource}>{children}</UserContext.Provider>
  );
};

export function useUser() {
  const resource = useContext(UserContext);
  if (!resource) {
    throw new Error("useUser must be used within a UserProvider");
  }
  return resource;
}
