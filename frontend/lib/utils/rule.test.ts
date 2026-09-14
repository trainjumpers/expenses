import { describe, expect, it } from "vitest";

import {
  getEffectiveScopeAndDate,
  normalizeRuleActions,
  normalizeRuleConditions,
} from "./rule";

const categories = [
  { id: 1, name: "Food" },
  { id: 2, name: "Travel" },
];
const accounts = [{ id: 7, name: "HDFC Savings" }];

describe("getEffectiveScopeAndDate", () => {
  it("treats a missing date as all time", () => {
    expect(getEffectiveScopeAndDate(undefined)).toEqual({
      effectiveScope: "all",
      effectiveFromDate: undefined,
    });
  });

  it("treats an invalid date as all time", () => {
    expect(getEffectiveScopeAndDate("not-a-date")).toEqual({
      effectiveScope: "all",
      effectiveFromDate: undefined,
    });
  });

  it("treats the epoch as all time", () => {
    expect(getEffectiveScopeAndDate(new Date(0).toISOString())).toEqual({
      effectiveScope: "all",
      effectiveFromDate: undefined,
    });
  });

  it("keeps a real date as the starting point", () => {
    const { effectiveScope, effectiveFromDate } = getEffectiveScopeAndDate(
      "2026-09-01T00:00:00.000Z"
    );

    expect(effectiveScope).toBe("from");
    expect(effectiveFromDate?.toISOString()).toBe("2026-09-01T00:00:00.000Z");
  });
});

describe("normalizeRuleActions", () => {
  it("maps category and transfer values to ids", () => {
    expect(
      normalizeRuleActions(
        [
          { action_type: "category", action_value: "Food" },
          { action_type: "transfer", action_value: "HDFC Savings" },
          { action_type: "name", action_value: "Renamed" },
        ],
        categories,
        accounts
      )
    ).toEqual([
      { action_type: "category", action_value: "1" },
      { action_type: "transfer", action_value: "7" },
      { action_type: "name", action_value: "Renamed" },
    ]);
  });

  it("keeps unknown values untouched", () => {
    expect(
      normalizeRuleActions(
        [{ action_type: "category", action_value: "Unknown" }],
        categories
      )
    ).toEqual([{ action_type: "category", action_value: "Unknown" }]);
  });
});

describe("normalizeRuleConditions", () => {
  it("maps category and transfer values to ids", () => {
    expect(
      normalizeRuleConditions(
        [
          {
            condition_type: "category",
            condition_operator: "equals",
            condition_value: "2",
          },
          {
            condition_type: "transfer",
            condition_operator: "equals",
            condition_value: "HDFC Savings",
          },
          {
            condition_type: "name",
            condition_operator: "contains",
            condition_value: "coffee",
          },
        ],
        categories,
        accounts
      )
    ).toEqual([
      {
        condition_type: "category",
        condition_operator: "equals",
        condition_value: "2",
      },
      {
        condition_type: "transfer",
        condition_operator: "equals",
        condition_value: "7",
      },
      {
        condition_type: "name",
        condition_operator: "contains",
        condition_value: "coffee",
      },
    ]);
  });
});
