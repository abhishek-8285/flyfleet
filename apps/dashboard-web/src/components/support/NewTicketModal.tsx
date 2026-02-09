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
    const token = localStorage.getItem('token');
    try {
      const res = await fetch('http://localhost:8081/api/support/incidents/create', {
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
        const newIncident = await res.json();
        onCreated(newIncident);
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
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div className="bg-[#16202a] w-full max-w-lg rounded-2xl border border-[#233648] shadow-2xl overflow-hidden flex flex-col animate-in fade-in zoom-in duration-200">
        <header className="px-6 py-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
          <h3 className="text-white font-bold text-lg">Create New Ticket</h3>
          <button onClick={onClose} className="text-[#92adc9] hover:text-white transition-colors">
            <span className="material-symbols-outlined">close</span>
          </button>
        </header>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div className="space-y-1">
            <label className="text-[10px] uppercase font-bold text-[#92adc9] tracking-wider">Subject</label>
            <input 
              required
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              className="w-full bg-[#0b1219] border border-[#233648] rounded-lg px-4 py-2.5 text-sm text-white focus:outline-none focus:border-primary transition-colors"
              placeholder="e.g. Telemetry lag in sector 4"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-[10px] uppercase font-bold text-[#92adc9] tracking-wider">Priority</label>
              <select 
                value={priority}
                onChange={(e) => setPriority(e.target.value)}
                className="w-full bg-[#0b1219] border border-[#233648] rounded-lg px-4 py-2.5 text-sm text-white focus:outline-none focus:border-primary transition-colors"
              >
                <option value="CRITICAL">Critical</option>
                <option value="HIGH">High</option>
                <option value="MEDIUM">Medium</option>
                <option value="LOW">Low</option>
              </select>
            </div>
            <div className="space-y-1">
              <label className="text-[10px] uppercase font-bold text-[#92adc9] tracking-wider">Organization</label>
              <input 
                required
                value={organization}
                onChange={(e) => setOrganization(e.target.value)}
                className="w-full bg-[#0b1219] border border-[#233648] rounded-lg px-4 py-2.5 text-sm text-white focus:outline-none focus:border-primary transition-colors"
                placeholder="Client name"
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-[10px] uppercase font-bold text-[#92adc9] tracking-wider">Description</label>
            <textarea 
              required
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full bg-[#0b1219] border border-[#233648] rounded-lg px-4 py-2.5 text-sm text-white focus:outline-none focus:border-primary transition-colors min-h-[120px]"
              placeholder="Provide full context for the support team..."
            ></textarea>
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button 
              type="button"
              onClick={onClose}
              className="px-6 py-2 border border-[#233648] text-[#92adc9] text-xs font-bold rounded-lg hover:bg-[#1c2a38] hover:text-white transition-colors"
            >
              Cancel
            </button>
            <button 
              disabled={submitting}
              className="px-6 py-2 bg-[#137fec] text-white text-xs font-bold rounded-lg hover:bg-[#0f6bd0] transition-colors shadow-lg shadow-blue-500/20 disabled:opacity-50"
            >
              {submitting ? 'Creating...' : 'Create Ticket'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default NewTicketModal;
