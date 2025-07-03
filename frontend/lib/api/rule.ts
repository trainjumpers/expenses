import { apiRequest, authHeaders } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import {
  BaseRuleAction,
  BaseRuleCondition,
  CreateRuleInput,
  DescribeRuleResponse,
  ExecuteRulesResponse,
  Rule,
  RuleAction,
  RuleCondition,
  UpdateRuleInput,
} from "@/lib/models/rule";

// List all rules
export async function listRules(): Promise<Rule[]> {
  return apiRequest<Rule[]>(
    `${API_BASE_URL}/rule`,
    {
      headers: authHeaders(),
      credentials: "include",
    },
    "rule",
    [],
    "Failed to fetch rules"
  );
}

// Get a single rule with actions and conditions
export async function getRule(id: number): Promise<DescribeRuleResponse> {
  return apiRequest<DescribeRuleResponse>(
    `${API_BASE_URL}/rule/${id}`,
    {
      headers: authHeaders(),
      credentials: "include",
    },
    "rule",
    [],
    "Failed to fetch rule"
  );
}

// Create a new rule (with actions and conditions)
export async function createRule(input: CreateRuleInput): Promise<Rule> {
  return apiRequest<Rule>(
    `${API_BASE_URL}/rule`,
    {
      method: "POST",
      headers: authHeaders(),
      credentials: "include",
      body: JSON.stringify(input),
    },
    "rule",
    [],
    "Failed to create rule"
  );
}

// Update a rule (partial update)
export async function updateRule(
  id: number,
  input: UpdateRuleInput
): Promise<Rule> {
  return apiRequest<Rule>(
    `${API_BASE_URL}/rule/${id}`,
    {
      method: "PATCH",
      headers: authHeaders(),
      credentials: "include",
      body: JSON.stringify(input),
    },
    "rule",
    [],
    "Failed to update rule"
  );
}

// Delete a rule
export async function deleteRule(id: number): Promise<void> {
  return apiRequest<void>(
    `${API_BASE_URL}/rule/${id}`,
    {
      method: "DELETE",
      headers: authHeaders(),
      credentials: "include",
    },
    "rule",
    [],
    "Failed to delete rule"
  );
}

// List actions for a rule
export async function listRuleActions(ruleId: number): Promise<RuleAction[]> {
  return apiRequest<RuleAction[]>(
    `${API_BASE_URL}/rule/${ruleId}/actions`,
    {
      headers: authHeaders(),
      credentials: "include",
    },
    "rule",
    [],
    "Failed to fetch rule actions"
  );
}

// Update actions for a rule
export async function updateRuleActions(
  ruleId: number,
  actions: BaseRuleAction[]
): Promise<void> {
  return apiRequest<void>(
    `${API_BASE_URL}/rule/${ruleId}/actions`,
    {
      method: "PUT",
      headers: authHeaders(),
      credentials: "include",
      body: JSON.stringify(actions),
    },
    "rule",
    [],
    "Failed to update rule actions"
  );
}

// List conditions for a rule
export async function listRuleConditions(
  ruleId: number
): Promise<RuleCondition[]> {
  return apiRequest<RuleCondition[]>(
    `${API_BASE_URL}/rule/${ruleId}/conditions`,
    {
      headers: authHeaders(),
      credentials: "include",
    },
    "rule",
    [],
    "Failed to fetch rule conditions"
  );
}

// Update conditions for a rule
export async function updateRuleConditions(
  ruleId: number,
  conditions: BaseRuleCondition[]
): Promise<void> {
  return apiRequest<void>(
    `${API_BASE_URL}/rule/${ruleId}/conditions`,
    {
      method: "PUT",
      headers: authHeaders(),
      credentials: "include",
      body: JSON.stringify(conditions),
    },
    "rule",
    [],
    "Failed to update rule conditions"
  );
}

// Execute rules (example endpoint, adjust as needed)
export async function executeRules(): Promise<ExecuteRulesResponse> {
  return apiRequest<ExecuteRulesResponse>(
    `${API_BASE_URL}/rule/execute`,
    {
      method: "POST",
      headers: authHeaders(),
      credentials: "include",
    },
    "rule",
    [],
    "Failed to execute rules"
  );
}
