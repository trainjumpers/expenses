import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ThemeProvider } from "./ThemeProvider";

describe("ThemeProvider", () => {
  it("renders its children through the themes provider", async () => {
    render(
      <ThemeProvider attribute="class" defaultTheme="system">
        <p>Themed</p>
      </ThemeProvider>
    );

    expect(await screen.findByText("Themed")).toBeInTheDocument();
  });
});
