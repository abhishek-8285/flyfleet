'use client';

import Footer from '@/components/landing/Footer';
import Navbar from '@/components/landing/Navbar';
import TopInfoBar from '@/components/landing/TopInfoBar';
import { useState } from 'react';
import branding from '../branding.json';

import { useTheme } from '@/context/ThemeContext';

export default function TermsPage() {
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
                    <h1 className="text-4xl font-black mb-10 font-display">Terms of Service</h1>
                    <div className="prose prose-slate dark:prose-invert max-w-none space-y-8 text-slate-600 dark:text-text-secondary">
                        <section>
                            <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">1. Agreement to Terms</h2>
                            <p>
                                By accessing or using the {branding.companyName} platform, you agree to be bound by these terms of service and our privacy policy.
                            </p>
                        </section>
                        <section>
                            <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">2. License to Use</h2>
                            <p>
                                We grant you a limited, non-exclusive, non-transferable license to use our platform for your internal business operations.
                            </p>
                        </section>
                        <section>
                            <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">3. Service Availability (SLA)</h2>
                            <p>
                                Enterprise plans include a 99.9% uptime SLA as documented in our Service Level Agreement.
                            </p>
                        </section>
                        <section>
                            <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">4. Hardware Terms</h2>
                            <p>
                                All hardware remains property of {branding.companyName} unless purchased outright under the Enterprise hardware buyout plan.
                            </p>
                        </section>
                    </div>
                </div>
            </main>

            <Footer companyName={branding.companyName} />
        </div>
    );
}
