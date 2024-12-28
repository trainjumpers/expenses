import { getUserToken } from "./cookies";

export const checkIfAuth = () => {
  const token = getUserToken();
  if (!token) {
    return false;
  }
  return true;
};
