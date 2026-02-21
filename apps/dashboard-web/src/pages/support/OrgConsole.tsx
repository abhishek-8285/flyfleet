import React, { useEffect, useMemo, useState } from 'react';

const toNumber = (value: any): number => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
};

const OrgConsole: React.FC = () => {
  const [org, setOrg] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const orgId = localStorage.getItem('org_id') || 'default';

  useEffect(() => {
    const fetchOrg = async () => {
      const token = localStorage.getItem('access_token');
      const headers = { Authorization: `Bearer ${token}` };

      setLoading(true);
      setError(null);

      try {
        const summaryRes = await fetch(`/api/v1/support/orgs/${orgId}/summary`, { headers });
        if (summaryRes.ok) {
          const data = await summaryRes.json();
          setOrg(data);
          return;
        }

        const fallbackRes = await fetch('/api/v1/support/orgs/overview', { headers });
        if (fallbackRes.ok) {
          setOrg(await fallbackRes.json());
          return;
        }

        setError('Failed to load organization overview');
      } catch (err) {
        console.error('Failed to fetch org console data', err);
        setError('Failed to load organization overview');
      } finally {
        setLoading(false);
      }
    };

    fetchOrg();
  }, [orgId]);

  const normalized = useMemo(() => {
    return {
      name: org?.name || org?.org_name || '--',
      plan: org?.plan || org?.subscription_plan || '--',
      renewal: org?.renewal || org?.renewal_date || '--',
      totalDevices: toNumber(org?.total_devices ?? org?.devices_total),
      onlineDevices: toNumber(org?.online_devices ?? org?.devices_online),
      apiHealth: toNumber(org?.api_health ?? org?.integration_health),
    };
  }, [org]);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <h2 className="text-white text-lg font-bold">Organization Console</h2>
            <div className="flex items-center gap-2 text-[11px] text-[#92adc9] mt-0.5">
              <span>SUPPORT</span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="text-white">{normalized.name}</span>
            </div>
          </div>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        {error && (
          <div className="mb-4 rounded border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-300">
            {error}
          </div>
        )}

        {loading ? (
          <div className="h-[320px] flex items-center justify-center text-[#92adc9] text-sm">
            Loading organization data...
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
              <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 relative overflow-hidden group">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1 tracking-wider">Subscription Plan</div>
                <div className="text-2xl font-bold text-white mb-1">{normalized.plan}</div>
                <div className="text-xs text-[#92adc9]">Next renewal: <span className="text-white">{normalized.renewal}</span></div>
              </div>

              <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 relative overflow-hidden group">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1 tracking-wider">Total Devices</div>
                <div className="text-2xl font-bold text-white mb-1">{normalized.totalDevices.toLocaleString()}</div>
                <div className="text-xs text-emerald-500 flex items-center gap-1 font-semibold">{normalized.onlineDevices.toLocaleString()} Online</div>
              </div>

              <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5 relative overflow-hidden group">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1 tracking-wider">Integration Health</div>
                <div className="text-2xl font-bold text-emerald-500 mb-1">{normalized.apiHealth}%</div>
                <div className="text-xs text-[#92adc9]">API Success Rate (24h)</div>
              </div>
            </div>

            <div className="bg-[#16202a] rounded-xl border border-[#233648] h-[400px] flex flex-col overflow-hidden relative">
              <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-primary text-[18px]">map</span>
                  Real-time Fleet Distribution
                </h3>
              </div>
              <div className="flex-1 bg-[#1a2632] flex items-center justify-center">
                <span className="text-[#92adc9] text-sm">Fleet map endpoint is pending frontend map integration.</span>
              </div>
            </div>
          </>
        )}
      </div>
    </main>
  );
};

export default OrgConsole;
