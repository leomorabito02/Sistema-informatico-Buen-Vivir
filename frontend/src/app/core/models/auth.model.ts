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
  role_data?: RoleData;
  image_url?: string;
}

export interface RoleData {
  birthdate?: string;
  dni?: string;
  address?: string;
  gender?: string;
  emergency_contact_name?: string;
  emergency_contact_phone?: string;
  specialization?: string;
  is_external?: boolean;
  business_name?: string;
  cuit?: string;
  approval_status?: string;
  approval_date?: string;
  approved_by?: string;
  [key: string]: string | boolean | null | undefined;
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

export interface MeResponse {
  user: User;
  role_id?: string;
  approval_status?: string;
  user_status?: string;
}

export interface ProfileResponse {
  user: User;
  role_data?: RoleData;
}

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

export interface UpdateBaseProfile {
  email?: string;
  phone?: string;
}

export interface UpdateParticipantProfile extends UpdateBaseProfile {
  birthdate?: string;
  dni?: string;
  address?: string;
  gender?: string;
  emergency_contact_name?: string;
  emergency_contact_phone?: string;
}

export interface UpdateCompanyProfile extends UpdateBaseProfile {
  business_name?: string;
  cuit?: string;
}

export interface UpdateTeacherProfile extends UpdateBaseProfile {
  specialization?: string;
  is_external?: boolean;
}

export interface UpdateVolunteerProfile extends UpdateBaseProfile {
}
