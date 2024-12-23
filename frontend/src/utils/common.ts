export const convertErrorToString = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error).replace("Error: ", "");
};
