'use client';

import CTA from '@/components/landing/CTA';
import Footer from '@/components/landing/Footer';
import Navbar from '@/components/landing/Navbar';
import PricingComponent from '@/components/landing/Pricing';
import TopInfoBar from '@/components/landing/TopInfoBar';
import { useState } from 'react';
import branding from '../branding.json';

import { useTheme } from '@/context/ThemeContext';

export default function PricingPage() {
    const { theme, toggleTheme } = useTheme();
    const isDark = theme === 'dark';
    const [language, setLanguage] = useState('US (EN)');
    const [showLangDropdown, setShowLangDropdown] = useState(false);

    return (
        <div className="min-h-screen font-sans bg-white dark:bg-background-dark text-slate-900 dark:text-white">
            <TopInfoBar 
                phone={branding.contact.phone}
                language={language}
                onLanguageChange={setLanguage}
                showLangDropdown={showLangDropdown}
                setShowLangDropdown={setShowLangDropdown}
            />
            <Navbar 
                companyName={branding.companyName}
                isDark={isDark}
                toggleTheme={toggleTheme}
            />
            
            <main className="pt-32 pb-20">
                <div className="max-w-[1280px] mx-auto px-4 md:px-10 text-center mb-16">
                    <h1 className="text-4xl md:text-6xl font-black mb-6 font-display">Simple, <span className="text-primary italic">Scalable</span> Pricing</h1>
                    <p className="text-slate-600 dark:text-text-secondary text-lg max-w-2xl mx-auto">
                        Choose the plan that fits your fleet's needs. All plans include our core tracking engine and 24/7 reliability.
                    </p>
                </div>

                <PricingComponent />
                
                <section className="py-20 bg-slate-50 dark:bg-background-dark/50">
                    <div className="max-w-[1280px] mx-auto px-4 md:px-10">
                        <h2 className="text-3xl font-bold text-center mb-12 font-display">Feature Comparison</h2>
                        <div className="overflow-x-auto">
                            <table className="w-full border-collapse">
                                <thead className="border-b-2 border-slate-200 dark:border-border-dark">
                                    <tr>
                                        <th className="p-4 text-left font-bold">Feature</th>
                                        <th className="p-4 text-center">Starter</th>
                                        <th className="p-4 text-center">Pro</th>
                                        <th className="p-4 text-center">Enterprise</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-100 dark:divide-border-dark">
                                    {[
                                        ['Real-time GPS', true, true, true],
                                        ['Geofencing', '5 Zones', 'Unlimited', 'Unlimited'],
                                        ['History', '30 Days', '1 Year', 'Unlimited'],
                                        ['Driver Scorecards', false, true, true],
                                        ['Fuel Monitoring', false, true, true],
                                        ['Custom Webhooks', false, true, true],
                                        ['SAML/SSO', false, false, true],
                                        ['Dedicated Manager', false, false, true],
                                        ['On-site Training', false, false, true],
                                    ].map(([feature, s, p, e]) => (
                                        <tr key={feature as string} className="hover:bg-primary/5 transition-colors">
                                            <td className="p-4 text-sm font-medium">{feature}</td>
                                            <td className="p-4 text-center text-sm">{typeof s === 'boolean' ? (s ? '✓' : '—') : s}</td>
                                            <td className="p-4 text-center text-sm">{typeof p === 'boolean' ? (p ? '✓' : '—') : p}</td>
                                            <td className="p-4 text-center text-sm">{typeof e === 'boolean' ? (e ? '✓' : '—') : e}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </section>

                <CTA />
            </main>

            <Footer companyName={branding.companyName} />
        </div>
    );
}
