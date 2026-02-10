import React, { useState } from 'react';

interface NewTicketModalProps {
    isOpen: boolean;
    onClose: () => void;
    onCreated: (newIncident: any) => void;
}

const NewTicketModal: React.FC<NewTicketModalProps> = ({ isOpen, onClose, onCreated }) => {
    const [subject, setSubject] = useState('');
    const [description, setDescription] = useState('');
    const [priority, setPriority] = useState('MEDIUM');
    const [organization, setOrganization] = useState('Global Fleet Ops');
    const [submitting, setSubmitting] = useState(false);

    if (!isOpen) return null;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setSubmitting(true);
        const token = localStorage.getItem('access_token');
        try {
            const res = await fetch('/bff/support/incidents/create', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    subject,
                    description,
                    priority,
                    organization
                })
            });
            if (res.ok) {
                const data = await res.json();
                onCreated(data);
                onClose();
                // Reset form
                setSubject('');
                setDescription('');
                setPriority('MEDIUM');
            }
        } catch (err) {
            console.error('Failed to create ticket', err);
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
            <div className="absolute inset-0 bg-black/60 backdrop-blur-md animate-in fade-in duration-300" onClick={onClose}></div>
            <div className="bg-[#16202a] w-full max-w-lg rounded-2xl border border-[#233648] shadow-2xl overflow-hidden flex flex-col relative z-10 animate-in zoom-in-95 duration-200">
                <header className="px-6 py-5 border-b border-[#233648] flex justify-between items-center bg-[#111a22]/80 backdrop-blur-sm">
                    <div>
                        <h3 className="text-white font-bold text-lg">Initialize Support Protocol</h3>
                        <p className="text-[10px] text-[#92adc9] uppercase font-bold tracking-widest mt-0.5">New Incident Ticket</p>
                    </div>
                    <button onClick={onClose} className="w-8 h-8 rounded-full flex items-center justify-center text-[#92adc9] hover:text-white hover:bg-[#233648] transition-all">
                        <span className="material-symbols-outlined text-[20px]">close</span>
                    </button>
                </header>

                <form onSubmit={handleSubmit} className="p-6 space-y-6">
                    <div className="space-y-2">
                        <label className="text-[10px] uppercase font-bold text-[#92adc9] tracking-widest flex items-center gap-2">
                             Subject Line
                             <span className="w-1 h-1 rounded-full bg-primary/40"></span>
                        </label>
                        <input 
                            required
                            value={subject}
                            onChange={(e) => setSubject(e.target.value)}
                            className="w-full bg-[#0b1219] border border-[#233648] rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary/50 focus:border-primary transition-all placeholder:text-[#92adc9]/30"
                            placeholder="e.g. Critical: Telemetry sync failure on Node-X"
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-6">
                        <div className="space-y-2">
                            <label className="text-[10px] uppercase font-bold text-[#92adc9] tracking-widest flex items-center gap-2">
                                Priority Level
                                <span className={`w-1.5 h-1.5 rounded-full ${priority === 'CRITICAL' ? 'bg-rose-500 animate-pulse' : 'bg-primary'}`}></span>
                            </label>
                            <select 
                                value={priority}
                                onChange={(e) => setPriority(e.target.value)}
                                className="w-full bg-[#0b1219] border border-[#233648] rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary/50 focus:border-primary transition-all appearance-none cursor-pointer"
                            >
                                <option value="CRITICAL">P0 - Critical</option>
                                <option value="HIGH">P1 - High</option>
                                <option value="MEDIUM">P2 - Medium</option>
                                <option value="LOW">P3 - Low</option>
                            </select>
                        </div>
                        <div className="space-y-2">
                            <label className="text-[10px] uppercase font-bold text-[#92adc9] tracking-widest">Target Organization</label>
                            <input 
                                required
                                value={organization}
                                onChange={(e) => setOrganization(e.target.value)}
                                className="w-full bg-[#0b1219] border border-[#233648] rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary/50 focus:border-primary transition-all placeholder:text-[#92adc9]/30"
                                placeholder="..."
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label className="text-[10px] uppercase font-bold text-[#92adc9] tracking-widest flex items-center gap-2">
                            Incident Narrative
                            <span className="text-[8px] text-[#92adc9]/40 normal-case font-medium">(Markdown supported)</span>
                        </label>
                        <textarea 
                            required
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            className="w-full bg-[#0b1219] border border-[#233648] rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary/50 focus:border-primary transition-all min-h-[140px] resize-none placeholder:text-[#92adc9]/30"
                            placeholder="Detailed breakdown of the observed anomaly..."
                        ></textarea>
                    </div>

                    <div className="flex gap-4 pt-4">
                        <button 
                            type="button"
                            onClick={onClose}
                            className="flex-1 py-3 border border-[#233648] text-[#92adc9] text-xs font-bold rounded-xl hover:bg-[#1c2a38] hover:text-white transition-all uppercase tracking-widest"
                        >
                            Abort
                        </button>
                        <button 
                            disabled={submitting}
                            className="flex-[2] py-3 bg-primary hover:bg-primary-dark text-white text-xs font-bold rounded-xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center gap-2 uppercase tracking-widest"
                        >
                            {submitting ? (
                                <span className="flex items-center gap-2">
                                    <span className="w-3 h-3 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
                                    Processing...
                                </span>
                            ) : 'Authorize Incident Creation'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default NewTicketModal;
