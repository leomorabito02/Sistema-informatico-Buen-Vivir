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
  twofa_method?: string;
  is_email_verified?: boolean;
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
  approval_status?: string;
  is_email_verified?: boolean;
}

export interface CheckCredentialsResponse {
  accounts: AvailableAccount[];
  login_nonce: string;
  is_email_verified?: boolean;
  email?: string;
  requires_2fa?: boolean;
  twofa_method?: string;
  twofa_nonce?: string;
}

export interface VerifyTwofaRequest {
  twofa_nonce: string;
  code: string;
  trust_device?: boolean;
}

export interface SetupTwofaResponse {
  method: string;
  secret?: string;
  otpauth_url?: string;
  message?: string;
}

export interface EnableTwofaRequest {
  method: string;
  code: string;
}

export interface DisableTwofaRequest {
  password?: string;
  code?: string;
}

export interface MeResponse {
  user: User;
  role_id?: string;
  approval_status?: string;
  user_status?: string;
  /** Docente externo (viene del token): solo ve "Mis contenidos" (HU-23). */
  is_external?: boolean;
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
  /** Solo con role "teacher": docente externo (biblioteca pública). */
  is_external?: boolean;
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

export interface UpdateVolunteerProfile extends UpdateBaseProfile {}
