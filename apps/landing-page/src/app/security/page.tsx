'use client';

import CTA from '@/components/landing/CTA';
import Footer from '@/components/landing/Footer';
import Navbar from '@/components/landing/Navbar';
import SecurityComponent from '@/components/landing/Security';
import TopInfoBar from '@/components/landing/TopInfoBar';
import { motion } from 'framer-motion';
import { useState } from 'react';
import branding from '../branding.json';

import { useTheme } from '@/context/ThemeContext';

export default function SecurityPage() {
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
                <div className="max-w-[1280px] mx-auto px-4 md:px-10 mb-20 text-center">
                    <motion.div 
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-green-500/10 border border-green-500/20 text-green-500 text-xs font-bold uppercase mb-6"
                    >
                      <span className="material-symbols-outlined !text-sm">verified</span> Verified Security Standards
                    </motion.div>
                    <h1 className="text-4xl md:text-6xl font-black mb-6 font-display">Trust & <span className="text-primary italic">Resilience</span></h1>
                    <p className="text-slate-600 dark:text-text-secondary text-lg max-w-2xl mx-auto">
                        Your fleet data is your most valuable asset. We protect it with banking-grade encryption, rigorous audits, and 99.9% uptime guarantees.
                    </p>
                </div>

                <SecurityComponent />

                <section className="py-24 bg-slate-50 dark:bg-background-dark/30">
                    <div className="max-w-[1280px] mx-auto px-4 md:px-10">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-16">
                            <div>
                                <h2 className="text-3xl font-bold mb-8 font-display">Data Residency & Privacy</h2>
                                <p className="text-slate-600 dark:text-text-secondary mb-6 leading-relaxed">
                                    We provide enterprise customers with choice over where their data lives. Our multi-region architecture supports data residency requirements in the US, EU, and Asia Pacific.
                                </p>
                                <ul className="space-y-4">
                                    {[
                                        'End-to-end AES-256 encryption at rest',
                                        'TLS 1.3 encryption in transit',
                                        'Zero-knowledge architecture available',
                                        'GDPR, CCPA, and HIPAA compliance ready'
                                    ].map(item => (
                                        <li key={item} className="flex gap-3 items-center font-bold text-slate-700 dark:text-slate-300">
                                            <span className="material-symbols-outlined text-primary">check_circle</span>
                                            {item}
                                        </li>
                                    ))}
                                </ul>
                            </div>
                            <div className="bg-white dark:bg-surface-dark p-8 rounded-3xl border border-slate-200 dark:border-border-dark shadow-xl">
                                <h3 className="text-xl font-bold mb-6 font-display">Responsible Disclosure</h3>
                                <p className="text-slate-500 text-sm mb-6">
                                    We run a private Bug Bounty program with leading security researchers to ensure our systems remain bulletproof.
                                </p>
                                <button className="w-full h-12 border border-primary text-primary font-bold rounded-xl hover:bg-primary/5 transition-colors">
                                    Report a Vulnerability
                                </button>
                                <div className="mt-8 pt-8 border-t border-slate-100 dark:border-border-dark grid grid-cols-3 gap-4 grayscale opacity-50">
                                     <div className="text-center font-black text-xs">SOC2</div>
                                     <div className="text-center font-black text-xs">ISO27001</div>
                                     <div className="text-center font-black text-xs">HIPAA</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                <CTA />
            </main>

            <Footer companyName={branding.companyName} />
        </div>
    );
}
