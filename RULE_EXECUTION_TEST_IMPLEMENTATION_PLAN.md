# Rule Execution Test Implementation Plan

## Recommended Test Implementation Strategy

Based on the comprehensive test plan, here's a prioritized approach for implementing tests for the rule execution functionality.

## Phase 1: Core Functionality Tests (High Priority)

### 1.1 Basic ExecuteRules Method Tests (Estimated: 2-3 hours)
```go
Describe("ExecuteRules", func() {
    Context("when no rules exist", func() {
        It("should return empty result with zero counts")
    })
    
    Context("when no transactions exist", func() {
        It("should return result with rule count but zero processed transactions")
    })
    
    Context("when both rules and transactions exist", func() {
        It("should execute rules and return modified transactions")
    })
    
    Context("with specific rule IDs", func() {
        It("should only execute specified rules")
    })
    
    Context("with specific transaction IDs", func() {
        It("should only process specified transactions")
    })
})
```

### 1.2 Rule Condition Evaluation Tests (Estimated: 3-4 hours)
```go
Describe("Rule Condition Evaluation", func() {
    Context("Amount field conditions", func() {
        It("should evaluate equals condition correctly")
        It("should evaluate greater than condition correctly")
        It("should evaluate less than condition correctly")
        It("should handle invalid amount values gracefully")
    })
    
    Context("Name field conditions", func() {
        It("should evaluate equals condition (case insensitive)")
        It("should evaluate contains condition")
        It("should handle empty names")
    })
    
    Context("Description field conditions", func() {
        It("should evaluate equals condition")
        It("should evaluate contains condition")
        It("should handle nil descriptions")
    })
    
    Context("Category field conditions", func() {
        It("should evaluate equals condition with single category")
        It("should evaluate equals condition with multiple categories")
        It("should handle invalid category IDs")
    })
    
    Context("Multiple conditions (AND logic)", func() {
        It("should apply rule when all conditions match")
        It("should not apply rule when some conditions don't match")
    })
})
```

### 1.3 Rule Action Application Tests (Estimated: 2-3 hours)
```go
Describe("Rule Action Application", func() {
    Context("Name actions", func() {
        It("should update transaction name")
        It("should respect first-rule-wins policy")
    })
    
    Context("Description actions", func() {
        It("should update transaction description")
        It("should respect first-rule-wins policy")
    })
    
    Context("Category actions", func() {
        It("should add category to transaction")
        It("should preserve existing categories (additive)")
        It("should skip invalid category IDs")
        It("should skip categories from different users")
    })
})
```

### 1.4 Rule Effective Date Tests (Estimated: 1-2 hours)
```go
Describe("Rule Effective Date Validation", func() {
    It("should apply rule when effective date is before transaction date")
    It("should not apply rule when effective date is after transaction date")
    It("should apply rule when effective date equals transaction date")
    It("should not apply future-effective rules")
})
```

## Phase 2: Error Handling & Edge Cases (Medium Priority)

### 2.1 Error Handling Tests (Estimated: 2-3 hours)
```go
Describe("Error Handling", func() {
    Context("Database errors", func() {
        It("should handle rule fetching errors")
        It("should handle transaction fetching errors")
        It("should handle transaction update errors")
    })
    
    Context("Validation errors", func() {
        It("should handle invalid rule IDs gracefully")
        It("should handle invalid transaction IDs gracefully")
    })
})
```

### 2.2 ExecuteRulesForTransaction Tests (Estimated: 1-2 hours)
```go
Describe("ExecuteRulesForTransaction", func() {
    It("should execute rules for valid transaction")
    It("should return error for non-existent transaction")
    It("should return error for transaction of different user")
})
```

### 2.3 ExecuteRulesForRule Tests (Estimated: 1-2 hours)
```go
Describe("ExecuteRulesForRule", func() {
    It("should execute specific rule for all transactions")
    It("should handle non-existent rule gracefully")
    It("should handle rule of different user gracefully")
})
```

