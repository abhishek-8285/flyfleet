import branding from '@flyfleet/config';
import { motion } from 'framer-motion';
import React, { useState } from 'react';
import Modal from '../common/Modal';

interface EnterpriseAuthLayoutProps {
  children: React.ReactNode;
  title: string;
  subtitle?: string;
}

const EnterpriseAuthLayout: React.FC<EnterpriseAuthLayoutProps> = ({ children, title, subtitle }) => {
  const [activeModal, setActiveModal] = useState<'privacy' | 'terms' | 'support' | null>(null);

  return (
    <>
      <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4 selection:bg-primary/30">
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/10 rounded-full blur-[120px] animate-pulse"></div>
          <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-600/10 rounded-full blur-[120px] animate-pulse" style={{ animationDelay: '2s' }}></div>
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: "easeOut" }}
          className="max-w-md w-full relative z-10"
        >
          <div className="bg-slate-900/40 backdrop-blur-2xl border border-white/5 p-6 rounded-3xl shadow-[0_20px_50px_rgba(0,0,0,0.3)] border-t-white/10">
            <div className="text-center mb-6">
              <div className="flex justify-center mb-4">
                <div className="h-12 w-12 bg-gradient-to-br from-primary to-blue-600 rounded-2xl flex items-center justify-center shadow-lg shadow-primary/20 rotate-3 hover:rotate-0 transition-transform duration-300">
                  <span className="material-symbols-outlined text-white text-2xl">local_shipping</span>
                </div>
              </div>
              <h1 className="text-2xl font-bold text-white tracking-tight mb-1">{branding.companyName}</h1>
              <div className="h-1 w-10 bg-primary mx-auto rounded-full mb-3"></div>
              <h2 className="text-lg font-semibold text-slate-200 mb-1">{title}</h2>
              {subtitle && <p className="text-slate-400 text-sm font-medium">{subtitle}</p>}
            </div>

            {children}

            <div className="mt-6 pt-4 border-t border-slate-800/50 text-center">
              <p className="text-slate-500 text-xs tracking-wide uppercase font-semibold mb-3">
                Secure Enterprise Gateway
              </p>
              <div className="flex justify-center gap-4 text-xs text-slate-500">
                <button
                  className="hover:text-primary transition-colors"
                  onClick={() => setActiveModal('privacy')}
                >
                  Privacy
                </button>
                <span className="text-slate-800">•</span>
                <button
                  className="hover:text-primary transition-colors"
                  onClick={() => setActiveModal('terms')}
                >
                  Terms
                </button>
                <span className="text-slate-800">•</span>
                <button
                  className="hover:text-primary transition-colors"
                  onClick={() => setActiveModal('support')}
                >
                  Support
                </button>
              </div>
            </div>
          </div>

          <p className="text-center mt-6 text-slate-600 text-[10px] uppercase tracking-[0.2em] font-bold">
            Powered by FleetCommand Core v4.0
          </p>
        </motion.div>
      </div>

      {/* Privacy Modal */}
      <Modal
        isOpen={activeModal === 'privacy'}
        onClose={() => setActiveModal(null)}
        title="Privacy Policy"
        size="lg"
      >
        <div className="prose prose-invert max-w-none space-y-6">
          <div className="bg-slate-800/20 border border-white/5 rounded-2xl p-5">
            <h4 className="text-white mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">policy</span>
              Last Updated: January 15, 2026
            </h4>
            <p className="text-slate-400 text-sm">
              This Privacy Policy describes how FlyFleet Global Operations collects, uses, and protects your personal information.
            </p>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">Information We Collect</h4>

            <div className="space-y-4">
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Account Information</h5>
                <ul className="text-slate-400 space-y-1 text-sm ml-4">
                  <li className="list-disc">Full name and job title</li>
                  <li className="list-disc">Corporate email address</li>
                  <li className="list-disc">Phone number and contact details</li>
                  <li className="list-disc">Company name and organization details</li>
                  <li className="list-disc">Role and department information</li>
                </ul>
              </div>

              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Usage Data</h5>
                <ul className="text-slate-400 space-y-1 text-sm ml-4">
                  <li className="list-disc">Login times and session duration</li>
                  <li className="list-disc">Pages visited and features used</li>
                  <li className="list-disc">API calls and data requests</li>
                  <li className="list-disc">Device information and IP addresses</li>
                  <li className="list-disc">Performance metrics and error logs</li>
                </ul>
              </div>

              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Fleet Data</h5>
                <ul className="text-slate-400 space-y-1 text-sm ml-4">
                  <li className="list-disc">Vehicle identification numbers (VINs)</li>
                  <li className="list-disc">GPS location and tracking data</li>
                  <li className="list-disc">Maintenance records and service history</li>
                  <li className="list-disc">Fuel consumption and efficiency metrics</li>
                  <li className="list-disc">Driver behavior and safety data</li>
                </ul>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">How We Use Your Information</h4>
            <div className="space-y-3 text-slate-400">
              <p><span className="font-semibold text-slate-300">Service Provision:</span> To operate and maintain our fleet management platform, including vehicle tracking, maintenance scheduling, and reporting.</p>
              <p><span className="font-semibold text-slate-300">Communication:</span> To send you important updates, security alerts, product announcements, and customer support responses.</p>
              <p><span className="font-semibold text-slate-300">Analytics:</span> To understand usage patterns, improve our services, and develop new features based on customer needs.</p>
              <p><span className="font-semibold text-slate-300">Compliance:</span> To meet legal requirements, industry regulations, and audit obligations.</p>
              <p><span className="font-semibold text-slate-300">Security:</span> To protect against unauthorized access, fraud, and other security threats.</p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">Data Protection & Security</h4>
            <div className="space-y-3 text-slate-400">
              <p>We implement robust security measures to protect your data:</p>
              <ul className="space-y-2 ml-4">
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-green-500 mt-0.5 text-sm">lock</span>
                  <span><span className="font-semibold">Encryption:</span> All data is encrypted using AES-256 encryption both in transit and at rest.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-green-500 mt-0.5 text-sm">security</span>
                  <span><span className="font-semibold">Access Controls:</span> Role-based access controls and multi-factor authentication for all accounts.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-green-500 mt-0.5 text-sm">cloud_done</span>
                  <span><span className="font-semibold">Infrastructure:</span> Secure cloud infrastructure with regular security audits and penetration testing.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-green-500 mt-0.5 text-sm">backup</span>
                  <span><span className="font-semibold">Backups:</span> Automated daily backups with 90-day retention and disaster recovery capabilities.</span>
                </li>
              </ul>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">Your Rights & Choices</h4>
            <div className="space-y-3 text-slate-400">
              <p>You have the following rights regarding your personal information:</p>
              <ul className="space-y-2 ml-4">
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">visibility</span>
                  <span><span className="font-semibold">Access:</span> Request a copy of the personal information we hold about you.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">edit</span>
                  <span><span className="font-semibold">Correction:</span> Request correction of inaccurate or incomplete personal information.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">delete</span>
                  <span><span className="font-semibold">Deletion:</span> Request deletion of your personal information (subject to legal requirements).</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">block</span>
                  <span><span className="font-semibold">Objection:</span> Object to processing of your personal information for marketing purposes.</span>
                </li>
              </ul>
              <p className="mt-3">To exercise these rights, contact our Data Protection Officer at <a href="mailto:dpo@flyfleet.com" className="text-primary hover:underline">dpo@flyfleet.com</a></p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">Data Retention</h4>
            <div className="space-y-3 text-slate-400">
              <p>We retain your information for as long as necessary to provide our services and comply with legal obligations:</p>
              <ul className="space-y-1 ml-4 text-sm">
                <li className="list-disc">Account data: While your account is active, plus 7 years for legal compliance</li>
                <li className="list-disc">Usage logs: 2 years for analytics and troubleshooting</li>
                <li className="list-disc">Financial records: 7 years as required by tax laws</li>
                <li className="list-disc">Audit logs: 5 years for security and compliance purposes</li>
              </ul>
            </div>
          </div>

          <div className="bg-amber-500/10 border border-amber-500/30 rounded-2xl p-4">
            <h4 className="text-amber-500 mb-2 flex items-center gap-2">
              <span className="material-symbols-outlined">info</span>
              International Data Transfers
            </h4>
            <p className="text-amber-200 text-sm">
              Your data may be transferred to and processed in countries outside your residence. We ensure adequate protection through standard contractual clauses and compliance with GDPR requirements.
            </p>
          </div>
        </div>
      </Modal>

      {/* Terms Modal */}
      <Modal
        isOpen={activeModal === 'terms'}
        onClose={() => setActiveModal(null)}
        title="Terms of Service"
        size="lg"
      >
        <div className="prose prose-invert max-w-none space-y-6">
          <div className="bg-slate-800/20 border border-white/5 rounded-2xl p-5">
            <h4 className="text-white mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">description</span>
              Effective Date: January 1, 2026
            </h4>
            <p className="text-slate-400 text-sm">
              These Terms of Service govern your access to and use of FlyFleet's fleet management platform and services.
            </p>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">1. Acceptance of Terms</h4>
            <div className="space-y-3 text-slate-400">
              <p>By accessing, browsing, or using the FlyFleet platform ("Service"), you acknowledge that you have read, understood, and agree to be bound by these Terms of Service ("Terms"). If you are using the Service on behalf of an organization, you represent and warrant that you have the authority to bind that organization to these Terms.</p>
              <p className="bg-red-500/10 border border-red-500/30 rounded-lg p-3">
                <span className="font-semibold text-red-300">Important:</span> If you do not agree to these Terms, you must not access or use the Service.
              </p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">2. Service Description</h4>
            <div className="space-y-3 text-slate-400">
              <p>FlyFleet provides a comprehensive fleet management platform that includes:</p>
              <ul className="space-y-2 ml-4">
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">directions_car</span>
                  <span>Real-time vehicle tracking and GPS monitoring</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">build</span>
                  <span>Maintenance scheduling and service management</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">analytics</span>
                  <span>Fuel management and cost analytics</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">people</span>
                  <span>Driver management and performance tracking</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">report</span>
                  <span>Comprehensive reporting and compliance tools</span>
                </li>
              </ul>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">3. Account Registration & Security</h4>
            <div className="space-y-4 text-slate-400">
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Registration Requirements</h5>
                <ul className="space-y-1 ml-4 text-sm">
                  <li className="list-disc">Provide accurate and complete registration information</li>
                  <li className="list-disc">Maintain and update your information regularly</li>
                  <li className="list-disc">Use a valid corporate email address</li>
                  <li className="list-disc">Be at least 18 years old or have parental consent</li>
                </ul>
              </div>

              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Security Responsibilities</h5>
                <ul className="space-y-1 ml-4 text-sm">
                  <li className="list-disc">Maintain the confidentiality of your account credentials</li>
                  <li className="list-disc">Enable two-factor authentication when available</li>
                  <li className="list-disc">Notify us immediately of any unauthorized access</li>
                  <li className="list-disc">Log out of your account after each session</li>
                </ul>
              </div>

              <div className="bg-amber-500/10 border border-amber-500/30 rounded-xl p-3">
                <p className="text-amber-200 text-sm">
                  <span className="font-semibold">Note:</span> You are solely responsible for all activities that occur under your account, whether authorized or not.
                </p>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">4. Acceptable Use Policy</h4>
            <div className="space-y-3 text-slate-400">
              <p>You agree to use the Service only for lawful purposes and in accordance with these Terms. You shall not:</p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-3">
                <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4">
                  <h5 className="text-red-300 font-semibold mb-2">Prohibited Activities</h5>
                  <ul className="space-y-1 text-sm">
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Violate any applicable laws or regulations</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Infringe intellectual property rights</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Reverse engineer or hack the Service</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Use the Service for illegal activities</span>
                    </li>
                  </ul>
                </div>

                <div className="bg-green-500/10 border border-green-500/30 rounded-xl p-4">
                  <h5 className="text-green-300 font-semibold mb-2">Permitted Uses</h5>
                  <ul className="space-y-1 text-sm">
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Manage your organization's fleet</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Generate legitimate business reports</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Track vehicles for operational purposes</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Monitor driver performance metrics</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">5. Data Ownership & Rights</h4>
            <div className="space-y-3 text-slate-400">
              <p><span className="font-semibold text-slate-300">Your Data:</span> You retain all rights, title, and interest in your data uploaded to the Service. We obtain no rights to your data except as necessary to provide the Service.</p>
              <p><span className="font-semibold text-slate-300">Our Rights:</span> FlyFleet retains all rights, title, and interest in and to the Service, including all software, technology, and intellectual property.</p>
              <p><span className="font-semibold text-slate-300">Feedback:</span> If you provide feedback or suggestions about the Service, we may use that feedback without obligation to you.</p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">6. Subscription Plans & Payment</h4>
            <div className="space-y-3 text-slate-400">
              <div className="bg-slate-800/30 border border-white/10 rounded-xl p-4">
                <h5 className="text-slate-200 font-semibold mb-3">Plan Details</h5>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">
                  <div>
                    <h6 className="font-medium text-white mb-1">Starter Plan</h6>
                    <p className="text-slate-400">Up to 10 vehicles • $49/month</p>
                  </div>
                  <div>
                    <h6 className="font-medium text-white mb-1">Professional Plan</h6>
                    <p className="text-slate-400">Up to 100 vehicles • $199/month</p>
                  </div>
                  <div>
                    <h6 className="font-medium text-white mb-1">Enterprise Plan</h6>
                    <p className="text-slate-400">Unlimited vehicles • Custom pricing</p>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <p><span className="font-semibold text-slate-300">Payment Terms:</span> Fees are billed in advance on a monthly or annual basis. All fees are non-refundable.</p>
                <p><span className="font-semibold text-slate-300">Auto-Renewal:</span> Subscriptions automatically renew unless canceled 30 days prior to renewal.</p>
                <p><span className="font-semibold text-slate-300">Taxes:</span> You are responsible for all applicable taxes, duties, and government charges.</p>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">7. Termination</h4>
            <div className="space-y-3 text-slate-400">
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">By You</h5>
                <p>You may terminate your account at any time by contacting support. Upon termination, you will lose access to the Service and your data may be deleted according to our retention policy.</p>
              </div>

              <div>
                <h5 className="text-slate-200 font-semibold mb-2">By Us</h5>
                <p>We may suspend or terminate your account immediately without notice if you:</p>
                <ul className="space-y-1 ml-4 mt-2 text-sm">
                  <li className="list-disc">Violate these Terms</li>
                  <li className="list-disc">Fail to pay fees</li>
                  <li className="list-disc">Engage in fraudulent activity</li>
                  <li className="list-disc">Pose a security risk</li>
                </ul>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">8. Disclaimer of Warranties</h4>
            <div className="space-y-3 text-slate-400">
              <p className="bg-red-500/10 border border-red-500/30 rounded-xl p-4">
                THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED. FLYFLEET DISCLAIMS ALL WARRANTIES, INCLUDING MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.
              </p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">9. Limitation of Liability</h4>
            <div className="space-y-3 text-slate-400">
              <p>TO THE MAXIMUM EXTENT PERMITTED BY LAW, FLYFLEET SHALL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, OR CONSEQUENTIAL DAMAGES RESULTING FROM YOUR USE OF THE SERVICE.</p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">10. Governing Law</h4>
            <div className="space-y-3 text-slate-400">
              <p>These Terms shall be governed by and construed in accordance with the laws of the State of Delaware, without regard to its conflict of law principles.</p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">11. Changes to Terms</h4>
            <div className="space-y-3 text-slate-400">
              <p>We reserve the right to modify these Terms at any time. Changes will be effective immediately upon posting. Your continued use of the Service constitutes acceptance of any modifications.</p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">12. Contact Information</h4>
            <div className="space-y-3 text-slate-400">
              <p>For questions about these Terms, please contact us at:</p>
              <div className="bg-slate-800/30 border border-white/10 rounded-xl p-4">
                <ul className="space-y-2 text-sm">
                  <li><span className="font-semibold">Email:</span> legal@flyfleet.com</li>
                  <li><span className="font-semibold">Address:</span> 1234 Fleet Street, Wilmington, DE 19801</li>
                  <li><span className="font-semibold">Phone:</span> +1 (555) 123-4567</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </Modal>

      {/* Support Modal */}
      <Modal
        isOpen={activeModal === 'support'}
        onClose={() => setActiveModal(null)}
        title="Support Center"
        size="md"
      >
        <div className="space-y-6">
          <div className="bg-slate-800/20 border border-white/5 rounded-2xl p-5">
            <h4 className="text-white mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">support_agent</span>
              24/7 Enterprise Support
            </h4>
            <p className="text-slate-400 text-sm">
              Our dedicated support team is available around the clock to assist with any technical issues or questions.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4">
            <div className="bg-slate-800/30 border border-white/10 rounded-xl p-4 hover:bg-slate-800/50 transition-all cursor-pointer group">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-blue-400 text-xl group-hover:scale-110 transition-transform">phone</span>
                <div>
                  <h5 className="text-white font-semibold">Phone Support</h5>
                  <p className="text-slate-400 text-sm">+1 (555) FLEET-HELP</p>
                </div>
              </div>
            </div>

            <div className="bg-slate-800/30 border border-white/10 rounded-xl p-4 hover:bg-slate-800/50 transition-all cursor-pointer group">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-green-400 text-xl group-hover:scale-110 transition-transform">mail</span>
                <div>
                  <h5 className="text-white font-semibold">Email Support</h5>
                  <p className="text-slate-400 text-sm">support@flyfleet.com</p>
                </div>
              </div>
            </div>

            <div className="bg-slate-800/30 border border-white/10 rounded-xl p-4 hover:bg-slate-800/50 transition-all cursor-pointer group">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-purple-400 text-xl group-hover:scale-110 transition-transform">forum</span>
                <div>
                  <h5 className="text-white font-semibold">Live Chat</h5>
                  <p className="text-slate-400 text-sm">Available in your dashboard</p>
                </div>
              </div>
            </div>

            <div className="bg-slate-800/30 border border-white/10 rounded-xl p-4 hover:bg-slate-800/50 transition-all cursor-pointer group">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-amber-400 text-xl group-hover:scale-110 transition-transform">emergency</span>
                <div>
                  <h5 className="text-white font-semibold">Emergency Hotline</h5>
                  <p className="text-slate-400 text-sm">Critical system issues only</p>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-amber-500/10 border border-amber-500/30 rounded-xl p-4">
            <h4 className="text-amber-500 mb-2 flex items-center gap-2">
              <span className="material-symbols-outlined">priority_high</span>
              Priority Support Levels
            </h4>
            <div className="space-y-2 text-amber-200 text-sm">
              <p><span className="font-semibold">Enterprise:</span> 15-minute response time</p>
              <p><span className="font-semibold">Professional:</span> 1-hour response time</p>
              <p><span className="font-semibold">Starter:</span> 24-hour response time</p>
            </div>
          </div>
        </div>
      </Modal>
    </>
  );
};

export default EnterpriseAuthLayout;
