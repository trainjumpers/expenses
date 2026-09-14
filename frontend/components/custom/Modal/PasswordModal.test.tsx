import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { afterAll, describe, expect, it, vi } from "vitest";

import { PasswordModal } from "./PasswordModal";

function setup() {
  const onOpenChange = vi.fn();
  renderWithProviders(<PasswordModal isOpen onOpenChange={onOpenChange} />);
  return { onOpenChange };
}

async function fillPasswords(
  user: ReturnType<typeof userEvent.setup>,
  {
    current = "password123",
    next = "password456",
    confirm = "password456",
  } = {}
) {
  await user.type(screen.getByLabelText("Current"), current);
  await user.type(screen.getByLabelText("New"), next);
  await user.type(screen.getByLabelText("Confirm"), confirm);
}

describe("PasswordModal", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("rejects mismatched passwords", async () => {
    const user = userEvent.setup();
    setup();
    await fillPasswords(user, { confirm: "password789" });

    await user.click(screen.getByRole("button", { name: /update password/i }));

    expect(toast.error).toHaveBeenCalledWith("Passwords don't match");
  });

  it("rejects short passwords", async () => {
    const user = userEvent.setup();
    setup();
    await fillPasswords(user, { next: "short", confirm: "short" });

    fireEvent.submit(
      screen.getByRole("button", { name: /update password/i }).closest("form")!
    );

    expect(toast.error).toHaveBeenCalledWith(
      "Password must be at least 8 characters long"
    );
  });

  it("updates the password and closes", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.post("*/api/v1/user/password", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json({ message: "Password updated" });
      })
    );
    const { onOpenChange } = setup();
    await fillPasswords(user);

    await user.click(screen.getByRole("button", { name: /update password/i }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith(
        "Password updated successfully!"
      )
    );
    expect(body).toEqual({
      old_password: "password123",
      new_password: "password456",
    });
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("keeps the dialog open when the update fails", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/user/password", () =>
        HttpResponse.json(
          { error: "current password is wrong" },
          { status: 400 }
        )
      )
    );
    const { onOpenChange } = setup();
    await fillPasswords(user);

    await user.click(screen.getByRole("button", { name: /update password/i }));

    await waitFor(() =>
      expect(consoleError).toHaveBeenCalledWith("current password is wrong")
    );
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });

  it("clears the form when the dialog closes", async () => {
    const user = userEvent.setup();
    const { onOpenChange } = setup();
    await user.type(screen.getByLabelText("Current"), "password123");

    await user.keyboard("{Escape}");

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
