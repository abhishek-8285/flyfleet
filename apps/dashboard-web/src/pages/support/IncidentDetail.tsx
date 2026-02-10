import React, { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

const IncidentDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [incident, setIncident] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [isEditingOwner, setIsEditingOwner] = useState(false);
  const [newOwner, setNewOwner] = useState('');

  const fetchIncident = async () => {
    const token = localStorage.getItem('token');
    try {
      const res = await fetch(`/bff/support/incidents/detail?id=${id}`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.ok) {
        const data = await res.json();
        setIncident(data);
        setNewOwner(data.assigned_to);
      }
    } catch (err) {
      console.error('Failed to fetch incident', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchIncident();
  }, [id]);

  const handlePostComment = async () => {
    if (!comment.trim()) return;
    setSubmitting(true);
    const token = localStorage.getItem('token');
    try {
      const res = await fetch(`/api/support/incidents/comment?id=${id}`, {
        method: 'POST',
        headers: { 
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ message: comment, user: 'Abhishek (Support)' })
      });
      if (res.ok) {
        setComment('');
        await fetchIncident();
      }
    } catch (err) {
      console.error('Failed to post comment', err);
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdate = async (updates: any) => {
    const token = localStorage.getItem('token');
    try {
      const res = await fetch(`/api/support/incidents/update?id=${id}`, {
        method: 'PATCH',
        headers: { 
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(updates)
      });
      if (res.ok) {
        await fetchIncident();
      }
    } catch (err) {
      console.error('Failed to update incident', err);
    }
  };

  const handleResolve = () => {
    handleUpdate({ status: 'RESOLVED' });
  };

  const handleOwnerChange = () => {
    handleUpdate({ assigned_to: newOwner });
    setIsEditingOwner(false);
  };

  if (loading) return <div className="p-8 text-white">Loading incident details...</div>;
  if (!incident) return <div className="p-8 text-white">Incident not found.</div>;

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <Link to="/support/incidents" className="text-[#92adc9] hover:text-white transition-colors flex items-center gap-1">
            <span className="material-symbols-outlined text-[20px]">arrow_back</span>
            <span className="text-xs font-semibold">Back to List</span>
          </Link>
          <div className="h-6 w-px bg-[#233648]"></div>
          <div>
            <div className="flex items-center gap-2 text-[10px] text-[#92adc9] uppercase font-bold tracking-wider mb-0.5">
              <span>Support</span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span>Incidents</span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="text-white">{incident.id}</span>
            </div>
            <h2 className="text-white text-lg font-bold truncate max-w-md">{incident.subject}</h2>
          </div>
        </div>
        <div className="flex items-center gap-3">
            {incident.status !== 'RESOLVED' && (
              <button 
                onClick={handleResolve}
                className="px-4 py-2 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded transition-colors"
              >
                  Resolve Ticket
              </button>
            )}
            {incident.status === 'RESOLVED' && (
              <button 
                onClick={() => handleUpdate({ status: 'OPEN' })}
                className="px-4 py-2 bg-[#233648] hover:bg-[#2c4054] text-white text-xs font-semibold rounded transition-colors"
              >
                  Reopen Ticket
              </button>
            )}
        </div>
      </header>

      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="grid grid-cols-12 gap-6">
          {/* Main Info */}
          <div className="col-span-12 lg:col-span-8 space-y-6">
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-6 shadow-lg shadow-black/20">
                <h3 className="text-[#92adc9] text-[10px] uppercase font-bold tracking-widest mb-4">Description</h3>
                <p className="text-white text-sm leading-relaxed">
                    {incident.description}
                </p>
            </div>

            <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden shadow-lg shadow-black/20">
                <div className="p-4 border-b border-[#233648] bg-[#111a22]">
                    <h3 className="text-white font-semibold text-sm">Activity Timeline</h3>
                </div>
                <div className="p-6 space-y-6">
                    {incident.updates?.map((update: any, idx: number) => (
                        <div key={idx} className="flex gap-4 relative">
                            {idx !== incident.updates.length - 1 && (
                                <div className="absolute left-[15px] top-8 bottom-[-24px] w-px bg-[#233648]"></div>
                            )}
                            <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 z-10 border border-[#233648] ${
                              update.user.includes('Abhishek') ? 'bg-[#137fec]/20' : 'bg-[#233648]'
                            }`}>
                                <span className={`material-symbols-outlined text-[16px] ${
                                  update.user.includes('Abhishek') ? 'text-primary' : 'text-[#92adc9]'
                                }`}>
                                  {update.user.includes('Abhishek') ? 'person' : 'history'}
                                </span>
                            </div>
                            <div className="flex-1 pb-4">
                                <div className="flex justify-between items-center mb-1">
                                    <span className={`text-xs font-bold ${update.user.includes('Abhishek') ? 'text-[#137fec]' : 'text-white'}`}>
                                      {update.user}
                                    </span>
                                    <span className="text-[10px] text-[#92adc9]">{update.time}</span>
                                </div>
                                <p className="text-xs text-[#92adc9]">{update.message}</p>
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-6 shadow-lg shadow-black/20">
                <h3 className="text-white font-semibold text-sm mb-4">Add Update</h3>
                <textarea 
                    value={comment}
                    onChange={(e) => setComment(e.target.value)}
                    className="w-full bg-[#111a22] border border-[#233648] rounded-lg p-4 text-sm text-white focus:outline-none focus:border-primary transition-colors min-h-[120px]"
                    placeholder="Type your update here..."
                ></textarea>
                <div className="flex justify-end mt-4">
                    <button 
                      onClick={handlePostComment}
                      disabled={submitting || !comment.trim()}
                      className="px-4 py-2 bg-[#137fec] text-white text-xs font-semibold rounded hover:bg-[#0f6bd0] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {submitting ? 'Posting...' : 'Post Internal Comment'}
                    </button>
                </div>
            </div>
          </div>

          {/* Sidebar Info */}
          <div className="col-span-12 lg:col-span-4 space-y-6">
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 shadow-lg shadow-black/20">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-white font-semibold text-sm">Ticket Details</h3>
                  <button className="text-[10px] text-[#137fec] hover:underline font-bold uppercase">Edit</button>
                </div>
                <div className="space-y-4">
                    <div className="flex justify-between items-center">
                        <span className="text-xs text-[#92adc9]">Status</span>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase border ${
                          incident.status === 'RESOLVED' 
                          ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' 
                          : 'bg-blue-500/10 text-blue-400 border-blue-500/20'
                        }`}>
                          {incident.status}
                        </span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-xs text-[#92adc9]">Priority</span>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase border ${
                          incident.priority === 'CRITICAL'
                          ? 'bg-rose-500/10 text-rose-500 border-rose-500/20'
                          : 'bg-amber-500/10 text-amber-500 border-amber-500/20'
                        }`}>
                          {incident.priority}
                        </span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-xs text-[#92adc9]">Assigned To</span>
                        {isEditingOwner ? (
                          <div className="flex gap-2">
                             <input 
                              type="text" 
                              value={newOwner}
                              onChange={(e) => setNewOwner(e.target.value)}
                              className="bg-[#111a22] border border-[#233648] rounded px-2 py-1 text-xs text-white focus:outline-none focus:border-primary w-24"
                            />
                            <button onClick={handleOwnerChange} className="material-symbols-outlined text-emerald-500 text-sm">check</button>
                            <button onClick={() => setIsEditingOwner(false)} className="material-symbols-outlined text-rose-500 text-sm">close</button>
                          </div>
                        ) : (
                          <span 
                            onClick={() => setIsEditingOwner(true)}
                            className="text-xs text-white font-medium cursor-pointer hover:text-primary transition-colors flex items-center gap-1"
                          >
                            {incident.assigned_to}
                            <span className="material-symbols-outlined text-[14px]">edit</span>
                          </span>
                        )}
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-xs text-[#92adc9]">Organization</span>
                        <span className="text-xs text-[#137fec] font-medium">{incident.organization}</span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-xs text-[#92adc9]">Reported At</span>
                        <span className="text-xs text-white font-medium">{incident.created_at}</span>
                    </div>
                </div>
            </div>

            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 shadow-lg shadow-black/20">
                <h3 className="text-white font-semibold text-sm mb-4">Reporter Information</h3>
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold">
                        {incident.reporter?.[0] || 'S'}
                    </div>
                    <div>
                        <div className="text-sm font-bold text-white">{incident.reporter}</div>
                        <div className="text-[11px] text-[#92adc9]">Verified Customer Support</div>
                    </div>
                </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
};

export default IncidentDetail;
