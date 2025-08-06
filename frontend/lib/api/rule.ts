import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import {
  BaseRuleAction,
  BaseRuleCondition,
  CreateRuleInput,
  DescribeRuleResponse,
  ExecuteRulesResponse,
  PaginatedRulesResponse,
  Rule,
  RuleAction,
  RuleCondition,
  RuleListQuery,
  UpdateRuleInput,
} from "@/lib/models/rule";

// List rules with optional pagination and search
export async function listRules(
  query?: RuleListQuery
): Promise<PaginatedRulesResponse> {
  const params = new URLSearchParams();
  
  if (query?.page) {
    params.append("page", query.page.toString());
  }
  if (query?.page_size) {
    params.append("page_size", query.page_size.toString());
  }
  if (query?.search) {
    params.append("search", query.search);
  }

  const url = `${API_BASE_URL}/rule${params.toString() ? `?${params.toString()}` : ""}`;
  
  return apiRequest<PaginatedRulesResponse>(
    url,
    {
      credentials: "include",
    },
    "data",
    [],
    "Failed to fetch rules"
  );
}

// Get a single rule with actions and conditions
export async function getRule(id: number): Promise<DescribeRuleResponse> {
  return apiRequest<DescribeRuleResponse>(
    `${API_BASE_URL}/rule/${id}`,
    {
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
): Promise<RuleAction[]> {
  return apiRequest<RuleAction[]>(
    `${API_BASE_URL}/rule/${ruleId}/actions`,
    {
      method: "PUT",
      credentials: "include",
      body: JSON.stringify({ actions }),
    },
    "actions",
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
): Promise<RuleCondition[]> {
  return apiRequest<RuleCondition[]>(
    `${API_BASE_URL}/rule/${ruleId}/conditions`,
    {
      method: "PUT",
      credentials: "include",
      body: JSON.stringify({ conditions }),
    },
    "conditions",
    [],
    "Failed to update rule conditions"
  );
}

// Execute rules (example endpoint, adjust as needed)
export async function executeRules(payload?: {
  transaction_ids?: number[];
}): Promise<ExecuteRulesResponse> {
  return apiRequest<ExecuteRulesResponse>(
    `${API_BASE_URL}/rule/execute`,
    {
      method: "POST",
      credentials: "include",
      body: payload ? JSON.stringify(payload) : undefined,
    },
    "rule",
    [],
    "Failed to execute rules"
  );
}
