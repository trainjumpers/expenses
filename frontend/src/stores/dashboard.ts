import { defineStore } from "pinia";
import { ref } from "vue";

export const useDashboardStore = defineStore("dashboard", () => {
  const startTime = ref(new Date(Date.now() - 12 * 30 * 24 * 60 * 60 * 1000).toISOString());
  const endTime = ref(new Date().toISOString());

  const setStartTime = (time: Date) => {
    startTime.value = time.toISOString();
  };
  const setEndTime = (time: Date) => {
    endTime.value = time.toISOString();
  };
  return { startTime, endTime, setStartTime, setEndTime };
});
