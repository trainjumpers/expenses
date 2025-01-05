<script setup lang="ts">
import DailyHeatmap from "@/components/Statistics/DailyHeatmap.vue";
import Navbar from "@/components/Navbar/NavBar.vue";
import CategoryBreakdown from "@/components/Statistics/CategoryBreakdown.vue";
import MonthlyTrend from "@/components/Statistics/MonthlyTrend.vue";
</script>

<template>
  <Suspense>
    <div>
      <Navbar />
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-4">Welcome to Expense Tracker</h1>
        <p className="text-lg mb-4">
          This is a simple expense tracker application built with Vue.js and
          Tailwind CSS.
        </p>
      </div>
      <router-view />
      <div class="flex flex-row gap-4">
        <CategoryBreakdown />
        <MonthlyTrend />
      </div>
      <DailyHeatmap />
    </div>

    <template #fallback>
      <div className="flex justify-center items-center h-screen">
        <div
          className="animate-spin rounded-full h-16 w-16 border-t-4 border-blue-500 border-solid"
        ></div>
      </div>
    </template>
  </Suspense>
</template>
