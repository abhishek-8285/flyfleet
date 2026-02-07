'use client';

import DemoRequest from '@/components/landing/DemoRequest';
import Footer from '@/components/landing/Footer';
import Navbar from '@/components/landing/Navbar';
import TopInfoBar from '@/components/landing/TopInfoBar';
import { useState } from 'react';
import branding from '../branding.json';

import { useTheme } from '@/context/ThemeContext';

export default function DemoPage() {
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
                    <h1 className="text-4xl md:text-6xl font-black mb-6 font-display">Experience the <span className="text-primary italic">Future</span></h1>
                    <p className="text-slate-600 dark:text-text-secondary text-lg max-w-2xl mx-auto">
                        See how our AI-driven fleet management platform can transform your operations in a personalized 30-minute walkthrough.
                    </p>
                </div>

                <DemoRequest />
            </main>

            <Footer companyName={branding.companyName} />
        </div>
    );
}
