# Rule Execution Service Test Plan

## Overview
This document outlines comprehensive tests for the rule execution functionality in the `RuleService`. The tests are organized by functionality and cover various scenarios including success cases, error cases, edge cases, and integration scenarios.

## Test Categories

### 1. ExecuteRules Method Tests

#### 1.1 Basic Functionality Tests
- **Test**: Execute rules with no rules available
  - **Scenario**: User has no rules defined
  - **Expected**: Return empty result with zero counts
  - **Priority**: High

- **Test**: Execute rules with no transactions available
  - **Scenario**: Rules exist but no transactions to process
  - **Expected**: Return empty result with rule count but zero processed transactions
  - **Priority**: High

- **Test**: Execute rules with both rules and transactions
  - **Scenario**: Normal execution with matching conditions
  - **Expected**: Return modified transactions with applied rules
  - **Priority**: High

#### 1.2 Request Parameter Tests
- **Test**: Execute specific rules by IDs
  - **Scenario**: Request contains specific rule IDs
  - **Expected**: Only specified rules are executed
  - **Priority**: High

- **Test**: Execute rules on specific transactions
  - **Scenario**: Request contains specific transaction IDs
  - **Expected**: Only specified transactions are processed
  - **Priority**: High

- **Test**: Execute with custom page size
  - **Scenario**: Request specifies page size
  - **Expected**: Transactions processed in specified batch size
  - **Priority**: Medium

- **Test**: Execute with invalid page size (too large)
  - **Scenario**: Page size > 1000
  - **Expected**: Default to maximum allowed (1000)
  - **Priority**: Medium

- **Test**: Execute with invalid page size (zero/negative)
  - **Scenario**: Page size <= 0
  - **Expected**: Default to 100
  - **Priority**: Medium

#### 1.3 Pagination Tests
- **Test**: Execute rules with multiple pages of transactions
  - **Scenario**: More transactions than page size
  - **Expected**: All transactions processed across multiple pages
  - **Priority**: High

- **Test**: Execute rules with exact page size match
  - **Scenario**: Transaction count equals page size
  - **Expected**: All transactions processed in single page
  - **Priority**: Medium

### 2. ExecuteRulesForTransaction Method Tests

#### 2.1 Basic Functionality Tests
- **Test**: Execute rules for existing transaction
  - **Scenario**: Valid transaction ID and user ID
  - **Expected**: Rules applied to single transaction
  - **Priority**: High

- **Test**: Execute rules for non-existent transaction
  - **Scenario**: Invalid transaction ID
  - **Expected**: Return appropriate error
  - **Priority**: High

- **Test**: Execute rules for transaction of different user
  - **Scenario**: Transaction belongs to different user
  - **Expected**: Return appropriate error
  - **Priority**: High

### 3. ExecuteRulesForRule Method Tests

#### 3.1 Basic Functionality Tests
- **Test**: Execute specific rule for all transactions
  - **Scenario**: Valid rule ID and user ID
  - **Expected**: Specific rule applied to all matching transactions
  - **Priority**: High

- **Test**: Execute non-existent rule
  - **Scenario**: Invalid rule ID
  - **Expected**: Return empty result or appropriate error
  - **Priority**: High

- **Test**: Execute rule of different user
  - **Scenario**: Rule belongs to different user
  - **Expected**: Return empty result or appropriate error
  - **Priority**: High

### 4. Rule Condition Evaluation Tests

#### 4.1 Amount Field Tests
- **Test**: Amount equals condition
  - **Scenario**: Transaction amount exactly matches condition value
  - **Expected**: Condition evaluates to true
  - **Priority**: High

- **Test**: Amount greater than condition
  - **Scenario**: Transaction amount is greater than condition value
  - **Expected**: Condition evaluates to true for 'greater' operator
  - **Priority**: High

- **Test**: Amount less than condition
  - **Scenario**: Transaction amount is less than condition value
  - **Expected**: Condition evaluates to true for 'lower' operator
  - **Priority**: High

- **Test**: Amount condition with invalid value
  - **Scenario**: Condition value is not a valid number
  - **Expected**: Condition evaluates to false
  - **Priority**: Medium

#### 4.2 Name Field Tests
- **Test**: Name equals condition (case insensitive)
  - **Scenario**: Transaction name matches condition value
  - **Expected**: Condition evaluates to true
  - **Priority**: High

- **Test**: Name contains condition
  - **Scenario**: Transaction name contains condition substring
  - **Expected**: Condition evaluates to true for 'contains' operator
  - **Priority**: High

- **Test**: Name condition with empty transaction name
  - **Scenario**: Transaction has empty name
  - **Expected**: Condition evaluates appropriately
  - **Priority**: Medium

