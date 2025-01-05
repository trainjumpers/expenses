<template>
  <div class="card w-full bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title">Monthly Spending Trend</h2>
      <div id="trend-chart">
        <apexchart
          type="line"
          :options="chartOptions"
          :series="series"
          height="350"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getMonthlyTrends } from "@/api/stats";
import { useDashboardStore } from "@/stores/dashboard";
import type { MonthlyTrendData } from "@/types/stats";
import { onMounted, ref } from "vue";

const {startTime, endTime} = useDashboardStore();
const series = ref([
  {
    name: "Monthly Spending",
    data: [] as number[],
  },
]);
const chartOptions = ref({
  chart: {
    type: "line",
    zoom: { enabled: false },
  },
  xaxis: {
    categories: [] as string[],
  },
  stroke: {
    curve: "smooth",
  },
});

const fetchData = async () => {

  const data: MonthlyTrendData[] = await getMonthlyTrends(startTime, endTime);
  series.value[0].data = data.map((item) => Math.trunc(item.total_amount));
  chartOptions.value.xaxis.categories = data.map((item) => item.month);
};

onMounted(fetchData);
</script>
