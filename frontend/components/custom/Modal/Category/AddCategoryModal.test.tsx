import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { afterAll, describe, expect, it, vi } from "vitest";

import { AddCategoryModal } from "./AddCategoryModal";

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
  const onCategoryAdded = vi.fn();
  renderWithProviders(
    <AddCategoryModal
      isOpen
      onOpenChange={onOpenChange}
      onCategoryAdded={onCategoryAdded}
    />
  );
  return { onOpenChange, onCategoryAdded };
}

describe("AddCategoryModal", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("creates a category", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.post("*/api/v1/category", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json(
          {
            message: "Category created successfully",
            data: { id: 9, name: "Travel", created_by: 1 },
          },
          { status: 201 }
        );
      })
    );
    const { onOpenChange, onCategoryAdded } = setup();

    await user.type(
      screen.getByPlaceholderText("Enter category name"),
      "Travel"
    );
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith(
        "Category created successfully"
      )
    );
    expect(body).toEqual({ name: "Travel", icon: "circle-dashed" });
    expect(onCategoryAdded).toHaveBeenCalledWith(
      expect.objectContaining({ id: 9 })
    );
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("keeps the dialog open when creation fails", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/category", () =>
        HttpResponse.json({ error: "duplicate" }, { status: 409 })
      )
    );
    const { onOpenChange } = setup();

    await user.type(
      screen.getByPlaceholderText("Enter category name"),
      "Travel"
    );
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() =>
      expect(consoleError).toHaveBeenCalledWith(
        "Failed to create category:",
        expect.any(Error)
      )
    );
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });

  it("closes without saving", async () => {
    const user = userEvent.setup();
    const { onOpenChange } = setup();

    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
