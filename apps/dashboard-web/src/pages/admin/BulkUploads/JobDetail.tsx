import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import { api } from "../../../services/api";

interface JobRecord {
  id: string;
  row_number: number;
  record_data: any;
  status: string;
  error_message?: string;
}

interface BulkJob {
  id: string;
  file_name: string;
  status: string;
  total_records: number;
  success_count: number;
  failed_count: number;
  created_at: string;
  updated_at: string;
}

const JobDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [job, setJob] = useState<BulkJob | null>(null);
  const [records, setRecords] = useState<JobRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"all" | "failed" | "success">(
    "all",
  );
  const { uiStrings: rawUiStrings } = useUIStrings("bulk_uploads");
  const uiStrings = rawUiStrings || {};

  useEffect(() => {
    if (id) {
      fetchJobDetails();
    }
  }, [id]);

  const fetchJobDetails = async () => {
    setLoading(true);
    try {
      const jobData = await api.getBulkJob(id!);
      setJob(jobData);

      // For now fetch all records, pagination needed for large jobs
      const recordsData = await api.getBulkJobRecords(id!);
      setRecords(recordsData || []);
    } catch (error) {
      console.error("Failed to fetch job details", error);
    } finally {
      setLoading(false);
    }
  };

  const filteredRecords = records.filter((r) => {
    if (activeTab === "all") return true;
    return r.status === activeTab;
  });

  if (loading)
    return (
      <div className="text-white p-6">
        {uiStrings?.msg_loading_details || "Loading job details..."}
      </div>
    );
  if (!job)
    return (
      <div className="text-white p-6">
        {uiStrings?.msg_job_not_found || "Job not found"}
      </div>
    );

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <Link
          to="/admin/bulk-uploads"
          className="text-slate-400 hover:text-white flex items-center gap-1 mb-2"
        >
          <span className="material-symbols-outlined text-sm">arrow_back</span>
          {uiStrings?.link_back_to_jobs || "Back to Jobs"}
        </Link>
        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-2xl font-bold text-white flex items-center gap-3">
              {job.file_name}
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
            </h1>
            <p className="text-slate-400 text-sm mt-1">
              {uiStrings?.msg_uploaded_on?.replace(
                "{date}",
                new Date(job.created_at).toLocaleString(),
              ) || `Uploaded on ${new Date(job.created_at).toLocaleString()}`}
            </p>
          </div>
          <button
            onClick={fetchJobDetails}
            className="bg-white/5 hover:bg-white/10 text-white p-2 rounded-lg"
          >
            <span className="material-symbols-outlined">refresh</span>
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <div className="bg-[#1c222b] rounded-xl border border-white/5 p-4">
          <div className="text-slate-400 text-sm mb-1">
            {uiStrings?.stat_total_records || "Total Records"}
          </div>
          <div className="text-2xl font-bold text-white">
            {job.total_records || "-"}
          </div>
        </div>
        <div className="bg-[#1c222b] rounded-xl border border-white/5 p-4">
          <div className="text-emerald-500 text-sm mb-1">
            {uiStrings?.stat_successful || "Successful"}
          </div>
          <div className="text-2xl font-bold text-white">
            {job.success_count}
          </div>
        </div>
        <div className="bg-[#1c222b] rounded-xl border border-white/5 p-4">
          <div className="text-red-500 text-sm mb-1">
            {uiStrings?.stat_failed || "Failed"}
          </div>
          <div className="text-2xl font-bold text-white">
            {job.failed_count}
          </div>
        </div>
        <div className="bg-[#1c222b] rounded-xl border border-white/5 p-4">
          <div className="text-blue-500 text-sm mb-1">
            {uiStrings?.stat_success_rate || "Success Rate"}
          </div>
          <div className="text-2xl font-bold text-white">
            {job.total_records > 0
              ? ((job.success_count / job.total_records) * 100).toFixed(1)
              : 0}
            %
          </div>
        </div>
      </div>

      {/* Records List */}
      <div className="bg-[#1c222b] rounded-xl border border-white/5 overflow-hidden">
        <div className="border-b border-white/5 flex">
          <button
            onClick={() => setActiveTab("all")}
            className={`px-6 py-3 text-sm font-medium transition-colors border-b-2 
                            ${activeTab === "all" ? "border-blue-500 text-white" : "border-transparent text-slate-400 hover:text-white"}`}
          >
            {uiStrings?.tab_all_records || "All Records"}
          </button>
          <button
            onClick={() => setActiveTab("failed")}
            className={`px-6 py-3 text-sm font-medium transition-colors border-b-2 
                            ${activeTab === "failed" ? "border-red-500 text-white" : "border-transparent text-slate-400 hover:text-white"}`}
          >
            {uiStrings?.tab_failed_records?.replace(
              "{count}",
              job.failed_count.toString(),
            ) || `Failed (${job.failed_count})`}
          </button>
          <button
            onClick={() => setActiveTab("success")}
            className={`px-6 py-3 text-sm font-medium transition-colors border-b-2 
                            ${activeTab === "success" ? "border-emerald-500 text-white" : "border-transparent text-slate-400 hover:text-white"}`}
          >
            {uiStrings?.tab_success_records?.replace(
              "{count}",
              job.success_count.toString(),
            ) || `Successful (${job.success_count})`}
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-sm text-left">
            <thead className="text-xs text-slate-400 uppercase bg-[#161b22]">
              <tr>
                <th className="px-6 py-3">
                  {uiStrings?.th_row_num || "Row #"}
                </th>
                <th className="px-6 py-3">
                  {uiStrings?.th_status || "Status"}
                </th>
                <th className="px-6 py-3">
                  {uiStrings?.th_data_preview || "Data Preview"}
                </th>
                <th className="px-6 py-3">
                  {uiStrings?.th_error_info || "Error / Info"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5 text-slate-300">
              {filteredRecords.map((record) => (
                <tr
                  key={record.id}
                  className="hover:bg-white/5 transition-colors"
                >
                  <td className="px-6 py-3 font-mono text-xs">
                    {record.row_number}
                  </td>
                  <td className="px-6 py-3">
                    <span
                      className={`px-2 py-0.5 rounded text-xs font-medium 
                                            ${record.status === "success" ? "bg-emerald-500/10 text-emerald-500" : "bg-red-500/10 text-red-500"}`}
                    >
                      {record.status.toUpperCase()}
                    </span>
                  </td>
                  <td className="px-6 py-3">
                    <div className="max-w-xs truncate text-xs text-slate-400 font-mono">
                      {JSON.stringify(record.record_data)}
                    </div>
                  </td>
                  <td className="px-6 py-3 text-red-400 text-xs">
                    {record.error_message || "-"}
                  </td>
                </tr>
              ))}
              {filteredRecords.length === 0 && (
                <tr>
                  <td
                    colSpan={4}
                    className="px-6 py-8 text-center text-slate-500"
                  >
                    {uiStrings?.msg_no_records_filter ||
                      "No records found for this filter."}
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

export default JobDetail;
