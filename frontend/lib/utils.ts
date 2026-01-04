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

export const formatShortCurrency = (
  amount: number,
  currency: string = "INR"
): string => {
  const abs = Math.abs(amount);
  const sign = amount < 0 ? "-" : "";

  // Extract a narrow currency symbol (e.g., ₹)
  const currencySymbol = new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency,
    currencyDisplay: "narrowSymbol",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  })
    .format(0)
    .replace(/0/g, "")
    .trim();

  if (!isFinite(abs)) return formatCurrency(amount, currency);

  if (abs >= 1e7) {
    const v = (abs / 1e7).toFixed(1).replace(/\.0$/, "");
    return `${sign}${currencySymbol}${v} Cr`;
  }

  if (abs >= 1e5) {
    const v = (abs / 1e5).toFixed(1).replace(/\.0$/, "");
    return `${sign}${currencySymbol}${v}L`;
  }

  if (abs >= 1e3) {
    const v = (abs / 1e3).toFixed(1).replace(/\.0$/, "");
    return `${sign}${currencySymbol}${v}K`;
  }

  // For small numbers, show the standard formatted currency
  return formatCurrency(amount, currency);
};

export const formatPercentage = (percentage: number): string => {
  const sign = percentage > 0 ? "+" : "";
  return `${sign}${percentage.toFixed(1)}%`;
};

export const getTransactionColor = (amount: number): string => {
  if (amount < 0) {
    return `text-emerald-600 dark:text-emerald-400`;
  }
  return "text-rose-600 dark:text-rose-400";
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
