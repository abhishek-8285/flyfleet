'use client';

import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface MegaMenuProps {
  isOpen: boolean;
  onClose: () => void;
}

// 1. Hardware Data (mimics Image 1 Dropdown Options)
const HARDWARE_DEVICES = [
  { id: '2g', name: '2G Telematics Devices', icon: 'sensors', desc: 'Robust and economical telemetry node for tracking basic location, ignition triggers, and dry-contact logs.', speed: '2G cellular, fallback SMS' },
  { id: '4g', name: '4G Telematics Devices', icon: 'router', desc: 'Enterprise high-speed gateway with 5-second polling rates, CAN-bus parsing, driver behavior tracking, and wifi-hotspot options.', speed: '4G LTE-M / NB-IoT' },
  { id: 'fuel', name: 'Fuel Monitoring Devices', icon: 'water_drop', desc: 'Precision capacitive fuel rods linked to telemetry switch, detecting refuels, drainage, and fuel thefts instantly.', speed: '99.2% Accuracy, OTA Calibrated' },
  { id: 'lock', name: 'Digital Lock', icon: 'lock', desc: 'High-security smart Padlocks supporting offline TOTP, Bluetooth unlocking, and dual-server handshake audits.', speed: 'IP68 Waterproof, Cryptographic signing' },
  { id: 'balancer', name: 'Load Balancer', icon: 'scale', desc: 'Axle load sensors validating cargo load balance and preventing overloading penalties dynamically.', speed: 'Real-time telemetry streams' },
  { id: 'weighbridge', name: 'Weighbridge System', icon: 'analytics', desc: 'Automated integration with physical terminal weighbridge gates, cross-referencing cargo slips with digital waybills.', speed: 'Modbus / RS232 API Bridge' },
  { id: 'telematics_tech', name: 'Telematics Technology', icon: 'settings_input_antenna', desc: 'Multi-constellation GPS/GLONASS receivers supporting sub-meter positioning and dynamic geofencing constraints.', speed: 'Active GNSS Telemetry' },
  { id: 'fintech_tech', name: 'Fintech Technology', icon: 'payments', desc: 'Automated partner payouts, FASTag toll reconciliation switches, and penny-drop driver bank verification.', speed: 'T+0 Settled, ISO-8583 Compliant' },
  { id: 'business_auto', name: 'Business Automation', icon: 'precision_manufacturing', desc: 'Automated WhatsApp driver dispatches, WhatsApp lead recovery campaigns, and instant vehicle gate onboarding logs.', speed: '100% Event-Driven Webhooks' }
];

// 2. Core Products Data (mimics Image 2 Product Grid)
const CORE_PRODUCTS = [
  { title: 'Fleet Management System', desc: 'One stop solution for all your fleet management needs and streamline logistics operations.', icon: 'directions_car', color: 'from-blue-500/10 to-indigo-500/10 text-blue-500' },
  { title: 'Transport Management System', desc: 'Drive operational efficiency using Transporter Management System.', icon: 'local_shipping', color: 'from-cyan-500/10 to-blue-500/10 text-cyan-500' },
  { title: 'Transport ERP', desc: 'Explore the power of Cloud based digitisation and reduce costs.', icon: 'business', color: 'from-emerald-500/10 to-teal-500/10 text-emerald-500' },
  { title: 'Data Integration', desc: 'Scale systems with modular API integrations and connect with industry leading OEMs.', icon: 'api', color: 'from-amber-500/10 to-orange-500/10 text-amber-500' },
  { title: 'Realtime Visibility & Control Tower', desc: 'An end to end Vehicle Telematics tool designed to offer Realtime visibility at scale.', icon: 'visibility', color: 'from-purple-500/10 to-pink-500/10 text-purple-500' },
  { title: 'Video Telematics', desc: 'Harness the power of AI and IoT to drive safety and visibility into your Fleet Operations.', icon: 'videocam', color: 'from-rose-500/10 to-red-500/10 text-rose-500' },
  { title: 'Fuel Management System', desc: 'Scale systems with modular API integrations and connect with industry leading OEMs.', icon: 'oil_barrel', color: 'from-orange-500/10 to-yellow-500/10 text-orange-500' },
  { title: 'Fleet Maintenance', desc: 'Harness the power of AI to integrate systems that ensures a profitable fleet performance.', icon: 'build', color: 'from-indigo-500/10 to-purple-500/10 text-indigo-500' },
  { title: 'Logistics Management System', desc: 'Utilize the power of automation and analytics to ensure reliable operations.', icon: 'grid_view', color: 'from-teal-500/10 to-emerald-500/10 text-teal-500' }
];

