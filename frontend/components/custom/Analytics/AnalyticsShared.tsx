import { cn, formatCurrency, formatShortCurrency } from "@/lib/utils";
import type { ReactNode } from "react";

export function AnalyticsSection({
  title,
  description,
  actions,
  children,
  className,
}: {
  title: string;
  description?: string;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section className={cn("border-t pt-6", className)}>
      <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-base font-semibold tracking-tight text-foreground">
            {title}
          </h2>
          {description ? (
            <p className="mt-1 max-w-prose text-pretty text-xs text-muted-foreground">
              {description}
            </p>
          ) : null}
        </div>
        {actions}
      </div>
      {children}
    </section>
  );
}

// Renders a compact value while exposing the exact amount to assistive
// technology and as a hover title.
export function Money({
  value,
  className,
  full = false,
}: {
  value: number;
  className?: string;
  full?: boolean;
}) {
  return (
    <span
      className={cn("tabular-nums", className)}
      title={formatCurrency(value)}
    >
      <span aria-hidden="true">
        {full ? formatCurrency(value) : formatShortCurrency(value)}
      </span>
      <span className="sr-only">{formatCurrency(value)}</span>
    </span>
  );
}

export function InlineNote({ children }: { children: ReactNode }) {
  return (
    <p className="text-pretty text-sm text-muted-foreground">{children}</p>
  );
}
