"use client";

import { useUser } from "@/components/custom/Provider/UserProvider";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useState } from "react";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useUser();

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
      console.log(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background bg-center">
      <div className="w-full max-w-md p-8 space-y-6 bg-card rounded-xl shadow-lg border border-border">
        <h2 className="text-2xl font-bold text-center mb-6">
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
            />
          </div>
          <Button type="submit" className="w-full" disabled={loading}>
            {loading && <Spinner />}
            Sign In
          </Button>
        </form>
        <p className="text-center text-sm mt-4">
          Don&apos;t have an account?{" "}
          <Link href="/signup" className="text-primary hover:underline">
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
}
