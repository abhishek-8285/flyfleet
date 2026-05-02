import React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";

export const TenantSettings: React.FC = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("admin_orgs");

  const [sections, setSections] = React.useState<any[]>([]);

  React.useEffect(() => {
    const fetchSections = async () => {
      try {
        // Fetch dynamic sections for the current organization
        const res = await fetch(
          `${import.meta.env.VITE_API_URL || "/api/v1"}/tenants/${id}/settings/sections`,
        );
        if (res.ok) {
          const data = await res.json();
          setSections(data.data || []);
        }
      } catch (err) {
        console.error("Failed to fetch abstract org settings", err);
      }
    };
    fetchSections();
  }, [id]);

  return (
    <div className="p-8 max-w-4xl">
      <div className="mb-10">
        <h1 className="text-white text-3xl font-bold mb-2">
          {uiStrings?.settings_title || "Tenant Settings"}
        </h1>
        <p className="text-text-muted">
          {uiStrings?.settings_subtitle ||
            "Manage security protocols, workspace preferences, and compliance standards."}
        </p>
      </div>

      <div className="space-y-12">
        {sections.map((section: any, idx: number) => (
          <div
            key={idx}
            className="animate-in fade-in slide-in-from-bottom-4 duration-500"
            style={{ animationDelay: `${idx * 100}ms` }}
          >
            <h2 className="text-lg font-bold text-slate-800 mb-4 px-1">
              {section.title}
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {(section.items || []).map((item: any, i: number) => (
                <div
                  key={i}
                  onClick={item.action}
                  className="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm hover:shadow-md hover:border-blue-200 transition-all cursor-pointer group"
                >
                  <div className="flex items-start space-x-4">
                    <div className={`p-3 rounded-xl ${item.bgColor}`}>
                      <item.icon className={`h-6 w-6 ${item.color}`} />
                    </div>
                    <div>
                      <h3 className="font-semibold text-slate-900 group-hover:text-blue-600 transition-colors">
                        {item.name}
                      </h3>
                      <p className="text-sm text-slate-500 mt-1 leading-relaxed">
                        {item.description}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Danger Zone */}
      <div className="mt-12 pt-12 border-t border-border-color">
        <h2 className="text-red-500 font-bold mb-4">
          {uiStrings?.danger_zone || "Danger Zone"}
        </h2>
        <div className="bg-red-500/5 border border-red-500/10 rounded-2xl p-6 flex items-center justify-between">
          <div>
            <h3 className="text-white font-bold mb-1">
              {uiStrings?.archive_title || "Archive Tenant"}
            </h3>
            <p className="text-slate-500 text-sm max-w-md">
              {uiStrings?.archive_desc ||
                "This will restrict access for all users immediately. Tenant data will be retained for 30 days."}
            </p>
          </div>
          <button className="bg-red-500 text-white px-6 py-2 rounded-xl font-bold hover:bg-red-600 transition-colors">
            {uiStrings?.btn_archive || "Archive Workspace"}
          </button>
        </div>
      </div>
    </div>
  );
};
