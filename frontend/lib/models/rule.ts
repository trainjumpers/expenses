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

export interface BaseRule {
  name: string;
  description?: string;
  effective_from: string;
}

export interface BaseRuleAction {
  action_type: RuleFieldType;
  action_value: string;
}

export interface BaseRuleCondition {
  condition_type: RuleFieldType;
  condition_value: string;
  condition_operator: RuleOperator;
}

export type CreateBaseRuleInput = BaseRule;
export type CreateRuleActionInput = BaseRuleAction;
export type CreateRuleConditionInput = BaseRuleCondition;

export interface CreateRuleInput {
  rule: CreateBaseRuleInput;
  actions: CreateRuleActionInput[];
  conditions: CreateRuleConditionInput[];
}

export type UpdateRuleInput = Partial<CreateBaseRuleInput>;
export type UpdateRuleActionInput = Partial<CreateRuleActionInput>;
export type UpdateRuleConditionInput = Partial<CreateRuleConditionInput>;

export interface Rule extends BaseRule {
  id: number;
  created_by: number;
}

export interface RuleAction extends BaseRuleAction {
  id: number;
  rule_id: number;
}

export interface RuleCondition extends BaseRuleCondition {
  id: number;
  rule_id: number;
}

export interface DescribeRuleResponse {
  rule: Rule;
  actions: RuleAction[];
  conditions: RuleCondition[];
}

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
