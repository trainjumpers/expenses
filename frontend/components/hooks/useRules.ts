import { createRule, deleteRule, listRules, updateRule } from "@/lib/api/rule";
import { CreateRuleInput, Rule, UpdateRuleInput } from "@/lib/models/rule";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export function useRules() {
  return useQuery<Rule[]>({
    queryKey: ["rules"],
    queryFn: listRules,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useCreateRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateRuleInput) => createRule(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["rules"] });
    },
  });
}

export function useDeleteRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteRule(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["rules"] });
    },
  });
}

export function useUpdateRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: UpdateRuleInput }) =>
      updateRule(id, input),
    onSuccess: (updatedRule) => {
      queryClient.setQueryData<Rule[]>(["rules"], (old) => {
        if (!old) return [updatedRule];
        return old.map((rule) =>
          rule.id === updatedRule.id ? updatedRule : rule
        );
      });
    },
  });
}
