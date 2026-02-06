export enum Screen {
  LOGIN = 'LOGIN',
  MFA = 'MFA',
  ORG_SELECTION = 'ORG_SELECTION',
}

export interface Organization {
  id: string;
  name: string;
  role: string;
  lastUsed?: boolean;
}

export interface User {
  name: string;
  email: string;
  avatarUrl: string;
}