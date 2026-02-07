'use client';

import Footer from '@/components/landing/Footer';
import Navbar from '@/components/landing/Navbar';
import TopInfoBar from '@/components/landing/TopInfoBar';
import { useState } from 'react';
import branding from '../branding.json';

import { useTheme } from '@/context/ThemeContext';

export default function PrivacyPage() {
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
            
            <main className="pt-40 pb-20">
                <div className="max-w-4xl mx-auto px-4 md:px-10">
                    <h1 className="text-4xl font-black mb-10 font-display">Privacy Policy</h1>
                    <div className="prose prose-slate dark:prose-invert max-w-none space-y-8 text-slate-600 dark:text-text-secondary">
                        <section>
                            <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">1. Introduction</h2>
                            <p>
                                At {branding.companyName}, we take your privacy seriously. This policy outlines how we collect, use, and protect your data across our fleet management platform and mobile applications.
                            </p>
                        </section>
                        <section>
                            <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">2. Data We Collect</h2>
                            <p>
                                We collect vehicle telemetry, driver behavior data, GPS coordinates, and administrative information required to provide our services.
                            </p>
                            <ul className="list-disc pl-6 space-y-2 mt-4">
                                <li>Real-time Location Data</li>
                                <li>Vehicle Health Diagnostics</li>
                                <li>Driver Identification & Performance</li>
                                <li>Account & Billing Information</li>
                            </ul>
                        </section>
                        <section>
                            <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">3. Data Usage</h2>
                            <p>
                                Your data is used exclusively to optimize your fleet operations, provide safety alerts, and improve our AI routing algorithms. We never sell your data to third-party advertisers.
                            </p>
                        </section>
                        <section>
                            <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">4. Compliance & Security</h2>
                            <p>
                                We are GDPR compliant and ISO 27001 certified. All data is encrypted using AES-256 at rest and TLS 1.3 in transit.
                            </p>
                        </section>
                    </div>
                </div>
            </main>

            <Footer companyName={branding.companyName} />
        </div>
    );
}
