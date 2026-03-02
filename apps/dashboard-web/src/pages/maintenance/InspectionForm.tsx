import React, { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";

interface InspectionTemplate {
  id: string;
  name: string;
  description: string;
  items: ChecklistItem[];
}

interface ChecklistItem {
  id: string;
  category: string;
  item: string;
  is_mandatory: boolean;
}

interface InspectionResult {
  item_id: string;
  status: "PASS" | "FAIL" | "NA";
  notes?: string;
  images?: string[];
}

const InspectionForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>(); // If editing existing inspection
  const { uiStrings: rawUiStrings } = useUIStrings("maintenance");
  const uiStrings = rawUiStrings || {};

  const [templates, setTemplates] = useState<InspectionTemplate[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<string>("");
  const [vehicleId, setVehicleId] = useState("");
  const [results, setResults] = useState<Record<string, InspectionResult>>({});
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const [vehicles, setVehicles] = useState<{ id: string; name: string }[]>([]);

  const fetchTemplates = useCallback(async () => {
    try {
      const data = (await apiClient.get("/admin/inspections/templates")) as any;
      const items = data?.templates || data?.data || [];
      setTemplates(items);
      if (items.length > 0 && !selectedTemplate) {
        setSelectedTemplate(items[0].id);
      }
    } catch (error) {
      console.error("Failed to fetch templates:", error);
    }
  }, [selectedTemplate]);

  const fetchVehicles = useCallback(async () => {
    try {
      const data = (await apiClient.get("/admin/vehicles")) as any;
      const items = data?.vehicles || data?.data || [];
      setVehicles(
        items.map((v: any) => ({
          id: v.id,
          name: `${v.make} ${v.model} (${v.license_plate})`,
        })) || [],
      );
    } catch (error) {
      console.error("Failed to fetch vehicles:", error);
    }
  }, []);

  useEffect(() => {
    fetchTemplates();
    fetchVehicles();
  }, [fetchTemplates, fetchVehicles]);

  const handleResultChange = (
    itemId: string,
    status: "PASS" | "FAIL" | "NA",
  ) => {
    setResults((prev) => ({
      ...prev,
      [itemId]: { ...prev[itemId], item_id: itemId, status },
    }));
  };

  const handleNoteChange = (itemId: string, notes: string) => {
    setResults((prev) => ({
      ...prev,
      [itemId]: { ...prev[itemId], item_id: itemId, notes },
    }));
  };

  const handleSubmit = async () => {
    if (!vehicleId || !selectedTemplate) {
      alert(
        uiStrings.err_select_vehicle_template ||
          "Please select a vehicle and a template",
      );
      return;
    }

    setSubmitting(true);
    try {
      const currentTemplate = templates.find((t) => t.id === selectedTemplate);
      const formattedResults =
        currentTemplate?.items.map((item) => ({
          item_id: item.id,
          status: results[item.id]?.status || "PASS", // Default to PASS if skipped? Or validation error?
          notes: results[item.id]?.notes || "",
          images: [],
        })) || [];

      // Validate mandatory items
      const missingMandatory = currentTemplate?.items.filter(
        (item) => item.is_mandatory && !results[item.id]?.status,
      );
      if (missingMandatory && missingMandatory.length > 0) {
        const itemNames = missingMandatory.map((i) => i.item).join(", ");
        alert(
          (
            uiStrings.err_mandatory_items ||
            `Please complete all mandatory items: {items}`
          ).replace("{items}", itemNames),
        );
        setSubmitting(false);
        return;
      }

      await apiClient.post("/admin/inspections", {
        vehicle_id: vehicleId,
        template_id: selectedTemplate,
        inspector_id: localStorage.getItem("user_id") || "",
        results: formattedResults,
        status: "COMPLETED",
      });
      navigate("/maintenance/work-orders");
    } catch (error) {
      console.error("Failed to submit inspection:", error);
    } finally {
      setSubmitting(false);
    }
  };

  const currentTemplateObj = templates.find((t) => t.id === selectedTemplate);

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      {/* Header */}
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate("/maintenance/work-orders")}
              className="p-2 hover:bg-surface-dark rounded-lg transition-colors"
            >
              <span className="material-symbols-outlined text-text-muted">
                arrow_back
              </span>
            </button>
            <div>
              <h1 className="text-2xl font-bold text-white">
                {uiStrings.inspection_title || "New Inspection"}
              </h1>
              <p className="text-sm text-text-muted mt-1">
                {uiStrings.inspection_sub || "Perform a vehicle inspection"}
              </p>
            </div>
          </div>
          <button
            onClick={handleSubmit}
            disabled={submitting}
            className="bg-primary hover:bg-primary/90 text-white text-sm font-medium py-2 px-6 rounded-lg shadow-[0_0_15px_rgba(19,127,236,0.3)] transition-all disabled:opacity-50 flex items-center gap-2"
          >
            {submitting ? (
              <>
                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                {uiStrings.label_submitting || "Submitting..."}
              </>
            ) : (
              <>
                <span className="material-symbols-outlined text-lg">
                  check_circle
                </span>
                {uiStrings.btn_submit_inspection || "Submit Inspection"}
              </>
            )}
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-4xl mx-auto space-y-6">
          {/* Metadata */}
          <div className="bg-surface-dark border border-border-color rounded-xl p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                {uiStrings.label_vehicle || "Vehicle"}
              </label>
              <select
                value={vehicleId}
                onChange={(e) => setVehicleId(e.target.value)}
                className="w-full bg-background-dark border border-border-color rounded-lg px-4 py-2 text-white focus:ring-1 focus:ring-primary focus:border-primary"
              >
                <option value="">
                  {uiStrings.label_select_vehicle || "Select Vehicle"}
                </option>
                {vehicles.map((v) => (
                  <option key={v.id} value={v.id}>
                    {v.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-text-muted mb-2">
                {uiStrings.label_inspection_template || "Inspection Template"}
              </label>
              <select
                value={selectedTemplate}
                onChange={(e) => setSelectedTemplate(e.target.value)}
                className="w-full bg-background-dark border border-border-color rounded-lg px-4 py-2 text-white focus:ring-1 focus:ring-primary focus:border-primary"
              >
                {templates.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Checklist */}
          {currentTemplateObj && (
            <div className="space-y-6">
              {/* Group by category if we wanted, but simplistic list for now */}
              <div className="bg-surface-dark border border-border-color rounded-xl overflow-hidden">
                <div className="bg-surface-darker px-6 py-4 border-b border-border-color">
                  <h3 className="font-semibold text-white">
                    {(
                      uiStrings.label_checklist_header || "{name} Checklist"
                    ).replace("{name}", currentTemplateObj.name)}
                  </h3>
                  <p className="text-sm text-text-muted">
                    {currentTemplateObj.description}
                  </p>
                </div>
                <div className="divide-y divide-border-color">
                  {currentTemplateObj.items.map((item) => (
                    <div
                      key={item.id}
                      className="p-6 hover:bg-surface-darker/30 transition-colors"
                    >
                      <div className="flex flex-col md:flex-row md:items-start gap-4">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="font-medium text-white">
                              {item.item}
                            </span>
                            {item.is_mandatory && (
                              <span className="text-xs text-red-400 bg-red-500/10 px-2 py-0.5 rounded">
                                {uiStrings.label_required || "Required"}
                              </span>
                            )}
                          </div>
                          <p className="text-sm text-text-muted">
                            {item.category}
                          </p>
                        </div>

                        <div className="flex flex-col gap-3 min-w-[200px]">
                          <div className="flex bg-background-dark rounded-lg p-1 border border-border-color">
                            {(["PASS", "FAIL", "NA"] as const).map((status) => (
                              <button
                                key={status}
                                onClick={() =>
                                  handleResultChange(item.id, status)
                                }
                                className={`flex-1 py-1 px-3 text-xs font-medium rounded transition-all ${
                                  results[item.id]?.status === status
                                    ? status === "PASS"
                                      ? "bg-emerald-500 text-white shadow-sm"
                                      : status === "FAIL"
                                        ? "bg-red-500 text-white shadow-sm"
                                        : "bg-gray-500 text-white shadow-sm"
                                    : "text-text-muted hover:text-white hover:bg-surface-light"
                                }`}
                              >
                                {status}
                              </button>
                            ))}
                          </div>

                          {results[item.id]?.status === "FAIL" && (
                            <textarea
                              placeholder={
                                uiStrings.placeholder_describe_issue ||
                                "Describe the issue..."
                              }
                              value={results[item.id]?.notes || ""}
                              onChange={(e) =>
                                handleNoteChange(item.id, e.target.value)
                              }
                              className="w-full bg-background-dark border border-border-color rounded-lg px-3 py-2 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary min-h-[60px]"
                            />
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default InspectionForm;
