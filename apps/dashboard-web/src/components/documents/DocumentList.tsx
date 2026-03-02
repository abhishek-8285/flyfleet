import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { Document } from "../../pages/admin/FleetCommand/types";
import apiClient from "../../services/api";

interface DocumentListProps {
  vehicleId: string;
  onAdd: () => void;
}

const DocumentList: React.FC<DocumentListProps> = ({ vehicleId, onAdd }) => {
  const { uiStrings } = useUIStrings("documents");
  const { uiStrings: commonStrings } = useUIStrings("common");

  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDocuments();
  }, [vehicleId]);

  const fetchDocuments = async () => {
    try {
      const data = await apiClient.getDocuments(vehicleId);
      setDocuments(data || []);
    } catch (error) {
      console.error("Failed to fetch documents", error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Are you sure you want to delete this document?"))
      return;
    try {
      await apiClient.deleteDocument(id);
      fetchDocuments();
    } catch (error) {
      console.error("Failed to delete document", error);
    }
  };

  if (loading)
    return <div className="text-white text-sm">Loading documents...</div>;

  return (
    <div className="bg-[#1c222b] rounded-xl border border-white/5 overflow-hidden">
      <div className="p-4 border-b border-white/5 flex justify-between items-center">
        <h3 className="font-medium text-white">
          {uiStrings?.title_list || "Vehicle Documents"}
        </h3>
        <button
          onClick={onAdd}
          className="bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500/20 px-3 py-1.5 rounded-lg text-sm transition-colors flex items-center gap-1"
        >
          <span className="material-symbols-outlined text-sm">add</span>
          {uiStrings?.btn_add_document || "Add Document"}
        </button>
      </div>

      {documents.length === 0 ? (
        <div className="p-8 text-center text-slate-500 text-sm">
          {uiStrings?.empty_state || "No documents uploaded yet."}
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm text-left">
            <thead className="text-xs text-slate-400 uppercase bg-[#161b22]">
              <tr>
                <th className="px-6 py-3">{uiStrings?.col_type || "Type"}</th>
                <th className="px-6 py-3">
                  {uiStrings?.col_number || "Doc Number"}
                </th>
                <th className="px-6 py-3">
                  {uiStrings?.col_issue || "Issue Date"}
                </th>
                <th className="px-6 py-3">
                  {uiStrings?.col_expiry || "Expiry"}
                </th>
                <th className="px-6 py-3">
                  {uiStrings?.col_status || "Status"}
                </th>
                <th className="px-6 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5 text-slate-300">
              {documents.map((doc) => (
                <tr key={doc.id} className="hover:bg-white/5 transition-colors">
                  <td className="px-6 py-4 font-medium text-white capitalize">
                    {doc.type}
                  </td>
                  <td className="px-6 py-4 font-mono">{doc.doc_number}</td>
                  <td className="px-6 py-4">
                    {new Date(doc.issue_date).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      {new Date(doc.expiry_date).toLocaleDateString()}
                      {new Date(doc.expiry_date) < new Date() && (
                        <span
                          className="text-red-500 material-symbols-outlined text-sm"
                          title={uiStrings?.status_expired_title || "Expired"}
                        >
                          error
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-2 py-0.5 rounded text-xs font-medium 
                                            ${
                                              doc.status === "valid"
                                                ? "bg-emerald-500/10 text-emerald-500"
                                                : doc.status === "expired"
                                                  ? "bg-red-500/10 text-red-500"
                                                  : "bg-amber-500/10 text-amber-500"
                                            }`}
                    >
                      {doc.status.toUpperCase()}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right space-x-2">
                    {doc.file_url && (
                      <a
                        href={doc.file_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-400 hover:text-blue-300"
                      >
                        <span className="material-symbols-outlined text-sm">
                          download
                        </span>
                      </a>
                    )}
                    <button
                      onClick={() => handleDelete(doc.id)}
                      className="text-red-400 hover:text-red-300"
                    >
                      <span className="material-symbols-outlined text-sm">
                        delete
                      </span>
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default DocumentList;
