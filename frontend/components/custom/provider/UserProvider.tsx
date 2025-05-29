"use client";

import { login as loginApi, signup as signupApi } from "@/lib/api/auth";
import {
  getUser,
  updatePassword as updatePasswordApi,
  updateUser,
} from "@/lib/api/user";
import {
  ACCESS_TOKEN_EXPIRY,
  ACCESS_TOKEN_NAME,
  REFRESH_TOKEN_EXPIRY,
  REFRESH_TOKEN_NAME,
} from "@/lib/constants/cookie";
import { User } from "@/lib/models/user";
import { deleteCookie, getCookie, setCookie } from "@/lib/utils/cookies";
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
  login: (email: string, password: string) => Promise<User>;
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
  const [token, setToken] = useState<string | undefined>(() =>
    getCookie(ACCESS_TOKEN_NAME)
  );
  const [resource, setResource] = useState<UserResource | null>(null);
  const [loading, setLoading] = useState(true);

  const login = async (email: string, password: string) => {
    const authResponse = await loginApi(email, password);
    setCookie(ACCESS_TOKEN_NAME, authResponse.access_token, {
      maxAge: ACCESS_TOKEN_EXPIRY,
    });
    setCookie(REFRESH_TOKEN_NAME, authResponse.refresh_token, {
      maxAge: REFRESH_TOKEN_EXPIRY,
    });
    setToken(authResponse.access_token);
    const user = await getUser();
    return user;
  };

  const logout = () => {
    deleteCookie(ACCESS_TOKEN_NAME);
    deleteCookie(REFRESH_TOKEN_NAME);
    setToken(undefined);
    window.location.href = "/login";
  };

  const update = async (user: Partial<User>) => {
    const updatedUser = await updateUser(user);
    setResource((prev) =>
      prev
        ? {
            ...prev,
            read: () => updatedUser,
          }
        : prev
    );
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
    setCookie(ACCESS_TOKEN_NAME, authResponse.access_token, {
      maxAge: ACCESS_TOKEN_EXPIRY,
    });
    setCookie(REFRESH_TOKEN_NAME, authResponse.refresh_token, {
      maxAge: REFRESH_TOKEN_EXPIRY,
    });
    setToken(authResponse.access_token);
    const user = await getUser();
    return user;
  };

  useEffect(() => {
    setLoading(true);
    if (!token) {
      if (
        typeof window !== "undefined" &&
        window.location.pathname !== "/login"
      ) {
        window.location.href = "/login";
        setResource({
          read: () => ({ id: 0, name: "", email: "" }),
          login,
          logout,
          update,
          updatePassword,
          signup,
        });
        setLoading(false);
        return;
      }
      setResource({
        read: () => {
          throw new Error("No access token. Redirecting to login.");
        },
        login,
        logout,
        update,
        updatePassword,
        signup,
      });
      setLoading(false);
      return;
    }
    const userResource = createResource<User>(getUser);
    setResource({
      read: () => {
        const user = userResource.read();
        if (!user) throw new Error("User not found");
        return user;
      },
      login,
      logout,
      update,
      updatePassword,
      signup,
    });
    setLoading(false);
  }, [token]);

  useEffect(() => {
    const interval = setInterval(() => {
      const currentToken = getCookie(ACCESS_TOKEN_NAME);
      if (currentToken !== token) {
        setToken(currentToken);
      }
    }, 10000);
    return () => clearInterval(interval);
  }, [token]);

  if (loading || !resource) {
    return null; // or a loading spinner
  }

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
