"use client";

import { AnalyticsView } from "@/components/custom/Analytics/AnalyticsView";
import Dashboard from "@/components/custom/Dashboard/Dashboard";

export default function AnalyticsPage() {
  return (
    <Dashboard>
      <AnalyticsView />
    </Dashboard>
  );
}
