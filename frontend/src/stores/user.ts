import { getUser as getUserApi } from "@/api/user";
import type { User } from "@/types/user";
import { defineStore } from "pinia";
import { reactive } from "vue";

export const useUserStore = defineStore("user", () => {
  const user: User = reactive({
    id: 0,
    name: "",
    email: "",
  });

  let isFetched = false;

  const getUser = async () => {
    if (isFetched) return user;
    const data = await getUserApi();
    user.id = data.id;
    user.name = data.name;
    user.email = data.email;
    isFetched = true;
    return user;
  };

  return { getUser };
});
