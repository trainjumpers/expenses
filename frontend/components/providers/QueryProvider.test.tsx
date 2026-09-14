import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { QueryProvider } from "./QueryProvider";

describe("QueryProvider", () => {
  it("renders its children", () => {
    render(
      <QueryProvider>
        <p>Provided</p>
      </QueryProvider>
    );

    expect(screen.getByText("Provided")).toBeInTheDocument();
  });
});
