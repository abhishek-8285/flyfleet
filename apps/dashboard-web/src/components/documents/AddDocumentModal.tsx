import React, { useState } from 'react';
import apiClient from '../../services/api';

interface AddDocumentModalProps {
    vehicleId: string;
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

const AddDocumentModal: React.FC<AddDocumentModalProps> = ({ vehicleId, isOpen, onClose, onSuccess }) => {
    const [formData, setFormData] = useState({
        type: 'insurance',
        doc_number: '',
        issue_date: '',
        expiry_date: '',
        file_url: '',
        status: 'valid',
        notes: ''
    });
    const [loading, setLoading] = useState(false);
    const [file, setFile] = useState<File | null>(null);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            setFile(e.target.files[0]);
            // Mock upload: in real app, upload here or on submit
            // For now, we'll set a fake URL or use a blob URL for preview
            setFormData({ ...formData, file_url: "https://example.com/mock-doc.pdf" });
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        try {
            // If we had real upload, we'd do it here.
            
            const payload = {
                vehicle_id: vehicleId,
                ...formData,
                metadata: { notes: formData.notes }
            };
            
            await apiClient.createDocument(vehicleId, payload);
            onSuccess();
            onClose();
        } catch (error) {
            console.error("Failed to add document", error);
            alert("Failed to add document");
        } finally {
            setLoading(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
            <div className="bg-[#1c222b] rounded-xl border border-white/10 w-full max-w-lg shadow-2xl">
                <div className="p-6 border-b border-white/5 flex justify-between items-center bg-[#161b22] rounded-t-xl">
                    <h2 className="text-xl font-bold text-white">Add Document</h2>
                    <button onClick={onClose} className="text-slate-400 hover:text-white">
                        <span className="material-symbols-outlined">close</span>
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-xs font-medium text-slate-400 mb-1">Document Type</label>
                            <select 
                                name="type" 
                                value={formData.type} 
                                onChange={handleChange}
                                className="w-full bg-[#0f111a] border border-white/10 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-emerald-500/50"
                            >
                                <option value="insurance">Insurance</option>
                                <option value="registration">Registration</option>
                                <option value="permit">Permit</option>
                                <option value="fitness">Fitness Certificate</option>
                                <option value="puc">PUC (Pollution)</option>
                                <option value="other">Other</option>
                            </select>
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-slate-400 mb-1">Document Number</label>
                            <input 
                                type="text" 
                                name="doc_number" 
                                value={formData.doc_number} 
                                onChange={handleChange}
                                required
                                className="w-full bg-[#0f111a] border border-white/10 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-emerald-500/50"
                            />
                        </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-xs font-medium text-slate-400 mb-1">Issue Date</label>
                            <input 
                                type="date" 
                                name="issue_date" 
                                value={formData.issue_date} 
                                onChange={handleChange}
                                required
                                className="w-full bg-[#0f111a] border border-white/10 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-emerald-500/50"
                            />
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-slate-400 mb-1">Expiry Date</label>
                            <input 
                                type="date" 
                                name="expiry_date" 
                                value={formData.expiry_date} 
                                onChange={handleChange}
                                required
                                className="w-full bg-[#0f111a] border border-white/10 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-emerald-500/50"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-xs font-medium text-slate-400 mb-1">Upload File</label>
                        <input 
                            type="file" 
                            onChange={handleFileChange}
                            className="w-full bg-[#0f111a] border border-white/10 rounded-lg px-3 py-2 text-white file:mr-4 file:py-1 file:px-3 file:rounded-full file:border-0 file:text-xs file:bg-emerald-500/10 file:text-emerald-500 hover:file:bg-emerald-500/20"
                        />
                        <p className="text-xs text-slate-500 mt-1">Mock upload only. Will save dummy URL.</p>
                    </div>

                    <div>
                        <label className="block text-xs font-medium text-slate-400 mb-1">Notes</label>
                        <textarea 
                            name="notes" 
                            value={formData.notes} 
                            onChange={handleChange}
                            rows={3}
                            className="w-full bg-[#0f111a] border border-white/10 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-emerald-500/50"
                        />
                    </div>

                    <div className="flex justify-end gap-3 pt-4">
                        <button 
                            type="button" 
                            onClick={onClose}
                            className="px-4 py-2 rounded-lg text-slate-300 hover:bg-white/5 transition-colors"
                        >
                            Cancel
                        </button>
                        <button 
                            type="submit" 
                            disabled={loading}
                            className="bg-emerald-500 hover:bg-emerald-600 text-white px-6 py-2 rounded-lg font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
                        >
                            {loading && <span className="material-symbols-outlined animate-spin text-sm">progress_activity</span>}
                            Save Document
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default AddDocumentModal;
