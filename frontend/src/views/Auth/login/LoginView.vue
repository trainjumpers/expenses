<template>
  <div class="min-h-screen bg-base-200 flex items-center justify-center">
    <div class="card w-96 bg-base-100 shadow-xl">
      <div class="card-body">
        <div v-if="error" className="toast toast-top toast-end">
          <div className="alert alert-error">
            <span>{{ error }}</span>
          </div>
        </div>
        <h2 class="card-title text-center justify-center mb-4 text-3xl">
          Welcome Back
        </h2>
        <form @submit.prevent="handleLogin">
          <div class="form-control">
            <label class="label">
              <span class="label-text">Email</span>
            </label>
            <input
              type="email"
              v-model="email"
              placeholder="email@example.com"
              class="input input-bordered"
              required
            />
          </div>
          <div class="form-control">
            <label class="label">
              <span class="label-text">Password</span>
            </label>
            <input
              type="password"
              v-model="password"
              placeholder="Enter your password"
              class="input input-bordered"
              required
            />
            <label class="label">
              <a href="#" class="label-text-alt link link-hover"
                >Forgot password?</a
              >
            </label>
          </div>
          <div class="form-control mt-6">
            <button class="btn btn-primary" v-on:click="{ handleLogin }">
              Login
            </button>
          </div>
        </form>
        <div class="divider">OR</div>
        <div class="text-center">
          <router-link to="/register" class="btn w-full">
            Create new account
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLogin } from "./login";

const { email, password, error, handleLogin } = useLogin();
</script>
