import { testCategory } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { afterAll, describe, expect, it, vi } from "vitest";

import { UpdateCategoryModal } from "./UpdateCategoryModal";

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

function setup() {
  const onOpenChange = vi.fn();
  const onCategoryUpdated = vi.fn();
  renderWithProviders(
    <UpdateCategoryModal
      isOpen
      onOpenChange={onOpenChange}
      category={testCategory}
      onCategoryUpdated={onCategoryUpdated}
    />
  );
  return { onOpenChange, onCategoryUpdated };
}

describe("UpdateCategoryModal", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("saves the edited category", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.patch("*/api/v1/category/1", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json({
          data: { ...testCategory, name: "Groceries" },
        });
      })
    );
    const { onOpenChange, onCategoryUpdated } = setup();

    const name = screen.getByDisplayValue("Food");
    await user.clear(name);
    await user.type(name, "Groceries");
    await user.click(screen.getByRole("button", { name: "Update" }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith(
        "Category updated successfully"
      )
    );
    expect(body).toEqual({ name: "Groceries", icon: "circle-dashed" });
    expect(onCategoryUpdated).toHaveBeenCalled();
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("keeps the dialog open when the update fails", async () => {
    const user = userEvent.setup();
    server.use(
      http.patch("*/api/v1/category/1", () =>
        HttpResponse.json({ error: "duplicate" }, { status: 409 })
      )
    );
    const { onOpenChange } = setup();

    await user.click(screen.getByRole("button", { name: "Update" }));

    await waitFor(() =>
      expect(consoleError).toHaveBeenCalledWith(
        "Failed to update category:",
        expect.any(Error)
      )
    );
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });
});
