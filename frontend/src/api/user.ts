import { BACKEND_URL } from "@/constants/web";
import type { User } from "@/types/user";
import { getUserToken } from "@/utils/cookies";

export const getUser = async (): Promise<User> => {
  const response = await fetch(`${BACKEND_URL}/user`, {
    headers: {
      Authorization: `Bearer ${getUserToken()}`,
    },
  });
  if (!response.ok) {
    throw new Error("Something went wrong");
  }
  const data = (await response.json()).data as User;
  return data;
};
