import { User } from "./user";

export interface AuthResponse {
  access_token: string;
  data: User;
  message: string;
  refresh_token: string;
}
