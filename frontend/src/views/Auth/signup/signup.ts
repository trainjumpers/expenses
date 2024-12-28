import { register } from "@/api/auth";
import { convertErrorToString } from "@/utils/common";
import { setUserToken } from "@/utils/cookies";
import { ref } from "vue";
import { useRouter } from "vue-router";



import { ERROR_TIMEOUT, NAME_LENGTH, PASSWORD_LENGTH } from "../constants";

export const useSignup = () => {
  const router = useRouter();
  const name = ref("");
  const email = ref("");
  const password = ref("");
  const confirmPassword = ref("");
  const error = ref("");
  const loading = ref(false);

  const setTimeoutForError = () => {
    setTimeout(() => {
      error.value = "";
    }, ERROR_TIMEOUT);
  };

  const handleSignup = async () => {
    if (password.value !== confirmPassword.value) {
      error.value = "Passwords do not match";
      setTimeoutForError();
      return;
    }
    if (password.value.length < PASSWORD_LENGTH) {
      error.value = "Password must be at least 6 characters long";
      setTimeoutForError();
      return;
    }
    if (name.value.length < NAME_LENGTH) {
      error.value = "Name must be at least 3 characters long";
      setTimeoutForError();
      return;
    }
    try {
      const res = await register(email.value, password.value, name.value);
      setUserToken(res.access_token);
      error.value = "";
      router.push("/");
    } catch (err) {
      error.value = convertErrorToString(err);
      setTimeoutForError();
    } finally {
      loading.value = false;
    }
  };

  return {
    name,
    email,
    password,
    confirmPassword,
    error,
    loading,
    handleSignup,
  };
};
