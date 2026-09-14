import type { RuleFieldType } from "@/lib/models/rule";
import { renderWithProviders } from "@/test/render";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { RuleInput } from "./RuleInput";

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/",
  useSearchParams: () => new URLSearchParams(),
}));

function ControlledInput({ fieldType }: { fieldType: RuleFieldType }) {
  const [value, setValue] = useState("");
  return (
    <RuleInput
      fieldType={fieldType}
      value={value}
      onChange={setValue}
      placeholder="Enter value"
    />
  );
}

describe("RuleInput", () => {
  it("sanitises the amount field", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ControlledInput fieldType="amount" />);

    const input = screen.getByRole("spinbutton");
    await user.type(input, "12.5x");

    expect(input).toHaveValue(12.5);
  });

  it("picks a category", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ControlledInput fieldType="category" />);

    const select = screen.getByRole("combobox");
    await user.selectOptions(
      select,
      await screen.findByRole("option", { name: "Food" })
    );

    expect(select).toHaveValue("1");
  });

  it("picks a transfer account", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ControlledInput fieldType="transfer" />);

    const select = screen.getByRole("combobox");
    await user.selectOptions(
      select,
      await screen.findByRole("option", { name: "HDFC Savings" })
    );

    expect(select).toHaveValue("1");
  });

  it("renders a free text input for other fields", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ControlledInput fieldType="name" />);

    const input = screen.getByPlaceholderText("Enter value");
    await user.type(input, "coffee");

    expect(input).toHaveValue("coffee");
  });
});