// 3. Motive-style Left Columns Products (mimics Image 3 Products Menu)
const DIRECTORY_PRODUCTS = [
  { title: 'Driver Safety', desc: 'Prevent accidents and reduce risk with accurate AI.', icon: 'security' },
  { title: 'Fleet Management', desc: 'Get unparalleled visibility into your vehicles\' location, usage, and health.', icon: 'hub' },
  { title: 'Equipment Monitoring', desc: 'Gain comprehensive insights into your assets\' location, usage, and health.', icon: 'dashboard' },
  { title: 'Spend Management', desc: 'Unlock savings and productivity with the Motive fleet card.', icon: 'credit_card' },
  { title: 'Workforce Management', desc: 'Qualifications, time tracking, training, and coaching – all in one place.', icon: 'groups' },
  { title: 'AI Vision', desc: 'Customized AI solutions for your business\'s unique operational needs.', icon: 'psychology' },
  { title: 'Platform Overview', desc: 'Your operations. One platform. With AI-powered automation at its core.', icon: 'language' }
];

const DIRECTORY_FEATURED = [
  'AI Dashcam', 'Fleet Compliance', 'Fleet Telematics', 'GPS Fleet Tracking',
  'Maintenance', 'Sustainability', 'Reefer Monitoring', 'IFTA Fuel Tax Reporting',
  'Electronic Logging Device (ELD)', 'GPS Asset Tracking', 'Accident Management', 'Driver Coaching'
];

