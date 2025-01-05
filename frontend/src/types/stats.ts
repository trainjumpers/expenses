export type CategoryBreakdownData = {
  category_name: string;
  category_color: string;
  subcategory_color: string;
  subcategory_name: string;
  total_amount: number;
  transaction_count: number;
}

export type MonthlyTrendData = {
  month: string;
  total_amount: number
}

export type HeatmapData = {
  day: string
  week_number: number
  total_amount: number
}
