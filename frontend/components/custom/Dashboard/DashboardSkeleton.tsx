import { Skeleton } from "@/components/ui/skeleton";

import { AccountsAnalyticsSidepanelSkeleton } from "./AccountsAnalyticsSidepanelSkeleton";

export default function DashboardSkeleton() {
  return (
    <div className="flex gap-4 px-2 py-4">
      <aside className="hidden lg:block flex-shrink-0">
        <AccountsAnalyticsSidepanelSkeleton />
      </aside>
      <main className="flex-1 min-w-0 px-2">
        <div className="flex items-center justify-between px-8 py-8 bg-background rounded-xl shadow-sm mb-8">
          <div>
            <Skeleton className="h-10 w-64 mb-2" />
            <Skeleton className="h-6 w-80" />
          </div>
          <Skeleton className="h-12 w-32 rounded-xl" />
        </div>
        <Skeleton className="h-8 w-48 mb-6" />
        <div className="bg-white rounded-lg shadow">
          <Skeleton className="h-[700px] w-full rounded-lg" />
        </div>
      </main>
    </div>
  );
}
