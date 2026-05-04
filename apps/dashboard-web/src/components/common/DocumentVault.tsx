import React, { useState, useEffect, useRef } from 'react';
import { 
  FileText, 
  Upload, 
  Trash2, 
  CheckCircle, 
  Clock, 
  AlertTriangle, 
  ExternalLink,
  Plus,
  X
} from 'lucide-react';
import apiClient from '../../services/apiClient';

interface Document {
  id: string;
  document_type: string;
  file_name: string;
  file_path: string;
  file_size: number;
  mime_type: string;
  expiry_date?: string;
  status: string;
  created_at: string;
}

interface DocumentVaultProps {
  entityType: 'VEHICLE' | 'DRIVER';
  entityId: string;
}

const DOCUMENT_TYPES = [
  { id: 'INSURANCE', label: 'Insurance Policy' },
  { id: 'PERMIT', label: 'Route Permit' },
  { id: 'RC_BOOK', label: 'Registration Certificate (RC)' },
  { id: 'LICENSE', label: 'Driving License' },
  { id: 'AADHAAR', label: 'Aadhaar Card' },
  { id: 'PAN', label: 'PAN Card' },
  { id: 'FITNESS', label: 'Fitness Certificate' },
  { id: 'POLLUTION', label: 'Pollution (PUC)' },
  { id: 'MEDICAL', label: 'Medical Certificate' },
  { id: 'POLICE_VERIFICATION', label: 'Police Verification' },
];

