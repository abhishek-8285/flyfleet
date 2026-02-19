import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../../services/apiClient';

interface Settings {
    schedule_enabled: boolean;
    frequency: string;
    schedule_day: string;
    schedule_time: string;
    intensity: string;
    notification_emails: string[];
}

const DRTestSettingsView: React.FC = () => {
    const navigate = useNavigate();
    const [settings, setSettings] = useState<Settings>({
        schedule_enabled: false,
        frequency: 'WEEKLY',
        schedule_day: 'SATURDAY',
        schedule_time: '02:00',
        intensity: 'NORMAL',
        notification_emails: []
    });
    const [email, setEmail] = useState('');
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        apiClient.get('/admin/dr/tests/settings')
            .then(data => {
                if (data.data) setSettings(data.data);
            })
            .catch(() => setError('Failed to load settings.'))
            .finally(() => setLoading(false));
    }, []);

    const handleSave = async () => {
        setSaving(true);
        try {
            await apiClient.put('/admin/dr/tests/settings', settings);
            alert('Settings saved successfully');
        } catch (error) {
            console.error('Failed to save settings:', error);
            alert('Failed to save settings');
        } finally {
            setSaving(false);
        }
    };

    const handleAddEmail = () => {
        if (email && !settings.notification_emails.includes(email)) {
            setSettings({ ...settings, notification_emails: [...settings.notification_emails, email] });
            setEmail('');
        }
    };

    const handleRemoveEmail = (e: string) => {
        setSettings({ ...settings, notification_emails: settings.notification_emails.filter(x => x !== e) });
    };

    if (loading) {
        return <div className="flex items-center justify-center h-full bg-[#0a0a0f] text-white">Loading...</div>;
    }
    if (error) {
        return <div className="flex items-center justify-center h-full bg-[#0a0a0f] text-red-400">{error}</div>;
    }

    return (
        <div className="flex flex-col h-full w-full bg-[#0a0a0f] text-white font-sans overflow-hidden">
            {/* Header */}
            <header className="h-16 flex items-center justify-between px-6 border-b border-white/10 bg-[#12121a] flex-shrink-0">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/admin/dr/tests')} className="text-white/60 hover:text-white">
                        ← Back
                    </button>
                    <h1 className="text-lg font-semibold">DR Testing Settings</h1>
                </div>
                <button
                    onClick={handleSave}
                    disabled={saving}
                    className="px-4 py-2 bg-gradient-to-r from-emerald-500 to-cyan-500 hover:opacity-90 rounded-lg text-sm font-medium disabled:opacity-50"
                >
                    {saving ? 'Saving...' : 'Save Changes'}
                </button>
            </header>

            {/* Content */}
            <div className="flex-1 overflow-auto p-6">
                <div className="max-w-2xl space-y-6">
                    {/* Scheduling */}
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-6 space-y-4">
                        <div className="flex items-center justify-between">
                            <div>
                                <h3 className="font-medium">Automated Testing</h3>
                                <p className="text-sm text-white/50">Schedule recurring DR tests</p>
                            </div>
                            <button
                                onClick={() => setSettings({ ...settings, schedule_enabled: !settings.schedule_enabled })}
                                className={`w-12 h-6 rounded-full transition-colors ${settings.schedule_enabled ? 'bg-emerald-500' : 'bg-white/20'}`}
                            >
                                <span className={`block w-5 h-5 rounded-full bg-white transition-transform mx-0.5 ${settings.schedule_enabled ? 'translate-x-6' : ''}`} />
                            </button>
                        </div>

                        {settings.schedule_enabled && (
                            <div className="grid grid-cols-2 gap-4 pt-4 border-t border-white/10">
                                <div>
                                    <label className="block text-sm text-white/70 mb-2">Frequency</label>
                                    <select
                                        value={settings.frequency}
                                        onChange={(e) => setSettings({ ...settings, frequency: e.target.value })}
                                        className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2"
                                    >
                                        <option value="WEEKLY">Weekly</option>
                                        <option value="MONTHLY">Monthly</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="block text-sm text-white/70 mb-2">Day</label>
                                    <select
                                        value={settings.schedule_day}
                                        onChange={(e) => setSettings({ ...settings, schedule_day: e.target.value })}
                                        className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2"
                                    >
                                        <option value="SATURDAY">Saturday</option>
                                        <option value="SUNDAY">Sunday</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="block text-sm text-white/70 mb-2">Time</label>
                                    <input
                                        type="time"
                                        value={settings.schedule_time}
                                        onChange={(e) => setSettings({ ...settings, schedule_time: e.target.value })}
                                        className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm text-white/70 mb-2">Intensity</label>
                                    <select
                                        value={settings.intensity}
                                        onChange={(e) => setSettings({ ...settings, intensity: e.target.value })}
                                        className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2"
                                    >
                                        <option value="NORMAL">Normal</option>
                                        <option value="HIGH">High</option>
                                    </select>
                                </div>
                            </div>
                        )}
                    </div>

                    {/* Notifications */}
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-6 space-y-4">
                        <h3 className="font-medium">Notification Recipients</h3>
                        <div className="flex gap-2">
                            <input
                                type="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                placeholder="email@example.com"
                                className="flex-1 bg-white/5 border border-white/10 rounded-lg px-4 py-2"
                            />
                            <button
                                onClick={handleAddEmail}
                                className="px-4 py-2 bg-white/10 hover:bg-white/20 rounded-lg"
                            >
                                Add
                            </button>
                        </div>
                        <div className="flex flex-wrap gap-2">
                            {settings.notification_emails.map((e) => (
                                <span key={e} className="px-3 py-1 bg-white/5 rounded-full text-sm flex items-center gap-2">
                                    {e}
                                    <button onClick={() => handleRemoveEmail(e)} className="text-white/40 hover:text-white">×</button>
                                </span>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default DRTestSettingsView;
