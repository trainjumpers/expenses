import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import React from "react";

interface LoadingButtonProps extends React.ComponentProps<typeof Button> {
  loading?: boolean;
  children: React.ReactNode;
  fixedWidth?: string; // e.g. '120px'
}

export const LoadingButton: React.FC<LoadingButtonProps> = ({
  loading = false,
  children,
  fixedWidth = "120px",
  disabled,
  ...props
}) => {
  return (
    <Button
      disabled={loading || disabled}
      style={{ minWidth: fixedWidth, maxWidth: fixedWidth, ...props.style }}
      {...props}
    >
      {loading ? (
        <span className="flex items-center justify-center w-full">
          <Spinner className="mr-2 w-4 h-4" />
          <span>Loading...</span>
        </span>
      ) : (
        children
      )}
    </Button>
  );
};
