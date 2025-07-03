import { Skeleton } from "@/components/ui/skeleton";

/**
 * Skeleton for a single rule card in the rule list view.
 */
export function RuleListCardSkeleton() {
  return (
    <div className="flex items-center justify-between p-4 rounded-lg border border-border">
      <div>
        <Skeleton className="h-4 w-32 mb-2" />
        <Skeleton className="h-3 w-48" />
      </div>
      <div className="flex gap-2">
        <Skeleton className="h-8 w-16 rounded" />
        <Skeleton className="h-8 w-8 rounded" />
      </div>
    </div>
  );
}

/**
 * Skeleton for the rule list (shows multiple cards).
 */
export function RuleListSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="grid gap-4">
      {Array.from({ length: count }).map((_, i) => (
        <RuleListCardSkeleton key={i} />
      ))}
    </div>
  );
}

/**
 * Skeleton for the Edit Rule modal form.
 */
export function EditRuleModalSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-6 w-1/3 mb-2" />
      <Skeleton className="h-10 w-full mb-2" />
      <Skeleton className="h-10 w-full mb-2" />
      <Skeleton className="h-6 w-1/4 mt-6 mb-2" />
      <Skeleton className="h-10 w-full mb-2" />
      <Skeleton className="h-10 w-full mb-2" />
      <Skeleton className="h-6 w-1/4 mt-6 mb-2" />
      <Skeleton className="h-10 w-full mb-2" />
      <Skeleton className="h-10 w-full mb-2" />
      <Skeleton className="h-6 w-1/4 mt-6 mb-2" />
      <Skeleton className="h-10 w-full mb-2" />
      <Skeleton className="h-10 w-1/2 mt-8" />
    </div>
  );
}
