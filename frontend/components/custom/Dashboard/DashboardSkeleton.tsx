import { Skeleton } from "@/components/ui/skeleton";

export default function DashboardSkeleton() {
  return (
    <main className="container mx-auto px-4 py-8">
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
  );
}
