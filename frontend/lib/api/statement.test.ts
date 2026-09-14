import type { Statement } from "@/lib/models/statement";
import { server } from "@/test/msw/server";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import {
  getStatement,
  listStatements,
  previewStatement,
  uploadStatement,
} from "./statement";

const statement: Statement = {
  id: 5,
  account_id: 1,
  created_by: 1,
  original_filename: "statement.csv",
  file_type: "csv",
  status: "pending",
  created_at: "2026-09-01T00:00:00.000Z",
};

function makeFile() {
  return new File(["date,amount"], "statement.csv", { type: "text/csv" });
}

describe("uploadStatement", () => {
  it("uploads the file with every optional field", async () => {
    let body = "";
    server.use(
      http.post("*/api/v1/statement", async ({ request }) => {
        body = await request.text();
        return HttpResponse.json(
          { message: "Statement uploaded", data: statement },
          { status: 201 }
        );
      })
    );

    const result = await uploadStatement({
      account_id: 1,
      file: makeFile(),
      bank_type: "hdfc",
      metadata: "{}",
      password: "secret",
    });

    expect(body).toMatch(/name="account_id"\r\n\r\n1\r\n/);
    expect(body).toMatch(/name="bank_type"\r\n\r\nhdfc\r\n/);
    expect(body).toMatch(/name="metadata"\r\n\r\n\{\}\r\n/);
    expect(body).toMatch(/name="password"\r\n\r\nsecret\r\n/);
    expect(body).toContain('name="file"');
    expect(result.statement).toEqual(statement);
    expect(result.message).toBe("Statement uploaded");
  });

  it("omits optional fields when they are not provided", async () => {
    let body = "";
    server.use(
      http.post("*/api/v1/statement", async ({ request }) => {
        body = await request.text();
        return HttpResponse.json({ data: statement });
      })
    );

    const result = await uploadStatement({ account_id: 2, file: makeFile() });

    expect(body).toMatch(/name="account_id"\r\n\r\n2\r\n/);
    expect(body).not.toContain('name="bank_type"');
    expect(body).not.toContain('name="metadata"');
    expect(body).not.toContain('name="password"');
    expect(result.message).toBe("Statement uploaded successfully");
  });

  it("flags a password-required failure", async () => {
    server.use(
      http.post("*/api/v1/statement", () =>
        HttpResponse.json(
          { message: "statement password required" },
          { status: 400 }
        )
      )
    );

    await expect(
      uploadStatement({ account_id: 1, file: makeFile() })
    ).rejects.toMatchObject({ status: 400, isPasswordRequired: true });
  });

  it("flags a password-required error variant", async () => {
    server.use(
      http.post("*/api/v1/statement", () =>
        HttpResponse.json(
          { error: "statement password required" },
          { status: 422 }
        )
      )
    );

    await expect(
      uploadStatement({ account_id: 1, file: makeFile() })
    ).rejects.toMatchObject({ status: 422, isPasswordRequired: true });
  });

  it("falls back to a generic failure message", async () => {
    server.use(
      http.post(
        "*/api/v1/statement",
        () => new HttpResponse(null, { status: 500 })
      )
    );

    await expect(
      uploadStatement({ account_id: 1, file: makeFile() })
    ).rejects.toMatchObject({ message: "Failed to upload statement" });
  });
});

describe("previewStatement", () => {
  it("returns the preview with the paging fields", async () => {
    let body = "";
    server.use(
      http.post("*/api/v1/statement/preview", async ({ request }) => {
        body = await request.text();
        return HttpResponse.json({
          message: "ok",
          data: { headers: ["Date", "Amount"], rows: [["2026-09-01", "12"]] },
        });
      })
    );

    const preview = await previewStatement(makeFile(), 2, 10, "secret");

    expect(body).toMatch(/name="skip_rows"\r\n\r\n2\r\n/);
    expect(body).toMatch(/name="row_size"\r\n\r\n10\r\n/);
    expect(body).toMatch(/name="password"\r\n\r\nsecret\r\n/);
    expect(preview.headers).toEqual(["Date", "Amount"]);
  });

  it("flags a password-required preview failure", async () => {
    server.use(
      http.post("*/api/v1/statement/preview", () =>
        HttpResponse.json(
          { error: "statement password required" },
          { status: 400 }
        )
      )
    );

    await expect(previewStatement(makeFile(), 0, 5)).rejects.toMatchObject({
      isPasswordRequired: true,
    });
  });
});

describe("listStatements", () => {
  it("builds the query string from the provided params", async () => {
    let url = "";
    server.use(
      http.get("*/api/v1/statement", ({ request }) => {
        url = request.url;
        return HttpResponse.json({
          data: { statements: [statement], total: 1, page: 1, page_size: 10 },
        });
      })
    );

    const result = await listStatements(undefined, {
      page: 1,
      page_size: 10,
      account_id: 2,
      date_from: "2026-09-01",
      date_to: "2026-09-30",
      search: "coffee beans",
    });

    const params = new URL(url).searchParams;
    expect(params.get("account_id")).toBe("2");
    expect(params.get("date_from")).toBe("2026-09-01");
    expect(params.get("date_to")).toBe("2026-09-30");
    expect(params.get("search")).toBe("coffee beans");
    expect(result.total).toBe(1);
  });

  it("omits the query string without params", async () => {
    let url = "";
    server.use(
      http.get("*/api/v1/statement", ({ request }) => {
        url = request.url;
        return HttpResponse.json({
          data: { statements: [], total: 0, page: 1, page_size: 10 },
        });
      })
    );

    await listStatements();

    expect(new URL(url).search).toBe("");
  });
});

describe("getStatement", () => {
  it("fetches a single statement", async () => {
    server.use(
      http.get("*/api/v1/statement/5", () =>
        HttpResponse.json({ message: "ok", data: statement })
      )
    );

    await expect(getStatement(5)).resolves.toEqual(statement);
  });
});
