import React, { useEffect, useState } from 'react';
import apiClient from '../../../services/apiClient';

interface LocalizationSetting {
    id: string;
    language_code: string;
    is_active: boolean;
    is_default: boolean;
}

const Localization: React.FC = () => {
    const [settings, setSettings] = useState<LocalizationSetting[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetchSettings();
    }, []);

    const fetchSettings = async () => {
        try {
            const response = await apiClient.get('/settings/localization');
            setSettings(response.data || []);
        } catch (error) {
            console.error('Failed to fetch localization settings', error);
        } finally {
            setLoading(false);
        }
    };

    const handleToggle = async (code: string) => {
        const setting = settings.find(s => s.language_code === code);
        if (!setting) return;

        const updated = { ...setting, is_active: !setting.is_active };
        
        // Optimistic update
        setSettings(settings.map(s => s.language_code === code ? updated : s));

        try {
            await apiClient.put('/settings/localization', updated);
        } catch (error) {
            console.error('Failed to update localization', error);
            fetchSettings();
        }
    };

    const handleSetDefault = async (code: string) => {
        // Find current default and unset it, set new default
        const updatedSettings = settings.map(s => ({
            ...s,
            is_default: s.language_code === code
        }));
        setSettings(updatedSettings);

        try {
             // We might need a separate endpoint for "set default" or just update the one
             const setting = settings.find(s => s.language_code === code);
             if (setting) {
                await apiClient.put('/settings/localization', { ...setting, is_default: true });
             }
        } catch (error) {
            console.error('Failed to set default language', error);
            fetchSettings();
        }
    };

    // Mock data if empty for UI demo
    const displaySettings = settings.length > 0 ? settings : [
        { id: '1', language_code: 'en-US', is_active: true, is_default: true },
        { id: '2', language_code: 'es-ES', is_active: true, is_default: false },
        { id: '3', language_code: 'fr-FR', is_active: false, is_default: false },
        { id: '4', language_code: 'de-DE', is_active: false, is_default: false },
    ];

    if (loading) return <div className="p-6 text-white">Loading...</div>;

    return (
        <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
             <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <div>
                        <h2 className="text-white text-lg font-bold">Localization & Regions</h2>
                        <div className="flex items-center gap-2 text-[11px] text-[#92adc9] mt-0.5 uppercase tracking-wide font-medium">
                            <span>System Settings</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span className="text-white">Localization</span>
                        </div>
                    </div>
                </div>
                <div className="flex items-center gap-4">
                     <button className="px-4 py-2 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-[#137fec]/20">
                        <span className="material-symbols-outlined text-[16px]">add</span>
                        Add Language
                    </button>
                </div>
            </header>
            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                <div className="max-w-5xl mx-auto grid grid-cols-1 gap-6">
                    <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
                        <div className="p-4 border-b border-[#233648] bg-[#111a22]/50">
                            <h3 className="text-white font-semibold text-sm">Supported Languages</h3>
                        </div>
                        <div className="divide-y divide-[#233648]">
                            {displaySettings.map((lang) => (
                                <div key={lang.language_code} className="p-4 flex items-center justify-between hover:bg-[#111a22]/50 transition-colors">
                                    <div className="flex items-center gap-4">
                                        <div className="w-10 h-10 rounded-full bg-[#111a22] border border-[#233648] flex items-center justify-center text-white font-bold text-xs">
                                            {lang.language_code.split('-')[1]}
                                        </div>
                                        <div>
                                            <div className="text-white text-sm font-medium flex items-center gap-2">
                                                {lang.language_code}
                                                {lang.is_default && (
                                                    <span className="px-1.5 py-0.5 rounded text-[10px] bg-[#137fec]/20 text-[#137fec] border border-[#137fec]/30">Default</span>
                                                )}
                                            </div>
                                            <div className="text-xs text-[#92adc9]">
                                                {lang.language_code === 'en-US' ? 'English (United States)' : 
                                                 lang.language_code === 'es-ES' ? 'Spanish (Spain)' :
                                                 lang.language_code === 'fr-FR' ? 'French (France)' : 'German (Germany)'}
                                            </div>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-4">
                                        {!lang.is_default && (
                                            <button 
                                                onClick={() => handleSetDefault(lang.language_code)}
                                                className="text-xs text-[#92adc9] hover:text-white underline decoration-dotted"
                                            >
                                                Make Default
                                            </button>
                                        )}
                                        <label className="relative inline-flex items-center cursor-pointer">
                                            <input 
                                                type="checkbox" 
                                                checked={lang.is_active} 
                                                onChange={() => handleToggle(lang.language_code)}
                                                className="sr-only peer" 
                                            />
                                            <div className="w-9 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#10b981]"></div>
                                        </label>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </main>
    );
};

export default Localization;