export const DocumentVault: React.FC<DocumentVaultProps> = ({ entityType, entityId }) => {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [showUploadModal, setShowUploadModal] = useState(false);
  
  // Upload Form State
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [docType, setDocType] = useState('INSURANCE');
  const [expiryDate, setExpiryDate] = useState('');
  
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchDocuments = async () => {
    try {
      const response = await apiClient.get(`/admin/documents?entity_type=${entityType}&entity_id=${entityId}`);
      setDocuments(response.data.data || []);
    } catch (error) {
      console.error('Failed to fetch documents:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (entityId) fetchDocuments();
  }, [entityId, entityType]);

  const handleFileUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedFile) return;

    setUploading(true);
    const formData = new FormData();
    formData.append('file', selectedFile);
    formData.append('entity_type', entityType);
    formData.append('entity_id', entityId);
    formData.append('document_type', docType);
    if (expiryDate) formData.append('expiry_date', new Date(expiryDate).toISOString());

    try {
      await apiClient.post('/admin/documents/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      });
      setShowUploadModal(false);
      setSelectedFile(null);
      setExpiryDate('');
      fetchDocuments();
    } catch (error) {
      console.error('Upload failed:', error);
      alert('Upload failed. Please try again.');
    } finally {
      setUploading(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'VERIFIED': return <CheckCircle className="w-4 h-4 text-emerald-500" />;
      case 'REJECTED': return <AlertTriangle className="w-4 h-4 text-rose-500" />;
      default: return <Clock className="w-4 h-4 text-amber-500" />;
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (loading) return <div className="p-4 text-slate-400">Loading document vault...</div>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-slate-100 flex items-center gap-2">
          <FileText className="w-5 h-5 text-indigo-400" />
          Document Vault
        </h3>
        <button
          onClick={() => setShowUploadModal(true)}
          className="flex items-center gap-2 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm transition-colors"
        >
          <Plus className="w-4 h-4" />
          Add Document
        </button>
      </div>

      {documents.length === 0 ? (
        <div className="border-2 border-dashed border-slate-800 rounded-xl p-8 text-center">
          <Upload className="w-10 h-10 text-slate-700 mx-auto mb-3" />
          <p className="text-slate-500 text-sm">No documents uploaded yet.</p>
          <button 
            onClick={() => setShowUploadModal(true)}
            className="mt-4 text-indigo-400 hover:text-indigo-300 text-sm font-medium"
          >
            Upload your first document
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {documents.map((doc) => (
            <div key={doc.id} className="bg-slate-900/50 border border-slate-800 rounded-xl p-4 flex items-start gap-4 group">
              <div className="w-12 h-12 bg-indigo-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                <FileText className="w-6 h-6 text-indigo-400" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between gap-2">
                  <h4 className="text-sm font-medium text-slate-200 truncate">{doc.file_name}</h4>
                  {getStatusIcon(doc.status)}
                </div>
                <p className="text-xs text-slate-500 mt-1">{doc.document_type} • {formatFileSize(doc.file_size)}</p>
                {doc.expiry_date && (
                  <p className={`text-xs mt-2 flex items-center gap-1 ${
                    new Date(doc.expiry_date) < new Date() ? 'text-rose-400' : 'text-slate-400'
                  }`}>
                    <Clock className="w-3 h-3" />
                    Expires: {new Date(doc.expiry_date).toLocaleDateString()}
                  </p>
                )}
              </div>
              <div className="flex flex-col gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <a 
                  href={`${apiClient.defaults.baseURL}/uploads/${doc.file_path}`} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="p-1.5 hover:bg-slate-800 rounded-lg text-slate-400 hover:text-white transition-colors"
                >
                  <ExternalLink className="w-4 h-4" />
                </a>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in duration-200">
            <div className="p-6 border-b border-slate-800 flex items-center justify-between">
              <h3 className="text-xl font-bold text-white">Upload Document</h3>
              <button onClick={() => setShowUploadModal(false)} className="text-slate-400 hover:text-white">
                <X className="w-6 h-6" />
              </button>
            </div>
            
            <form onSubmit={handleFileUpload} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1.5">Document Type</label>
                <select 
                  value={docType}
                  onChange={(e) => setDocType(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-2.5 text-slate-200 focus:ring-2 focus:ring-indigo-500 outline-none transition-all"
                >
                  {DOCUMENT_TYPES.map(t => (
                    <option key={t.id} value={t.id}>{t.label}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1.5">Expiry Date (Optional)</label>
                <input 
                  type="date"
                  value={expiryDate}
                  onChange={(e) => setExpiryDate(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-2.5 text-slate-200 focus:ring-2 focus:ring-indigo-500 outline-none transition-all"
                />
              </div>

              <div 
                onClick={() => fileInputRef.current?.click()}
                className={`border-2 border-dashed rounded-2xl p-8 text-center cursor-pointer transition-all ${
                  selectedFile ? 'border-indigo-500 bg-indigo-500/5' : 'border-slate-800 hover:border-slate-700 hover:bg-slate-900'
                }`}
              >
                <input 
                  type="file"
                  ref={fileInputRef}
                  onChange={(e) => setSelectedFile(e.target.files?.[0] || null)}
                  className="hidden"
                />
                {selectedFile ? (
                  <div className="space-y-2">
                    <FileText className="w-10 h-10 text-indigo-400 mx-auto" />
                    <p className="text-sm text-slate-200 font-medium">{selectedFile.name}</p>
                    <p className="text-xs text-slate-500">{formatFileSize(selectedFile.size)}</p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    <div className="w-12 h-12 bg-slate-800 rounded-full flex items-center justify-center mx-auto mb-2 text-slate-400">
                      <Upload className="w-6 h-6" />
                    </div>
                    <p className="text-sm text-slate-300 font-medium">Click to select or drag and drop</p>
                    <p className="text-xs text-slate-500">PDF, PNG, JPG (max 10MB)</p>
                  </div>
                )}
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowUploadModal(false)}
                  className="flex-1 px-4 py-2.5 bg-slate-800 hover:bg-slate-700 text-white rounded-xl font-medium transition-all"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={!selectedFile || uploading}
                  className="flex-1 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-800 disabled:text-slate-600 text-white rounded-xl font-bold transition-all flex items-center justify-center gap-2"
                >
                  {uploading ? (
                    <>
                      <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                      Uploading...
                    </>
                  ) : (
                    <>
                      <Upload className="w-4 h-4" />
                      Upload
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
