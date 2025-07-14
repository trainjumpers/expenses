# Rule Execution Endpoint

## Overview

The `/rule/execute` endpoint has been implemented to provide automated rule execution functionality. This endpoint allows users to execute rules against transactions in a scalable and extensible manner.

## Endpoint Details

**URL:** `POST /api/v1/rule/execute`  
**Authentication:** Required (JWT token)  
**Content-Type:** `application/json`

## Request Body

```json
{
  "rule_ids": [1, 2, 3],           // Optional: specific rule IDs to execute
  "transaction_ids": [10, 20, 30], // Optional: specific transaction IDs to process
  "page_size": 100                 // Optional: batch size for processing (default: 100, max: 1000)
}
```

### Request Parameters

- `rule_ids` (optional): Array of rule IDs to execute. If not provided, all active rules for the user will be executed.
- `transaction_ids` (optional): Array of transaction IDs to process. If not provided, all transactions will be processed in batches.
- `page_size` (optional): Number of transactions to process per batch. Defaults to 100, maximum 1000.

## Response

```json
{
  "status": "success",
  "message": "Rules executed successfully",
  "data": {
    "modified": [
      {
        "transaction_id": 10,
        "applied_rules": [1, 2],
        "updated_fields": ["name", "category"]
      }
    ],
    "skipped": [
      {
        "transaction_id": 20,
        "reason": "failed to apply updates: validation error"
      }
    ],
    "total_rules": 5,
    "processed_transactions": 150
  }
}
```

### Response Fields

- `modified`: Array of transactions that were successfully modified
  - `transaction_id`: ID of the modified transaction
  - `applied_rules`: Array of rule IDs that were applied
  - `updated_fields`: Array of fields that were updated
- `skipped`: Array of transactions that were skipped
  - `transaction_id`: ID of the skipped transaction
  - `reason`: Reason why the transaction was skipped
- `total_rules`: Total number of rules that were considered for execution
- `processed_transactions`: Total number of transactions that were processed

## Usage Examples

### Execute All Rules on All Transactions

```bash
curl -X POST http://localhost:8080/api/v1/rule/execute \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

### Execute Specific Rules

```bash
curl -X POST http://localhost:8080/api/v1/rule/execute \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "rule_ids": [1, 2, 3],
    "page_size": 50
  }'
```

### Execute Rules on Specific Transactions

```bash
curl -X POST http://localhost:8080/api/v1/rule/execute \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_ids": [10, 20, 30, 40, 50]
  }'
```

## Rule Execution Logic

### Rule Conditions

Rules are executed based on the following field types and operators:

#### Field Types
- `amount`: Transaction amount
- `name`: Transaction name
- `description`: Transaction description
- `category`: Transaction category

#### Operators
- `equals`: Exact match
- `contains`: Substring match (for text fields)
- `greater`: Greater than (for amount)
- `lower`: Less than (for amount)

### Rule Actions

When conditions are met, the following actions can be performed:

- `name`: Update transaction name
- `description`: Update transaction description
- `category`: Add category to transaction (preserves existing categories)

### Execution Rules

1. **Effective Date**: Rules are only applied to transactions on or after the rule's `effective_from` date
2. **Condition Evaluation**: ALL conditions in a rule must be true for the rule to be applied
3. **Action Priority**: If multiple rules would update the same field, the first rule wins (no overwriting)
4. **Category Handling**: Category actions add to existing categories rather than replacing them
5. **Error Handling**: Individual transaction failures don't stop the entire execution

## Architecture

### Scalability Features

1. **Batch Processing**: Transactions are processed in configurable batches to manage memory usage
2. **Pagination**: Large transaction sets are processed page by page
3. **Selective Execution**: Can target specific rules or transactions for efficiency
4. **Error Isolation**: Individual transaction failures don't affect other transactions

### Extensibility Features

1. **Modular Design**: Rule execution logic is separated into its own service
2. **Interface-Based**: Uses interfaces for easy testing and mocking
3. **Condition Engine**: Easy to add new field types and operators
4. **Action Engine**: Easy to add new action types

### Future Extensions

The architecture supports easy addition of:

- **New Field Types**: Add to `RuleFieldType` enum and implement evaluation logic
- **New Operators**: Add to `RuleOperator` enum and implement comparison logic
- **New Actions**: Add action types and implement application logic
- **Async Processing**: Can be extended to run as background jobs
- **Webhooks**: Can trigger external systems after rule execution

## Integration Points

### Automatic Execution Triggers

The rule execution service provides methods for automatic triggering:

1. **New Transaction**: `ExecuteRulesForTransaction(transactionId, userId)`
2. **New Rule**: `ExecuteRulesForRule(ruleId, userId)`
3. **Bulk Execution**: `ExecuteRules(userId, request)`

These can be integrated into:
- Transaction creation workflow
- Rule creation workflow
- Scheduled batch jobs
- Manual user-triggered execution

## Error Handling

The endpoint handles various error scenarios:

1. **Authentication Errors**: Returns 401 if JWT token is invalid
2. **Validation Errors**: Returns 400 for invalid request parameters
3. **Rule Not Found**: Logs warning and continues with other rules
4. **Transaction Not Found**: Logs warning and continues with other transactions
5. **Database Errors**: Returns 500 with appropriate error message
6. **Category Validation**: Skips invalid category assignments with logging

## Performance Considerations

1. **Batch Size**: Default 100 transactions per batch, configurable up to 1000
2. **Memory Usage**: Processes transactions in batches to avoid memory issues
3. **Database Queries**: Optimized to minimize database round trips
4. **Logging**: Comprehensive logging for monitoring and debugging
5. **Transaction Safety**: Uses database transactions where appropriate

## Testing

The implementation includes comprehensive test coverage:

1. **Unit Tests**: Individual service method testing
2. **Integration Tests**: End-to-end workflow testing
3. **Mock Testing**: External dependency mocking
4. **Error Scenario Testing**: Various failure condition testing

## Security

1. **Authentication**: Requires valid JWT token
2. **Authorization**: Users can only execute rules on their own data
3. **Input Validation**: All request parameters are validated
4. **SQL Injection Protection**: Uses parameterized queries
5. **Rate Limiting**: Can be added at the API gateway level
