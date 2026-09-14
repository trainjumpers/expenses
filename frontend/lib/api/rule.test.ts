import { server } from "@/test/msw/server";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { listRuleActions, listRuleConditions } from "./rule";

describe("rule api", () => {
  it("lists the actions of a rule", async () => {
    server.use(
      http.get("*/api/v1/rule/1/actions", () =>
        HttpResponse.json({
          message: "ok",
          data: [
            { id: 1, rule_id: 1, action_type: "category", action_value: "1" },
          ],
        })
      )
    );

    await expect(listRuleActions(1)).resolves.toHaveLength(1);
  });

  it("lists the conditions of a rule", async () => {
    server.use(
      http.get("*/api/v1/rule/1/conditions", () =>
        HttpResponse.json({
          message: "ok",
          data: [
            {
              id: 1,
              rule_id: 1,
              condition_type: "name",
              condition_operator: "contains",
              condition_value: "coffee",
            },
          ],
        })
      )
    );

    await expect(listRuleConditions(1)).resolves.toHaveLength(1);
  });
});
