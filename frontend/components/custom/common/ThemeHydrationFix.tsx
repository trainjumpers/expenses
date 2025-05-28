"use client";

import { Skeleton } from "@/components/ui/skeleton";
import { useEffect, useState } from "react";

export function ThemeHydrationFix({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = useState(false);
  useEffect(() => {
    setMounted(true);
  }, []);
  if (!mounted) {
    return (
      <div
        style={{
          minHeight: "100vh",
          width: "100vw",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "#18181b",
        }}
      >
        <Skeleton className="h-20 w-20 rounded-2xl" />
      </div>
    );
  }
  return <>{children}</>;
}
