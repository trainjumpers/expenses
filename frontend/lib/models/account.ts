export type BankType = "investment" | "axis" | "sbi" | "hdfc" | "icici" | "others";
export type Currency = "inr" | "usd";

export interface CreateAccountInput {
  name: string;
  bank_type: BankType;
  currency: Currency;
  balance?: number;
}

export interface Account extends CreateAccountInput {
  id: number;
  created_by: number;
}
