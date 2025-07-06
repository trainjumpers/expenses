import DashboardSkeleton from "@/components/custom/Dashboard/DashboardSkeleton";
import { AccountProvider } from "@/components/custom/Provider/AccountProvider";
import { CategoryProvider } from "@/components/custom/Provider/CategoryProvider";
import { SessionProvider } from "@/components/custom/Provider/SessionProvider";
import { ThemeProvider } from "@/components/custom/Provider/ThemeProvider";
import { UserProvider } from "@/components/custom/Provider/UserProvider";
import { Toaster } from "@/components/ui/sonner";
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Suspense } from "react";

import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "NeuroSpend",
  description: "Smart expense tracker with automated statement parsing",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <SessionProvider>
            <UserProvider>
              <AccountProvider>
                <CategoryProvider>
                  <Suspense fallback={<DashboardSkeleton />}>
                    {children}
                  </Suspense>
                </CategoryProvider>
              </AccountProvider>
            </UserProvider>
          </SessionProvider>
          <Toaster
            position="top-right"
            richColors
            toastOptions={{
              classNames: {
                toast:
                  "shadow-lg rounded-lg flex items-center p-4 text-xs gap-1.5",
                error: "[&>button]:!bg-red-300",
                info: "[&>button]:!bg-blue-300",
                success: "[&>button]:!bg-green-300",
                warning: "[&>button]:!bg-yellow-300",
              },
            }}
          />
        </ThemeProvider>
      </body>
    </html>
  );
}
