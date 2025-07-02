export type RuleFieldType = "amount" | "name" | "description" | "category";
export type RuleOperator = "equals" | "contains" | "greater" | "lower";

export const RULE_FIELD_TYPES: { label: string; value: RuleFieldType }[] = [
  { label: "Amount", value: "amount" },
  { label: "Name", value: "name" },
  { label: "Description", value: "description" },
  { label: "Category", value: "category" },
];

export const RULE_OPERATORS: { label: string; value: RuleOperator }[] = [
  { label: "Equals", value: "equals" },
  { label: "Contains", value: "contains" },
  { label: "Greater", value: "greater" },
  { label: "Lower", value: "lower" },
];

// --- Rule Creation ---

export interface CreateBaseRuleInput {
  name: string;
  description?: string;
  effective_from: string; // ISO date string
  created_by: number;
}

export interface CreateRuleActionInput {
  action_type: RuleFieldType;
  action_value: string;
}

export interface CreateRuleConditionInput {
  condition_type: RuleFieldType;
  condition_value: string;
  condition_operator: RuleOperator;
}

export interface CreateRuleInput {
  rule: CreateBaseRuleInput;
  actions: CreateRuleActionInput[];
  conditions: CreateRuleConditionInput[];
}

// --- Rule Update ---

export interface UpdateRuleInput {
  name?: string;
  description?: string;
  effective_from?: string;
}

export interface UpdateRuleActionInput {
  action_type?: RuleFieldType;
  action_value?: string;
}

export interface UpdateRuleConditionInput {
  condition_type?: RuleFieldType;
  condition_value?: string;
  condition_operator?: RuleOperator;
}

// --- Rule Response Types ---

export interface Rule {
  id: number;
  name: string;
  description?: string;
  effective_from: string;
  created_by: number;
}

export interface RuleAction {
  id: number;
  rule_id: number;
  action_type: RuleFieldType;
  action_value: string;
}

export interface RuleCondition {
  id: number;
  rule_id: number;
  condition_type: RuleFieldType;
  condition_value: string;
  condition_operator: RuleOperator;
}

export interface DescribeRuleResponse {
  rule: Rule;
  actions: RuleAction[];
  conditions: RuleCondition[];
}

// --- Execute Rules Response ---

export interface ExecuteRulesResponse {
  modified: ModifiedResult[];
  skipped: SkippedResult[];
}

export interface ModifiedResult {
  transaction_id: number;
  applied_rules: number[];
  updated_fields: RuleFieldType[];
}

export interface SkippedResult {
  transaction_id: number;
  reason: string;
}
