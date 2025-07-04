import { toast } from "sonner";

export const handleApiError = (status: number, resource: string) => {
  switch (status) {
    case 400:
      toast.error("Please check your input and try again", {
        id: "bad-user-input",
      });
      break;
    case 401:
      toast.error("Please login again", {
        id: "unauthorized",
        action: {
          label: "Login",
          onClick: () => {
            window.location.href = "/login";
          },
        },
      });
      break;
    case 404:
      toast.warning(
        `${resource.charAt(0).toUpperCase() + resource.slice(1)} does not exist`
      );
      break;
    case 409:
      toast.error(
        `${resource.charAt(0).toUpperCase() + resource.slice(1)} already exists`
      );
      break;
    default:
      toast.error(
        "Something went wrong. Contact support if the problem persists",
        { id: "generic-error" }
      );
      break;
  }
};
