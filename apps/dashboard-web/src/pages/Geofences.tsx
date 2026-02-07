import React from 'react';

const Geofences: React.FC = () => {
  return (
    <div className="h-full flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">Geofences</h2>
        <p className="text-gray-600">Geofence management with map drawing will be displayed here</p>
        <p className="text-sm text-gray-500 mt-4">Requires Leaflet Draw integration</p>
      </div>
    </div>
  );
};

export default Geofences;