export const MegaMenu: React.FC<MegaMenuProps> = ({ isOpen, onClose }) => {
  const [activeTab, setActiveTab] = useState<'hardware' | 'products' | 'directory'>('products');
  const [selectedHardware, setSelectedHardware] = useState(HARDWARE_DEVICES[0]);
  const [hwDropdownOpen, setHwDropdownOpen] = useState(false);

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -10 }}
        transition={{ duration: 0.2 }}
        className="absolute top-16 left-0 right-0 w-full bg-white dark:bg-slate-900 border-b border-solid border-slate-200 dark:border-slate-800 shadow-2xl z-40 overflow-hidden backdrop-blur-xl"
        onMouseLeave={onClose}
      >
        <div className="max-w-[1440px] mx-auto grid grid-cols-12 min-h-[520px]">
          
          {/* LEFT COLUMN - SUB TAB NAVIGATION (Aesthetic Sidebar) */}
          <div className="col-span-3 border-r border-solid border-slate-200 dark:border-slate-800 p-6 bg-slate-50 dark:bg-slate-950/50 flex flex-col gap-2 justify-between">
            <div className="space-y-2">
              <h3 className="text-xs font-black uppercase text-slate-400 tracking-widest mb-4">Explore Ecosystem</h3>
              
              <button
                onClick={() => setActiveTab('products')}
                className={`w-full flex items-center gap-3 p-3 rounded-xl text-left transition-all ${
                  activeTab === 'products'
                    ? 'bg-primary text-white shadow-lg shadow-primary/20'
                    : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800/50'
                }`}
              >
                <span className="material-symbols-outlined">grid_view</span>
                <div>
                  <div className="font-bold text-sm">Core Products Matrix</div>
                  <div className={`text-[10px] ${activeTab === 'products' ? 'text-white/80' : 'text-slate-400'}`}>9 Core Modules</div>
                </div>
              </button>

              <button
                onClick={() => setActiveTab('directory')}
                className={`w-full flex items-center gap-3 p-3 rounded-xl text-left transition-all ${
                  activeTab === 'directory'
                    ? 'bg-primary text-white shadow-lg shadow-primary/20'
                    : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800/50'
                }`}
              >
                <span className="material-symbols-outlined">menu_open</span>
                <div>
                  <div className="font-bold text-sm">Motive-Style Directory</div>
                  <div className={`text-[10px] ${activeTab === 'directory' ? 'text-white/80' : 'text-slate-400'}`}>B2B Mega Catalog</div>
                </div>
              </button>

              <button
                onClick={() => setActiveTab('hardware')}
                className={`w-full flex items-center gap-3 p-3 rounded-xl text-left transition-all ${
                  activeTab === 'hardware'
                    ? 'bg-primary text-white shadow-lg shadow-primary/20'
                    : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800/50'
                }`}
              >
                <span className="material-symbols-outlined">developer_board</span>
                <div>
                  <div className="font-bold text-sm">Hardware & Telematics</div>
                  <div className={`text-[10px] ${activeTab === 'hardware' ? 'text-white/80' : 'text-slate-400'}`}>9 IoT Devices</div>
                </div>
              </button>
            </div>

            <div className="p-4 rounded-xl bg-primary/5 border border-primary/10 text-center">
              <span className="material-symbols-outlined text-primary !text-3xl mb-2">support_agent</span>
              <div className="font-bold text-xs text-slate-800 dark:text-white">Need custom IoT?</div>
              <p className="text-[10px] text-slate-500 mt-1">Our hardware team designs specialized telematics modules.</p>
              <button className="text-[10px] text-primary font-bold hover:underline mt-2">Speak to Architect &rarr;</button>
            </div>
          </div>

          {/* MAIN CONTAINER PANEL */}
          <div className="col-span-9 p-8 bg-white dark:bg-slate-900 overflow-y-auto max-h-[600px]">
            
            {/* TAB 1: CORE PRODUCTS MATRIX (Image 2) */}
            {activeTab === 'products' && (
              <motion.div
                initial={{ opacity: 0, x: 10 }}
                animate={{ opacity: 1, x: 0 }}
                className="space-y-6"
              >
                <div className="flex justify-between items-center border-b border-solid border-slate-100 dark:border-slate-800 pb-4">
                  <div>
                    <h2 className="text-xl font-black text-slate-900 dark:text-white font-display">Logistics Products Suite</h2>
                    <p className="text-xs text-slate-500">Premium cloud-based transport and vehicle control tower modules.</p>
                  </div>
                  <span className="text-[10px] bg-emerald-500/10 text-emerald-500 px-2 py-1 rounded-full font-bold uppercase tracking-widest">SaaS Enabled</span>
                </div>

                <div className="grid grid-cols-3 gap-6">
                  {CORE_PRODUCTS.map((prod, idx) => (
                    <motion.div
                      key={idx}
                      whileHover={{ y: -3, scale: 1.01 }}
                      className="p-5 rounded-xl border border-solid border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/20 hover:border-primary/40 hover:shadow-xl transition-all cursor-pointer group"
                    >
                      <div className={`w-10 h-10 rounded-lg flex items-center justify-center bg-gradient-to-br ${prod.color} mb-3 group-hover:scale-110 transition-transform`}>
                        <span className="material-symbols-outlined !text-xl">{prod.icon}</span>
                      </div>
                      <h4 className="font-bold text-sm text-slate-900 dark:text-white group-hover:text-primary transition-colors">{prod.title}</h4>
                      <p className="text-slate-500 text-[11px] mt-1 leading-relaxed">{prod.desc}</p>
                    </motion.div>
                  ))}
                </div>
              </motion.div>
            )}

            {/* TAB 2: MOTIVE-STYLE DIRECTORY (Image 3) */}
            {activeTab === 'directory' && (
              <motion.div
                initial={{ opacity: 0, x: 10 }}
                animate={{ opacity: 1, x: 0 }}
                className="grid grid-cols-12 gap-8"
              >
                {/* Products Directory Columns */}
                <div className="col-span-5 space-y-4">
                  <h3 className="text-xs font-black uppercase text-slate-400 tracking-wider border-b border-solid border-slate-100 dark:border-slate-800 pb-2">Business Modules</h3>
                  <div className="space-y-1">
                    {DIRECTORY_PRODUCTS.map((item, idx) => (
                      <div
                        key={idx}
                        className="flex gap-3 p-2 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors cursor-pointer group"
                      >
                        <span className="material-symbols-outlined text-slate-400 group-hover:text-primary transition-colors !text-xl mt-0.5">{item.icon}</span>
                        <div>
                          <h4 className="font-bold text-xs text-slate-800 dark:text-white group-hover:text-primary transition-colors">{item.title}</h4>
                          <p className="text-[10px] text-slate-400 mt-0.5 leading-snug">{item.desc}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Featured Checklist column */}
                <div className="col-span-3 space-y-4">
                  <h3 className="text-xs font-black uppercase text-slate-400 tracking-wider border-b border-solid border-slate-100 dark:border-slate-800 pb-2">Featured Tech</h3>
                  <div className="grid grid-cols-1 gap-2.5">
                    {DIRECTORY_FEATURED.map((item, idx) => (
                      <div
                        key={idx}
                        className="flex items-center gap-2 text-[11px] text-slate-600 dark:text-slate-400 hover:text-primary cursor-pointer transition-colors"
                      >
                        <span className="w-1.5 h-1.5 rounded-full bg-primary/60" />
                        <span>{item}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Featured Brochure Magazine (Image 3 right pane) */}
                <div className="col-span-4 space-y-4">
                  <h3 className="text-xs font-black uppercase text-slate-400 tracking-wider border-b border-solid border-slate-100 dark:border-slate-800 pb-2">Special Report</h3>
                  
                  <div className="relative rounded-2xl bg-gradient-to-br from-slate-900 to-slate-950 p-4 border border-solid border-slate-800/80 shadow-2xl overflow-hidden group">
                    {/* Glowing effect */}
                    <div className="absolute top-0 right-0 w-32 h-32 bg-primary/20 rounded-full blur-3xl pointer-events-none group-hover:scale-150 transition-transform duration-500" />
                    
                    {/* Simulated visual brochure */}
                    <div className="w-full aspect-[4/3] rounded-xl overflow-hidden relative mb-4 border border-solid border-slate-800 bg-slate-900 flex items-center justify-center">
                      <img 
                        src="/home/abhishek/.gemini/antigravity/brain/0893f9de-07b9-492a-b369-587c50d1869f/ai_accuracy_brochure_1779782769732.png"
                        alt="AI Safety Brochure Booklet" 
                        className="w-full h-full object-cover object-center group-hover:scale-105 transition-transform duration-500"
                      />
                    </div>

                    <h4 className="font-bold text-xs text-white leading-snug">Why AI accuracy and benchmarking matter in safety-critical environments</h4>
                    <p className="text-[9px] text-slate-400 mt-1 leading-snug">Explore accurate AI\'s impact on fleet safety and operations, and see the measurable results customers achieved.</p>
                    
                    <button className="flex items-center gap-1.5 text-[10px] font-bold text-primary hover:underline mt-4 group-hover:translate-x-1 transition-transform">
                      <span>View All Products</span>
                      <span className="material-symbols-outlined !text-xs">arrow_forward</span>
                    </button>
                  </div>
                </div>
              </motion.div>
            )}

            {/* TAB 3: IoT HARDWARE SELECTOR (Image 1) */}
            {activeTab === 'hardware' && (
              <motion.div
                initial={{ opacity: 0, x: 10 }}
                animate={{ opacity: 1, x: 0 }}
                className="space-y-6"
              >
                <div className="flex justify-between items-center border-b border-solid border-slate-100 dark:border-slate-800 pb-4">
                  <div>
                    <h2 className="text-xl font-black text-slate-900 dark:text-white font-display">IoT Telematics & Hardware Matrix</h2>
                    <p className="text-xs text-slate-500">Over-the-air rolling key authentication & secure physical terminal integrations.</p>
                  </div>
                  <span className="text-[10px] bg-primary/10 text-primary px-2.5 py-1 rounded-full font-bold uppercase tracking-wider">IoT Certified</span>
                </div>

                <div className="grid grid-cols-12 gap-8">
                  {/* Selector listbox panel (exact matching Image 1 list dropdown) */}
                  <div className="col-span-5 space-y-3">
                    <h4 className="text-[10px] font-black uppercase text-slate-400 tracking-wider">Device Hardware Selector</h4>
                    
                    {/* Simulated dropdown element */}
                    <div className="relative">
                      <button
                        onClick={() => setHwDropdownOpen(!hwDropdownOpen)}
                        className="w-full flex items-center justify-between p-3 rounded-xl border border-solid border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-850 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors font-bold text-xs text-slate-900 dark:text-white text-left"
                      >
                        <div className="flex items-center gap-2">
                          <span className="material-symbols-outlined text-primary !text-lg">{selectedHardware.icon}</span>
                          <span>{selectedHardware.name}</span>
                        </div>
                        <span className="material-symbols-outlined !text-sm">{hwDropdownOpen ? 'expand_less' : 'expand_more'}</span>
                      </button>

                      {/* Dropdown Options matching Image 1 exactly */}
                      {hwDropdownOpen && (
                        <div className="absolute top-12 left-0 right-0 w-full bg-white dark:bg-slate-900 border border-solid border-slate-200 dark:border-slate-800 rounded-xl shadow-2xl z-50 overflow-hidden divide-y divide-solid divide-slate-100 dark:divide-slate-800">
                          <div className="p-2 text-[10px] text-slate-400 font-bold bg-slate-50/50 dark:bg-slate-950/20">----------Select option----------</div>
                          {HARDWARE_DEVICES.map((item) => (
                            <button
                              key={item.id}
                              onClick={() => {
                                setSelectedHardware(item);
                                setHwDropdownOpen(false);
                              }}
                              className={`w-full flex items-center gap-2.5 p-3 text-left text-xs transition-colors hover:bg-primary/5 hover:text-primary ${
                                selectedHardware.id === item.id 
                                  ? 'bg-primary text-white hover:bg-primary/95 hover:text-white font-bold' 
                                  : 'text-slate-700 dark:text-slate-300'
                              }`}
                            >
                              <span className="material-symbols-outlined !text-sm">{item.icon}</span>
                              <span>{item.name}</span>
                            </button>
                          ))}
                        </div>
                      )}
                    </div>

                    <div className="p-3.5 rounded-xl border border-solid border-slate-100 dark:border-slate-800 bg-slate-50/30 dark:bg-slate-950/20 text-slate-500 text-[10px] leading-relaxed">
                      * Supported firmware modules automatically compile secure rolling master/session cryptograms signed under HSM-derived hardware credentials OTA.
                    </div>
                  </div>

                  {/* Active selected hardware detail card */}
                  <div className="col-span-7">
                    <motion.div
                      key={selectedHardware.id}
                      initial={{ opacity: 0, y: 5 }}
                      animate={{ opacity: 1, y: 0 }}
                      className="h-full p-6 rounded-2xl border border-solid border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-800/10 flex flex-col justify-between"
                    >
                      <div className="space-y-4">
                        <div className="flex items-center gap-3">
                          <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
                            <span className="material-symbols-outlined !text-2xl">{selectedHardware.icon}</span>
                          </div>
                          <div>
                            <h3 className="font-bold text-base text-slate-900 dark:text-white">{selectedHardware.name}</h3>
                            <span className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">{selectedHardware.speed}</span>
                          </div>
                        </div>

                        <p className="text-slate-600 dark:text-slate-400 text-xs leading-relaxed">{selectedHardware.desc}</p>
                      </div>

                      <div className="pt-6 border-t border-solid border-slate-100 dark:border-slate-800 flex justify-between items-center mt-6">
                        <div className="flex gap-4">
                          <div>
                            <div className="text-[9px] uppercase text-slate-400 font-bold">Certification</div>
                            <div className="text-xs font-bold text-emerald-500">FMCSA Compliant</div>
                          </div>
                          <div>
                            <div className="text-[9px] uppercase text-slate-400 font-bold">Encryption</div>
                            <div className="text-xs font-bold text-slate-800 dark:text-white">DUKPT / TLE</div>
                          </div>
                        </div>

                        <button className="flex items-center gap-1 bg-primary hover:bg-primary/95 text-white font-bold text-xs px-3 py-1.5 rounded-lg transition-colors">
                          <span>Request Spec Sheet</span>
                          <span className="material-symbols-outlined !text-sm">download</span>
                        </button>
                      </div>
                    </motion.div>
                  </div>
                </div>
              </motion.div>
            )}

          </div>

        </div>
      </motion.div>
    </AnimatePresence>
  );
};
