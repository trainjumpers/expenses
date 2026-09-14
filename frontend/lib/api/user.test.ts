import { testUser } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it } from "vitest";

import { getUser, updatePassword, updateUser } from "./user";

describe("user api", () => {
  it("loads the current user", async () => {
    await expect(getUser()).resolves.toMatchObject({
      email: "test1@example.com",
    });
  });

  it("patches the profile", async () => {
    server.use(
      http.patch("*/api/v1/user", async ({ request }) => {
        const body = (await request.json()) as { name?: string };
        return HttpResponse.json({
          message: "ok",
          data: { ...testUser, ...body },
        });
      })
    );

    await expect(updateUser({ name: "Renamed" })).resolves.toMatchObject({
      name: "Renamed",
    });
  });

  it("explains an incorrect current password", async () => {
    server.use(
      http.post("*/api/v1/user/password", () =>
        HttpResponse.json({ error: "wrong" }, { status: 401 })
      )
    );

    await expect(updatePassword("wrong", "password456")).rejects.toThrow(
      "wrong"
    );
    expect(toast.error).toHaveBeenCalledWith("Current password is incorrect", {
      id: "password-error",
    });
  });
});
