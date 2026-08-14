export interface Role {
  role_id: string;
  name: string;
  description?: string;
  created_at?: string;
  updated_at?: string;
}

export interface User {
  user_id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  status: string;
  google_auth_id?: string;
  role_id: string;
  role?: Role;
  twofa_enabled: boolean;
  created_at: string;
  updated_at: string;
}

/** Step 1 response: list of accounts associated with the email */
export interface AvailableAccount {
  user_id: string;
  role_id: string;
  role_name: string;
  status: string;
}

export interface CheckCredentialsResponse {
  accounts: AvailableAccount[];
  login_nonce: string;
}

/** Step 2 response: JWT token after selecting account/role */
export interface LoginResponse {
  token: string;
  refresh_token?: string;
  user: User;
  approval_status?: string;
  user_status?: string;
}

export interface RegisterRequest {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  password?: string;
  role: string;
}

export interface UpdateParticipantProfile {
  birthdate?: string;
  dni?: string;
  address?: string;
  gender?: string;
  emergency_contact_name?: string;
  emergency_contact_phone?: string;
}

export interface UpdateCompanyProfile {
  business_name?: string;
  cuit?: string;
}

export interface UpdateTeacherProfile {
  specialization?: string;
  is_external?: boolean;
}
