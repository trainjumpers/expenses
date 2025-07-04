"use client";

import { useUser } from "@/components/custom/Provider/UserProvider";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { useTheme } from "next-themes";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useUser();
  const { theme } = useTheme();

  const [formData, setFormData] = useState({ email: "", password: "" });
  const [loading, setLoading] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await login(formData.email, formData.password);
      router.push("/");
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
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
          Sign in to your account
        </h2>
        <form className="space-y-4" onSubmit={handleSubmit}>
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
              className="bg-white/40 dark:bg-input/40 backdrop-blur-md border border-border focus:border-primary/60 shadow-inner focus:shadow-lg transition-all duration-200"
            />
          </div>
          <Button
            type="submit"
            className="w-full font-semibold shadow-lg shadow-primary/10 hover:shadow-xl transition-all duration-200"
            disabled={loading}
          >
            {loading && <Spinner />}
            Sign In
          </Button>
        </form>
        <p className="text-center text-sm mt-4 text-foreground/70">
          Don&apos;t have an account?{" "}
          <Link href="/signup" className="text-primary hover:underline">
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
}
