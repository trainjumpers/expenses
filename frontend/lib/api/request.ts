import { handleApiError } from "@/lib/utils/toast";
import { toast } from "sonner";

export type CustomErrorHandler = (response: Response, data: unknown) => boolean;

export async function apiRequest<T>(
  url: string,
  options: RequestInit,
  resourceName: string,
  customErrorHandlers: CustomErrorHandler[] = [],
  errorMessage: string = "Something went wrong. Please try again."
): Promise<T> {
  let toastShown = false;
  try {
    const response = await fetch(url, options);
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
      throw new Error(
        typeof errorMsg === "string" ? errorMsg : "Request failed"
      );
    }
    return (data as { data: T }).data;
  } catch (err) {
    if (!toastShown) toast.error(errorMessage);
    throw err;
  }
}
