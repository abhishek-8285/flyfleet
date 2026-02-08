import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../../../services/api';

const NewUpload: React.FC = () => {
    const navigate = useNavigate();
    const [file, setFile] = useState<File | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            setFile(e.target.files[0]);
            setError(null);
        }
    };

    const handleUpload = async () => {
        if (!file) {
            setError("Please select a file.");
            return;
        }

        setLoading(true);
        setError(null);

        try {
            await api.createBulkJob(file);
            navigate('/admin/bulk-uploads');
        } catch (err: any) {
            console.error("Upload failed", err);
            setError(err.response?.data?.error || "Failed to upload file. Please try again.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="p-6 max-w-2xl mx-auto">
             <div className="mb-6">
                <button onClick={() => navigate('/admin/bulk-uploads')} className="text-slate-400 hover:text-white flex items-center gap-1 mb-2">
                    <span className="material-symbols-outlined text-sm">arrow_back</span>
                    Back to Jobs
                </button>
                <h1 className="text-2xl font-bold text-white">New Bulk Upload</h1>
                <p className="text-slate-400 mt-1">Upload a CSV or Excel file to import vehicles.</p>
            </div>

            <div className="bg-[#1c222b] rounded-xl border border-white/5 p-8">
                <div className="border-2 border-dashed border-white/10 rounded-xl p-10 flex flex-col items-center justify-center text-center hover:border-emerald-500/30 transition-colors bg-[#161b22]/50">
                    <span className="material-symbols-outlined text-5xl text-emerald-500 mb-4">cloud_upload</span>
                    
                    {file ? (
                        <div className="mb-6">
                            <p className="text-white font-medium text-lg">{file.name}</p>
                            <p className="text-slate-500 text-sm">{(file.size / 1024).toFixed(2)} KB</p>
                        </div>
                    ) : (
                        <div className="mb-6">
                            <p className="text-white font-medium text-lg mb-1">Drag and drop your file here</p>
                            <p className="text-slate-500 text-sm">or click to browse from your computer</p>
                        </div>
                    )}

                    <input 
                        type="file" 
                        accept=".csv,.xlsx" 
                        onChange={handleFileChange}
                        className="hidden" 
                        id="file-upload"
                    />
                    
                    {!file && (
                        <label 
                            htmlFor="file-upload"
                            className="bg-white/5 hover:bg-white/10 text-white px-6 py-2.5 rounded-lg transition-colors cursor-pointer border border-white/10"
                        >
                            Browse Files
                        </label>
                    )}
                    
                    {file && (
                        <button 
                            onClick={() => setFile(null)}
                            className="text-red-400 hover:text-red-300 text-sm underline"
                        >
                            Remove File
                        </button>
                    )}
                </div>

                <div className="mt-6">
                    <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4 mb-6">
                        <div className="flex items-start gap-3">
                            <span className="material-symbols-outlined text-blue-400">info</span>
                            <div>
                                <h4 className="text-blue-400 font-medium text-sm mb-1">File Requirements</h4>
                                <ul className="text-slate-400 text-sm list-disc pl-4 space-y-1">
                                    <li>Supported formats: CSV, XLSX</li>
                                    <li>Maximum file size: 50MB</li>
                                    <li>Required columns: vin, license_plate</li>
                                </ul>
                            </div>
                        </div>
                    </div>

                    {error && (
                         <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 mb-6 text-red-500 flex items-center gap-2">
                             <span className="material-symbols-outlined">error</span>
                             {error}
                         </div>
                    )}

                    <button 
                        onClick={handleUpload}
                        disabled={!file || loading}
                        className={`w-full py-3 rounded-lg font-medium transition-colors flex items-center justify-center gap-2
                            ${!file || loading 
                                ? 'bg-slate-700 text-slate-400 cursor-not-allowed' 
                                : 'bg-emerald-500 hover:bg-emerald-600 text-white'}`}
                    >
                        {loading && <span className="material-symbols-outlined animate-spin text-sm">progress_activity</span>}
                        {loading ? 'Uploading...' : 'Start Upload'}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default NewUpload;
