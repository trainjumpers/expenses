"use client";

import { useSignup } from "@/components/hooks/useUser";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { useTheme } from "next-themes";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

export default function SignupPage() {
  const router = useRouter();
  const { theme } = useTheme();
  const signupMutation = useSignup();

  const [formData, setFormData] = useState({
    name: "",
    email: "",
    password: "",
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    signupMutation.mutate(
      {
        name: formData.name,
        email: formData.email,
        password: formData.password,
      },
      {
        onSuccess: () => {
          router.push("/");
        },
      }
    );
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center bg-center bg-cover"
      style={{
        backgroundImage: `url(${theme === "dark" ? "/dark-bg.png" : "/light-bg.png"})`,
      }}
    >
      <div className="w-full max-w-md p-8 space-y-6 rounded-2xl border border-border shadow-2xl bg-white/30 dark:bg-card/40 backdrop-blur-xl backdrop-saturate-150 transition-all duration-300">
        <h2 className="text-2xl font-bold text-center mb-6 drop-shadow-md text-foreground/90">
          Create your account
        </h2>
        <form className="space-y-3" onSubmit={handleSubmit}>
          <div>
            <Input
              type="text"
              id="name"
              name="name"
              required
              placeholder="Name"
              value={formData.name}
              onChange={handleChange}
              className="bg-white/40 dark:bg-input/40 backdrop-blur-md border border-border focus:border-primary/60 shadow-inner focus:shadow-lg transition-all duration-200"
              disabled={signupMutation.isPending}
            />
          </div>
          <div>
            <Input
              type="email"
              id="email"
              name="email"
              required
              placeholder="Email"
              value={formData.email}
              onChange={handleChange}
              className="bg-white/40 dark:bg-input/40 backdrop-blur-md border border-border focus:border-primary/60 shadow-inner focus:shadow-lg transition-all duration-200"
              disabled={signupMutation.isPending}
            />
          </div>
          <div>
            <Input
              type="password"
              id="password"
              name="password"
              required
              placeholder="Password"
              value={formData.password}
              onChange={handleChange}
              minLength={8}
              className="bg-white/40 dark:bg-input/40 backdrop-blur-md border border-border focus:border-primary/60 shadow-inner focus:shadow-lg transition-all duration-200"
              disabled={signupMutation.isPending}
            />
          </div>
          <Button
            type="submit"
            className="w-full font-semibold shadow-lg shadow-primary/10 hover:shadow-xl transition-all duration-200"
            disabled={signupMutation.isPending}
          >
            {signupMutation.isPending && <Spinner />}
            Sign Up
          </Button>
        </form>
        <p className="text-center text-sm mt-4 text-foreground/70">
          Already have an account?{" "}
          <Link href="/login" className="text-primary hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
