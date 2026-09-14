import { testCategory } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { CategoryAnalytics } from "./CategoryAnalytics";

const travelCategory = { id: 2, name: "Travel", created_by: 1 };

vi.mock("@/components/ui/icon-picker", () => ({
  Icon: () => null,
  IconPicker: () => null,
}));

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

const categoryData = [
  { category_id: 1, category_name: "Food", total_amount: -300 },
  { category_id: 2, category_name: "Travel", total_amount: -600 },
];

const recentTransaction = {
  id: 9,
  date: "2026-09-01T00:00:00.000Z",
  name: "Lunch",
  description: null,
  amount: 120,
  category_ids: [1],
  account_id: 1,
};

describe("CategoryAnalytics", () => {
  it("invites the user to create the first category", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CategoryAnalytics />);

    expect(screen.getByText("No categories yet")).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", { name: /add your first category/i })
    );

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("applies a category filter from the dropdown", async () => {
    const user = userEvent.setup();
    const onFilterChange = vi.fn();
    renderWithProviders(
      <CategoryAnalytics
        data={[]}
        categories={[testCategory]}
        onCategoryFilterChange={onFilterChange}
      />
    );

    expect(
      screen.getByText("No category activity for the selected filter.")
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "All categories" }));
    await user.click(
      await screen.findByRole("menuitemcheckbox", { name: "Food" })
    );
    const apply = screen.getByRole("button", { name: "Apply" });
    expect(apply).toBeEnabled();
    await user.click(apply);

    expect(onFilterChange).toHaveBeenCalledWith([1]);
  });

  it("renders shares and expands recent transactions", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/transaction", () =>
        HttpResponse.json({
          message: "ok",
          data: {
            transactions: [recentTransaction],
            total: 1,
            page: 1,
            page_size: 5,
          },
        })
      )
    );
    renderWithProviders(<CategoryAnalytics data={categoryData} />);

    expect(screen.getByText("Food:")).toBeInTheDocument();
    expect(screen.getByText("Travel:")).toBeInTheDocument();
    expect(screen.getByText("33.3%")).toBeInTheDocument();
    expect(screen.getByText("66.7%")).toBeInTheDocument();
    expect(screen.getByText("₹300.00")).toBeInTheDocument();
    expect(screen.getByText("₹600.00")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Expand Travel" }));

    expect(
      await screen.findByText("Latest 5 transactions")
    ).toBeInTheDocument();
    expect(await screen.findByText("Lunch")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "View all" })).toHaveAttribute(
      "href",
      "/transaction?category_id=2"
    );
  });

  it("hides the uncategorized pseudo row", () => {
    renderWithProviders(
      <CategoryAnalytics
        data={[
          {
            category_id: -1,
            category_name: "Uncategorized",
            total_amount: -50,
          },
        ]}
        categories={[testCategory]}
      />
    );

    expect(
      screen.getByText("No category activity for the selected filter.")
    ).toBeInTheDocument();
  });

  it("selects every category at once", async () => {
    const user = userEvent.setup();
    const onFilterChange = vi.fn();
    renderWithProviders(
      <CategoryAnalytics
        data={[]}
        categories={[testCategory, travelCategory]}
        onCategoryFilterChange={onFilterChange}
      />
    );

    await user.click(screen.getByRole("button", { name: "All categories" }));
    await user.click(await screen.findByRole("button", { name: "Select all" }));
    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(onFilterChange).toHaveBeenCalledWith([1, 2]);
  });

  it("clears the selection when everything is selected", async () => {
    const user = userEvent.setup();
    const onFilterChange = vi.fn();
    renderWithProviders(
      <CategoryAnalytics
        data={[]}
        categories={[testCategory, travelCategory]}
        selectedCategoryIds={[1, 2]}
        onCategoryFilterChange={onFilterChange}
      />
    );

    await user.click(screen.getByRole("button", { name: "All categories" }));
    expect(screen.getByText("Deselect all")).toBeInTheDocument();
    await user.click(
      await screen.findByRole("button", { name: "Deselect all" })
    );
    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(onFilterChange).toHaveBeenCalledWith([]);
  });

  it("collapses an expanded category again", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CategoryAnalytics data={categoryData} />);

    await user.click(screen.getByRole("button", { name: "Expand Travel" }));
    await user.click(
      await screen.findByRole("button", { name: "Collapse Travel" })
    );

    expect(
      screen.getByRole("button", { name: "Expand Travel" })
    ).toBeInTheDocument();
  });

  it("reports a failed transaction lookup", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/transaction", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );
    renderWithProviders(<CategoryAnalytics data={categoryData} />);

    await user.click(screen.getByRole("button", { name: "Expand Travel" }));

    expect(
      await screen.findByText("Failed to load transactions.")
    ).toBeInTheDocument();
  });

  it("closes the add dialog after creating a category", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/category", () =>
        HttpResponse.json(
          { data: { id: 9, name: "Travel", created_by: 1 } },
          { status: 201 }
        )
      )
    );
    renderWithProviders(<CategoryAnalytics data={categoryData} />);

    await user.click(screen.getByRole("button", { name: "Add category" }));
    await user.type(
      await screen.findByPlaceholderText("Enter category name"),
      "Travel"
    );
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument()
    );
  });

  it("filters from the data view", async () => {
    const user = userEvent.setup();
    const onFilterChange = vi.fn();
    renderWithProviders(
      <CategoryAnalytics
        data={categoryData}
        categories={[testCategory, travelCategory]}
        onCategoryFilterChange={onFilterChange}
      />
    );

    await user.click(screen.getByRole("button", { name: "All categories" }));
    await user.click(
      await screen.findByRole("menuitemcheckbox", { name: "Food" })
    );
    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(onFilterChange).toHaveBeenCalledWith([1]);
  });

  it("unchecks a selected category", async () => {
    const user = userEvent.setup();
    const onFilterChange = vi.fn();
    renderWithProviders(
      <CategoryAnalytics
        data={[]}
        categories={[testCategory, travelCategory]}
        selectedCategoryIds={[1]}
        onCategoryFilterChange={onFilterChange}
      />
    );

    await user.click(screen.getByRole("button", { name: "1 selected" }));
    await user.click(
      await screen.findByRole("menuitemcheckbox", { name: "Food" })
    );
    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(onFilterChange).toHaveBeenCalledWith([]);
  });

  it("creates the first category from the empty state", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/category", () =>
        HttpResponse.json(
          { data: { id: 9, name: "Travel", created_by: 1 } },
          { status: 201 }
        )
      )
    );
    renderWithProviders(<CategoryAnalytics />);

    await user.click(
      screen.getByRole("button", { name: /add your first category/i })
    );
    await user.type(
      await screen.findByPlaceholderText("Enter category name"),
      "Travel"
    );
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument()
    );
  });

  it("creates a category from the no-activity state", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/category", () =>
        HttpResponse.json(
          { data: { id: 9, name: "Travel", created_by: 1 } },
          { status: 201 }
        )
      )
    );
    renderWithProviders(
      <CategoryAnalytics data={[]} categories={[testCategory]} />
    );

    await screen.findByText("No category activity for the selected filter.");
    await user.click(screen.getByRole("button", { name: "Add category" }));
    await user.type(
      await screen.findByPlaceholderText("Enter category name"),
      "Travel"
    );
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument()
    );
  });
});
