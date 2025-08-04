import { Navbar } from "@/components/custom/Navbar/Navbar";

import { AccountsAnalyticsSidepanel } from "./AccountsAnalyticsSidepanel";

export default function Dashboard({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <div className="flex flex-1 gap-4 px-2">
        <aside className="hidden lg:block flex-shrink-0">
          <AccountsAnalyticsSidepanel className="h-[calc(100vh-4rem)] sticky top-4" />
        </aside>
        <main className="flex-1 min-w-0 px-2 py-4">{children}</main>
      </div>
    </div>
  );
}
