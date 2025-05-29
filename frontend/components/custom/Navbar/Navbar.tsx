"use client";

import { ProfileDropdown } from "@/components/custom/Navbar/ProfileDropdown";
import { ToggleTheme } from "@/components/custom/Navbar/ToggleTheme";
import { ACCESS_TOKEN_NAME } from "@/lib/constants/cookie";
import { getCookie } from "@/lib/utils/cookies";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

export function Navbar() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const router = useRouter();

  useEffect(() => {
    const accessToken = getCookie(ACCESS_TOKEN_NAME);
    if (!accessToken) {
      router.push("/login");
    }
    setIsLoggedIn(!!accessToken);
  }, [router]);

  if (!isLoggedIn) {
    return null;
  }

  return (
    <nav className="border-b bg-background">
      <div className="flex h-16 mx-10 justify-between items-center">
        <div className="flex">
          <h1 className="text-xl font-bold">NeuroSpend</h1>
        </div>
        <div className="flex items-center gap-2">
          <ToggleTheme />
          <ProfileDropdown />
        </div>
      </div>
    </nav>
  );
}
