import { AddCategoryModal } from "@/components/custom/Modal/Category/AddCategoryModal";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { CategoryAnalyticsResponse } from "@/lib/models/analytics";
import { formatCurrency } from "@/lib/utils";
import { ChevronRight, Plus, Tag } from "lucide-react";
import { useState } from "react";

interface CategoryAnalyticsProps {
  data?: CategoryAnalyticsResponse["category_transactions"];
}

// Color palette for different categories
const categoryColors = [
  "bg-purple-500",
  "bg-blue-600",
  "bg-gray-600",
  "bg-blue-400",
  "bg-pink-400",
  "bg-green-500",
  "bg-orange-500",
  "bg-indigo-500",
  "bg-teal-500",
  "bg-red-500",
];

export function CategoryAnalytics({ data }: CategoryAnalyticsProps) {
  const [expandedCategories, setExpandedCategories] = useState<Set<number>>(
    new Set()
  );
  const [isAddCategoryModalOpen, setIsAddCategoryModalOpen] = useState(false);

  if (!data || data.length === 0) {
    return (
      <>
        <Card className="h-full">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span>Category Analytics</span>
              <Tag className="h-5 w-5 text-muted-foreground" />
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col items-center justify-center py-12 space-y-6">
              <div className="rounded-full bg-muted p-6">
                <Tag className="h-12 w-12 text-muted-foreground" />
              </div>
              <div className="text-center space-y-2">
                <h3 className="text-lg font-semibold">No categories yet</h3>
                <p className="text-sm text-muted-foreground max-w-sm">
                  Start organizing your expenses by creating your first
                  category. You can add categories for groceries, entertainment,
                  bills, and more.
                </p>
              </div>
              <Button
                onClick={() => setIsAddCategoryModalOpen(true)}
                className="flex items-center gap-2"
              >
                <Plus className="h-4 w-4" />
                Add Your First Category
              </Button>
            </div>
          </CardContent>
        </Card>

        <AddCategoryModal
          isOpen={isAddCategoryModalOpen}
          onOpenChange={setIsAddCategoryModalOpen}
          onCategoryAdded={() => {
            // The category list will automatically refresh due to React Query
            setIsAddCategoryModalOpen(false);
          }}
        />
      </>
    );
  }

  // Calculate total amount - use absolute values to handle negative amounts
  const totalAmount = data.reduce(
    (sum, category) => sum + Math.abs(category.total_amount),
    0
  );

  // Calculate percentages and prepare data
  const categoriesWithPercentages = data
    .map((category, index) => {
      const absoluteAmount = Math.abs(category.total_amount);
      const percentage =
        totalAmount > 0 ? (absoluteAmount / totalAmount) * 100 : 0;
      return {
        ...category,
        percentage,
        color: categoryColors[index % categoryColors.length],
      };
    })
    .sort((a, b) => b.percentage - a.percentage); // Sort by percentage descending

  const toggleCategoryExpansion = (categoryId: number) => {
    const newExpanded = new Set(expandedCategories);
    if (newExpanded.has(categoryId)) {
      newExpanded.delete(categoryId);
    } else {
      newExpanded.add(categoryId);
    }
    setExpandedCategories(newExpanded);
  };

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span>Categories</span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsAddCategoryModalOpen(true)}
              className="h-8 w-8 p-0"
            >
              <Plus className="h-4 w-4" />
            </Button>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Horizontal Progress Bar */}
          <div className="space-y-4">
            <div className="h-4 rounded-full overflow-hidden">
              {categoriesWithPercentages.map((category) => (
                <div
                  key={category.category_id}
                  className={`h-full ${category.color} inline-block`}
                  style={{ width: `${category.percentage}%` }}
                />
              ))}
            </div>

            {/* Legend */}
            <div className="flex flex-wrap gap-4 text-sm">
              {categoriesWithPercentages.map((category) => (
                <div
                  key={category.category_id}
                  className="flex items-center gap-2"
                >
                  <div className={`w-3 h-3 rounded-full ${category.color}`} />
                  <span className="text-muted-foreground">
                    {category.category_name}:
                  </span>
                  <span className="font-medium">
                    {category.percentage.toFixed(1)}%
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Detailed Table */}
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow className="border-b">
                  <TableHead className="w-12"></TableHead>
                  <TableHead className="text-left text-sm font-medium text-muted-foreground">
                    NAME
                  </TableHead>
                  <TableHead className="text-left text-sm font-medium text-muted-foreground">
                    WEIGHT
                  </TableHead>
                  <TableHead className="text-right text-sm font-medium text-muted-foreground">
                    VALUE
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {categoriesWithPercentages.map((category) => (
                  <TableRow key={category.category_id} className="border-b">
                    <TableCell className="w-12">
                      <button
                        onClick={() =>
                          toggleCategoryExpansion(category.category_id)
                        }
                        className="p-1 hover:bg-muted rounded transition-colors"
                      >
                        <ChevronRight
                          className={`h-4 w-4 transition-transform ${
                            expandedCategories.has(category.category_id)
                              ? "rotate-90"
                              : ""
                          }`}
                        />
                      </button>
                    </TableCell>
                    <TableCell>
                      <span className="font-medium">
                        {category.category_name}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className="w-16 h-2 flex">
                          {Array.from({ length: 5 }).map((_, i) => (
                            <div
                              key={i}
                              className={`flex-1 h-full ${
                                i < Math.floor(category.percentage / 20)
                                  ? category.color
                                  : "bg-gray-200"
                              }`}
                              style={{
                                marginRight: i < 4 ? "1px" : "0",
                              }}
                            />
                          ))}
                        </div>
                        <span className="text-sm">
                          {category.percentage.toFixed(2)}%
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <span className="font-medium">
                        {formatCurrency(Math.abs(category.total_amount))}
                      </span>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <AddCategoryModal
        isOpen={isAddCategoryModalOpen}
        onOpenChange={setIsAddCategoryModalOpen}
        onCategoryAdded={() => {
          // The category list will automatically refresh due to React Query
          setIsAddCategoryModalOpen(false);
        }}
      />
    </>
  );
}
