"use client";

import { AccountAnalytics } from "@/components/custom/AccountAnalytics/AccountAnalytics";
import { CategoryAnalytics } from "@/components/custom/CategoryAnalytics/CategoryAnalytics";
import { AnalyticsSkeleton } from "@/components/custom/Dashboard/AnalyticsSkeleton";
import Dashboard from "@/components/custom/Dashboard/Dashboard";
import { NetWorth } from "@/components/custom/Dashboard/NetWorth";
import { CommandCenterModal } from "@/components/custom/Modal/CommandCenterModal";
import { InfoCenterModal } from "@/components/custom/Modal/InfoCenterModal";
import { useAccountAnalytics } from "@/components/hooks/useAnalytics";
import { useCategoryAnalytics } from "@/components/hooks/useCategoryAnalytics";
import { useUser } from "@/components/hooks/useUser";
import { Button } from "@/components/ui/button";
import { format } from "date-fns";
import { Eye, Plus } from "lucide-react";
import { useState } from "react";

export default function Page() {
  const { data: user } = useUser();
  const [isNewModalOpen, setIsNewModalOpen] = useState(false);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);

  // Date range for category analytics (last month to now)
  const [dateRange, setDateRange] = useState({
    from: new Date(new Date().setMonth(new Date().getMonth() - 1)),
    to: new Date(),
  });

  const {
    data: categoryData,
    isLoading: categoryLoading,
    isError: categoryError,
  } = useCategoryAnalytics(
    format(dateRange.from, "yyyy-MM-dd"),
    format(dateRange.to, "yyyy-MM-dd")
  );

  const {
    data: accountData,
    isLoading: accountLoading,
    isError: accountError,
  } = useAccountAnalytics();

  const isLoading = categoryLoading || accountLoading;
  const isError = categoryError || accountError;

  return (
    <Dashboard>
      <div className="flex items-center justify-between px-8 py-8 bg-background rounded-xl">
        <div>
          <h1 className="text-4xl font-bold text-foreground mb-2">
            Welcome back, {user?.name?.split(" ")[0] || "Human"}
          </h1>
          <p className="text-lg text-muted-foreground">
            Here&apos;s what&apos;s happening with your finances
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={() => setIsViewModalOpen(true)} variant="outline">
            <Eye className="h-4 w-4 mr-2" /> View
          </Button>
          <Button onClick={() => setIsNewModalOpen(true)}>
            <Plus className="h-4 w-4 mr-2" /> New
          </Button>
        </div>
      </div>

      {/* Net Worth Chart */}
      <div className="mb-8">
        <NetWorth dateRange={dateRange} onDateRangeChange={setDateRange} />
      </div>

      {/* Analytics Section */}
      <div className="mb-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {isLoading ? (
            <>
              <AnalyticsSkeleton />
              <AnalyticsSkeleton />
            </>
          ) : isError ? (
            <div className="col-span-2 text-center py-8">
              <p className="text-muted-foreground">
                Error loading analytics data.
              </p>
            </div>
          ) : (
            <>
              <CategoryAnalytics data={categoryData?.category_transactions} />
              <AccountAnalytics data={accountData?.account_analytics} />
            </>
          )}
        </div>
      </div>

      <CommandCenterModal
        isOpen={isNewModalOpen}
        onOpenChange={setIsNewModalOpen}
      />
      <InfoCenterModal
        isOpen={isViewModalOpen}
        onOpenChange={setIsViewModalOpen}
      />
    </Dashboard>
  );
}
