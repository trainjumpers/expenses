import { handleApiError } from "@/lib/utils/toast";
import { toast } from "sonner";

export type CustomErrorHandler = (response: Response, data: unknown) => boolean;

export async function apiRequest<T>(
  url: string,
  options: RequestInit & { signal?: AbortSignal },
  resourceName: string,
  customErrorHandlers: CustomErrorHandler[] = [],
  errorMessage: string = "Something went wrong. Please try again."
): Promise<T> {
  let toastShown = false;
  try {
    const response = await fetch(url, {
      ...options,
      credentials: "include", // Always send cookies
    });
    let data: unknown;
    try {
      data = await response.json();
    } catch {
      data = {};
    }

    if (!response.ok) {
      for (const handler of customErrorHandlers) {
        if (handler(response, data)) {
          toastShown = true;
          break;
        }
      }
      if (!toastShown) {
        handleApiError(response.status, resourceName);
        toastShown = true;
      }
      const errorMsg = (data as Record<string, unknown>)["error"];
      const error = new Error(
        typeof errorMsg === "string" ? errorMsg : "Request failed"
      ) as Error & { status: number; statusText: string; data: unknown };
      error.status = response.status;
      error.statusText = response.statusText;
      error.data = data;
      throw error;
    }
    return (data as { data: T }).data;
  } catch (err) {
    if (
      !toastShown &&
      !(err instanceof DOMException && err.name === "AbortError")
    ) {
      toast.error(errorMessage);
    }
    throw err;
  }
}
