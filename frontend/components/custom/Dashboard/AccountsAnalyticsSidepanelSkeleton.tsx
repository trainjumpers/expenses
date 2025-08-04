import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function AccountsAnalyticsSidepanelSkeleton() {
  return (
    <Card className="w-80 h-[calc(100vh-4rem)]">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-6 w-16" />
          <Skeleton className="h-8 w-8" />
        </div>
        <Skeleton className="h-4 w-20" />
      </CardHeader>
      <CardContent className="space-y-1">
        {Array.from({ length: 6 }).map((_, index) => (
          <div key={index} className="flex items-center justify-between py-3 px-2">
            <div className="flex items-center space-x-3">
              <Skeleton className="h-4 w-4" />
              <Skeleton className="h-4 w-20" />
            </div>
            <div className="text-right space-y-1">
              <Skeleton className="h-4 w-16" />
              <Skeleton className="h-3 w-10" />
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}