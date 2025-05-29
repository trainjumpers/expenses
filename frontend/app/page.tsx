"use client";

import Dashboard from "@/components/custom/Dashboard/Dashboard";
import { useUser } from "@/components/custom/Provider/UserProvider";

export default function Page() {
  const { read: user } = useUser();
  return (
    <Dashboard>
      <main className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-between px-8 py-8 bg-background rounded-xl mb-8">
          <div>
            <h1 className="text-4xl font-bold text-foreground mb-2">
              Welcome back, {user().name}
            </h1>
            <p className="text-lg text-muted-foreground">
              Here&apos;s what&apos;s happening with your finances
            </p>
          </div>
        </div>
      </main>
    </Dashboard>
  );
}
