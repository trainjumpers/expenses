"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { InsightsCategory } from "@/lib/models/analytics";
import { formatCurrency } from "@/lib/utils";
import Link from "next/link";

interface CategoryBreakdownProps {
  categories: InsightsCategory[];
}

export function CategoryBreakdown({ categories }: CategoryBreakdownProps) {
  const sorted = [...categories].sort(
    (a, b) => Math.abs(b.total_amount) - Math.abs(a.total_amount)
  );
  const maxAmount = sorted.reduce(
    (max, category) => Math.max(max, Math.abs(category.total_amount)),
    0
  );

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-muted-foreground">
          Categories
        </CardTitle>
      </CardHeader>
      <CardContent>
        {sorted.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No category activity in this range.
          </p>
        ) : (
          <div className="space-y-3">
            {sorted.map((category) => {
              const href =
                category.category_id === -1
                  ? "/transaction?uncategorized=true"
                  : `/transaction?category_id=${category.category_id}`;
              const width =
                maxAmount > 0
                  ? (Math.abs(category.total_amount) / maxAmount) * 100
                  : 0;

              return (
                <div key={category.category_id} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <Link href={href} className="font-medium hover:underline">
                      {category.category_name}
                    </Link>
                    <span className="tabular-nums">
                      {formatCurrency(Math.abs(category.total_amount))}
                    </span>
                  </div>
                  <div className="h-2 rounded-full bg-muted overflow-hidden">
                    <div
                      className="h-full rounded-full bg-primary"
                      style={{ width: `${width}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
