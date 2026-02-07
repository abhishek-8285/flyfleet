'use client';

import { useState } from 'react';
import branding from './branding.json';

// Components
import AppShowcase from '@/components/landing/AppShowcase';
import ChatButton from '@/components/landing/ChatButton';
import CTA from '@/components/landing/CTA';
import DemoRequest from '@/components/landing/DemoRequest';
import FAQ from '@/components/landing/FAQ';
import Features from '@/components/landing/Features';
import Footer from '@/components/landing/Footer';
import Hero from '@/components/landing/Hero';
import ImplementationRoadmap from '@/components/landing/ImplementationRoadmap';
import Integrations from '@/components/landing/Integrations';
import InteractiveIntelligence from '@/components/landing/InteractiveIntelligence';
import Navbar from '@/components/landing/Navbar';
import Pricing from '@/components/landing/Pricing';
import ROICalculator from '@/components/landing/ROICalculator';
import Security from '@/components/landing/Security';
import SuccessStories from '@/components/landing/SuccessStories';
import TopInfoBar from '@/components/landing/TopInfoBar';
import TrustedBy from '@/components/landing/TrustedBy';

import { useTheme } from '@/context/ThemeContext';

export default function LandingPage() {
    const { theme, toggleTheme } = useTheme();
    const isDark = theme === 'dark';
    const [language, setLanguage] = useState('US (EN)');
    const [showLangDropdown, setShowLangDropdown] = useState(false);

    const handleDemoRequest = () => {
        document.getElementById('contact')?.scrollIntoView({ behavior: 'smooth' });
    };

    const handleTourRequest = () => {
        window.open('https://youtube.com', '_blank'); // Placeholder
    };

    return (
        <div className="min-h-screen font-sans bg-white dark:bg-background-dark text-slate-900 dark:text-white overflow-x-hidden selection:bg-primary selection:text-white">
            

            <Navbar 
                companyName={branding.companyName}
                isDark={isDark}
                toggleTheme={toggleTheme}
                language={language}
            />

            <main className="flex min-h-screen w-full flex-col pt-24">
                <Hero 
                    onDemoRequest={handleDemoRequest} 
                    onTourRequest={handleTourRequest} 
                    language={language}
                />
                <TrustedBy />
                <ROICalculator companyName={branding.companyName} />
                <InteractiveIntelligence />
                <SuccessStories />
                <AppShowcase />
                <Features />
                <Integrations />
                <Security />
                <ImplementationRoadmap />
                <Pricing />
                <DemoRequest />
                <FAQ />
                <CTA />
            </main>

            <Footer companyName={branding.companyName} />
            <ChatButton />
            <TopInfoBar 
                phone={branding.contact.phone}
                language={language}
                onLanguageChange={setLanguage}
                showLangDropdown={showLangDropdown}
                setShowLangDropdown={setShowLangDropdown}
            />
        </div>
    );
}