#### 4.3 Description Field Tests
- **Test**: Description equals condition
  - **Scenario**: Transaction description matches condition value
  - **Expected**: Condition evaluates to true
  - **Priority**: High

- **Test**: Description contains condition
  - **Scenario**: Transaction description contains condition substring
  - **Expected**: Condition evaluates to true
  - **Priority**: High

- **Test**: Description condition with nil description
  - **Scenario**: Transaction has nil description
  - **Expected**: Condition evaluates against empty string
  - **Priority**: High

#### 4.4 Category Field Tests
- **Test**: Category equals condition
  - **Scenario**: Transaction has matching category ID
  - **Expected**: Condition evaluates to true
  - **Priority**: High

- **Test**: Category condition with multiple categories
  - **Scenario**: Transaction has multiple categories, one matches
  - **Expected**: Condition evaluates to true
  - **Priority**: High

- **Test**: Category condition with no matching categories
  - **Scenario**: Transaction categories don't match condition
  - **Expected**: Condition evaluates to false
  - **Priority**: High

- **Test**: Category condition with invalid category ID
  - **Scenario**: Condition value is not a valid category ID
  - **Expected**: Condition evaluates to false
  - **Priority**: Medium

#### 4.5 Multiple Conditions Tests
- **Test**: All conditions match (AND logic)
  - **Scenario**: Rule has multiple conditions, all match transaction
  - **Expected**: Rule is applied
  - **Priority**: High

- **Test**: Some conditions match, some don't
  - **Scenario**: Rule has multiple conditions, only some match
  - **Expected**: Rule is not applied
  - **Priority**: High

- **Test**: No conditions match
  - **Scenario**: Rule has multiple conditions, none match
  - **Expected**: Rule is not applied
  - **Priority**: High

### 5. Rule Action Application Tests

#### 5.1 Name Action Tests
- **Test**: Apply name action to transaction
  - **Scenario**: Rule action updates transaction name
  - **Expected**: Transaction name is updated
  - **Priority**: High

- **Test**: Name action with existing name (first-rule-wins)
  - **Scenario**: Multiple rules try to update same transaction name
  - **Expected**: First rule wins, subsequent rules don't override
  - **Priority**: High

#### 5.2 Description Action Tests
- **Test**: Apply description action to transaction
  - **Scenario**: Rule action updates transaction description
  - **Expected**: Transaction description is updated
  - **Priority**: High

- **Test**: Description action with existing description
  - **Scenario**: Multiple rules try to update same transaction description
  - **Expected**: First rule wins
  - **Priority**: High

#### 5.3 Category Action Tests
- **Test**: Apply category action to transaction
  - **Scenario**: Rule action adds category to transaction
  - **Expected**: Category is added to transaction
  - **Priority**: High

- **Test**: Category action with existing categories (additive)
  - **Scenario**: Transaction already has categories, rule adds more
  - **Expected**: New categories are added, existing ones preserved
  - **Priority**: High

- **Test**: Category action with duplicate category
  - **Scenario**: Rule tries to add category that already exists
  - **Expected**: No duplicate categories added
  - **Priority**: Medium

- **Test**: Category action with invalid category ID
  - **Scenario**: Rule action references non-existent category
  - **Expected**: Action is skipped with warning
  - **Priority**: High

- **Test**: Category action with category from different user
  - **Scenario**: Rule references category belonging to different user
  - **Expected**: Action is skipped with warning
  - **Priority**: High

### 6. Rule Effective Date Tests

#### 6.1 Date Validation Tests
- **Test**: Rule effective date before transaction date
  - **Scenario**: Rule is effective before transaction occurred
  - **Expected**: Rule is applied
  - **Priority**: High

- **Test**: Rule effective date after transaction date
  - **Scenario**: Rule is effective after transaction occurred
  - **Expected**: Rule is not applied
  - **Priority**: High

- **Test**: Rule effective date equals transaction date
  - **Scenario**: Rule effective date is same as transaction date
  - **Expected**: Rule is applied
  - **Priority**: High

- **Test**: Rule effective date in future
  - **Scenario**: Rule effective date is in the future
  - **Expected**: Rule is not applied to any transactions
  - **Priority**: High

### 7. Error Handling Tests

#### 7.1 Database Error Tests
- **Test**: Database error when fetching rules
  - **Scenario**: Database returns error when listing rules
  - **Expected**: Return appropriate error
  - **Priority**: High

- **Test**: Database error when fetching transactions
  - **Scenario**: Database returns error when listing transactions
  - **Expected**: Return appropriate error
  - **Priority**: High

- **Test**: Database error when updating transaction
  - **Scenario**: Database returns error during transaction update
  - **Expected**: Transaction is skipped with error reason
  - **Priority**: High

