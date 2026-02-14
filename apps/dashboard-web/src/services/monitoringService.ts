import axios from 'axios';

// Create a specific instance for monitoring to use port 8081 (web-bff) directly if needed
// or rely on the main apiClient if proxy is set up.
// For now, we will use a direct instance to ensure we hit the new BFF on 8081
// since the main apiClient defaults to 8080 (which might be another gateway).

const MONITORING_API_URL = process.env.REACT_APP_MONITORING_API_URL || '/api/v1/admin/monitoring';

const monitoringClient = axios.create({
  baseURL: MONITORING_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Attach token from localStorage
monitoringClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export interface SystemHealthData {
  status: string;
  uptime_percent: number;
  avg_latency_ms: number;
  error_rate_percent: number;
  database_status: string;
  queue_backlog: number;
  active_integrations: {
    total: number;
    active: number;
  };
  top_errors: Array<{
    service: string;
    code: number;
    type: string;
    count: number;
  }>;
}

export interface MetricsData {
  throughput: number;
  db_latency_ms: number;
  cache_hit_ratio: number;
  cluster_usage: Array<{
    service: string;
    cpu: number;
    memory_gb: number;
    status: string;
  }>;
}

export interface QueueData {
  running_jobs: number;
  pending_queue: number;
  failed_24h: number;
  dlq_count: number;
  active_jobs: Array<{
    id: string;
    type: string;
    progress: number;
    duration: string;
  }>;
}

export interface SLAData {
  global_uptime_90d: number;
  mttr_minutes: number;
  error_budget_remaining: string;
  active_incidents: number;
  heatmap: number[];
}

export interface LogEntry {
  timestamp: string;
  service: string;
  severity: 'CRITICAL' | 'ERROR' | 'WARNING' | 'INFO';
  type: string;
  message: string;
  count?: number;
}

export interface LogsResponse {
  logs: LogEntry[];
  total: number;
}

export const monitoringService = {
  getSystemHealth: async (): Promise<SystemHealthData> => {
    const response = await monitoringClient.get('/health');
    return response.data;
  },

  getMetrics: async (): Promise<MetricsData> => {
    const response = await monitoringClient.get('/metrics');
    return response.data;
  },

  getQueues: async (): Promise<QueueData> => {
    const response = await monitoringClient.get('/queues');
    return response.data;
  },

  getSLA: async (): Promise<SLAData> => {
    const response = await monitoringClient.get('/sla');
    return response.data;
  },

  getLogs: async (filters?: any): Promise<LogsResponse> => {
    const response = await monitoringClient.get('/logs', { params: filters });
    return response.data;
  }
};
