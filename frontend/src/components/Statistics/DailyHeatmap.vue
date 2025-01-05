<template>
  <div class="card w-full bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title">Daily Spending Heatmap</h2>
      <div id="heatmap-chart">
        <apexchart
          type="heatmap"
          :options="chartOptions"
          :series="series"
          height="350"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getHeatMapData } from "@/api/stats";
import { useDashboardStore } from "@/stores/dashboard";
import type { HeatmapData } from "@/types/stats";
import { onMounted, ref } from "vue";

const {startTime, endTime} = useDashboardStore();

const series = ref([
  {
    name: "Spending",
    data: [] as { x: string; y: number }[],
  },
]);

const chartOptions = ref({
  chart: {
    type: "heatmap",
  },
  dataLabels: {
    enabled: false,
  },
  colors: ["#008FFB"],
  plotOptions: {
    heatmap: {
      colorScale: {
        ranges: [
          { from: 0, to: 1000, color: "#00A100" },
          { from: 1001, to: 2000, color: "#128FD9" },
          { from: 2001, to: 5000, color: "#FFB200" },
          { from: 5001, to: Infinity, color: "#FF0000" },
        ],
      },
    },
  },
});

const fetchData = async () => {
  const data: HeatmapData[] = await getHeatMapData(startTime, endTime);
  series.value[0].data = data.map((item) => ({
    x: item.day,
    y: item.total_amount,
  }));
};

onMounted(fetchData);
</script>