#### 7.2 Validation Error Tests
- **Test**: Invalid rule ID in request
  - **Scenario**: Request contains non-existent rule ID
  - **Expected**: Rule is skipped with warning
  - **Priority**: Medium

- **Test**: Invalid transaction ID in request
  - **Scenario**: Request contains non-existent transaction ID
  - **Expected**: Transaction is skipped with warning
  - **Priority**: Medium

### 8. Integration Tests

#### 8.1 End-to-End Scenarios
- **Test**: Complete rule execution workflow
  - **Scenario**: Create rule, create transaction, execute rules
  - **Expected**: Transaction is modified according to rule
  - **Priority**: High

- **Test**: Multiple rules on single transaction
  - **Scenario**: Multiple rules match same transaction
  - **Expected**: All applicable rules are applied correctly
  - **Priority**: High

- **Test**: Single rule on multiple transactions
  - **Scenario**: One rule matches multiple transactions
  - **Expected**: All matching transactions are modified
  - **Priority**: High

- **Test**: Complex scenario with mixed results
  - **Scenario**: Some transactions modified, some skipped
  - **Expected**: Correct modified and skipped results returned
  - **Priority**: High

#### 8.2 Performance Tests
- **Test**: Large number of transactions
  - **Scenario**: Execute rules on 1000+ transactions
  - **Expected**: Efficient processing with proper batching
  - **Priority**: Medium

- **Test**: Large number of rules
  - **Scenario**: Execute 100+ rules on transactions
  - **Expected**: Efficient rule evaluation
  - **Priority**: Medium

### 9. Edge Case Tests

#### 9.1 Boundary Conditions
- **Test**: Empty rule conditions
  - **Scenario**: Rule has no conditions defined
  - **Expected**: Rule is not applied (no conditions = false)
  - **Priority**: Medium

- **Test**: Empty rule actions
  - **Scenario**: Rule has no actions defined
  - **Expected**: Rule evaluation succeeds but no changes made
  - **Priority**: Medium

- **Test**: Transaction with all nil/empty fields
  - **Scenario**: Transaction has minimal data
  - **Expected**: Rules evaluate correctly against empty values
  - **Priority**: Medium

#### 9.2 Concurrent Access Tests
- **Test**: Concurrent rule execution
  - **Scenario**: Multiple rule executions running simultaneously
  - **Expected**: No data corruption or race conditions
  - **Priority**: Low (if time permits)

### 10. Helper Method Tests

#### 10.1 Utility Function Tests
- **Test**: getRulesForExecution with specific IDs
  - **Scenario**: Test rule filtering by IDs
  - **Expected**: Only specified rules returned
  - **Priority**: Medium

- **Test**: getSpecificTransactions with IDs
  - **Scenario**: Test transaction filtering by IDs
  - **Expected**: Only specified transactions returned
  - **Priority**: Medium

- **Test**: categoryExists validation
  - **Scenario**: Test category existence checking
  - **Expected**: Correct validation results
  - **Priority**: Medium

- **Test**: deduplicateFields functionality
  - **Scenario**: Test field deduplication
  - **Expected**: Duplicate fields removed
  - **Priority**: Low

## Test Implementation Priority

### High Priority (Must Implement)
1. Basic ExecuteRules functionality
2. Rule condition evaluation (all field types)
3. Rule action application (all action types)
4. Error handling for database errors
5. Rule effective date validation
6. End-to-end integration scenarios

### Medium Priority (Should Implement)
1. Request parameter validation
2. Pagination testing
3. Edge cases and boundary conditions
4. Helper method testing
5. Performance scenarios

### Low Priority (Nice to Have)
1. Concurrent access testing
2. Utility function edge cases
3. Complex integration scenarios

## Test Structure Recommendation

```go
Describe("RuleService - Rule Execution", func() {
    Describe("ExecuteRules", func() {
        Context("when no rules exist", func() {
            It("should return empty result")
        })
        Context("when rules exist but no transactions", func() {
            It("should return result with rule count but zero processed")
        })
        // ... more contexts
    })
    
    Describe("Rule Condition Evaluation", func() {
        Context("Amount conditions", func() {
            It("should evaluate equals condition correctly")
            It("should evaluate greater than condition correctly")
            // ... more tests
        })
        // ... more contexts for other field types
    })
    
    Describe("Rule Action Application", func() {
        Context("Name actions", func() {
            It("should update transaction name")
            It("should respect first-rule-wins policy")
        })
        // ... more contexts
    })
})
```

This comprehensive test plan ensures thorough coverage of all rule execution functionality, from basic operations to complex edge cases and error scenarios.