## Phase 3: Integration & Performance Tests (Lower Priority)

### 3.1 Integration Tests (Estimated: 2-3 hours)
```go
Describe("Integration Scenarios", func() {
    It("should handle complete workflow: create rule -> create transaction -> execute")
    It("should handle multiple rules on single transaction")
    It("should handle single rule on multiple transactions")
    It("should return correct modified and skipped results")
})
```

### 3.2 Pagination Tests (Estimated: 1-2 hours)
```go
Describe("Pagination", func() {
    It("should process multiple pages of transactions")
    It("should handle custom page sizes")
    It("should default invalid page sizes appropriately")
})
```

## Implementation Recommendations

### Test Data Setup Strategy
```go
// Helper functions for test data creation
func createTestRule(name string, effectiveFrom time.Time, userId int64) models.CreateRuleRequest
func createTestTransaction(name string, amount float64, date time.Time, userId int64) models.TransactionResponse
func createTestCategory(name string, userId int64) models.CategoryResponse
```

### Mock Setup Patterns
```go
// Common mock setups
func setupMockRuleRepo(rules []models.RuleResponse, actions []models.RuleActionResponse, conditions []models.RuleConditionResponse)
func setupMockTransactionRepo(transactions []models.TransactionResponse)
func setupMockCategoryRepo(categories []models.CategoryResponse)
```

### Test Scenarios to Focus On

#### High-Value Test Cases
1. **Rule with amount condition and name action**
   - Transaction amount > 1000 → Set name to "Large Purchase"
   
2. **Rule with description condition and category action**
   - Description contains "grocery" → Add "Food" category
   
3. **Multiple rules on same transaction**
   - Rule 1: amount > 500 → Add "Expensive" category
   - Rule 2: name contains "restaurant" → Add "Dining" category
   
4. **Rule effective date scenarios**
   - Rule effective from Jan 1, 2024
   - Transaction from Dec 2023 → Rule not applied
   - Transaction from Feb 2024 → Rule applied

#### Error Scenarios to Test
1. **Database failures during execution**
2. **Invalid category references in actions**
3. **Malformed condition values**
4. **Transaction update failures**

## Estimated Total Implementation Time

- **Phase 1 (Core)**: 8-12 hours
- **Phase 2 (Error Handling)**: 4-7 hours  
- **Phase 3 (Integration)**: 3-5 hours
- **Total**: 15-24 hours

## Success Criteria

### Code Coverage Goals
- **Minimum**: 80% coverage of rule execution methods
- **Target**: 90% coverage of rule execution methods
- **Stretch**: 95% coverage including edge cases

### Test Quality Metrics
- All happy path scenarios covered
- All error conditions tested
- All business rules validated
- Performance characteristics verified
- Integration scenarios working

## Recommended Implementation Order

1. **Start with Phase 1.1** - Basic ExecuteRules functionality
2. **Move to Phase 1.2** - Condition evaluation (most complex logic)
3. **Implement Phase 1.3** - Action application
4. **Add Phase 1.4** - Effective date validation
5. **Continue with Phase 2** - Error handling
6. **Finish with Phase 3** - Integration tests

This approach ensures we have solid coverage of the core functionality first, then build up to more complex scenarios and edge cases.

## Decision Points

### Questions to Consider:
1. **How much time do we want to invest?** (Phases 1-2 vs all phases)
2. **What's our coverage target?** (80% vs 90% vs 95%)
3. **Should we include performance tests?** (Large datasets)
4. **Do we need concurrent access tests?** (Multi-user scenarios)
5. **How detailed should error testing be?** (Every error path vs major ones)

### Recommended Minimal Implementation:
If time is limited, implement **Phase 1 only** (8-12 hours) which covers:
- Basic execution functionality
- All condition evaluation logic
- All action application logic  
- Effective date validation

This provides solid coverage of the core business logic while being achievable in a reasonable timeframe.
