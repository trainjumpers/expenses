"use client";

import { ProfileDropdown } from "@/components/custom/Navbar/ProfileDropdown";
import { ToggleTheme } from "@/components/custom/Navbar/ToggleTheme";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { usePathname } from "next/navigation";

export function Navbar() {
  const pathname = usePathname();

  return (
    <nav className="border-b bg-background">
      <div className="flex h-16 mx-10 justify-between items-center">
        <div className="flex items-center">
          <h1 className="text-xl font-bold">NeuroSpend</h1>
        </div>
        <div className="flex-1 flex justify-center">
          <div className="flex gap-4">
            <Button
              asChild
              variant="ghost"
              className={
                `border-1 border-transparent hover:border-primary ` +
                (pathname === "/"
                  ? "border-primary bg-primary text-primary-foreground font-bold shadow-md"
                  : "border-border bg-background text-foreground")
              }
            >
              <Link href="/">Dashboard</Link>
            </Button>
            <Button
              asChild
              variant="ghost"
              className={
                `border-1 border-transparent hover:border-primary ` +
                (pathname.startsWith("/transaction")
                  ? "border-primary bg-primary text-primary-foreground font-bold shadow-md"
                  : "border-border bg-background text-foreground")
              }
            >
              <Link href="/transaction">Transactions</Link>
            </Button>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <ToggleTheme />
          <ProfileDropdown />
        </div>
      </div>
    </nav>
  );
}
