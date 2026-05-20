import React, { useState, useEffect } from 'react';
import { 
  Plus, 
  Search, 
  Map as MapIcon, 
  Circle, 
  Pentagon, 
  Trash2, 
  Edit,
  Shield,
  Activity,
  Navigation,
  AlertTriangle,
  Mail,
  MessageSquare,
  Bell,
  Play,
  X,
  Check
} from 'lucide-react';
import { api } from '../../services/api';
import { Geofence } from '../../modules/dashboard/components/LiveOperationsMap';
import { LoadingSpinner } from '../../components/common/LoadingSpinner';

interface NotificationRule {
  id: string;
  name: string;
  geofenceId: string;
  geofenceName: string;
  event: 'enter' | 'exit' | 'both';
  channels: {
    email: boolean;
    sms: boolean;
    push: boolean;
  };
  recipients: string;
  messageTemplate: string;
  active: boolean;
}

const GeofenceManagement: React.FC = () => {
  const [geofences, setGeofences] = useState<Geofence[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  // Notification Rules State
  const [activeTab, setActiveTab] = useState<'boundaries' | 'notifications'>('boundaries');
  const [rules, setRules] = useState<NotificationRule[]>([]);
  const [showAddRuleModal, setShowAddRuleModal] = useState(false);
  const [editingRuleId, setEditingRuleId] = useState<string | null>(null);
  const [ruleForm, setRuleForm] = useState({
    name: '',
    geofenceId: '',
    geofenceName: '',
    event: 'enter' as 'enter' | 'exit' | 'both',
    channels: { email: true, sms: false, push: true },
    recipients: '',
    messageTemplate: 'Vehicle {vehicle_name} has {event_type}ed the {geofence_name} geofence.'
  });

  // Simulated Alert Toast State
  const [alertToast, setAlertToast] = useState<{
    medium: 'EMAIL' | 'SMS' | 'PUSH';
    recipient: string;
    message: string;
    ruleName: string;
  } | null>(null);

  // Simulated Vehicle inside/outside state tracker to ensure exit only happens after entry
  const [simulatedStates, setSimulatedStates] = useState<Record<string, boolean>>({});

  const fetchGeofences = async () => {
    try {
      setLoading(true);
      const res = await api.getGeofences();
      const data = res?.data || res;
      setGeofences(Array.isArray(data) ? data : data?.geofences || []);
    } catch (err) {
      console.error('Failed to fetch geofences:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchGeofences();
  }, []);

  // Load and seed Notification Rules
  useEffect(() => {
    const stored = localStorage.getItem('flyfleet_notification_rules');
    if (stored) {
      setRules(JSON.parse(stored));
    } else {
      const mockRules: NotificationRule[] = [
        {
          id: '1',
          name: 'Burari Terminal Arrival Alert',
          geofenceId: 'b66cc463-11f4-41d5-ac6a-ca02ac1a625e',
          geofenceName: 'BURARI',
          event: 'enter',
          channels: { email: true, sms: false, push: true },
          recipients: 'dispatch-alert@flyfleet.com',
          messageTemplate: 'ALERT: Vehicle {vehicle_name} entered the BURARI yard zone at {timestamp}. Preparing unloading bay.',
          active: true
        },
        {
          id: '2',
          name: 'Home Depot Yard Security',
          geofenceId: '728778ff-2092-411a-85ee-f3310461be61',
          geofenceName: 'HOME',
          event: 'exit',
          channels: { email: false, sms: true, push: true },
          recipients: '+91 98765 43210',
          messageTemplate: 'Yard Out: Fleet unit {vehicle_name} has exited HOME geofence boundaries.',
          active: true
        }
      ];
      setRules(mockRules);
      localStorage.setItem('flyfleet_notification_rules', JSON.stringify(mockRules));
    }
  }, []);

  const confirmDelete = async () => {
    if (!deleteConfirmId) return;
    try {
      await api.deleteGeofence(deleteConfirmId);
      fetchGeofences();
    } catch (err) {
      console.error('Failed to delete geofence:', err);
    } finally {
      setDeleteConfirmId(null);
    }
  };

  const handleCloseModal = () => {
    setShowAddRuleModal(false);
    setEditingRuleId(null);
    setRuleForm({
      name: '',
      geofenceId: '',
      geofenceName: '',
      event: 'enter',
      channels: { email: true, sms: false, push: true },
      recipients: '',
      messageTemplate: 'Vehicle {vehicle_name} has {event_type}ed the {geofence_name} geofence.'
    });
  };

  const handleStartEditRule = (rule: NotificationRule) => {
    setEditingRuleId(rule.id);
    setRuleForm({
      name: rule.name,
      geofenceId: rule.geofenceId,
      geofenceName: rule.geofenceName,
      event: rule.event,
      channels: rule.channels,
      recipients: rule.recipients,
      messageTemplate: rule.messageTemplate
    });
    setShowAddRuleModal(true);
  };

  const handleCreateOrUpdateRule = (e: React.FormEvent) => {
    e.preventDefault();
    if (!ruleForm.name || !ruleForm.geofenceId) {
      alert('Please fill out the name and select a geofence.');
      return;
    }

    const selectedGeo = geofences.find(g => g.id === ruleForm.geofenceId);
    const geofenceName = selectedGeo?.name || ruleForm.geofenceName || 'UNKNOWN';

    let updatedRules: NotificationRule[];

    if (editingRuleId) {
      updatedRules = rules.map(r => r.id === editingRuleId ? {
        ...r,
        name: ruleForm.name,
        geofenceId: ruleForm.geofenceId,
        geofenceName: geofenceName,
        event: ruleForm.event,
        channels: ruleForm.channels,
        recipients: ruleForm.recipients,
        messageTemplate: ruleForm.messageTemplate
      } : r);
    } else {
      const newRule: NotificationRule = {
        id: Math.random().toString(36).substring(2, 9),
        name: ruleForm.name,
        geofenceId: ruleForm.geofenceId,
        geofenceName: geofenceName,
        event: ruleForm.event,
        channels: ruleForm.channels,
        recipients: ruleForm.recipients,
        messageTemplate: ruleForm.messageTemplate,
        active: true
      };
      updatedRules = [...rules, newRule];
    }

    setRules(updatedRules);
    localStorage.setItem('flyfleet_notification_rules', JSON.stringify(updatedRules));
    handleCloseModal();
  };

  const toggleRuleActive = (id: string) => {
    const updated = rules.map(r => r.id === id ? { ...r, active: !r.active } : r);
    setRules(updated);
    localStorage.setItem('flyfleet_notification_rules', JSON.stringify(updated));
  };

  const deleteRule = (id: string) => {
    if (!window.confirm('Are you sure you want to delete this notification rule?')) return;
    const updated = rules.filter(r => r.id !== id);
    setRules(updated);
    localStorage.setItem('flyfleet_notification_rules', JSON.stringify(updated));
  };

  // Simulate Alert Triggering
  const triggerSimulation = (rule: NotificationRule) => {
    const mockVehicles = ['TRUCK-DL3C-5542', 'SEMI-UP16-9908', 'VAN-HR26-1044', 'TANKER-MH12-9021'];
    const randomVehicle = mockVehicles[Math.floor(Math.random() * mockVehicles.length)];
    
    const stateKey = `${randomVehicle}_${rule.geofenceId}`;
    const isCurrentlyInside = !!simulatedStates[stateKey];

    let eventType: 'enter' | 'exit' = 'enter';
    if (rule.event === 'enter') {
      eventType = 'enter';
    } else if (rule.event === 'exit') {
      eventType = 'exit';
    } else {
      // For 'both', transition state: if inside, trigger exit. If outside, trigger enter.
      eventType = isCurrentlyInside ? 'exit' : 'enter';
      setSimulatedStates(prev => ({
        ...prev,
        [stateKey]: !isCurrentlyInside
      }));
    }

    const timeStr = new Date().toLocaleTimeString();

    const formattedMsg = rule.messageTemplate
      .replace('{vehicle_name}', randomVehicle)
      .replace('{event_type}', eventType)
      .replace('{geofence_name}', rule.geofenceName.toUpperCase())
      .replace('{timestamp}', timeStr);

    let primaryMedium: 'EMAIL' | 'SMS' | 'PUSH' = 'PUSH';
    if (rule.channels.email) primaryMedium = 'EMAIL';
    else if (rule.channels.sms) primaryMedium = 'SMS';

    setAlertToast({
      medium: primaryMedium,
      recipient: rule.recipients || 'Active Dispatch Console',
      message: formattedMsg,
      ruleName: rule.name
    });

    // Auto fade out
    setTimeout(() => {
      setAlertToast(null);
    }, 6000);
  };

  const filteredGeofences = geofences.filter(g => 
    g.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="p-6 space-y-6 bg-[#f8fafc] dark:bg-black min-h-screen relative overflow-x-hidden">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-gray-900 dark:text-white uppercase">
            Geofence Management
          </h1>
          <p className="text-sm text-gray-500 font-medium">
            Monitor and configure spatial boundaries for automated alerts.
          </p>
        </div>
        {activeTab === 'boundaries' ? (
          <button 
            onClick={() => window.location.href = '/admin/geofences/new'}
            className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2.5 rounded-xl font-bold transition-all shadow-lg shadow-emerald-600/20 text-sm uppercase tracking-tighter"
          >
            <Plus size={18} />
            Create New Geofence
          </button>
        ) : (
          <button 
            onClick={() => setShowAddRuleModal(true)}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-500 text-white px-4 py-2.5 rounded-xl font-bold transition-all shadow-lg shadow-blue-600/20 text-sm uppercase tracking-tighter"
          >
            <Plus size={18} />
            Create Alert Rule
          </button>
        )}
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[
          { label: 'Total Zones', value: geofences.length, icon: MapIcon, color: 'text-blue-500' },
          { label: 'Notification Rules', value: rules.length, icon: Bell, color: 'text-emerald-500' },
          { label: 'Detection Types', value: 'Circle / Polygon', icon: Shield, color: 'text-purple-500' }
        ].map((stat, i) => (
          <div key={i} className="bg-white dark:bg-white/[0.03] p-4 rounded-2xl border border-gray-100 dark:border-white/5 shadow-sm">
            <div className="flex items-center justify-between mb-1">
              <p className="text-[10px] font-black uppercase tracking-widest text-gray-400">{stat.label}</p>
              <stat.icon size={16} className={stat.color} />
            </div>
            <p className="text-2xl font-black text-gray-900 dark:text-white">{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Tab Switcher */}
      <div className="flex gap-2 border-b border-gray-100 dark:border-white/5 pb-px shrink-0">
        <button
          onClick={() => setActiveTab('boundaries')}
          className={`flex items-center gap-2 px-4 py-3 text-xs font-black uppercase tracking-widest border-b-2 transition-all ${activeTab === 'boundaries' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-white'}`}
        >
          <MapIcon size={16} />
          Active Boundaries
        </button>
        <button
          onClick={() => setActiveTab('notifications')}
          className={`flex items-center gap-2 px-4 py-3 text-xs font-black uppercase tracking-widest border-b-2 transition-all ${activeTab === 'notifications' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-white'}`}
        >
          <Bell size={16} />
          Alert Rules Config
        </button>
      </div>

      {/* Tab Content 1: Boundaries List */}
      {activeTab === 'boundaries' && (
        <div className="space-y-6 animate-in fade-in duration-200">
          {/* Toolbar */}
          <div className="flex flex-col md:flex-row gap-4 items-center bg-white dark:bg-white/[0.02] p-2 rounded-2xl border border-gray-100 dark:border-white/5">
            <div className="relative flex-1 w-full">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
              <input
                type="text"
                placeholder="SEARCH GEOFENCES..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full bg-transparent pl-10 pr-4 py-2 text-sm font-bold uppercase tracking-tighter text-gray-700 dark:text-white focus:outline-none"
              />
            </div>
          </div>

          {/* List */}
          {loading ? (
            <div className="flex justify-center py-20">
              <LoadingSpinner />
            </div>
          ) : filteredGeofences.length > 0 ? (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              {filteredGeofences.map((gf) => (
                <div key={gf.id} className="bg-white dark:bg-white/[0.03] border border-gray-100 dark:border-white/5 rounded-2xl overflow-hidden hover:shadow-xl transition-all group">
                  <div className="p-5 flex items-start justify-between">
                    <div className="flex gap-4">
                      <div className={`w-12 h-12 rounded-2xl flex items-center justify-center ${gf.type === 'circle' ? 'bg-blue-50 text-blue-600' : 'bg-purple-50 text-purple-600'} dark:bg-white/5`}>
                        {gf.type === 'circle' ? <Circle size={24} /> : <Pentagon size={24} />}
                      </div>
                      <div>
                        <h3 className="font-black text-gray-900 dark:text-white uppercase tracking-tighter text-lg leading-none mb-1">
                          {gf.name}
                        </h3>
                        <div className="flex items-center gap-3">
                          <span className={`text-[10px] font-black uppercase tracking-widest px-2 py-0.5 rounded ${gf.status === 'active' ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-600'}`}>
                            {gf.status}
                          </span>
                          <span className="text-[10px] font-black uppercase tracking-widest text-gray-400 flex items-center gap-1">
                            <Navigation size={10} />
                            {gf.type === 'circle' ? `${gf.radius_m}m Radius` : 'Polygonal Boundary'}
                          </span>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2 md:opacity-0 group-hover:opacity-100 transition-opacity">
                      <button 
                        onClick={() => window.location.href = `/admin/geofences/new?edit=${gf.id}`}
                        className="p-2 hover:bg-gray-100 dark:hover:bg-white/5 rounded-lg text-gray-400 hover:text-blue-500 transition-colors"
                      >
                        <Edit size={18} />
                      </button>
                      <button 
                        onClick={() => setDeleteConfirmId(gf.id)}
                        className="p-2 hover:bg-gray-100 dark:hover:bg-white/5 rounded-lg text-gray-400 hover:text-red-500 transition-colors"
                      >
                        <Trash2 size={18} />
                      </button>
                    </div>
                  </div>
                  
                  {/* Mini Preview Placeholder / Info */}
                  <div className="px-5 py-3 bg-gray-50 dark:bg-white/[0.01] border-t border-gray-100 dark:border-white/5 flex justify-between items-center text-[11px] font-bold uppercase tracking-tighter text-gray-400">
                    <div className="flex gap-4">
                      <span>ID: {gf.id.substring(0, 8)}</span>
                      {gf.center_lat && (
                        <span>{gf.center_lat.toFixed(4)}, {gf.center_lng?.toFixed(4)}</span>
                      )}
                    </div>
                    <div className="flex items-center gap-1 text-emerald-500">
                      <Activity size={12} />
                      Live Monitoring
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-white dark:bg-white/[0.02] border-2 border-dashed border-gray-200 dark:border-white/5 rounded-3xl py-20 flex flex-col items-center justify-center text-center">
              <div className="w-16 h-16 bg-gray-100 dark:bg-white/5 rounded-2xl flex items-center justify-center text-gray-400 mb-4">
                <MapIcon size={32} />
              </div>
              <h3 className="text-xl font-black text-gray-900 dark:text-white uppercase tracking-tighter mb-1">No Geofences Found</h3>
              <p className="text-gray-400 max-w-xs mb-6">Start by creating a new spatial boundary to monitor asset entry and exit events.</p>
              <button 
                onClick={() => window.location.href = '/admin/geofences/new'}
                className="bg-emerald-600 hover:bg-emerald-500 text-white px-6 py-3 rounded-xl font-bold transition-all uppercase tracking-tighter"
              >
                Create Your First Geofence
              </button>
            </div>
          )}
        </div>
      )}

      {/* Tab Content 2: Notification Rules */}
      {activeTab === 'notifications' && (
        <div className="space-y-6 animate-in fade-in duration-200">
          <div className="bg-white dark:bg-white/[0.02] border border-gray-100 dark:border-white/5 rounded-2xl p-5">
            <h2 className="text-sm font-black text-gray-900 dark:text-white uppercase tracking-wider mb-1">Automated Warning & Alert Triggers</h2>
            <p className="text-[11px] text-gray-400 font-medium">
              Configure alert distribution policies. When live GPS trackers match entry/exit criteria for target zones, FlyFleet automatically triggers direct notifications.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {rules.map((rule) => (
              <div key={rule.id} className="bg-white dark:bg-white/[0.02] border border-gray-100 dark:border-white/5 rounded-2xl p-5 hover:shadow-xl transition-all relative flex flex-col justify-between">
                <div>
                  <div className="flex items-start justify-between mb-4">
                    <div>
                      <h3 className="font-black text-gray-900 dark:text-white uppercase tracking-tighter text-base leading-none mb-1">
                        {rule.name}
                      </h3>
                      <div className="flex items-center gap-2">
                        <span className="text-[9px] font-black uppercase tracking-widest bg-blue-50 dark:bg-blue-500/10 text-blue-600 dark:text-blue-400 px-2 py-0.5 rounded">
                          ZONE: {rule.geofenceName.toUpperCase()}
                        </span>
                        <span className="text-[9px] font-black uppercase tracking-widest bg-purple-50 dark:bg-purple-500/10 text-purple-600 dark:text-purple-400 px-2 py-0.5 rounded">
                          {rule.event === 'both' ? 'ENTER & EXIT' : `ON ${rule.event.toUpperCase()}`}
                        </span>
                      </div>
                    </div>

                    {/* Active toggle */}
                    <button 
                      onClick={() => toggleRuleActive(rule.id)}
                      className={`w-9 h-5 rounded-full p-0.5 transition-colors ${rule.active ? 'bg-emerald-600' : 'bg-gray-200 dark:bg-white/10'}`}
                    >
                      <div className={`w-4 h-4 rounded-full bg-white transition-transform ${rule.active ? 'translate-x-4' : 'translate-x-0'}`} />
                    </button>
                  </div>

                  {/* Message Template */}
                  <div className="bg-gray-50 dark:bg-white/[0.01] border border-gray-100 dark:border-white/5 rounded-xl p-3 mb-4">
                    <p className="text-[10px] font-black uppercase tracking-widest text-gray-400 mb-1">ALERT MESSAGE TEMPLATE</p>
                    <p className="text-xs font-mono text-gray-600 dark:text-gray-300 leading-relaxed">
                      "{rule.messageTemplate}"
                    </p>
                  </div>
                </div>

                {/* Footer Controls */}
                <div className="border-t border-gray-100 dark:border-white/5 pt-4 flex items-center justify-between mt-2">
                  <div className="flex items-center gap-2 text-gray-400">
                    <span className="text-[10px] font-black uppercase tracking-wider text-gray-400 mr-1">CHANNELS:</span>
                    {rule.channels.email && <Mail size={14} className="text-blue-500" />}
                    {rule.channels.sms && <MessageSquare size={14} className="text-emerald-500" />}
                    {rule.channels.push && <Bell size={14} className="text-purple-500" />}
                  </div>

                  <div className="flex gap-2">
                    <button
                      onClick={() => triggerSimulation(rule)}
                      disabled={!rule.active}
                      className="flex items-center gap-1.5 bg-blue-50 hover:bg-blue-100 dark:bg-blue-500/10 dark:hover:bg-blue-500/20 text-blue-600 dark:text-blue-400 px-3 py-1.5 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all disabled:opacity-40"
                    >
                      <Play size={10} />
                      Test Alert
                    </button>
                    <button
                      onClick={() => handleStartEditRule(rule)}
                      className="p-1.5 hover:bg-blue-50 dark:hover:bg-blue-500/10 text-gray-400 hover:text-blue-500 rounded-xl transition-all"
                    >
                      <Edit size={14} />
                    </button>
                    <button
                      onClick={() => deleteRule(rule.id)}
                      className="p-1.5 hover:bg-red-50 dark:hover:bg-red-500/10 text-gray-400 hover:text-red-500 rounded-xl transition-all"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteConfirmId && (
        <div className="fixed inset-0 z-[2000] bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white dark:bg-[#1a2b3c] border border-gray-100 dark:border-white/10 rounded-[2rem] max-w-sm w-full p-6 shadow-2xl animate-in zoom-in-95 duration-200">
            <div className="w-12 h-12 rounded-2xl bg-red-50 dark:bg-red-500/10 text-red-500 flex items-center justify-center mx-auto mb-4">
              <AlertTriangle size={24} />
            </div>
            <h3 className="text-base font-black text-gray-900 dark:text-white uppercase tracking-tighter text-center mb-1">
              Delete Geofence?
            </h3>
            <p className="text-[11px] text-gray-500 dark:text-gray-400 font-medium text-center mb-6 leading-relaxed">
              Are you sure you want to delete this geofence? This action cannot be undone and will stop all active monitoring alerts for this zone.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setDeleteConfirmId(null)}
                className="flex-1 py-3 bg-gray-50 hover:bg-gray-100 dark:bg-white/5 dark:hover:bg-white/10 rounded-xl text-[10px] font-black uppercase tracking-widest text-gray-500 dark:text-gray-400 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={confirmDelete}
                className="flex-1 py-3 bg-red-600 hover:bg-red-500 rounded-xl text-[10px] font-black uppercase tracking-widest text-white transition-colors shadow-lg shadow-red-600/20"
              >
                Delete Zone
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Notification Rule Modal */}
      {showAddRuleModal && (
        <div className="fixed inset-0 z-[2000] bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white dark:bg-[#1a2b3c] border border-gray-100 dark:border-white/10 rounded-[2rem] max-w-lg w-full overflow-hidden shadow-2xl animate-in zoom-in-95 duration-200">
            <div className="h-14 bg-gray-50 dark:bg-white/[0.02] border-b border-gray-100 dark:border-white/5 flex items-center justify-between px-6">
              <h3 className="text-sm font-black text-gray-900 dark:text-white uppercase tracking-wider">
                {editingRuleId ? 'Edit Alarm & Notification Rule' : 'Create Alarm & Notification Rule'}
              </h3>
              <button 
                onClick={handleCloseModal}
                className="p-1 hover:bg-gray-200 dark:hover:bg-white/10 rounded-lg text-gray-400"
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleCreateOrUpdateRule} className="p-6 space-y-4">
              <div>
                <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Rule Name</label>
                <input
                  type="text"
                  required
                  value={ruleForm.name}
                  onChange={(e) => setRuleForm(prev => ({ ...prev, name: e.target.value }))}
                  placeholder="e.g. Warehouse Alpha Entry Alarm"
                  className="w-full bg-gray-50 dark:bg-white/5 border border-gray-100 dark:border-white/10 rounded-xl px-4 py-2.5 text-xs font-bold text-gray-900 dark:text-white outline-none focus:ring-2 ring-blue-500/20"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Target Geofence</label>
                  <select
                    required
                    value={ruleForm.geofenceId}
                    onChange={(e) => setRuleForm(prev => ({ ...prev, geofenceId: e.target.value }))}
                    className="w-full bg-gray-50 dark:bg-white/5 border border-gray-100 dark:border-white/10 rounded-xl px-4 py-2.5 text-xs font-bold text-gray-900 dark:text-white outline-none focus:ring-2 ring-blue-500/20"
                  >
                    <option value="">Select Zone...</option>
                    {geofences.map(g => (
                      <option key={g.id} value={g.id}>{g.name.toUpperCase()}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Trigger Condition</label>
                  <select
                    value={ruleForm.event}
                    onChange={(e) => setRuleForm(prev => ({ ...prev, event: e.target.value as 'enter' | 'exit' | 'both' }))}
                    className="w-full bg-gray-50 dark:bg-white/5 border border-gray-100 dark:border-white/10 rounded-xl px-4 py-2.5 text-xs font-bold text-gray-900 dark:text-white outline-none focus:ring-2 ring-blue-500/20"
                  >
                    <option value="enter">On Entry Only</option>
                    <option value="exit">On Exit Only</option>
                    <option value="both">On Both Enter & Exit</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Notification Channels</label>
                <div className="grid grid-cols-3 gap-3">
                  {[
                    { key: 'email', label: 'Email', icon: Mail },
                    { key: 'sms', label: 'SMS', icon: MessageSquare },
                    { key: 'push', label: 'In-App Toast', icon: Bell }
                  ].map((ch) => (
                    <button
                      key={ch.key}
                      type="button"
                      onClick={() => setRuleForm(prev => ({
                        ...prev,
                        channels: {
                          ...prev.channels,
                          [ch.key]: !prev.channels[ch.key as keyof typeof prev.channels]
                        }
                      }))}
                      className={`flex items-center justify-center gap-2 p-2.5 rounded-xl border text-[10px] font-black uppercase tracking-wider transition-all ${
                        ruleForm.channels[ch.key as keyof typeof ruleForm.channels]
                          ? 'bg-blue-50 border-blue-200 text-blue-600 dark:bg-blue-500/10 dark:border-blue-500/30 dark:text-blue-400'
                          : 'bg-transparent border-gray-100 dark:border-white/5 text-gray-400 hover:text-gray-600'
                      }`}
                    >
                      <ch.icon size={12} />
                      {ch.label}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Recipient Details</label>
                <input
                  type="text"
                  value={ruleForm.recipients}
                  onChange={(e) => setRuleForm(prev => ({ ...prev, recipients: e.target.value }))}
                  placeholder="e.g. dispatch@flyfleet.com OR +91 98765 43210"
                  className="w-full bg-gray-50 dark:bg-white/5 border border-gray-100 dark:border-white/10 rounded-xl px-4 py-2.5 text-xs font-bold text-gray-900 dark:text-white outline-none focus:ring-2 ring-blue-500/20"
                />
              </div>

              <div>
                <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Alert Message Template</label>
                <textarea
                  value={ruleForm.messageTemplate}
                  onChange={(e) => setRuleForm(prev => ({ ...prev, messageTemplate: e.target.value }))}
                  rows={3}
                  className="w-full bg-gray-50 dark:bg-white/5 border border-gray-100 dark:border-white/10 rounded-xl px-4 py-2.5 text-xs font-mono text-gray-900 dark:text-white outline-none focus:ring-2 ring-blue-500/20"
                />
                <p className="text-[9px] text-gray-400 font-bold mt-1">Placeholders: {"{vehicle_name}, {event_type}, {geofence_name}, {timestamp}"}</p>
              </div>

              <div className="flex gap-3 pt-4 border-t border-gray-100 dark:border-white/5 mt-4">
                <button
                  type="button"
                  onClick={handleCloseModal}
                  className="flex-1 py-3 bg-gray-50 hover:bg-gray-100 dark:bg-white/5 dark:hover:bg-white/10 rounded-xl text-[10px] font-black uppercase tracking-widest text-gray-500 dark:text-gray-400 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 py-3 bg-blue-600 hover:bg-blue-500 rounded-xl text-[10px] font-black uppercase tracking-widest text-white transition-colors shadow-lg shadow-blue-600/20"
                >
                  {editingRuleId ? 'Save Changes' : 'Create Rule'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Simulated Live Alert Toast notification */}
      {alertToast && (
        <div className="fixed bottom-6 right-6 z-[3000] max-w-sm w-full bg-white dark:bg-[#1a2b3c] border-2 border-blue-500/30 rounded-2xl shadow-2xl p-5 animate-in slide-in-from-right duration-300">
          <div className="flex items-start justify-between mb-2">
            <div className="flex items-center gap-2">
              <span className={`w-8 h-8 rounded-xl flex items-center justify-center ${
                alertToast.medium === 'EMAIL' ? 'bg-blue-50 text-blue-600 dark:bg-blue-500/10' :
                alertToast.medium === 'SMS' ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-500/10' :
                'bg-purple-50 text-purple-600 dark:bg-purple-500/10'
              }`}>
                {alertToast.medium === 'EMAIL' ? <Mail size={16} /> :
                 alertToast.medium === 'SMS' ? <MessageSquare size={16} /> :
                 <Bell size={16} />}
              </span>
              <div>
                <p className="text-[10px] font-black uppercase tracking-widest text-blue-500 leading-none mb-1">
                  {alertToast.medium} DISPATCH ALERT SENT
                </p>
                <h4 className="text-xs font-black text-gray-900 dark:text-white uppercase tracking-tighter">
                  {alertToast.ruleName}
                </h4>
              </div>
            </div>
            <button 
              onClick={() => setAlertToast(null)}
              className="p-1 hover:bg-gray-100 dark:hover:bg-white/10 rounded-lg text-gray-400"
            >
              <X size={14} />
            </button>
          </div>

          <div className="space-y-2 mt-3">
            <p className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">
              RECIPIENT: <span className="text-gray-600 dark:text-gray-200 font-mono">{alertToast.recipient}</span>
            </p>
            <div className="bg-gray-50 dark:bg-white/5 border border-gray-100 dark:border-white/5 rounded-xl p-3">
              <p className="text-xs font-mono text-gray-600 dark:text-gray-300 leading-relaxed italic">
                "{alertToast.message}"
              </p>
            </div>
            <div className="flex items-center gap-1.5 text-[9px] font-black text-emerald-500 uppercase tracking-wider justify-end pt-1">
              <Check size={12} />
              Delivered Successfully
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default GeofenceManagement;
