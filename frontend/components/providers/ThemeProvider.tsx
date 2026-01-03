"use client";

import { Skeleton } from "@/components/ui/skeleton";
import { ThemeProvider as NextThemesProvider, useTheme } from "next-themes";
import * as React from "react";

export function ThemeProvider({
  children,
  ...props
}: React.ComponentProps<typeof NextThemesProvider>) {
  const { resolvedTheme } = useTheme();
  
  if (!resolvedTheme) {
    return (
      <div className="min-h-screen w-screen flex items-center justify-center bg-zinc-900">
        <Skeleton className="h-20 w-20 rounded-2xl" />
      </div>
    );
  }
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>;
}
