import { getCategoryAnalytics } from "@/lib/api/analytics";
import { useQuery } from "@tanstack/react-query";

export function useCategoryAnalytics(startDate: string, endDate: string) {
  return useQuery({
    queryKey: ["categoryAnalytics", startDate, endDate],
    queryFn: ({ signal }) => getCategoryAnalytics(startDate, endDate, signal),
    enabled: !!startDate && !!endDate,
  });
}
