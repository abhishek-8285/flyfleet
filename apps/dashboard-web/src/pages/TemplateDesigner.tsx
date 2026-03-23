import React from 'react';
import SvelteWrapper from '../components/builder/SvelteWrapper';

const TemplateDesignerPage: React.FC = () => {
  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-slate-900">Billing Infrastructure</h1>
        <p className="text-slate-500">Design and customize your enterprise invoice templates.</p>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <SvelteWrapper customerName="Global Logistics Inc." initialAmount={125000} />
      </div>
      
      <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="p-4 bg-blue-50 rounded-lg border border-blue-100">
          <h4 className="font-bold text-blue-900 mb-2">Auto-Scaling</h4>
          <p className="text-sm text-blue-700">Templates automatically adjust for mobile and A4 print formats.</p>
        </div>
        <div className="p-4 bg-green-50 rounded-lg border border-green-100">
          <h4 className="font-bold text-green-900 mb-2">GST Compliant</h4>
          <p className="text-sm text-green-700">All designs are validated against the Indian GST schema.</p>
        </div>
        <div className="p-4 bg-purple-50 rounded-lg border border-purple-100">
          <h4 className="font-bold text-purple-900 mb-2">Svelte Performance</h4>
          <p className="text-sm text-purple-700">Real-time reactive updates with zero latency.</p>
        </div>
      </div>
    </div>
  );
};

export default TemplateDesignerPage;
