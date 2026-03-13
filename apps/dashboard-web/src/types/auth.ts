export interface SignupRequest {
  name: string;
  email: string;
  password?: string;
}

export interface VerifyEmailRequest {
  token: string;
}

export interface ResetPasswordRequest {
  token: string;
  newPassword?: string;
}

export interface ChangePasswordRequest {
  currentPassword?: string;
  newPassword?: string;
}

export interface DeviceApprovalRequest {
  fingerprint: string;
  approve: boolean;
}

export interface SSOCallbackRequest {
  code?: string;
  state?: string;
  token?: string;
}

export interface AuthErrorResponse {
  error: string;
  message?: string;
  code?: string;
}

export interface MFASetupResponse {
  secret: string;
  qrCode: string;
}

export interface TrustedDevice {
  id: string;
  name: string;
  ip: string;
  location: string;
  lastUsed: string;
  isCurrent: boolean;
}

export interface SecurityKey {
  id: string;
  nickname: string;
  lastUsed: string;
}

export interface PasswordPolicy {
  minLength: number;
  requireUppercase: boolean;
  requireLowercase: boolean;
  requireNumbers: boolean;
  requireSpecial: boolean;
  historyCount: number;
  expiryDays: number;
  breachDetection: boolean;
}

export interface AccountLockInfo {
  isLocked: boolean;
  reason: string;
  remainingSeconds: number;
}

export interface RiskChallengeResponse {
  challengeId: string;
  type: "otp" | "device_verify" | "identity_confirm";
}

export interface APIKey {
  id: string;
  name: string;
  key?: string; // Only present on creation
  createdAt: string;
  expiresAt: string;
  status: "active" | "revoked" | "expired";
}

export interface SecurityTelemetry {
  loginFailures: { timestamp: string; count: number }[];
  mfaChallenges: { method: string; count: number }[];
  riskEvents: {
    type: string;
    severity: "low" | "medium" | "high";
    timestamp: string;
  }[];
  authMethodDistribution: { method: string; percentage: number }[];
}

export interface AuthStats {
  totalLogins: number;
  mfaAdoptionRate: number;
  activeSessions: number;
  failedAuthRate: number;
  deviceDistribution: { type: string; count: number }[];
  dailyLoginHistory: { date: string; logins: number }[];
}

export interface UserSession {
  id: string;
  userId: string;
  userName: string;
  deviceName: string;
  deviceType: "mobile" | "desktop" | "tablet";
  ipAddress: string;
  location: string;
  lastActive: string;
  status: "active" | "inactive";
}
