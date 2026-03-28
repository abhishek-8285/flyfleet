import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

interface IPRule {
  id: string;
  org_id: string;
  cidr: string;
  action: "ALLOW" | "BLOCK";
  description: string;
  added_by: string;
  created_at: string;
}

const API_BASE = import.meta.env.VITE_API_URL || "/api/v1";

export default function IPAccess() {
  const [rules, setRules] = useState<IPRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [uiStrings, setUiStrings] = useState<any>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [newRule, setNewRule] = useState({
    cidr: "",
    action: "BLOCK" as "ALLOW" | "BLOCK",
    description: "",
  });

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) =>
        setUiStrings(
          data.ui_strings?.security_v2 ||
            data.system_metadata?.ui_strings?.security_v2,
        ),
      )
      .catch((err) => console.error("Failed to fetch UI strings", err));
  }, []);
  const [searchParams] = useSearchParams();
  const orgId = searchParams.get("org_id") || "";

  useEffect(() => {
    fetchRules();
  }, [orgId]);

  const fetchRules = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/security/ip-rules?org_id=${orgId}`);
      const data = await res.json();
      setRules(data || []);
    } catch (err) {
      console.error("Failed to fetch IP rules:", err);
    }
    setLoading(false);
  };

  const handleAddRule = async () => {
    try {
      await fetch(`${API_BASE}/security/ip-rules?org_id=${orgId}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newRule),
      });
      setShowAddModal(false);
      setNewRule({ cidr: "", action: "BLOCK", description: "" });
      fetchRules();
    } catch (err) {
      console.error("Failed to add IP rule:", err);
    }
  };

  const handleDeleteRule = async (id: string) => {
    if (
      !window.confirm(
        uiStrings?.confirm_delete_rule ||
          "Are you sure you want to delete this rule?",
      )
    )
      return;
    try {
      await fetch(`${API_BASE}/security/ip-rules/${id}`, {
        method: "DELETE",
      });
      fetchRules();
    } catch (err) {
      console.error("Failed to delete IP rule:", err);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString();
  };

  const allowedRules = rules.filter((r) => r.action === "ALLOW");
  const blockedRules = rules.filter((r) => r.action === "BLOCK");

  return (
    <div className="p-6 bg-gray-900 min-h-screen text-white">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">
            {uiStrings?.ip_access_title || "IP Access Control"}
          </h1>
          <p className="text-gray-400">
            {uiStrings?.ip_access_subtitle ||
              "Manage IP allowlists and blocklists for your organization"}
          </p>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors flex items-center gap-2"
        >
          <span>+</span> {uiStrings?.add_rule_btn || "Add Rule"}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-green-400">
            {allowedRules.length}
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_allowed_ranges || "Allowed IPs/Ranges"}
          </div>
        </div>
        <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl p-4 border border-gray-700">
          <div className="text-3xl font-bold text-red-400">
            {blockedRules.length}
          </div>
          <div className="text-gray-400 text-sm">
            {uiStrings?.stat_blocked_ranges || "Blocked IPs/Ranges"}
          </div>
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      ) : (
        <div className="space-y-6">
          {/* Blocked Rules */}
          <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl border border-gray-700 overflow-hidden">
            <div className="px-4 py-3 bg-red-900/20 border-b border-gray-700">
              <h2 className="font-semibold text-red-300">
                {uiStrings?.blocked_ips_title || "Blocked IPs"}
              </h2>
            </div>
            <table className="w-full">
              <thead>
                <tr className="bg-gray-700/50 text-left">
                  <th className="px-4 py-3 text-gray-300">
                    {uiStrings?.th_ip_cidr_label || "IP/CIDR"}
                  </th>
                  <th className="px-4 py-3 text-gray-300">
                    {uiStrings?.th_description_label || "Description"}
                  </th>
                  <th className="px-4 py-3 text-gray-300">
                    {uiStrings?.th_added_label || "Added"}
                  </th>
                  <th className="px-4 py-3 text-gray-300">
                    {uiStrings?.th_actions_label || "Actions"}
                  </th>
                </tr>
              </thead>
              <tbody>
                {blockedRules.map((rule) => (
                  <tr
                    key={rule.id}
                    className="border-t border-gray-700 hover:bg-gray-700/30"
                  >
                    <td className="px-4 py-3 font-mono text-red-300">
                      {rule.cidr}
                    </td>
                    <td className="px-4 py-3 text-gray-300">
                      {rule.description || "-"}
                    </td>
                    <td className="px-4 py-3 text-gray-400">
                      {formatDate(rule.created_at)}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleDeleteRule(rule.id)}
                        className="px-3 py-1 text-red-400 hover:text-red-300"
                      >
                        {uiStrings?.delete_btn || "Delete"}
                      </button>
                    </td>
                  </tr>
                ))}
                {blockedRules.length === 0 && (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-4 py-6 text-center text-gray-400"
                    >
                      {uiStrings?.no_blocked_ips || "No blocked IPs configured"}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Allowed Rules */}
          <div className="bg-gray-800/50 backdrop-blur-lg rounded-xl border border-gray-700 overflow-hidden">
            <div className="px-4 py-3 bg-green-900/20 border-b border-gray-700">
              <h2 className="font-semibold text-green-300">
                {uiStrings?.allowed_ips_title || "Allowed IPs (Allowlist)"}
              </h2>
            </div>
            <table className="w-full">
              <thead>
                <tr className="bg-gray-700/50 text-left">
                  <th className="px-4 py-3 text-gray-300">
                    {uiStrings?.th_ip_cidr_label || "IP/CIDR"}
                  </th>
                  <th className="px-4 py-3 text-gray-300">
                    {uiStrings?.th_description_label || "Description"}
                  </th>
                  <th className="px-4 py-3 text-gray-300">
                    {uiStrings?.th_added_label || "Added"}
                  </th>
                  <th className="px-4 py-3 text-gray-300">
                    {uiStrings?.th_actions_label || "Actions"}
                  </th>
                </tr>
              </thead>
              <tbody>
                {allowedRules.map((rule) => (
                  <tr
                    key={rule.id}
                    className="border-t border-gray-700 hover:bg-gray-700/30"
                  >
                    <td className="px-4 py-3 font-mono text-green-300">
                      {rule.cidr}
                    </td>
                    <td className="px-4 py-3 text-gray-300">
                      {rule.description || "-"}
                    </td>
                    <td className="px-4 py-3 text-gray-400">
                      {formatDate(rule.created_at)}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleDeleteRule(rule.id)}
                        className="px-3 py-1 text-red-400 hover:text-red-300"
                      >
                        {uiStrings?.delete_btn || "Delete"}
                      </button>
                    </td>
                  </tr>
                ))}
                {allowedRules.length === 0 && (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-4 py-6 text-center text-gray-400"
                    >
                      {uiStrings?.no_allowed_ips ||
                        "No allowed IPs configured (all IPs allowed by default)"}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Add Rule Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md border border-gray-700">
            <h2 className="text-xl font-bold mb-4">
              {uiStrings?.modal_title_add_ip || "Add IP Rule"}
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-gray-400 mb-1">
                  IP Address or CIDR Range
                </label>
                <input
                  type="text"
                  value={newRule.cidr}
                  onChange={(e) =>
                    setNewRule({ ...newRule, cidr: e.target.value })
                  }
                  placeholder={
                    uiStrings?.placeholder_ip_cidr ||
                    "e.g., 192.168.1.0/24 or 10.0.0.1"
                  }
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg focus:outline-none focus:border-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm text-gray-400 mb-1">
                  Action
                </label>
                <select
                  value={newRule.action}
                  onChange={(e) =>
                    setNewRule({
                      ...newRule,
                      action: e.target.value as "ALLOW" | "BLOCK",
                    })
                  }
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg focus:outline-none focus:border-blue-500"
                >
                  <option value="BLOCK">
                    {uiStrings?.opt_block || "Block"}
                  </option>
                  <option value="ALLOW">
                    {uiStrings?.opt_allow || "Allow"}
                  </option>
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-400 mb-1">
                  Description
                </label>
                <input
                  type="text"
                  value={newRule.description}
                  onChange={(e) =>
                    setNewRule({ ...newRule, description: e.target.value })
                  }
                  placeholder={
                    uiStrings?.placeholder_description || "Optional description"
                  }
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg focus:outline-none focus:border-blue-500"
                />
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowAddModal(false)}
                className="px-4 py-2 text-gray-400 hover:text-white"
              >
                {uiStrings?.cancel_btn || "Cancel"}
              </button>
              <button
                onClick={handleAddRule}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg"
              >
                {uiStrings?.add_rule_btn || "Add Rule"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
