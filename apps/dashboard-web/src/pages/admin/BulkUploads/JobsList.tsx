import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import { api } from "../../../services/api";

interface BulkJob {
  id: string;
  file_name: string;
  status: string;
  total_records: number;
  success_count: number;
  failed_count: number;
  created_at: string;
}

const JobsList: React.FC = () => {
  const [jobs, setJobs] = useState<BulkJob[]>([]);
  const [loading, setLoading] = useState(true);
  const { uiStrings: rawUiStrings } = useUIStrings("bulk_uploads");
  const uiStrings = rawUiStrings || {};

  useEffect(() => {
    fetchJobs();
    const interval = setInterval(fetchJobs, 5000); // Auto refresh every 5s for sorting out stuck jobs or updating status
    return () => clearInterval(interval);
  }, []);

  const fetchJobs = async () => {
    try {
      const data = await api.getBulkJobs({ limit: 50 });
      // API returns { jobs: [], total: ... }
      if (data && data.jobs) {
        setJobs(data.jobs);
      } else {
        setJobs([]);
      }
    } catch (error) {
      console.error("Failed to fetch jobs", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-white">
          {uiStrings?.title || "Bulk Uploads"}
        </h1>
        <Link
          to="/admin/bulk-uploads/new"
          className="bg-emerald-500 hover:bg-emerald-600 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
        >
          <span className="material-symbols-outlined">upload_file</span>
          {uiStrings?.btn_new_upload || "New Upload"}
        </Link>
      </div>

      <div className="bg-[#1c222b] rounded-xl border border-white/5 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm text-left">
            <thead className="text-xs text-slate-400 uppercase bg-[#161b22]">
              <tr>
                <th className="px-6 py-4">
                  {uiStrings?.th_file_name || "File Name"}
                </th>
                <th className="px-6 py-4">
                  {uiStrings?.th_status || "Status"}
                </th>
                <th className="px-6 py-4">
                  {uiStrings?.th_progress || "Progress"}
                </th>
                <th className="px-6 py-4">
                  {uiStrings?.th_uploaded_at || "Uploaded At"}
                </th>
                <th className="px-6 py-4 text-right">
                  {uiStrings?.th_actions || "Actions"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5 text-slate-300">
              {jobs.map((job) => (
                <tr key={job.id} className="hover:bg-white/5 transition-colors">
                  <td className="px-6 py-4 font-medium text-white">
                    <div className="flex items-center gap-3">
                      <span className="material-symbols-outlined text-slate-500">
                        description
                      </span>
                      {job.file_name}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-2 py-1 rounded-full text-xs font-medium 
                                            ${
                                              job.status === "completed"
                                                ? "bg-emerald-500/10 text-emerald-500"
                                                : job.status === "processing"
                                                  ? "bg-blue-500/10 text-blue-500"
                                                  : job.status === "failed"
                                                    ? "bg-red-500/10 text-red-500"
                                                    : "bg-amber-500/10 text-amber-500"
                                            }`}
                    >
                      {job.status.toUpperCase().replace("_", " ")}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex flex-col gap-1 w-32">
                      <div className="flex justify-between text-xs text-slate-400">
                        <span>
                          {job.success_count + job.failed_count} /{" "}
                          {job.total_records || "-"}
                        </span>
                      </div>
                      {(job.total_records > 0 ||
                        job.status === "completed") && (
                        <div className="w-full bg-white/10 rounded-full h-1.5 overflow-hidden flex">
                          <div
                            className="bg-emerald-500 h-full"
                            style={{
                              width: `${(job.success_count / (job.total_records || 1)) * 100}%`,
                            }}
                          />
                          <div
                            className="bg-red-500 h-full"
                            style={{
                              width: `${(job.failed_count / (job.total_records || 1)) * 100}%`,
                            }}
                          />
                        </div>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    {new Date(job.created_at).toLocaleString()}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <Link
                      to={`/admin/bulk-uploads/${job.id}`}
                      className="text-blue-400 hover:text-blue-300"
                    >
                      {uiStrings?.link_view_details || "View Details"}
                    </Link>
                  </td>
                </tr>
              ))}
              {jobs.length === 0 && !loading && (
                <tr>
                  <td
                    colSpan={5}
                    className="px-6 py-8 text-center text-slate-500"
                  >
                    {uiStrings?.msg_no_jobs || "No upload jobs found."}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default JobsList;
