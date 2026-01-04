export type BankType =
  | "investment"
  | "axis"
  | "axis_credit"
  | "sbi"
  | "hdfc"
  | "icici"
  | "icici_credit"
  | "others";
export type Currency = "inr" | "usd";

export interface CreateAccountInput {
  name: string;
  bank_type: BankType;
  currency: Currency;
  balance?: number;
  current_value?: number;
}

export interface Account extends CreateAccountInput {
  id: number;
  created_by: number;
}
