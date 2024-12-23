import { login } from "@/api/auth";
import { ref } from "vue";
import { useRouter } from "vue-router";

export const useLogin = () => {
  const router = useRouter();
  const email = ref("");
  const password = ref("");
  const error = ref("");

  const handleLogin = async () => {
    const res = await login(email.value, password.value);
    console.log(res.access_token)
  };

  return {
    email,
    password,
    error,
    handleLogin,
  };
};
