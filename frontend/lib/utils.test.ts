import { describe, expect, it } from "vitest";

import {
  cn,
  formatCurrency,
  formatPercentage,
  formatShortCurrency,
  getSurplusTone,
  getTransactionColor,
  transformToChartData,
} from "./utils";

describe("cn", () => {
  it("merges conflicting tailwind classes", () => {
    expect(cn("px-2", "px-4")).toBe("px-4");
    expect(cn("text-sm", false && "hidden", "font-bold")).toBe(
      "text-sm font-bold"
    );
  });
});

describe("formatCurrency", () => {
  it("formats INR by default", () => {
    expect(formatCurrency(2000)).toBe("₹2,000.00");
  });

  it("formats other currencies", () => {
    expect(formatCurrency(12.5, "USD")).toBe("$12.50");
  });
});

describe("formatShortCurrency", () => {
  it("shortens crores and lakhs", () => {
    expect(formatShortCurrency(15000000)).toBe("₹1.5Cr");
    expect(formatShortCurrency(150000)).toBe("₹1.5L");
  });

  it("shortens thousands and keeps the sign", () => {
    expect(formatShortCurrency(1500)).toBe("₹1.5K");
    expect(formatShortCurrency(-1500)).toBe("-₹1.5K");
  });

  it("falls back to full currency for small or invalid amounts", () => {
    expect(formatShortCurrency(999)).toBe("₹999.00");
    expect(formatShortCurrency(Infinity)).toBe("₹∞");
  });
});

describe("formatPercentage", () => {
  it("formats one decimal place", () => {
    expect(formatPercentage(33.333)).toBe("33.3%");
    expect(formatPercentage(0)).toBe("0.0%");
  });
});

describe("transaction tones", () => {
  it("treats negative amounts as credits", () => {
    expect(getTransactionColor(-10)).toContain("emerald");
    expect(getTransactionColor(10)).toContain("rose");
  });

  it("treats negative surplus as a deficit", () => {
    expect(getSurplusTone(-10)).toContain("rose");
    expect(getSurplusTone(10)).toContain("emerald");
  });
});

describe("transformToChartData", () => {
  it("maps the series with short and long dates", () => {
    expect(
      transformToChartData([
        { date: "2026-09-01T00:00:00.000Z", cash_balance: 120 },
      ])
    ).toEqual([
      {
        date: "Sep 01",
        value: 120,
        formattedDate: "Sep 01, 2026",
      },
    ]);
  });
});
