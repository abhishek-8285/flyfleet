import apiClient from "./apiClient";

export const dashboardApi = {
  getOverview: () => apiClient.get("/dashboard/overview"),
  getStats: () => apiClient.get("/dashboard/stats"),
  getData: () => apiClient.get("/dashboard/data"),
  getTimeline: (params?: any) => apiClient.get("/timeline", { params }),
  getTimelineStats: () => apiClient.get("/timeline/stats"),
  getAlerts: (params?: any) => apiClient.get("/alerts", { params }),
  getAnalyticsTrips: () => apiClient.get("/admin/analytics/trips"),
  getAnalyticsDrivers: () => apiClient.get("/admin/analytics/drivers"),
  getAnalyticsFleet: (period: string) =>
    apiClient.get("/admin/analytics/fleet/period", { params: { period } }),
  getVehicles: () => apiClient.get("/admin/vehicles"),
  getTrips: () => apiClient.get("/admin/trips"),
  acknowledgeAlert: (id: string) => apiClient.post(`/alerts/${id}/acknowledge`),
};

export default dashboardApi;
