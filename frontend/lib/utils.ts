import { type ClassValue, clsx } from "clsx";
import { format } from "date-fns";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const formatCurrency = (
  amount: number,
  currency: string = "INR"
): string => {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency,
    minimumFractionDigits: 2,
  }).format(amount);
};

export const formatPercentage = (percentage: number): string => {
  const sign = percentage > 0 ? "+" : "";
  return `${sign}${percentage.toFixed(1)}%`;
};

interface ChartDataPoint {
  date: string;
  value: number;
  formattedDate: string;
}

// Transform API data to chart format
export const transformToChartData = (
  timeSeries: Array<{ date: string; networth: number }>
): ChartDataPoint[] => {
  return timeSeries.map((point) => ({
    date: format(new Date(point.date), "MMM dd"),
    value: point.networth,
    formattedDate: format(new Date(point.date), "MMM dd, yyyy"),
  }));
};
