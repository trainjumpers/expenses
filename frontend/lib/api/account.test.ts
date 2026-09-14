import { server } from "@/test/msw/server";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it } from "vitest";

import { deleteAccount, getAccount } from "./account";

describe("account api", () => {
  it("loads a single account", async () => {
    server.use(
      http.get("*/api/v1/account/1", () =>
        HttpResponse.json({
          message: "ok",
          data: { id: 1, name: "HDFC Savings", bank_type: "hdfc" },
        })
      )
    );

    await expect(getAccount(1)).resolves.toMatchObject({ id: 1 });
  });

  it("falls back to the generic message for a different 409", async () => {
    server.use(
      http.delete("*/api/v1/account/1", () =>
        HttpResponse.json({ message: "account is locked" }, { status: 409 })
      )
    );

    await expect(deleteAccount(1)).rejects.toThrow("Request failed");

    expect(toast.error).not.toHaveBeenCalledWith(
      "Cannot delete account with existing transactions",
      expect.anything()
    );
  });
});
