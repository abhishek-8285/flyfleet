import React, { useEffect, useState, useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../../services/api";
import { Driver, DriverDocument, DriverAnalytics } from "../../modules/drivers/types";

type TabType = "overview" | "behavior" | "documents";

/**
 * DriverDetail: Driver profile detail page with 2/3 + 1/3 layout.
 * Follows the VehicleDetail pattern with tabs (Overview/Behavior/Documents).
 */
export const DriverDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [driver, setDriver] = useState<Driver | null>(null);
  const [analytics, setAnalytics] = useState<DriverAnalytics | null>(null);
  const [documents, setDocuments] = useState<DriverDocument[]>([]);
  const [behaviorScore, setBehaviorScore] = useState<number>(0);
  const [activeTab, setActiveTab] = useState<TabType>("overview");
  const [loading, setLoading] = useState(true);

  const loadData = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const [driverData, analyticsData, docsData, scoreData] = await Promise.all([
        api.getDriver(id).catch(() => null),
        api.getAnalytics(id).catch(() => null),
        api.getDocuments(id).catch(() => []),
        api.getBehaviorScore(id).catch(() => null),
      ]);

      setDriver(driverData);
      setAnalytics(analyticsData);
      setDocuments(docsData || []);
      setBehaviorScore(scoreData?.behavior_score || 0);
    } catch (error) {
      console.error("Failed to load driver data:", error);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      active: "bg-green-500/20 text-green-400",
      inactive: "bg-gray-500/20 text-gray-400",
      suspended: "bg-red-500/20 text-red-400",
      resigned: "bg-yellow-500/20 text-yellow-400",
    };
    return styles[status] || "bg-gray-500/20 text-gray-400";
  };

  const getScoreColor = (score: number) => {
    if (score >= 90) return "text-green-400";
    if (score >= 70) return "text-yellow-400";
    if (score >= 50) return "text-orange-400";
    return "text-red-400";
  };

  const getScoreBarColor = (score: number) => {
    if (score >= 90) return "bg-green-400";
    if (score >= 70) return "bg-yellow-400";
    if (score >= 50) return "bg-orange-400";
    return "bg-red-400";
  };

  const getLicenseStatus = (expiry: string) => {
    const exp = new Date(expiry);
    const now = new Date();
    const diff = (exp.getTime() - now.getTime()) / (1000 * 60 * 60 * 24);
    if (diff < 0) return { label: "Expired", color: "text-red-400" };
    if (diff < 30) return { label: "Expiring Soon", color: "text-yellow-400" };
    return { label: "Valid", color: "text-green-400" };
  };

  if (loading) {
    return (
      <div className="w-full flex items-center justify-center py-20">
        <span className="material-symbols-outlined animate-spin text-primary text-4xl">progress_activity</span>
      </div>
    );
  }

  if (!driver) {
    return (
      <div className="w-full text-center py-20 text-text-muted">
        <p>Driver not found.</p>
        <button onClick={() => navigate("/admin/drivers")} className="text-primary hover:underline mt-2">
          Back to Drivers
        </button>
      </div>
    );
  }

  const licenseStatus = getLicenseStatus(driver.license_expiry);

  return (
    <div className="w-full max-w-6xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <button
          onClick={() => navigate("/admin/drivers")}
          className="text-text-muted hover:text-white text-sm mb-2 flex items-center gap-1"
        >
          <span className="material-symbols-outlined text-[18px]">arrow_back</span>
          Back to Drivers
        </button>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-primary/20 flex items-center justify-center text-primary text-xl font-semibold">
              {driver.first_name[0]}{driver.last_name[0]}
            </div>
            <div>
              <h1 className="text-white text-2xl font-bold">{driver.first_name} {driver.last_name}</h1>
              <div className="flex items-center gap-3 mt-1">
                <span className="text-text-muted text-sm">{driver.email}</span>
                <span className="text-text-muted">·</span>
                <span className="text-text-muted text-sm">{driver.phone}</span>
              </div>
              <div className="flex items-center gap-2 mt-2">
                <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${getStatusBadge(driver.status)}`}>
                  {driver.status}
                </span>
                {driver.employee_code && (
                  <span className="text-text-muted text-xs">{driver.employee_code}</span>
                )}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => navigate(`/admin/drivers/${id}/edit`)}
              className="px-3 py-2 bg-surface-dark border border-border-color text-white text-sm font-medium rounded-lg hover:bg-border-color transition-colors flex items-center gap-1"
            >
              <span className="material-symbols-outlined text-[16px]">edit</span>
              Edit
            </button>
          </div>
        </div>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-surface-dark border border-border-color rounded-lg p-4">
          <div className="text-text-muted text-xs">Behavior Score</div>
          <div className={`text-2xl font-bold ${getScoreColor(behaviorScore)}`}>{behaviorScore}</div>
          <div className="w-full bg-surface-darker rounded-full h-1.5 mt-2">
            <div className={`h-1.5 rounded-full ${getScoreBarColor(behaviorScore)}`} style={{ width: `${behaviorScore}%` }} />
          </div>
        </div>
        <div className="bg-surface-dark border border-border-color rounded-lg p-4">
          <div className="text-text-muted text-xs">Rating</div>
          <div className="text-2xl font-bold text-white">{driver.rating.toFixed(1)}</div>
          <div className="text-text-muted text-xs mt-1">out of 5.0</div>
        </div>
        <div className="bg-surface-dark border border-border-color rounded-lg p-4">
          <div className="text-text-muted text-xs">Total Trips</div>
          <div className="text-2xl font-bold text-white">{driver.total_trips}</div>
        </div>
        <div className="bg-surface-dark border border-border-color rounded-lg p-4">
          <div className="text-text-muted text-xs">Distance</div>
          <div className="text-2xl font-bold text-white">{driver.total_distance_km.toFixed(0)} km</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 mb-6 border-b border-border-color">
        {(["overview", "behavior", "documents"] as TabType[]).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium capitalize transition-colors relative ${
              activeTab === tab ? "text-white" : "text-text-muted hover:text-white"
            }`}
          >
            {tab}
            {activeTab === tab && (
              <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary" />
            )}
          </button>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content (2/3) */}
        <div className="lg:col-span-2">
          {activeTab === "overview" && (
            <div className="bg-surface-dark border border-border-color rounded-lg p-6 space-y-6">
              <h3 className="text-white font-semibold">Personal Information</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-text-muted text-xs">Date of Birth</div>
                  <div className="text-white">{new Date(driver.date_of_birth).toLocaleDateString()}</div>
                </div>
                <div>
                  <div className="text-text-muted text-xs">Employment Type</div>
                  <div className="text-white">{driver.employment_type}</div>
                </div>
                <div>
                  <div className="text-text-muted text-xs">Experience</div>
                  <div className="text-white">{driver.experience_years} years</div>
                </div>
                <div>
                  <div className="text-text-muted text-xs">Joining Date</div>
                  <div className="text-white">{new Date(driver.joining_date).toLocaleDateString()}</div>
                </div>
              </div>

              <h3 className="text-white font-semibold pt-4 border-t border-border-color">Address</h3>
              <div className="text-white">
                {driver.address?.street && <div>{driver.address.street}</div>}
                {(driver.address?.city || driver.address?.state) && (
                  <div>{[driver.address.city, driver.address.state].filter(Boolean).join(", ")}</div>
                )}
                {driver.address?.pincode && <div>{driver.address.pincode}</div>}
              </div>

              <h3 className="text-white font-semibold pt-4 border-t border-border-color">Emergency Contact</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-text-muted text-xs">Name</div>
                  <div className="text-white">{driver.emergency_contact?.name || "—"}</div>
                </div>
                <div>
                  <div className="text-text-muted text-xs">Phone</div>
                  <div className="text-white">{driver.emergency_contact?.phone || "—"}</div>
                </div>
                <div>
                  <div className="text-text-muted text-xs">Relationship</div>
                  <div className="text-white">{driver.emergency_contact?.relation || "—"}</div>
                </div>
              </div>
            </div>
          )}

          {activeTab === "behavior" && (
            <div className="bg-surface-dark border border-border-color rounded-lg p-6 space-y-6">
              <h3 className="text-white font-semibold">Behavior Analytics</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-text-muted text-xs">Behavior Score</div>
                  <div className={`text-3xl font-bold ${getScoreColor(behaviorScore)}`}>{behaviorScore}/100</div>
                </div>
                <div>
                  <div className="text-text-muted text-xs">Avg Behavior Score (30d)</div>
                  <div className="text-3xl font-bold text-white">{analytics?.avg_behavior_score?.toFixed(1) || "—"}</div>
                </div>
                <div>
                  <div className="text-text-muted text-xs">Total Hours Driven</div>
                  <div className="text-xl font-semibold text-white">{driver.total_hours_driven.toFixed(1)}h</div>
                </div>
                <div>
                  <div className="text-text-muted text-xs">Incidents Count</div>
                  <div className="text-xl font-semibold text-white">{analytics?.incidents_count || 0}</div>
                </div>
              </div>

              {analytics?.trend && (
                <div className="pt-4 border-t border-border-color">
                  <div className="text-text-muted text-xs mb-2">Trend</div>
                  <div className={`text-lg font-semibold capitalize ${
                    analytics.trend === "improving" ? "text-green-400" :
                    analytics.trend === "declining" ? "text-red-400" :
                    "text-yellow-400"
                  }`}>
                    {analytics.trend}
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === "documents" && (
            <div className="bg-surface-dark border border-border-color rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-white font-semibold">Documents</h3>
                <button className="px-3 py-1.5 bg-primary text-white text-sm rounded-lg hover:bg-primary-dark transition-colors flex items-center gap-1">
                  <span className="material-symbols-outlined text-[16px]">add</span>
                  Upload
                </button>
              </div>

              {documents.length === 0 ? (
                <div className="text-center py-8 text-text-muted">
                  <span className="material-symbols-outlined text-4xl mb-2">folder_open</span>
                  <p>No documents uploaded</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {documents.map((doc) => (
                    <div key={doc.id} className="flex items-center justify-between p-3 bg-surface-darker rounded-lg">
                      <div className="flex items-center gap-3">
                        <span className="material-symbols-outlined text-text-muted">description</span>
                        <div>
                          <div className="text-white text-sm font-medium">{doc.document_type}</div>
                          <div className="text-text-muted text-xs">{doc.document_number}</div>
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        <div className="text-right">
                          <div className={`text-xs ${
                            new Date(doc.expiry_date) < new Date() ? "text-red-400" : "text-green-400"
                          }`}>
                            {new Date(doc.expiry_date).toLocaleDateString()}
                          </div>
                          {doc.verified && (
                            <div className="text-green-400 text-xs flex items-center gap-1">
                              <span className="material-symbols-outlined text-[14px]">verified</span>
                              Verified
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Sidebar (1/3) */}
        <div className="space-y-4">
          {/* Current Assignment */}
          <div className="bg-surface-dark border border-border-color rounded-lg p-4">
            <h4 className="text-white font-semibold text-sm mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-[18px] text-text-muted">directions_car</span>
              Current Assignment
            </h4>
            {driver.current_vehicle_id ? (
              <div className="text-white text-sm">Vehicle: {driver.current_vehicle_id.slice(0, 8)}...</div>
            ) : (
              <div className="text-text-muted text-sm">No vehicle assigned</div>
            )}
            {driver.current_trip_id && (
              <div className="text-primary text-xs mt-1">On trip: {driver.current_trip_id.slice(0, 8)}...</div>
            )}
          </div>

          {/* License */}
          <div className="bg-surface-dark border border-border-color rounded-lg p-4">
            <h4 className="text-white font-semibold text-sm mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-[18px] text-text-muted">badge</span>
              License
            </h4>
            <div className="space-y-2 text-sm">
              <div>
                <div className="text-text-muted text-xs">Number</div>
                <div className="text-white font-mono">{driver.license_number}</div>
              </div>
              <div>
                <div className="text-text-muted text-xs">Type</div>
                <div className="text-white">{driver.license_type || "—"}</div>
              </div>
              <div>
                <div className="text-text-muted text-xs">Status</div>
                <div className={licenseStatus.color}>{licenseStatus.label}</div>
              </div>
              <div>
                <div className="text-text-muted text-xs">Expiry</div>
                <div className="text-white">{new Date(driver.license_expiry).toLocaleDateString()}</div>
              </div>
              <div>
                <div className="text-text-muted text-xs">Authority</div>
                <div className="text-white">{driver.license_authority || "—"}</div>
              </div>
            </div>
          </div>

          {/* Contact */}
          <div className="bg-surface-dark border border-border-color rounded-lg p-4">
            <h4 className="text-white font-semibold text-sm mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-[18px] text-text-muted">phone</span>
              Contact
            </h4>
            <div className="space-y-2 text-sm">
              <div>
                <div className="text-text-muted text-xs">Email</div>
                <div className="text-white">{driver.email}</div>
              </div>
              <div>
                <div className="text-text-muted text-xs">Phone</div>
                <div className="text-white">{driver.phone}</div>
              </div>
            </div>
          </div>

          {/* System Info */}
          <div className="bg-surface-dark border border-border-color rounded-lg p-4">
            <h4 className="text-white font-semibold text-sm mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-[18px] text-text-muted">info</span>
              System Info
            </h4>
            <div className="space-y-2 text-xs">
              <div>
                <div className="text-text-muted">ID</div>
                <div className="text-white font-mono">{driver.id.slice(0, 12)}...</div>
              </div>
              <div>
                <div className="text-text-muted">Created</div>
                <div className="text-white">{new Date(driver.created_at).toLocaleDateString()}</div>
              </div>
              <div>
                <div className="text-text-muted">Updated</div>
                <div className="text-white">{new Date(driver.updated_at).toLocaleDateString()}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
