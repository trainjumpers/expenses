import { register } from "@/api/auth";
import { convertErrorToString } from "@/utils/common";
import { setUserToken } from "@/utils/cookies";
import { ref } from "vue";
import { useRouter } from "vue-router";

export const useSignup = () => {
  const router = useRouter();
  const name = ref("");
  const email = ref("");
  const password = ref("");
  const confirmPassword = ref("");
  const error = ref("");

  const setTimeoutForError = () => {
    setTimeout(() => {
      error.value = "";
    }, 3000);
  };

  const handleSignup = async () => {
    if (password.value !== confirmPassword.value) {
      error.value = "Passwords do not match";
      setTimeoutForError();
      return;
    }
    if (password.value.length < 6) {
      error.value = "Password must be at least 6 characters long";
      setTimeoutForError();
      return;
    }
    if (name.value.length < 3) {
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
    }
  };

  return {
    name,
    email,
    password,
    confirmPassword,
    error,
    handleSignup,
  };
};
