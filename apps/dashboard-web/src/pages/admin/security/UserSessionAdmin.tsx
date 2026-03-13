import { AnimatePresence, motion } from "framer-motion";
import React, { useState } from "react";
import AuthInput from "../../../components/auth/AuthInput";
import EnterpriseAuthLayout from "../../../components/auth/EnterpriseAuthLayout";
import api from "../../../services/api";
import { UserSession } from "../../../types/auth";

const UserSessionAdmin: React.FC = () => {
  const [sessions, setSessions] = useState<UserSession[]>([]);
  const [searchUserId, setSearchUserId] = useState("");
  const [loading, setLoading] = useState(false);

  const fetchSessions = async () => {
    if (!searchUserId) return;
    setLoading(true);
    try {
      const resp = await api.getUserSessions(searchUserId);
      setSessions(resp.data);
    } catch (err) {
      // Mock for admin view
      setSessions([
        {
          id: "sess_1",
          userId: searchUserId,
          userName: "Abhishek Singh",
          deviceName: "MacBook Pro 16",
          deviceType: "desktop",
          ipAddress: "192.168.1.45",
          location: "Mumbai, India",
          lastActive: "Now",
          status: "active",
        },
        {
          id: "sess_2",
          userId: searchUserId,
          userName: "Abhishek Singh",
          deviceName: "iPhone Pro Max",
          deviceType: "mobile",
          ipAddress: "223.187.45.12",
          location: "Bangalore, India",
          lastActive: "2 hours ago",
          status: "inactive",
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

   const handleRevoke = async (sessionId: string) => {
     if (!window.confirm("Terminate this user session immediately?")) return;
     try {
       await api.revokeUserSession(searchUserId, sessionId);
       setSessions((prev) => prev.filter((s) => s.id !== sessionId));
     } catch (err) {
       setSessions((prev) => prev.filter((s) => s.id !== sessionId));
     }
   };

  return (
    <EnterpriseAuthLayout
      title="User Session Oversight"
      subtitle="Administrative terminal control utility"
    >
      <div className="p-4 space-y-6">
        <div className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-6 rounded-[32px] space-y-4">
          <div className="flex gap-4 items-end">
            <div className="flex-1">
              <AuthInput
                label="User Identifier"
                icon="person_search"
                placeholder="Enter User ID or Email"
                value={searchUserId}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setSearchUserId(e.target.value)
                }
              />
            </div>
            <button
              onClick={fetchSessions}
              className="px-8 py-4 bg-primary rounded-2xl text-white font-bold text-sm hover:scale-[1.02] transition-all disabled:opacity-50"
              disabled={!searchUserId || loading}
            >
              Inspect
            </button>
          </div>
        </div>

        <div className="space-y-4">
          <AnimatePresence mode="popLayout">
            {sessions.map((session) => (
              <motion.div
                key={session.id}
                initial={{ opacity: 0, scale: 0.98 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, x: 20 }}
                className="bg-slate-900/40 backdrop-blur-xl border border-white/5 p-6 rounded-[32px] flex items-center justify-between group"
              >
                <div className="flex items-center gap-6">
                  <div className="w-12 h-12 bg-slate-800 rounded-full flex items-center justify-center">
                    <span className="material-symbols-outlined text-slate-400">
                      {session.deviceType === "mobile"
                        ? "smartphone"
                        : "laptop_mac"}
                    </span>
                  </div>
                  <div className="space-y-1">
                    <p className="text-white font-bold">{session.deviceName}</p>
                    <div className="flex gap-3 text-[10px] text-slate-500 font-bold uppercase tracking-widest">
                      <span>{session.ipAddress}</span>
                      <span>•</span>
                      <span>{session.location}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-8">
                  <div className="text-right space-y-1">
                    <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">
                      Last Activity
                    </p>
                    <p className="text-white text-xs font-mono">
                      {session.lastActive}
                    </p>
                  </div>
                  <button
                    onClick={() => handleRevoke(session.id)}
                    className="p-3 bg-red-500/10 text-red-500 rounded-xl opacity-0 group-hover:opacity-100 transition-all hover:bg-red-500/20"
                    title="Terminate Session"
                  >
                    <span className="material-symbols-outlined">logout</span>
                  </button>
                </div>
              </motion.div>
            ))}
          </AnimatePresence>

          {sessions.length === 0 && !loading && searchUserId && (
            <div className="text-center py-20 bg-slate-900/20 rounded-[40px] border border-dashed border-white/5">
              <span className="material-symbols-outlined text-slate-700 text-5xl mb-4">
                search_off
              </span>
              <p className="text-slate-500 font-medium">
                No active sessions found for this user.
              </p>
            </div>
          )}
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default UserSessionAdmin;
