"use client";

import { getUser } from "@/lib/api/user";
import { User } from "@/lib/models/user";
import { getCookie } from "@/lib/utils/cookies";
import { createResource } from "@/lib/utils/suspense";
import { useRouter } from "next/navigation";
import React, { ReactNode, createContext, useContext, useState } from "react";

type UserResource = {
  read: () => User;
};

const UserContext = createContext<UserResource | null>(null);

export const UserProvider = ({ children }: { children: ReactNode }) => {
  const router = useRouter();
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
      };
    }
    const userResource = createResource<User>(getUser);
    return {
      read: () => {
        const user = userResource.read();
        if (!user) throw new Error("User not found");
        return user;
      },
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
  return resource.read();
}
