<template>
  <div class="card w-full bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title">Expense Distribution by Category</h2>
      <div id="category-chart">
        <apexchart
          type="pie"
          :options="chartOptions"
          :series="series"
          height="350"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getCategoryBreakdown } from "@/api/stats";
import { useDashboardStore } from "@/stores/dashboard";
import { onMounted, ref } from "vue";

const { startTime, endTime } = useDashboardStore();

const series = ref<number[]>([]);
const chartOptions = ref({
  labels: [] as string[],
  colors: ["#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF"],
  legend: {
    position: "bottom",
  },
});

const fetchData = async () => {
  const data = await getCategoryBreakdown(startTime, endTime);
  series.value = data.map((item) => item.total_amount);
  chartOptions.value.labels = data.map((item) => item.subcategory_name);
};

onMounted(fetchData);
</script>
