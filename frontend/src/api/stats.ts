import { BACKEND_URL } from "@/constants/web";
import type { CategoryBreakdownData, HeatmapData, MonthlyTrendData } from "@/types/stats";
import { getUserToken } from "@/utils/cookies";

export const getCategoryBreakdown = async (
  startTime: string,
  endTime: string
) => {
  const response = await fetch(
    `${BACKEND_URL}/statistics/category?start_time=${startTime}&end_time=${endTime}`,
    {
      headers: {
        Authorization: `Bearer ${getUserToken()}`,
      },
    }
  );
  const data = (await response.json()).data as CategoryBreakdownData[];
  return data ?? [];
};

export const getMonthlyTrends = async (
  startDate: string,
  endDate: string
) => {
  const response = await fetch(
    `${BACKEND_URL}/statistics/monthly?start_date=${startDate}&end_date=${endDate}`,
    {
      headers: {
        Authorization: `Bearer ${getUserToken()}`,
      },
    }
  );
  const data = (await response.json()).data as MonthlyTrendData[];
  return data ?? [];
};

export const getHeatMapData = async (startDate: string, endDate: string) => {
  const response = await fetch(
    `${BACKEND_URL}/statistics/heatmap?start_date=${startDate}&end_date=${endDate}`,
    {
      headers: {
        Authorization: `Bearer ${getUserToken()}`,
      },
    }
  );
  const data = (await response.json()).data as HeatmapData[]
  return data ?? [];
};
