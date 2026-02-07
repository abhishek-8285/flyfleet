'use client';

import Footer from '@/components/landing/Footer';
import Navbar from '@/components/landing/Navbar';
import TopInfoBar from '@/components/landing/TopInfoBar';
import { useState } from 'react';
import branding from '../branding.json';

import { useTheme } from '@/context/ThemeContext';

export default function ContactSalesPage() {
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
                <div className="max-w-[1280px] mx-auto px-4 md:px-10 mb-20">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-16">
                        <div>
                             <h1 className="text-4xl md:text-6xl font-black mb-8 font-display">Let's Talk <span className="text-primary italic">Scale</span></h1>
                             <p className="text-slate-600 dark:text-text-secondary text-lg mb-12 leading-relaxed">
                                Our enterprise team works with the world's largest fleets to optimize operations, reduce costs, and improve driver safety. Reach out for a tailored consultation.
                             </p>
                             
                             <div className="space-y-8">
                                <div className="flex gap-6 items-start">
                                    <div className="size-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                                        <span className="material-symbols-outlined">headset_mic</span>
                                    </div>
                                    <div>
                                        <h3 className="font-bold text-lg mb-1">Global Support</h3>
                                        <p className="text-sm text-slate-500">24/7 dedicated support for enterprise accounts across all timezones.</p>
                                    </div>
                                </div>
                                <div className="flex gap-6 items-start">
                                    <div className="size-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                                        <span className="material-symbols-outlined">handshake</span>
                                    </div>
                                    <div>
                                        <h3 className="font-bold text-lg mb-1">Partnership Focused</h3>
                                        <p className="text-sm text-slate-500">We don't just provide software; we provide a strategic partnership to grow your business.</p>
                                    </div>
                                </div>
                             </div>
                        </div>
                        <div className="bg-slate-50 dark:bg-surface-dark p-8 md:p-12 rounded-3xl border border-slate-200 dark:border-border-dark shadow-xl">
                            <h2 className="text-2xl font-bold mb-8 font-display">Contact Sales</h2>
                            <form className="space-y-6">
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div className="space-y-2">
                                        <label className="text-sm font-bold opacity-70">First Name</label>
                                        <input type="text" className="w-full h-12 px-4 rounded-xl border border-slate-200 dark:border-border-dark bg-white dark:bg-background-dark outline-none focus:border-primary transition-colors" />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-sm font-bold opacity-70">Last Name</label>
                                        <input type="text" className="w-full h-12 px-4 rounded-xl border border-slate-200 dark:border-border-dark bg-white dark:bg-background-dark outline-none focus:border-primary transition-colors" />
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-bold opacity-70">Business Email</label>
                                    <input type="email" className="w-full h-12 px-4 rounded-xl border border-slate-200 dark:border-border-dark bg-white dark:bg-background-dark outline-none focus:border-primary transition-colors" />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-bold opacity-70">Company</label>
                                    <input type="text" className="w-full h-12 px-4 rounded-xl border border-slate-200 dark:border-border-dark bg-white dark:bg-background-dark outline-none focus:border-primary transition-colors" />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-bold opacity-70">Message</label>
                                    <textarea className="w-full h-32 px-4 py-3 rounded-xl border border-slate-200 dark:border-border-dark bg-white dark:bg-background-dark outline-none focus:border-primary transition-colors resize-none" placeholder="Tell us about your fleet..."></textarea>
                                </div>
                                <button className="w-full h-14 bg-primary text-white font-bold rounded-xl shadow-lg shadow-primary/20 hover:bg-primary/90 transition-all">
                                    Send Inquiry
                                </button>
                            </form>
                        </div>
                    </div>
                </div>
            </main>

            <Footer companyName={branding.companyName} />
        </div>
    );
}
