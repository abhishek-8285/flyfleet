export type GovernanceRequestType = 'EXPORT' | 'DELETION' | 'ACCESS_REVIEW';
export type GovernanceRequestStatus = 'PENDING' | 'APPROVED' | 'REJECTED' | 'EXPIRED' | 'COMPLETED';

export interface GovernanceRequest {
  id: string;
  organization_id: string;
  requester_id: string;
  request_type: GovernanceRequestType;
  data_scope: string;
  purpose: string;
  format: string;
  delivery_method: string;
  status: GovernanceRequestStatus;
  expiry_date?: string;
  comments?: string;
  created_at: string;
  updated_at: string;
}

export interface GovernanceSettings {
  organization_id: string;
  retention_policies: Record<string, any>;
  masking_rules: Record<string, any>;
  auto_expiry_enabled: boolean;
  download_ttl_hours: number;
  updated_at: string;
}

export interface CreateGovernanceRequest {
  request_type: GovernanceRequestType;
  data_scope: string;
  purpose: string;
  format: string;
  delivery_method: string;
}

export interface GovernanceRequestFilter {
  type?: GovernanceRequestType;
  status?: GovernanceRequestStatus[];
  search?: string;
  page?: number;
  page_size?: number;
}

export interface GovernanceRequestListResponse {
  total: number;
  page: number;
  page_size: number;
  requests: GovernanceRequest[];
}
