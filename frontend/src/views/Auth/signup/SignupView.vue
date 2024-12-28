<template>
  <div class="min-h-screen bg-base-200 flex items-center justify-center">
    <div class="card w-96 bg-base-100 shadow-xl">
      <div class="card-body">
        <div v-if="error" className="toast toast-top toast-end">
          <div className="alert alert-error">
            <span>{{ error }}</span>
          </div>
        </div>

        <h2
          class="card-title text-2xl font-bold text-center justify-center mb-2"
        >
          Join us
        </h2>
        <form @submit.prevent="handleSignup">
          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">Name</span>
            </label>
            <input
              type="text"
              v-model="name"
              placeholder="Enter your name"
              class="input input-bordered w-full"
              required
            />
          </div>

          <div class="form-control w-full mt-2">
            <label class="label">
              <span class="label-text">Email</span>
            </label>
            <input
              type="email"
              v-model="email"
              placeholder="Enter your email"
              class="input input-bordered w-full"
              required
            />
          </div>

          <div class="form-control w-full mt-4">
            <label class="label">
              <span class="label-text">Password</span>
            </label>
            <input
              type="password"
              v-model="password"
              placeholder="Enter your password"
              class="input input-bordered w-full"
              required
            />
          </div>

          <div class="form-control w-full mt-4">
            <label class="label">
              <span class="label-text">Confirm Password</span>
            </label>
            <input
              type="password"
              v-model="confirmPassword"
              placeholder="Confirm your password"
              class="input input-bordered w-full"
              required
            />
          </div>

          <div class="mt-6">
            <LoadingButton
              className="btn btn-primary w-full"
              :loading="loading"
              :handleClick="handleSignup"
            >
              Signup
            </LoadingButton>
          </div>
          <div class="divider">OR</div>
          <div class="text-center mt-4">
            <router-link to="/login" class="btn w-full"
              >Already have an account? Login</router-link
            >
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import LoadingButton from "@/components/Button/LoadingButton.vue";

import { useSignup } from "./signup";

const { name, email, password, confirmPassword, error, loading, handleSignup } =
  useSignup();
</script>
