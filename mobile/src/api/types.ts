export interface User {
  id: string;
  email: string;
  full_name: string;
  household_id: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface HouseholdMember {
  id: string;
  full_name: string;
  email: string;
}

export interface Household {
  id: string;
  name: string;
  invite_code: string;
  members: HouseholdMember[];
}

export interface ReceiptDraft {
  receipt_id: string;
  image_url: string;
  status: string;
  parsed_vendor: string;
  parsed_date: string | null;
  parsed_amount: number | null;
  raw_ocr_text: string;
  confidence: number;
  suggested_category: string;
  warranty_expires_at: string;
  warranty_uncertain: boolean;
}

export interface Product {
  id: string;
  household_id: string;
  name: string;
  category: string;
  brand: string;
  purchase_date: string;
  price: number | null;
  room: string;
  warranty_expires_at: string;
  warranty_uncertain: boolean;
  photo_url: string;
  receipt_id: string | null;
  created_at: string;
  updated_at: string;
}

export type ClaimStatus = 'open' | 'in_progress' | 'closed';

export interface WarrantyClaim {
  id: string;
  product_id: string;
  issue_description: string;
  status: ClaimStatus;
  created_at: string;
  resolved_at: string | null;
}

export interface WarrantyResolution {
  warranty_expires_at: string;
  duration_months: number;
  uncertain: boolean;
  source: string;
}

export interface ManufacturerContact {
  brand: string;
  phone: string;
  website: string;
}

export interface Receipt {
  id: string;
  household_id: string;
  image_url: string;
  raw_ocr_text: string;
  parsed_vendor: string;
  parsed_date: string | null;
  parsed_amount: number | null;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface ApiError {
  error: string;
}
