'use client';

import React, { useState } from 'react';

interface ROICalculatorProps {
  companyName: string;
}

const ROICalculator: React.FC<ROICalculatorProps> = ({ companyName }) => {
  const [fleetSize, setFleetSize] = useState(50);
  const [fuelCost, setFuelCost] = useState(3000);
  
  const monthlySavings = fleetSize * (fuelCost * 0.15 + 120);
  const yearlySavings = monthlySavings * 12;

  return (
    <section className="py-20 px-4 md:px-10 bg-primary/5">
      <div className="max-w-[1280px] mx-auto">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
          <div>
            <h2 className="text-3xl md:text-5xl font-bold text-slate-900 dark:text-white mb-6 font-display">Calculate Your <span className="text-primary italic">Estimated ROI</span></h2>
            <p className="text-slate-600 dark:text-text-secondary text-lg mb-10">
              See how much you could save by switching to {companyName}'s AI-driven platform. Most enterprise fleets see a break-even within 4 months.
            </p>
            <div className="space-y-8">
              <div>
                <div className="flex justify-between items-center mb-4">
                  <label className="font-bold text-slate-700 dark:text-slate-300">Fleet size</label>
                  <span className="text-primary font-black text-xl">{fleetSize} Vehicles</span>
                </div>
                <input 
                  type="range" min="1" max="1000" step="5" value={fleetSize} 
                  onChange={(e) => setFleetSize(parseInt(e.target.value))}
                  className="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-primary"
                />
              </div>
              <div>
                <div className="flex justify-between items-center mb-4">
                  <label className="font-bold text-slate-700 dark:text-slate-300">Monthly fuel cost per vehicle</label>
                  <span className="text-primary font-black text-xl">${fuelCost}</span>
                </div>
                <input 
                  type="range" min="500" max="10000" step="100" value={fuelCost} 
                  onChange={(e) => setFuelCost(parseInt(e.target.value))}
                  className="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-primary"
                />
              </div>
            </div>
          </div>
          <div className="bg-white dark:bg-surface-dark p-8 md:p-12 rounded-3xl shadow-2xl border border-slate-200 dark:border-border-dark relative overflow-hidden group">
            <div className="absolute top-0 right-0 w-32 h-32 bg-primary/10 rounded-full -translate-y-1/2 translate-x-1/2 blur-2xl"></div>
            <div className="relative z-10">
              <h3 className="text-slate-500 dark:text-text-secondary font-bold uppercase tracking-widest text-xs mb-8">Estimated Annual Savings</h3>
              <div className="text-5xl md:text-7xl font-black text-slate-900 dark:text-white mb-4 font-display">
                ${yearlySavings.toLocaleString()}
              </div>
              <p className="text-green-500 font-bold flex items-center gap-2 mb-10">
                <span className="material-symbols-outlined">trending_up</span>
                Approx. 22% reduction in total operational costs
              </p>
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 rounded-xl bg-slate-50 dark:bg-background-dark/50 border border-slate-100 dark:border-border-dark">
                  <div className="text-slate-500 text-xs mb-1">Monthly Savings</div>
                  <div className="text-xl font-bold text-slate-900 dark:text-white">${monthlySavings.toLocaleString()}</div>
                </div>
                <div className="p-4 rounded-xl bg-slate-50 dark:bg-background-dark/50 border border-slate-100 dark:border-border-dark">
                  <div className="text-slate-500 text-xs mb-1">Payback Period</div>
                  <div className="text-xl font-bold text-slate-900 dark:text-white">~3.8 Months</div>
                </div>
              </div>
              <button className="w-full mt-10 h-14 bg-primary hover:bg-primary/90 text-white font-bold rounded-xl shadow-lg shadow-primary/30 transition-all flex items-center justify-center gap-3">
                Get Detailed ROI Report
                <span className="material-symbols-outlined !text-xl">description</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default ROICalculator;
