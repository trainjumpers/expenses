import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { ProfileModal } from "./ProfileModal";

function setup() {
  const onOpenChange = vi.fn();
  renderWithProviders(
    <ProfileModal
      isOpen
      onOpenChange={onOpenChange}
      formData={{ name: "Test User", email: "test1@example.com" }}
    />
  );
  return { onOpenChange };
}

describe("ProfileModal", () => {
  it("saves the edited profile", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.patch("*/api/v1/user", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json({ data: { name: "Renamed" } });
      })
    );
    const { onOpenChange } = setup();

    const name = screen.getByDisplayValue("Test User");
    await user.clear(name);
    await user.type(name, "Renamed");
    await user.click(screen.getByRole("button", { name: "Save changes" }));

    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
    expect(body).toEqual({
      name: "Renamed",
      email: "test1@example.com",
    });
  });
});
