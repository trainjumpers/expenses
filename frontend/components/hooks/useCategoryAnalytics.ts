import { getCategoryAnalytics } from "@/lib/api/analytics";
import { useQuery } from "@tanstack/react-query";

export function useCategoryAnalytics(
  startDate: string,
  endDate: string,
  categoryIds?: number[]
) {
  const categoryKey = categoryIds?.length ? categoryIds.join(",") : "all";

  return useQuery({
    queryKey: ["categoryAnalytics", startDate, endDate, categoryKey],
    queryFn: ({ signal }) =>
      getCategoryAnalytics(startDate, endDate, categoryIds, signal),
    enabled: !!startDate && !!endDate,
  });
}
