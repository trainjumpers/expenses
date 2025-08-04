'use client';

import React from 'react';
import { Card, CardContent } from '../ui/card';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { X, Filter } from 'lucide-react';

interface AnalyticsFiltersProps {
  selectedAccounts: number[];
  selectedCategories: number[];
  onAccountsChange: (accounts: number[]) => void;
  onCategoriesChange: (categories: number[]) => void;
}

// Mock data - in real app, these would come from API
const MOCK_ACCOUNTS = [
  { id: 1, name: 'HDFC Savings', bank_name: 'HDFC Bank' },
  { id: 2, name: 'SBI Current', bank_name: 'State Bank of India' },
  { id: 3, name: 'Axis Credit Card', bank_name: 'Axis Bank' },
];

const MOCK_CATEGORIES = [
  { id: 1, name: 'Food & Dining', icon: '🍽️' },
  { id: 2, name: 'Transportation', icon: '🚗' },
  { id: 3, name: 'Shopping', icon: '🛍️' },
  { id: 4, name: 'Entertainment', icon: '🎬' },
  { id: 5, name: 'Bills & Utilities', icon: '💡' },
];

export function AnalyticsFilters({
  selectedAccounts,
  selectedCategories,
  onAccountsChange,
  onCategoriesChange,
}: AnalyticsFiltersProps) {
  const handleAccountSelect = (accountId: string) => {
    const id = parseInt(accountId);
    if (!selectedAccounts.includes(id)) {
      onAccountsChange([...selectedAccounts, id]);
    }
  };

  const handleCategorySelect = (categoryId: string) => {
    const id = parseInt(categoryId);
    if (!selectedCategories.includes(id)) {
      onCategoriesChange([...selectedCategories, id]);
    }
  };

  const removeAccount = (accountId: number) => {
    onAccountsChange(selectedAccounts.filter(id => id !== accountId));
  };

  const removeCategory = (categoryId: number) => {
    onCategoriesChange(selectedCategories.filter(id => id !== categoryId));
  };

  const clearAllFilters = () => {
    onAccountsChange([]);
    onCategoriesChange([]);
  };

  const hasFilters = selectedAccounts.length > 0 || selectedCategories.length > 0;

  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Filter className="h-4 w-4" />
            <span className="font-medium">Filters</span>
          </div>
          {hasFilters && (
            <Button variant="ghost" size="sm" onClick={clearAllFilters}>
              Clear all
            </Button>
          )}
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          {/* Account Filter */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Accounts</label>
            <Select onValueChange={handleAccountSelect}>
              <SelectTrigger>
                <SelectValue placeholder="Select accounts..." />
              </SelectTrigger>
              <SelectContent>
                {MOCK_ACCOUNTS
                  .filter(account => !selectedAccounts.includes(account.id))
                  .map((account) => (
                    <SelectItem key={account.id} value={account.id.toString()}>
                      <div className="flex flex-col">
                        <span>{account.name}</span>
                        <span className="text-xs text-muted-foreground">{account.bank_name}</span>
                      </div>
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
            
            {selectedAccounts.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {selectedAccounts.map((accountId) => {
                  const account = MOCK_ACCOUNTS.find(a => a.id === accountId);
                  return account ? (
                    <Badge key={accountId} variant="secondary" className="text-xs">
                      {account.name}
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-auto p-0 ml-1"
                        onClick={() => removeAccount(accountId)}
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    </Badge>
                  ) : null;
                })}
              </div>
            )}
          </div>

          {/* Category Filter */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Categories</label>
            <Select onValueChange={handleCategorySelect}>
              <SelectTrigger>
                <SelectValue placeholder="Select categories..." />
              </SelectTrigger>
              <SelectContent>
                {MOCK_CATEGORIES
                  .filter(category => !selectedCategories.includes(category.id))
                  .map((category) => (
                    <SelectItem key={category.id} value={category.id.toString()}>
                      <div className="flex items-center gap-2">
                        <span>{category.icon}</span>
                        <span>{category.name}</span>
                      </div>
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
            
            {selectedCategories.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {selectedCategories.map((categoryId) => {
                  const category = MOCK_CATEGORIES.find(c => c.id === categoryId);
                  return category ? (
                    <Badge key={categoryId} variant="secondary" className="text-xs">
                      {category.icon} {category.name}
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-auto p-0 ml-1"
                        onClick={() => removeCategory(categoryId)}
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    </Badge>
                  ) : null;
                })}
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
