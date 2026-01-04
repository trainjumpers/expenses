import { AddCategoryModal } from "@/components/custom/Modal/Category/AddCategoryModal";
import { useTransactions } from "@/components/hooks/useTransactions";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { CategoryAnalyticsResponse } from "@/lib/models/analytics";
import type { Category } from "@/lib/models/category";
import { formatCurrency, getTransactionColor } from "@/lib/utils";
import { format } from "date-fns";
import { ChevronRight, FileQuestion, Plus, Tag } from "lucide-react";
import Link from "next/link";
import { Fragment, useEffect, useState } from "react";

interface CategoryAnalyticsProps {
  data?: CategoryAnalyticsResponse["category_transactions"];
  categories?: Category[];
  selectedCategoryIds?: number[];
  onCategoryFilterChange?: (value: number[]) => void;
}

interface CategoryTransactionsProps {
  categoryId: number;
  isUncategorized: boolean;
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

function CategoryTransactions({
  categoryId,
  isUncategorized,
}: CategoryTransactionsProps) {
  const { data, isLoading, error } = useTransactions({
    page: 1,
    page_size: 5,
    sort_by: "date",
    sort_order: "desc",
    ...(isUncategorized
      ? { uncategorized: true }
      : { category_id: categoryId }),
  });

  if (isLoading) {
    return (
      <TableRow>
        <TableCell colSpan={4} className="bg-muted/40">
          <div className="px-4 py-3 text-sm text-muted-foreground">
            Loading latest transactions...
          </div>
        </TableCell>
      </TableRow>
    );
  }

  if (error) {
    return (
      <TableRow>
        <TableCell colSpan={4} className="bg-muted/40">
          <div className="px-4 py-3 text-sm text-destructive">
            Failed to load transactions.
          </div>
        </TableCell>
      </TableRow>
    );
  }

  const transactions = data?.transactions ?? [];

  if (transactions.length === 0) {
    return (
      <TableRow>
        <TableCell colSpan={4} className="bg-muted/40">
          <div className="px-4 py-3 text-sm text-muted-foreground">
            No recent transactions for this category.
          </div>
        </TableCell>
      </TableRow>
    );
  }

  const transactionsUrl = isUncategorized
    ? "/transaction?uncategorized=true"
    : `/transaction?category_id=${categoryId}`;

  return (
    <TableRow>
      <TableCell colSpan={4} className="bg-muted/40">
        <div className="px-4 py-3">
          <div className="flex items-center justify-between gap-2">
            <div className="text-xs uppercase tracking-wide text-muted-foreground">
              Latest 5 transactions
            </div>
            <Button variant="ghost" size="sm" asChild>
              <Link href={transactionsUrl}>View all</Link>
            </Button>
          </div>
          <div className="mt-3 space-y-2">
            {transactions.map((transaction) => (
              <div
                key={transaction.id}
                className="flex items-center justify-between gap-3 rounded-md border border-border/60 bg-background px-3 py-2"
              >
                <div className="min-w-0">
                  <div className="text-sm font-medium text-foreground">
                    {transaction.name}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {format(new Date(transaction.date), "MMM d, yyyy")}
                  </div>
                </div>
                <div
                  className={`text-sm font-semibold ${getTransactionColor(
                    transaction.amount
                  )}`}
                >
                  {formatCurrency(Math.abs(transaction.amount))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </TableCell>
    </TableRow>
  );
}

export function CategoryAnalytics({
  data,
  categories,
  selectedCategoryIds = [],
  onCategoryFilterChange,
}: CategoryAnalyticsProps) {
  const [expandedCategories, setExpandedCategories] = useState<Set<number>>(
    new Set()
  );
  const [isAddCategoryModalOpen, setIsAddCategoryModalOpen] = useState(false);
  const [draftSelectedIds, setDraftSelectedIds] =
    useState<number[]>(selectedCategoryIds);

  useEffect(() => {
    setDraftSelectedIds(selectedCategoryIds);
  }, [selectedCategoryIds]);

  const hasData = !!data && data.length > 0;
  const hasCategoryList = !!categories && categories.length > 0;
  const showFilter = hasCategoryList && !!onCategoryFilterChange;
  const handleCategoryFilterChange =
    onCategoryFilterChange ?? (() => undefined);
  const allCategoryIds = hasCategoryList
    ? [...categories.map((category) => category.id), -1]
    : [-1];
  const hasAllSelectedApplied =
    selectedCategoryIds.length > 0 &&
    allCategoryIds.every((categoryId) =>
      selectedCategoryIds.includes(categoryId)
    );
  const hasAllSelectedDraft =
    draftSelectedIds.length > 0 &&
    allCategoryIds.every((categoryId) => draftSelectedIds.includes(categoryId));
  const selectedCategoryCount = selectedCategoryIds.length;
  const triggerLabel =
    selectedCategoryCount === 0 || hasAllSelectedApplied
      ? "All categories"
      : `${selectedCategoryCount} selected`;
  const isDirty =
    selectedCategoryIds.length !== draftSelectedIds.length ||
    selectedCategoryIds.some(
      (categoryId) => !draftSelectedIds.includes(categoryId)
    );

  const toggleCategorySelection = (categoryId: number, checked: boolean) => {
    if (checked) {
      if (draftSelectedIds.includes(categoryId)) {
        return;
      }
      setDraftSelectedIds([...draftSelectedIds, categoryId]);
      return;
    }

    setDraftSelectedIds(draftSelectedIds.filter((id) => id !== categoryId));
  };

  const toggleSelectAll = () => {
    if (hasAllSelectedDraft) {
      setDraftSelectedIds([]);
      return;
    }

    setDraftSelectedIds(allCategoryIds);
  };

  const applyCategoryFilter = () => {
    handleCategoryFilterChange(draftSelectedIds);
  };

  if (!hasData && !hasCategoryList) {
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

  if (!hasData) {
    return (
      <>
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between gap-3">
              <span>Categories</span>
              <div className="flex items-center gap-2">
                {showFilter && (
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="outline" size="sm" className="h-8">
                        {triggerLabel}
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-56">
                      <div className="flex items-center justify-between px-2 py-1.5 text-xs text-muted-foreground">
                        <span>Categories</span>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 px-2"
                          onClick={toggleSelectAll}
                        >
                          {hasAllSelectedDraft ? "Deselect all" : "Select all"}
                        </Button>
                      </div>
                      <DropdownMenuSeparator />
                      <DropdownMenuCheckboxItem
                        checked={draftSelectedIds.includes(-1)}
                        onCheckedChange={(checked) =>
                          toggleCategorySelection(-1, Boolean(checked))
                        }
                        onSelect={(event) => event.preventDefault()}
                      >
                        Uncategorized
                      </DropdownMenuCheckboxItem>
                      {categories?.map((category) => (
                        <DropdownMenuCheckboxItem
                          key={category.id}
                          checked={draftSelectedIds.includes(category.id)}
                          onCheckedChange={(checked) =>
                            toggleCategorySelection(
                              category.id,
                              Boolean(checked)
                            )
                          }
                          onSelect={(event) => event.preventDefault()}
                        >
                          {category.name}
                        </DropdownMenuCheckboxItem>
                      ))}
                      <DropdownMenuSeparator />
                      <div className="flex justify-end px-2 py-2">
                        <Button
                          size="sm"
                          onClick={applyCategoryFilter}
                          disabled={!isDirty}
                        >
                          Apply
                        </Button>
                      </div>
                    </DropdownMenuContent>
                  </DropdownMenu>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsAddCategoryModalOpen(true)}
                  className="h-8 w-8 p-0"
                >
                  <Plus className="h-4 w-4" />
                </Button>
              </div>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">
              No category activity for the selected filter.
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
      const isUncategorized = category.category_id === -1;
      return {
        ...category,
        percentage,
        color: isUncategorized
          ? "bg-gray-400"
          : categoryColors[index % categoryColors.length],
        isUncategorized,
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
          <CardTitle className="flex items-center justify-between gap-3">
            <span>Categories</span>
            <div className="flex items-center gap-2">
              {showFilter && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm" className="h-8">
                      {triggerLabel}
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-56">
                    <div className="flex items-center justify-between px-2 py-1.5 text-xs text-muted-foreground">
                      <span>Categories</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 px-2"
                        onClick={toggleSelectAll}
                      >
                        {hasAllSelectedDraft ? "Deselect all" : "Select all"}
                      </Button>
                    </div>
                    <DropdownMenuSeparator />
                    <DropdownMenuCheckboxItem
                      checked={draftSelectedIds.includes(-1)}
                      onCheckedChange={(checked) =>
                        toggleCategorySelection(-1, Boolean(checked))
                      }
                      onSelect={(event) => event.preventDefault()}
                    >
                      Uncategorized
                    </DropdownMenuCheckboxItem>
                    {categories?.map((category) => (
                      <DropdownMenuCheckboxItem
                        key={category.id}
                        checked={draftSelectedIds.includes(category.id)}
                        onCheckedChange={(checked) =>
                          toggleCategorySelection(category.id, Boolean(checked))
                        }
                        onSelect={(event) => event.preventDefault()}
                      >
                        {category.name}
                      </DropdownMenuCheckboxItem>
                    ))}
                    <DropdownMenuSeparator />
                    <div className="flex justify-end px-2 py-2">
                      <Button
                        size="sm"
                        onClick={applyCategoryFilter}
                        disabled={!isDirty}
                      >
                        Apply
                      </Button>
                    </div>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsAddCategoryModalOpen(true)}
                className="h-8 w-8 p-0"
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>
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
                  {category.isUncategorized ? (
                    <FileQuestion className="w-3 h-3 text-gray-500" />
                  ) : (
                    <div className={`w-3 h-3 rounded-full ${category.color}`} />
                  )}
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
                {categoriesWithPercentages.map((category) => {
                  const isExpanded = expandedCategories.has(
                    category.category_id
                  );

                  return (
                    <Fragment key={category.category_id}>
                      <TableRow className="border-b">
                        <TableCell className="w-12">
                          <button
                            onClick={() =>
                              toggleCategoryExpansion(category.category_id)
                            }
                            className="p-1 hover:bg-muted rounded transition-colors"
                          >
                            <ChevronRight
                              className={`h-4 w-4 transition-transform ${
                                isExpanded ? "rotate-90" : ""
                              }`}
                            />
                          </button>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {category.isUncategorized && (
                              <FileQuestion className="w-4 h-4 text-gray-500" />
                            )}
                            <span className="font-medium">
                              {category.category_name}
                            </span>
                          </div>
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
                      {isExpanded && (
                        <CategoryTransactions
                          categoryId={category.category_id}
                          isUncategorized={category.isUncategorized}
                        />
                      )}
                    </Fragment>
                  );
                })}
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
