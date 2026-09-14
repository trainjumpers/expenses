import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { ImportStatementModal } from "./ImportStatementModal";

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

const statement = {
  id: 5,
  account_id: 1,
  created_by: 1,
  original_filename: "statement.csv",
  file_type: "csv",
  status: "pending",
  created_at: "2026-09-01T00:00:00.000Z",
};

const previewData = {
  headers: ["Date", "Narration", "Amount"],
  rows: [["2026-09-01", "Coffee", "120"]],
};

function csvFile(name = "statement.csv") {
  return new File(["date,amount\n2026-09-01,120"], name, {
    type: "text/csv",
  });
}

function fileInput(id: string) {
  return document.querySelector(`#${id}`) as HTMLInputElement;
}

function setup() {
  const onOpenChange = vi.fn();
  renderWithProviders(
    <ImportStatementModal isOpen onOpenChange={onOpenChange} />
  );
  return { onOpenChange };
}

function uploadHandler(
  onBody?: (body: string) => void,
  attempts = { count: 0 }
) {
  return http.post("*/api/v1/statement", async ({ request }) => {
    attempts.count += 1;
    const body = await request.text();
    onBody?.(body);
    return HttpResponse.json(
      { message: "ok", data: statement },
      { status: 201 }
    );
  });
}

async function mapField(
  user: ReturnType<typeof userEvent.setup>,
  index: number,
  option: string
) {
  await user.click(screen.getAllByRole("combobox")[index]);
  await user.click(await screen.findByRole("option", { name: option }));
}

