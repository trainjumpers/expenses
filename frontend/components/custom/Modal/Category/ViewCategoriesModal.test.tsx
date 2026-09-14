import { testCategory } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

import { ViewCategoriesModal } from "./ViewCategoriesModal";

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

const travelCategory = { id: 2, name: "Travel", created_by: 1 } as const;

function render() {
  return renderWithProviders(
    <ViewCategoriesModal isOpen onOpenChange={vi.fn()} />
  );
}

describe("ViewCategoriesModal", () => {
  it("lists the available categories", async () => {
    render();

    expect(await screen.findByText("Food")).toBeInTheDocument();
  });

  it("filters categories by the search term", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/category", () =>
        HttpResponse.json({
          message: "ok",
          data: [testCategory, travelCategory],
        })
      )
    );
    render();
    await screen.findByText("Food");

    await user.type(screen.getByLabelText("Search categories"), "travel");

    await waitFor(() =>
      expect(screen.queryByText("Food")).not.toBeInTheDocument()
    );
    expect(screen.getByText("Travel")).toBeInTheDocument();
  });

  it("shows an empty state without categories", async () => {
    server.use(
      http.get("*/api/v1/category", () =>
        HttpResponse.json({ message: "ok", data: [] })
      )
    );
    render();

    expect(
      await screen.findByText(
        "No categories found. Add your first category to get started."
      )
    ).toBeInTheDocument();
  });

  it("deletes a category after confirming", async () => {
    const user = userEvent.setup();
    server.use(
      http.delete(
        "*/api/v1/category/1",
        () => new HttpResponse(null, { status: 204 })
      )
    );
    render();
    await screen.findByText("Food");

    const view = screen.getByRole("dialog", { name: "Categories" });
    await user.click(within(view).getByRole("button", { name: "Delete" }));

    const confirm = await screen.findByRole("dialog", {
      name: "Delete Category",
    });
    await user.click(within(confirm).getByRole("button", { name: "Confirm" }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith(
        "Category deleted successfully"
      )
    );
  });

  it("opens the edit dialog for a category", async () => {
    const user = userEvent.setup();
    render();
    await screen.findByText("Food");

    const view = screen.getByRole("dialog", { name: "Categories" });
    await user.click(within(view).getByRole("button", { name: "Edit" }));

    expect(
      await screen.findByRole("dialog", { name: "Update Category" })
    ).toBeInTheDocument();
  });

  it("pages through the categories", async () => {
    const user = userEvent.setup();
    const many = Array.from({ length: 6 }, (_, i) => ({
      id: i + 1,
      name: `Category ${i + 1}`,
      created_by: 1,
    }));
    server.use(
      http.get("*/api/v1/category", () =>
        HttpResponse.json({ message: "ok", data: many })
      )
    );
    render();
    await screen.findByText("Category 1");

    await user.click(screen.getByText("2"));
    expect(await screen.findByText("Category 6")).toBeInTheDocument();

    await user.click(screen.getByText("Previous"));
    expect(await screen.findByText("Category 1")).toBeInTheDocument();

    await user.click(screen.getByText("Next"));
    expect(await screen.findByText("Category 6")).toBeInTheDocument();
  });

  it("opens the create dialog from the list", async () => {
    const user = userEvent.setup();
    render();
    await screen.findByText("Food");

    await user.click(screen.getByRole("button", { name: "Add New Category" }));

    expect(
      await screen.findByRole("dialog", { name: "Add Category" })
    ).toBeInTheDocument();
  });

  it("closes the edit and delete dialogs", async () => {
    const user = userEvent.setup();
    render();
    await screen.findByText("Food");

    await user.click(screen.getByRole("button", { name: "Edit" }));
    await screen.findByRole("dialog", { name: "Update Category" });
    await user.keyboard("{Escape}");
    await waitFor(() =>
      expect(
        screen.queryByRole("dialog", { name: "Update Category" })
      ).not.toBeInTheDocument()
    );

    await user.click(screen.getByRole("button", { name: "Delete" }));
    const confirm = await screen.findByRole("dialog", {
      name: "Delete Category",
    });
    await user.click(within(confirm).getByRole("button", { name: "Cancel" }));
    await waitFor(() =>
      expect(
        screen.queryByRole("dialog", { name: "Delete Category" })
      ).not.toBeInTheDocument()
    );
  });
});
