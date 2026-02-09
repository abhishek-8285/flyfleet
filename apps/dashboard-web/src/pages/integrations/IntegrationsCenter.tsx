import {
    CommandLineIcon,
    GlobeAltIcon,
    KeyIcon,
    MagnifyingGlassIcon,
    PlusIcon,
    PuzzlePieceIcon,
    ShieldCheckIcon
} from '@heroicons/react/24/outline';
import { AnimatePresence, motion } from 'framer-motion';
import React, { useState } from 'react';
import ApiClientHub from './components/ApiClientHub';
import ConfigForm from './components/ConfigForm';
import IntegrationLogs from './components/IntegrationLogs';
import Marketplace from './components/Marketplace';
import WebhookManager from './components/WebhookManager';

const tabs = [
  { id: 'marketplace', name: 'Marketplace', icon: PuzzlePieceIcon },
  { id: 'apiclients', name: 'API Clients', icon: KeyIcon },
  { id: 'webhooks', name: 'Webhooks', icon: GlobeAltIcon },
  { id: 'logs', name: 'System Logs', icon: CommandLineIcon },
];

const IntegrationsCenter: React.FC = () => {
  const [activeTab, setActiveTab] = useState('marketplace');
  const [selectedIntegration, setSelectedIntegration] = useState<string | null>(null);

  const renderContent = () => {
    if (selectedIntegration) {
      return (
        <ConfigForm 
          id={selectedIntegration} 
          onBack={() => setSelectedIntegration(null)} 
        />
      );
    }

    switch (activeTab) {
      case 'marketplace':
        return <Marketplace onConfigure={(id) => setSelectedIntegration(id)} />;
      case 'apiclients':
        return <ApiClientHub />;
      case 'webhooks':
        return <WebhookManager />;
      case 'logs':
        return <IntegrationLogs />;
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-[#0b1219] text-white p-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
        <div>
          <div className="flex items-center gap-2 text-primary font-semibold text-xs uppercase tracking-widest mb-1">
            <ShieldCheckIcon className="w-4 h-4" />
            Advanced Fleet Security
          </div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
            Integrations & Access Control
          </h1>
        </div>
        
        <div className="flex items-center gap-3">
          <div className="relative">
            <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
            <input 
              type="text" 
              placeholder="Search integrations..."
              className="bg-slate-900/50 border border-slate-800 rounded-xl pl-10 pr-4 py-2 text-sm focus:ring-2 focus:ring-primary/50 focus:border-primary outline-none transition-all w-64"
            />
          </div>
          <button className="flex items-center gap-2 bg-primary hover:bg-primary-dark text-white px-4 py-2 rounded-xl text-sm font-semibold transition-all shadow-lg shadow-primary/20">
            <PlusIcon className="w-5 h-5" />
            Add Connector
          </button>
        </div>
      </div>

      {/* Navigation Tabs */}
      {!selectedIntegration && (
        <div className="flex items-center gap-2 bg-slate-900/30 p-1 rounded-2xl border border-slate-800/50 w-fit mb-8 backdrop-blur-sm">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-medium transition-all ${
                activeTab === tab.id 
                  ? 'bg-primary text-white shadow-lg shadow-primary/20' 
                  : 'text-slate-400 hover:text-white hover:bg-slate-800/50'
              }`}
            >
              <tab.icon className="w-4 h-4" />
              {tab.name}
            </button>
          ))}
        </div>
      )}

      {/* Main Content Area */}
      <AnimatePresence mode="wait">
        <motion.div
          key={selectedIntegration || activeTab}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -10 }}
          transition={{ duration: 0.2 }}
        >
          {renderContent()}
        </motion.div>
      </AnimatePresence>
    </div>
  );
};

export default IntegrationsCenter;
