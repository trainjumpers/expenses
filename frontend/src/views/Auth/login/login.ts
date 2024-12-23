import { login } from "@/api/auth";
import { convertErrorToString } from "@/utils/common";
import { setUserToken } from "@/utils/cookies";
import { ref } from "vue";
import { useRouter } from "vue-router";

export const useLogin = () => {
  const router = useRouter();
  const email = ref("");
  const password = ref("");
  const error = ref("");

  const handleLogin = async () => {
    try {
      const res = await login(email.value, password.value);
      setUserToken(res.access_token);
      router.push("/");
      error.value = "";
    } catch (err) {
      error.value = convertErrorToString(err);
      setTimeout(() => {
        error.value = "";
      }, 2000);
    }
  };

  return {
    email,
    password,
    error,
    handleLogin,
  };
};
