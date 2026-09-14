import { ViewStatementsModal } from "@/components/custom/Modal/Statement/ViewStatementsModal";
import type { Statement } from "@/lib/models/statement";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, delay, http } from "msw";
import { describe, expect, it, vi } from "vitest";

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

const statement: Statement = {
  id: 5,
  account_id: 1,
  created_by: 1,
  original_filename: "statement.csv",
  file_type: "csv",
  status: "pending",
  created_at: "2026-09-01T00:00:00.000Z",
};

function listHandler(
  statements: Statement[],
  total = statements.length,
  onRequest?: (params: URLSearchParams) => void
) {
  return http.get("*/api/v1/statement", ({ request }) => {
    const params = new URL(request.url).searchParams;
    onRequest?.(params);
    return HttpResponse.json({
      message: "ok",
      data: { statements, total, page: 1, page_size: 5 },
    });
  });
}

function render() {
  return renderWithProviders(
    <ViewStatementsModal isOpen onOpenChange={vi.fn()} />
  );
}

describe("ViewStatementsModal", () => {
  it("lists the uploaded statements with their account and status", async () => {
    server.use(listHandler([statement]));
    render();

    expect(await screen.findByText("statement.csv")).toBeInTheDocument();
    expect(await screen.findByText("HDFC Savings")).toBeInTheDocument();
    expect(screen.getByText("csv")).toBeInTheDocument();
    expect(screen.getByText("pending")).toBeInTheDocument();
  });

  it("renders every status", async () => {
    const statuses = ["pending", "processing", "done", "error"] as const;
    server.use(
      listHandler(
        statuses.map((status, i) => ({
          ...statement,
          id: i + 1,
          status,
          original_filename: `statement-${status}.csv`,
        }))
      )
    );
    render();

    await screen.findByText("statement-pending.csv");
    for (const status of statuses) {
      expect(screen.getByText(status)).toBeInTheDocument();
    }
  });

  it("shows the empty state without statements", async () => {
    server.use(listHandler([]));
    render();

    expect(
      await screen.findByText("No statements uploaded yet")
    ).toBeInTheDocument();
    expect(
      screen.getByText("Upload your first bank statement to get started")
    ).toBeInTheDocument();
  });

  it("shows a loading state", () => {
    server.use(
      http.get("*/api/v1/statement", async () => {
        await delay(100);
        return HttpResponse.json({
          data: { statements: [], total: 0, page: 1, page_size: 5 },
        });
      })
    );
    render();

    expect(screen.getByText("Loading statements...")).toBeInTheDocument();
  });

  it("shows a failure state", async () => {
    server.use(
      http.get("*/api/v1/statement", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );
    render();

    expect(
      await screen.findByText("Failed to load statements")
    ).toBeInTheDocument();
  });

  it("filters by search and clears the filters", async () => {
    const user = userEvent.setup();
    let lastSearch: string | null = "unset";
    server.use(
      http.get("*/api/v1/statement", ({ request }) => {
        lastSearch = new URL(request.url).searchParams.get("search");
        const statements = lastSearch ? [] : [statement];
        return HttpResponse.json({
          message: "ok",
          data: {
            statements,
            total: statements.length,
            page: 1,
            page_size: 5,
          },
        });
      })
    );
    render();
    await screen.findByText("statement.csv");

    await user.type(
      screen.getByPlaceholderText("Search by filename..."),
      "coffee"
    );

    expect(await screen.findByText("No statements found")).toBeInTheDocument();
    expect(screen.getByText("Try adjusting your filters")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /clear filters/i }));

    expect(await screen.findByText("statement.csv")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Search by filename...")).toHaveValue(
      ""
    );
    expect(
      screen.queryByRole("button", { name: /clear filters/i })
    ).not.toBeInTheDocument();
  });

  it("filters by account", async () => {
    const user = userEvent.setup();
    let accountId: string | null = null;
    server.use(
      listHandler([statement], 1, (params) => {
        accountId = params.get("account_id");
      })
    );
    render();
    await screen.findByText("statement.csv");

    await user.click(screen.getByRole("combobox"));
    await user.click(
      await screen.findByRole("option", { name: "HDFC Savings" })
    );

    await waitFor(() => expect(accountId).toBe("1"));
  });

  it("pages through the uploads", async () => {
    const user = userEvent.setup();
    let page = "1";
    const six = Array.from({ length: 6 }, (_, i) => ({
      ...statement,
      id: i + 1,
      original_filename: `statement-${i + 1}.csv`,
    }));
    server.use(
      http.get("*/api/v1/statement", ({ request }) => {
        page = new URL(request.url).searchParams.get("page") ?? "1";
        return HttpResponse.json({
          message: "ok",
          data: {
            statements: page === "2" ? six.slice(5) : six.slice(0, 5),
            total: 6,
            page: Number(page),
            page_size: 5,
          },
        });
      })
    );
    render();
    await screen.findByText("statement-1.csv");
    expect(screen.queryByText("statement-6.csv")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "2" }));

    expect(await screen.findByText("statement-6.csv")).toBeInTheDocument();
    expect(page).toBe("2");
  });

  it("falls back to Unknown Account for orphan statements", async () => {
    server.use(listHandler([{ ...statement, account_id: 99 }]));
    render();

    expect(await screen.findByText("Unknown Account")).toBeInTheDocument();
  });
});
