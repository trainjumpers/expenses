"use client";

import Dashboard from "@/components/custom/Dashboard/Dashboard";
import { NetWorth } from "@/components/custom/Dashboard/NetWorth";
import { CommandCenterModal } from "@/components/custom/Modal/CommandCenterModal";
import { InfoCenterModal } from "@/components/custom/Modal/InfoCenterModal";
import { useUser } from "@/components/hooks/useUser";
import { Button } from "@/components/ui/button";
import { Eye, Plus } from "lucide-react";
import { useState } from "react";

export default function Page() {
  const { data: user } = useUser();
  const [isNewModalOpen, setIsNewModalOpen] = useState(false);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);

  return (
    <Dashboard>
      <div className="flex items-center justify-between px-8 py-8 bg-background rounded-xl mb-8">
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
        <NetWorth />
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
