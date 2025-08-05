import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function AnalyticsSkeleton() {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Skeleton className="h-6 w-20" />
          <Skeleton className="h-4 w-2" />
          <Skeleton className="h-6 w-24" />
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Progress Bar Skeleton */}
        <div className="space-y-4">
          <Skeleton className="h-4 w-full rounded-full" />
          
          {/* Legend Skeleton */}
          <div className="flex flex-wrap gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="flex items-center gap-2">
                <Skeleton className="w-3 h-3 rounded-full" />
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-4 w-8" />
              </div>
            ))}
          </div>
        </div>

        {/* Table Skeleton */}
        <div className="border rounded-lg">
          <div className="p-4 space-y-3">
            {/* Table Header */}
            <div className="flex items-center gap-4 pb-2 border-b">
              <Skeleton className="w-12 h-4" />
              <Skeleton className="h-4 w-16" />
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-4 w-16 ml-auto" />
            </div>
            
            {/* Table Rows */}
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center gap-4 py-3 border-b last:border-b-0">
                <Skeleton className="w-12 h-4" />
                <Skeleton className="h-4 w-24" />
                <div className="flex items-center gap-2">
                  <Skeleton className="w-16 h-2" />
                  <Skeleton className="h-4 w-12" />
                </div>
                <Skeleton className="h-4 w-20 ml-auto" />
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
} 