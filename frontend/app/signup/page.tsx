"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { signup } from "@/lib/api/auth";
import { ACCESS_TOKEN_NAME, REFRESH_TOKEN_NAME } from "@/lib/constants/cookie";
import { setCookie } from "@/lib/utils/cookies";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

export default function SignupPage() {
  const router = useRouter();

  const [formData, setFormData] = useState({
    name: "",
    email: "",
    password: "",
  });
  const [loading, setLoading] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const data = await signup(
        formData.name,
        formData.email,
        formData.password
      );
      setCookie(ACCESS_TOKEN_NAME, data.access_token, {
        maxAge: 12 * 60 * 60 * 1000,
      });
      setCookie(REFRESH_TOKEN_NAME, data.refresh_token, {
        maxAge: 7 * 24 * 60 * 60 * 1000,
      });
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
            />
          </div>
          <Button type="submit" className="w-full" disabled={loading}>
            {loading && <Spinner />}
            Sign Up
          </Button>
        </form>
        <p className="text-center text-sm mt-4">
          Already have an account?{" "}
          <Link href="/login" className="text-primary hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
