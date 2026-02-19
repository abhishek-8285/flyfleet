import React, { useEffect, useState } from 'react';
import apiClient from '../../../services/apiClient';

interface Template {
    id: string;
    template_key: string;
    channel: string;
    subject: string;
    body: string;
    variables: string[];
    updated_at: string;
}

const Templates: React.FC = () => {
    const [templates, setTemplates] = useState<Template[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null);
    const [editMode, setEditMode] = useState(false);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        fetchTemplates();
    }, []);

    const fetchTemplates = async () => {
        setError(null);
        try {
            const response = await apiClient.get('/settings/templates') as any;
            const items = Array.isArray(response)
                ? response
                : Array.isArray(response?.data)
                    ? response.data
                    : [];
            setTemplates(items);
            setSelectedTemplate((prev) => (prev ? prev : items[0] || null));
        } catch (error) {
            console.error('Failed to fetch templates', error);
            setError('Failed to load templates');
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        if (!selectedTemplate) return;
        setSaving(true);
        try {
            await apiClient.put(`/settings/templates/${selectedTemplate.id}`, selectedTemplate);
            setTemplates(templates.map(t => t.id === selectedTemplate.id ? selectedTemplate : t));
            setEditMode(false);
        } catch (error) {
            console.error('Failed to save template', error);
            alert('Failed to save template');
        } finally {
            setSaving(false);
        }
    };

    if (loading) return <div className="p-6 text-white">Loading...</div>;

    return (
        <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
             <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <h2 className="text-white text-lg font-bold">Communication Templates</h2>
                        <div className="flex items-center gap-2 text-[11px] text-[#92adc9] mt-0.5 uppercase tracking-wide font-medium">
                            <span>System Settings</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span className="text-white">Templates</span>
                        </div>
                    </div>
                </div>
            </header>
            <div className="flex flex-1 overflow-hidden">
                <div className="w-1/3 border-r border-[#233648] overflow-y-auto bg-[#16202a]">
                    <div className="p-4 border-b border-[#233648] sticky top-0 bg-[#16202a] z-10">
                        <input 
                            className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-xs text-white focus:outline-none focus:border-[#137fec] placeholder-gray-600" 
                            placeholder="Search templates..."
                        />
                    </div>
                    {error ? (
                        <div className="p-4 text-sm text-red-200">{error}</div>
                    ) : (
                        <div className="divide-y divide-[#233648]">
                            {templates.length === 0 ? (
                                <div className="p-4 text-sm text-[#92adc9]">No templates available.</div>
                            ) : templates.map(template => (
                                <div 
                                    key={template.id} 
                                    onClick={() => { setSelectedTemplate(template); setEditMode(false); }}
                                    className={`p-4 cursor-pointer hover:bg-[#111a22] transition-colors ${selectedTemplate?.id === template.id ? 'bg-[#111a22] border-l-2 border-[#137fec]' : 'border-l-2 border-transparent'}`}
                                >
                                    <div className="flex items-center justify-between mb-1">
                                        <span className="text-xs font-mono text-[#137fec]">{template.template_key}</span>
                                        <span className={`px-1.5 py-0.5 rounded text-[9px] uppercase font-bold ${template.channel === 'EMAIL' ? 'bg-blue-500/20 text-blue-300' : 'bg-green-500/20 text-green-300'}`}>
                                            {template.channel}
                                        </span>
                                    </div>
                                    <div className="text-sm text-white font-medium truncate">{template.subject || '(No Subject)'}</div>
                                    <div className="text-[10px] text-[#92adc9] mt-1">Updated {new Date(template.updated_at).toLocaleDateString()}</div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
                <div className="flex-1 overflow-y-auto bg-[#0b1219] p-6">
                    {selectedTemplate ? (
                        <div className="max-w-3xl mx-auto space-y-6">
                            <div className="flex items-center justify-between mb-2">
                                <h3 className="text-white font-bold text-lg">Template Editor</h3>
                                <div className="flex gap-2">
                                     {editMode ? (
                                        <>
                                            <button 
                                                onClick={() => setEditMode(false)}
                                                className="px-3 py-1.5 border border-[#233648] text-[#92adc9] hover:text-white text-xs font-semibold rounded"
                                            >
                                                Cancel
                                            </button>
                                            <button 
                                                onClick={handleSave}
                                                disabled={saving}
                                                className="px-3 py-1.5 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded flex items-center gap-1"
                                            >
                                                {saving ? 'Saving...' : 'Save Changes'}
                                            </button>
                                        </>
                                     ) : (
                                        <button 
                                            onClick={() => setEditMode(true)}
                                            className="px-3 py-1.5 bg-[#233648] hover:bg-[#2d4458] text-white text-xs font-semibold rounded flex items-center gap-1"
                                        >
                                            <span className="material-symbols-outlined text-[14px]">edit</span>
                                            Edit Template
                                        </button>
                                     )}
                                </div>
                            </div>

                            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-6 space-y-4">
                                <div>
                                    <label className="block text-xs font-medium text-[#92adc9] mb-1.5">Subject Line</label>
                                    <input 
                                        disabled={!editMode}
                                        className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] disabled:opacity-50 disabled:cursor-not-allowed" 
                                        value={selectedTemplate.subject}
                                        onChange={(e) => setSelectedTemplate({...selectedTemplate, subject: e.target.value})}
                                    />
                                </div>
                                <div>
                                    <label className="block text-xs font-medium text-[#92adc9] mb-1.5">Content Body (HTML/Text)</label>
                                    <textarea 
                                        disabled={!editMode}
                                        rows={15}
                                        className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-[#137fec] disabled:opacity-50 disabled:cursor-not-allowed" 
                                        value={selectedTemplate.body}
                                        onChange={(e) => setSelectedTemplate({...selectedTemplate, body: e.target.value})}
                                    />
                                </div>
                                <div>
                                    <label className="block text-xs font-medium text-[#92adc9] mb-2">Available Variables</label>
                                    <div className="flex flex-wrap gap-2">
                                        {selectedTemplate.variables?.map(v => (
                                            <span key={v} className="px-2 py-1 bg-[#111a22] border border-[#233648] rounded text-[10px] text-[#137fec] font-mono">
                                                {`{{${v}}}`}
                                            </span>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </div>
                    ) : (
                        <div className="h-full flex flex-col items-center justify-center text-[#92adc9]">
                            <span className="material-symbols-outlined text-4xl mb-2 opacity-50">description</span>
                            <p className="text-sm">Select a template from the list to view or edit</p>
                        </div>
                    )}
                </div>
            </div>
        </main>
    );
};

export default Templates;
