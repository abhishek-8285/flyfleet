import React from 'react';

const SystemMaintenance: React.FC = () => {
  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-50 text-gray-900 font-sans">
      <div className="max-w-xl px-8 py-12 bg-white rounded-2xl shadow-xl text-center border border-gray-100">
        <div className="mb-6 flex justify-center">
          <div className="h-20 w-20 bg-yellow-100 rounded-full flex items-center justify-center">
            <svg 
              className="w-10 h-10 text-yellow-600" 
              fill="none" 
              stroke="currentColor" 
              viewBox="0 0 24 24" 
              xmlns="http://www.w3.org/2000/svg"
            >
              <path 
                strokeLinecap="round" 
                strokeLinejoin="round" 
                strokeWidth={2} 
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" 
              />
              <path 
                strokeLinecap="round" 
                strokeLinejoin="round" 
                strokeWidth={2} 
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" 
              />
            </svg>
          </div>
        </div>
        <h1 className="text-3xl font-extrabold text-gray-900 tracking-tight mb-4">
          Scheduled Maintenance
        </h1>
        <p className="text-lg text-gray-600 mb-8 leading-relaxed">
          The FlyFleet platform is currently undergoing scheduled maintenance or experiencing high load. 
          To ensure platform stability, we've temporarily restricted access. We will be back online shortly.
        </p>
        <button 
          onClick={() => window.location.href = '/'}
          className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-8 rounded-lg shadow-sm transition duration-150 ease-in-out"
        >
          Check Status & Reload
        </button>
      </div>
    </div>
  );
};

export default SystemMaintenance;
