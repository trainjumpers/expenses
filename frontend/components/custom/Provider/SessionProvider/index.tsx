"use client";

import {
  login as loginApi,
  logout as logoutApi,
  refresh as refreshApi,
} from "@/lib/api/auth";
import { checkUser } from "@/lib/api/user";
import { usePathname } from "next/navigation";
import React, {
  ReactNode,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";

type SessionContextType = {
  isTokenAvailable: boolean;
  login: (
    email: string,
    password: string
  ) => Promise<import("@/lib/models/user").User>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
  loading: boolean;
  error: string | null;
};

const SessionContext = createContext<SessionContextType | null>(null);

export const SessionProvider = ({ children }: { children: ReactNode }) => {
  const [isTokenAvailable, setIsTokenAvailable] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const pathname = usePathname();

  // Checks session validity and tries refresh if needed
  const refreshSession = useCallback(async () => {
    console.log(window, window.location.href);
    if (pathname === "/login" || pathname === "/signup") {
      setIsTokenAvailable(false);
      console.log("Here");
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const response = await checkUser();
      if (response && response.ok) {
        setIsTokenAvailable(true);
      } else if (response && response.status === 401) {
        // Try refresh
        try {
          const res = await refreshApi();
          if (res.ok) {
            // After refresh, check user again
            const userCheck = await checkUser();
            if (userCheck && userCheck.ok) {
              setIsTokenAvailable(true);
            } else {
              setIsTokenAvailable(false);
              window.location.href = "/login";
            }
          } else {
            setIsTokenAvailable(false);
            window.location.href = "/login";
          }
        } catch {
          setIsTokenAvailable(false);
          window.location.href = "/login";
        }
      } else {
        setError("Failed to validate session");
        setIsTokenAvailable(false);
      }
    } catch {
      setError("Failed to validate session");
      setIsTokenAvailable(false);
    } finally {
      setLoading(false);
    }
  }, [pathname]);

  useEffect(() => {
    refreshSession();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const login = async (
    email: string,
    password: string
  ): Promise<import("@/lib/models/user").User> => {
    setLoading(true);
    setError(null);
    try {
      const authResponse = await loginApi(email, password);
      setIsTokenAvailable(true);
      return authResponse.user;
    } catch (err: unknown) {
      setIsTokenAvailable(false);
      setError("Login failed");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    await logoutApi();
    setIsTokenAvailable(false);
    window.location.href = "/login";
  };

  return (
    <SessionContext.Provider
      value={{
        isTokenAvailable,
        login,
        logout,
        refreshSession,
        loading,
        error,
      }}
    >
      {children}
    </SessionContext.Provider>
  );
};

export function useSession() {
  const ctx = useContext(SessionContext);
  if (!ctx) throw new Error("useSession must be used within a SessionProvider");
  return ctx;
}