describe("ImportStatementModal", () => {
  it("starts on the import method step", () => {
    setup();

    expect(
      screen.getByRole("dialog", { name: "Select Import Method" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /import from bank/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /custom parsing/i })
    ).toBeInTheDocument();
  });

  it("preselects the first account and uploads a statement", async () => {
    const user = userEvent.setup();
    let body = "";
    server.use(uploadHandler((value) => (body = value)));
    const { onOpenChange } = setup();

    await user.click(screen.getByRole("button", { name: /import from bank/i }));
    expect(await screen.findByText("HDFC Savings (HDFC)")).toBeInTheDocument();

    await user.upload(fileInput("file-input"), csvFile());
    expect(await screen.findByText("statement.csv")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /import statement/i }));

    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
    expect(body).toMatch(/name="account_id"\r\n\r\n1\r\n/);
    expect(body).toContain('name="file"');
  });

  it("rejects unsupported and oversized files", async () => {
    const user = userEvent.setup({ applyAccept: false });
    setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));

    await user.upload(
      fileInput("file-input"),
      new File(["x"], "statement.pdf", { type: "application/pdf" })
    );
    expect(
      await screen.findByText(/File must be .csv, .xls, .xlsx, .txt format/)
    ).toBeInTheDocument();

    const big = csvFile("big.csv");
    Object.defineProperty(big, "size", { value: 6 * 1024 * 1024 });
    await user.upload(fileInput("file-input"), big);
    expect(
      await screen.findByText(/File size must be less than 5MB/)
    ).toBeInTheDocument();
  });

  it("limits bank imports to ten files", async () => {
    const user = userEvent.setup();
    setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));

    const files = Array.from({ length: 11 }, (_, i) =>
      csvFile(`statement-${i + 1}.csv`)
    );
    await user.upload(fileInput("file-input"), files);

    expect(
      await screen.findByText("Maximum 10 files allowed")
    ).toBeInTheDocument();
  });

  it("uploads several files in sequence", async () => {
    const user = userEvent.setup();
    const attempts = { count: 0 };
    server.use(uploadHandler(undefined, attempts));
    const { onOpenChange } = setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));

    await user.upload(fileInput("file-input"), csvFile("first.csv"));
    expect(
      await screen.findByText("Click to add more files (1/10)")
    ).toBeInTheDocument();

    await user.upload(
      fileInput("file-input-additional"),
      csvFile("second.csv")
    );
    expect(await screen.findByText("second.csv")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /import statement/i }));

    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
    expect(attempts.count).toBe(2);
  });

  it("asks for the file password when a bank file is locked", async () => {
    const user = userEvent.setup();
    let sawPassword = false;
    server.use(
      http.post("*/api/v1/statement", async ({ request }) => {
        const body = await request.text();
        if (body.includes('name="password"')) {
          sawPassword = true;
          return HttpResponse.json(
            { message: "ok", data: statement },
            { status: 201 }
          );
        }
        return HttpResponse.json(
          { error: "statement password required" },
          { status: 400 }
        );
      })
    );
    const { onOpenChange } = setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));
    await user.upload(fileInput("file-input"), csvFile());
    await user.click(screen.getByRole("button", { name: /import statement/i }));

    expect(
      await screen.findByText("File Password Required")
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Submit Password" }));
    expect(await screen.findByText("Password is required")).toBeInTheDocument();

    await user.type(
      screen.getByPlaceholderText("Enter file password"),
      "secret123"
    );
    await user.click(screen.getByRole("button", { name: "Submit Password" }));

    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
    expect(sawPassword).toBe(true);
  });

  it("shows upload failures without closing", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/statement", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );
    const { onOpenChange } = setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));
    await user.upload(fileInput("file-input"), csvFile());
    await user.click(screen.getByRole("button", { name: /import statement/i }));

    expect(
      await screen.findByText(/Failed to upload statement.csv: boom/)
    ).toBeInTheDocument();
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });

  it("needs an account for the upload", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/account", () =>
        HttpResponse.json({ message: "ok", data: [] })
      )
    );
    setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));
    await user.upload(fileInput("file-input"), csvFile());
    await user.click(screen.getByRole("button", { name: /import statement/i }));

    expect(
      await screen.findByText("Please select at least one file and an account")
    ).toBeInTheDocument();
  });

  it("previews a custom file, maps the columns and processes it", async () => {
    const user = userEvent.setup();
    let previewBody = "";
    let uploadBody = "";
    server.use(
      http.post("*/api/v1/statement/preview", async ({ request }) => {
        previewBody = await request.text();
        return HttpResponse.json({ message: "ok", data: previewData });
      }),
      http.post("*/api/v1/statement", async ({ request }) => {
        uploadBody = await request.text();
        return HttpResponse.json(
          { message: "ok", data: statement },
          { status: 201 }
        );
      })
    );
    const { onOpenChange } = setup();

    await user.click(screen.getByRole("button", { name: /custom parsing/i }));
    expect(
      await screen.findByRole("dialog", { name: "Preview" })
    ).toBeInTheDocument();

    await user.upload(fileInput("file-input-fallback"), csvFile());

    expect(await screen.findByText("Coffee")).toBeInTheDocument();

    await user.clear(screen.getByLabelText("Skip Rows"));
    await user.type(screen.getByLabelText("Skip Rows"), "2");
    await waitFor(() =>
      expect(previewBody).toMatch(/name="skip_rows"\r\n\r\n2\r\n/)
    );

    await user.click(screen.getByRole("button", { name: "Next" }));
    expect(
      screen.getByRole("dialog", { name: "Map Columns" })
    ).toBeInTheDocument();

    await mapField(user, 0, "Date");
    await mapField(user, 1, "Narration");
    await mapField(user, 3, "Amount");
    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );

    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
    expect(uploadBody).toContain('name="bank_type"');
    const metadata = /name="metadata"\r\n\r\n(.*)\r\n/.exec(uploadBody);
    expect(metadata).not.toBeNull();
    expect(JSON.parse(metadata![1])).toEqual({
      skipRows: 2,
      columnMapping: {
        txn_date: "Date",
        name: "Narration",
        amount: "Amount",
      },
    });
  });

  it("asks for a password while previewing", async () => {
    const user = userEvent.setup();
    let sawPassword = false;
    server.use(
      http.post("*/api/v1/statement/preview", async ({ request }) => {
        const body = await request.text();
        if (body.includes('name="password"')) {
          sawPassword = true;
          return HttpResponse.json({ message: "ok", data: previewData });
        }
        return HttpResponse.json(
          { error: "statement password required" },
          { status: 400 }
        );
      })
    );
    setup();

    await user.click(screen.getByRole("button", { name: /custom parsing/i }));
    await user.upload(fileInput("file-input-fallback"), csvFile());

    expect(
      await screen.findByText("File Password Required")
    ).toBeInTheDocument();

    await user.type(
      screen.getByPlaceholderText("Enter file password"),
      "secret123"
    );
    await user.click(screen.getByRole("button", { name: "Submit Password" }));

    await waitFor(() => expect(sawPassword).toBe(true));
    expect(await screen.findByText("Coffee")).toBeInTheDocument();
  });

  it("surfaces preview failures", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/statement/preview", () =>
        HttpResponse.json({ error: "invalid file" }, { status: 400 })
      )
    );
    setup();

    await user.click(screen.getByRole("button", { name: /custom parsing/i }));
    await user.upload(fileInput("file-input-fallback"), csvFile());

    expect(await screen.findByText("invalid file")).toBeInTheDocument();
  });

  it("walks back to the method step", async () => {
    const user = userEvent.setup();
    setup();

    await user.click(screen.getByRole("button", { name: /import from bank/i }));
    await user.click(screen.getByRole("button", { name: "Back" }));
    expect(
      screen.getByRole("dialog", { name: "Select Import Method" })
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /custom parsing/i }));
    await user.click(screen.getByRole("button", { name: "Back" }));
    expect(
      screen.getByRole("dialog", { name: "Select Import Method" })
    ).toBeInTheDocument();
  });

  it("submits the file password with the enter key", async () => {
    const user = userEvent.setup();
    let sawPassword = false;
    server.use(
      http.post("*/api/v1/statement", async ({ request }) => {
        const body = await request.text();
        if (body.includes('name="password"')) {
          sawPassword = true;
          return HttpResponse.json(
            { message: "ok", data: statement },
            { status: 201 }
          );
        }
        return HttpResponse.json(
          { error: "statement password required" },
          { status: 400 }
        );
      })
    );
    const { onOpenChange } = setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));
    await user.upload(fileInput("file-input"), csvFile());
    await user.click(screen.getByRole("button", { name: /import statement/i }));
    await screen.findByText("File Password Required");

    await user.type(
      screen.getByPlaceholderText("Enter file password"),
      "secret123{Enter}"
    );

    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
    expect(sawPassword).toBe(true);
  });

  it("removes a selected file", async () => {
    const user = userEvent.setup();
    setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));
    await user.upload(fileInput("file-input"), csvFile("first.csv"));
    await screen.findByText("first.csv");

    await user.click(screen.getByRole("button", { name: "Remove first.csv" }));

    expect(
      screen.getByText(/Drag & drop your bank statements here/)
    ).toBeInTheDocument();
  });

  it("rejects duplicate files when adding more", async () => {
    const user = userEvent.setup();
    setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));
    await user.upload(fileInput("file-input"), csvFile("first.csv"));
    await screen.findByText("first.csv");

    await user.upload(fileInput("file-input-additional"), csvFile("first.csv"));

    expect(
      await screen.findByText("All selected files are already added")
    ).toBeInTheDocument();
  });

  it("accepts a dropped file", async () => {
    const user = userEvent.setup();
    setup();
    await user.click(screen.getByRole("button", { name: /import from bank/i }));

    const dropzone = screen
      .getByText(/Drag & drop your bank statements here/)
      .closest("div")!;
    fireEvent.dragEnter(dropzone);
    expect(screen.getByText("Drop the files here...")).toBeInTheDocument();

    fireEvent.drop(dropzone, { dataTransfer: { files: [csvFile()] } });

    expect(await screen.findByText("statement.csv")).toBeInTheDocument();
  });

  it("removes the previewed file", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/statement/preview", () =>
        HttpResponse.json({ message: "ok", data: previewData })
      )
    );
    setup();
    await user.click(screen.getByRole("button", { name: /custom parsing/i }));
    await user.upload(fileInput("file-input-fallback"), csvFile());
    await screen.findByText("Coffee");

    await user.click(
      screen.getByRole("button", { name: "Remove statement.csv" })
    );

    expect(
      screen.getByText(/Drag & drop your bank statement here/)
    ).toBeInTheDocument();
  });

  it("refreshes the preview when the row size changes", async () => {
    const user = userEvent.setup();
    let previewBody = "";
    server.use(
      http.post("*/api/v1/statement/preview", async ({ request }) => {
        previewBody = await request.text();
        return HttpResponse.json({ message: "ok", data: previewData });
      })
    );
    setup();
    await user.click(screen.getByRole("button", { name: /custom parsing/i }));
    await user.upload(fileInput("file-input-fallback"), csvFile());
    await screen.findByText("Coffee");

    await user.clear(screen.getByLabelText("Row Size"));
    fireEvent.change(screen.getByLabelText("Row Size"), {
      target: { value: "3" },
    });

    await waitFor(() =>
      expect(previewBody).toMatch(/name="row_size"\r\n\r\n3\r\n/)
    );
  });

  it("reports a failed processing run", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/statement/preview", () =>
        HttpResponse.json({ message: "ok", data: previewData })
      ),
      http.post("*/api/v1/statement", () =>
        HttpResponse.json({ error: "bad file" }, { status: 400 })
      )
    );
    setup();
    await user.click(screen.getByRole("button", { name: /custom parsing/i }));
    await user.upload(fileInput("file-input-fallback"), csvFile());
    await screen.findByText("Coffee");
    await user.click(screen.getByRole("button", { name: "Next" }));
    await mapField(user, 0, "Date");
    await mapField(user, 1, "Narration");
    await mapField(user, 3, "Amount");

    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );

    expect(
      await screen.findByText(/Failed to process statement.csv: bad file/)
    ).toBeInTheDocument();
    expect(
      screen.getByRole("dialog", { name: "Map Columns" })
    ).toBeInTheDocument();
  });

  it("stays quiet when processing needs a password", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/statement/preview", () =>
        HttpResponse.json({ message: "ok", data: previewData })
      ),
      http.post("*/api/v1/statement", () =>
        HttpResponse.json(
          { error: "statement password required" },
          { status: 400 }
        )
      )
    );
    const { onOpenChange } = setup();
    await user.click(screen.getByRole("button", { name: /custom parsing/i }));
    await user.upload(fileInput("file-input-fallback"), csvFile());
    await screen.findByText("Coffee");
    await user.click(screen.getByRole("button", { name: "Next" }));
    await mapField(user, 0, "Date");
    await mapField(user, 1, "Narration");
    await mapField(user, 3, "Amount");

    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );

    await waitFor(() => expect(onOpenChange).not.toHaveBeenCalledWith(false));
    expect(
      screen.getByRole("dialog", { name: "Map Columns" })
    ).toBeInTheDocument();
  });
});
