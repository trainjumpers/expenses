"use client";

import { useSession } from "@/components/custom/Provider/SessionProvider";
import { logout as logoutApi, signup as signupApi } from "@/lib/api/auth";
import {
  getUser,
  updatePassword as updatePasswordApi,
  updateUser,
} from "@/lib/api/user";
import { User } from "@/lib/models/user";
import { createResource } from "@/lib/utils/suspense";
import React, {
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from "react";

type UserResource = {
  read: () => User;
  logout: () => void;
  update: (user: Partial<User>) => Promise<User>;
  updatePassword: (
    currentPassword: string,
    newPassword: string
  ) => Promise<User>;
  signup: (name: string, email: string, password: string) => Promise<User>;
};

const UserContext = createContext<UserResource | null>(null);

export const UserProvider = ({ children }: { children: ReactNode }) => {
  const { isTokenAvailable } = useSession();

  const logout = async () => {
    await logoutApi();
    window.location.href = "/login";
  };

  const update = async (user: Partial<User>) => {
    const updatedUser = await updateUser(user);
    setUserResource(() => updatedUser);
    return updatedUser;
  };

  const updatePassword = async (
    currentPassword: string,
    newPassword: string
  ) => {
    const updatedUser = await updatePasswordApi(currentPassword, newPassword);
    return updatedUser;
  };

  const signup = async (name: string, email: string, password: string) => {
    const authResponse = await signupApi(name, email, password);
    setUserResource(() => authResponse.user);
    return authResponse.user;
  };

  const setUserResource = (func: () => User) => {
    setResource((prev) =>
      prev
        ? {
            ...prev,
            read: func,
          }
        : {
            read: func,
            logout,
            update,
            updatePassword,
            signup,
          }
    );
  };

  const [resource, setResource] = useState<UserResource>(() => ({
    read: () => {
      return { id: 0, name: "", email: "" };
    },
    logout,
    update,
    updatePassword,
    signup,
  }));

  useEffect(() => {
    if (!isTokenAvailable) return;
    const userResource = createResource<User>(getUser);
    setUserResource(() => {
      const user = userResource.read();
      if (!user) throw new Error("User not found");
      return user;
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isTokenAvailable]);

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
