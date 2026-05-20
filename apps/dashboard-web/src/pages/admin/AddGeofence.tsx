import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { 
  ChevronLeft, 
  Save, 
  Map as MapIcon, 
  Circle as CircleIcon, 
  Pentagon, 
  Navigation,
  Info,
  AlertTriangle
} from 'lucide-react';
import { MapContainer, TileLayer, Marker, Circle, Polygon, useMapEvents } from 'react-leaflet';
import { api } from '../../services/api';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';

// Fix for default marker icon
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
});

const AddGeofence: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const query = new URLSearchParams(location.search);
  const editId = query.get('edit');

  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    type: 'circle' as 'circle' | 'polygon',
    radius_m: 500,
    center_lat: 0,
    center_lng: 0,
    polygon: [] as [number, number][]
  });

  const [mapCenter, setMapCenter] = useState<[number, number]>([28.6139, 77.2090]); // Delhi default

  useEffect(() => {
    if (editId) {
      const loadGeofence = async () => {
        try {
          const res = await api.getGeofences();
          const data = res?.data || res;
          const list = Array.isArray(data) ? data : data?.geofences || [];
          const gf = list.find((g: any) => g.id === editId);
          if (gf) {
            let polyCoords: [number, number][] = [];
            if (gf.type === 'polygon') {
              if (gf.polygon) {
                polyCoords = gf.polygon;
              } else if (gf.polygon_wkt) {
                try {
                  const cleanWkt = gf.polygon_wkt.trim();
                  if (cleanWkt.startsWith('[')) {
                    polyCoords = JSON.parse(cleanWkt);
                  } else if (cleanWkt.toUpperCase().startsWith('POLYGON')) {
                    const match = cleanWkt.match(/\(\((.*)\)\)/);
                    if (match && match[1]) {
                      const points = match[1].split(',');
                      polyCoords = points.map((pt: string) => {
                        const parts = pt.trim().split(/\s+/);
                        return [parseFloat(parts[1]), parseFloat(parts[0])] as [number, number];
                      });
                    }
                  }
                } catch (e) {
                  console.error('Failed to parse polygon_wkt:', e);
                }
              }
            }

            setFormData({
              name: gf.name || '',
              description: gf.description || '',
              type: gf.type || 'circle',
              radius_m: gf.radius_m || gf.radius || 500,
              center_lat: gf.center_lat || gf.center?.lat || 0,
              center_lng: gf.center_lng || gf.center?.lng || 0,
              polygon: polyCoords
            });

            if (gf.center_lat) {
              setMapCenter([gf.center_lat, gf.center_lng]);
            } else if (polyCoords.length > 0) {
              setMapCenter(polyCoords[0]);
            }
          }
        } catch (err) {
          console.error('Failed to load geofence for editing:', err);
        }
      };
      loadGeofence();
    }
  }, [editId]);

  const MapEvents = () => {
    useMapEvents({
      click(e) {
        if (formData.type === 'circle') {
          setFormData(prev => ({
            ...prev,
            center_lat: e.latlng.lat,
            center_lng: e.latlng.lng
          }));
        } else {
          setFormData(prev => ({
            ...prev,
            polygon: [...prev.polygon, [e.latlng.lat, e.latlng.lng]]
          }));
        }
      },
    });
    return null;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name) return alert('Please enter a name');
    
    setLoading(true);
    try {
      const payload = {
        name: formData.name,
        description: formData.description,
        type: formData.type,
        center_lat: formData.type === 'circle' ? formData.center_lat : null,
        center_lng: formData.type === 'circle' ? formData.center_lng : null,
        radius_m: formData.type === 'circle' ? formData.radius_m : null,
        polygon_wkt: formData.type === 'polygon' ? JSON.stringify(formData.polygon) : null,
        enabled: true
      };
      
      if (editId) {
        // Delete original geofence first, then save the updated one
        await api.deleteGeofence(editId);
      }
      await api.createGeofence(payload);
      navigate('/admin/geofences');
    } catch (err) {
      console.error('Failed to save geofence:', err);
      alert('Failed to save geofence. Please check console.');
    } finally {
      setLoading(false);
    }
  };

  const clearPolygon = () => setFormData(prev => ({ ...prev, polygon: [] }));
  const removeLastPoint = () => setFormData(prev => ({ ...prev, polygon: prev.polygon.slice(0, -1) }));

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] bg-[#f8fafc] dark:bg-black">
      {/* Top Header */}
      <div className="h-16 bg-white dark:bg-white/[0.02] border-b border-gray-100 dark:border-white/5 flex items-center justify-between px-6 shrink-0">
        <div className="flex items-center gap-4">
          <button 
            onClick={() => navigate('/admin/geofences')}
            className="p-2 hover:bg-gray-100 dark:hover:bg-white/5 rounded-xl text-gray-500 transition-all"
          >
            <ChevronLeft size={20} />
          </button>
          <div>
            <h1 className="text-lg font-black tracking-tighter text-gray-900 dark:text-white uppercase leading-none">
              {editId ? 'Edit Geofence' : 'Create Geofence'}
            </h1>
            <p className="text-[10px] font-black text-gray-400 uppercase tracking-widest mt-1">
              {editId ? 'Modify spatial monitoring zone' : 'Design spatial monitoring zone'}
            </p>
          </div>
        </div>
        <button 
          onClick={handleSubmit}
          disabled={loading}
          className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white px-5 py-2 rounded-xl font-bold transition-all shadow-lg shadow-emerald-600/20 text-xs uppercase tracking-tighter"
        >
          {loading ? 'SAVING...' : <><Save size={16} /> {editId ? 'UPDATE GEOFENCE' : 'SAVE GEOFENCE'}</>}
        </button>
      </div>

      <div className="flex flex-1 overflow-hidden">
        {/* Left Sidebar Form */}
        <div className="w-80 border-r border-gray-100 dark:border-white/5 bg-white dark:bg-white/[0.01] overflow-y-auto p-6 space-y-6 shrink-0">
          <div className="space-y-4">
            <div>
              <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Zone Name</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                placeholder="e.g. Warehouse Alpha"
                className="w-full bg-gray-50 dark:bg-white/5 border border-gray-100 dark:border-white/10 rounded-xl px-4 py-2.5 text-sm font-bold text-gray-900 dark:text-white focus:ring-2 ring-emerald-500/20 outline-none transition-all"
              />
            </div>

            <div>
              <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Description</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                placeholder="Zone details..."
                rows={3}
                className="w-full bg-gray-50 dark:bg-white/5 border border-gray-100 dark:border-white/10 rounded-xl px-4 py-2.5 text-sm font-bold text-gray-900 dark:text-white focus:ring-2 ring-emerald-500/20 outline-none transition-all resize-none"
              />
            </div>

            <div>
              <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Geometry Type</label>
              <div className="grid grid-cols-2 gap-2">
                {[
                  { id: 'circle', label: 'Circle', icon: CircleIcon },
                  { id: 'polygon', label: 'Polygon', icon: Pentagon }
                ].map((t) => (
                  <button
                    key={t.id}
                    onClick={() => setFormData(prev => ({ ...prev, type: t.id as any }))}
                    className={`flex items-center justify-center gap-2 py-3 rounded-xl border transition-all ${
                      formData.type === t.id 
                        ? 'bg-blue-600 border-blue-600 text-white shadow-lg shadow-blue-600/20' 
                        : 'bg-white dark:bg-white/5 border-gray-100 dark:border-white/10 text-gray-500 hover:border-gray-200'
                    }`}
                  >
                    <t.icon size={16} />
                    <span className="text-[10px] font-black uppercase tracking-tighter">{t.label}</span>
                  </button>
                ))}
              </div>
            </div>

            {formData.type === 'circle' ? (
              <div>
                <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5 flex justify-between">
                  Radius (Meters)
                  <span className="text-blue-500 font-mono">{formData.radius_m}m</span>
                </label>
                <input
                  type="range"
                  min="50"
                  max="5000"
                  step="50"
                  value={formData.radius_m}
                  onChange={(e) => setFormData(prev => ({ ...prev, radius_m: parseInt(e.target.value) }))}
                  className="w-full h-1.5 bg-gray-200 dark:bg-white/10 rounded-lg appearance-none cursor-pointer accent-blue-600"
                />
              </div>
            ) : (
              <div>
                <label className="text-[10px] font-black uppercase tracking-widest text-gray-400 block mb-1.5">Polygon Points</label>
                <div className="bg-gray-50 dark:bg-white/5 rounded-xl p-3 border border-gray-100 dark:border-white/10">
                  <div className="grid grid-cols-2 gap-2">
                    <button 
                      type="button"
                      onClick={removeLastPoint}
                      disabled={formData.polygon.length === 0}
                      className="py-2.5 bg-gray-50 hover:bg-gray-100 dark:bg-white/5 dark:hover:bg-white/10 text-gray-700 dark:text-gray-300 rounded-xl text-[10px] font-black uppercase tracking-tighter disabled:opacity-50 transition-colors border border-gray-100 dark:border-white/5"
                    >
                      Undo Last
                    </button>
                    <button 
                      type="button"
                      onClick={clearPolygon}
                      disabled={formData.polygon.length === 0}
                      className="py-2.5 bg-red-50 hover:bg-red-100 dark:bg-red-500/10 dark:hover:bg-red-500/20 text-red-600 dark:text-red-400 rounded-xl text-[10px] font-black uppercase tracking-tighter disabled:opacity-50 transition-colors border border-red-100/50 dark:border-red-500/10"
                    >
                      Clear All
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>

          <div className="bg-blue-50 dark:bg-blue-500/10 p-4 rounded-2xl border border-blue-100 dark:border-blue-500/20">
            <div className="flex gap-3 text-blue-700 dark:text-blue-400">
              <Info size={18} className="shrink-0" />
              <div>
                <p className="text-[11px] font-black uppercase tracking-tighter leading-tight mb-1">How to define</p>
                <p className="text-[10px] font-medium leading-normal opacity-80 whitespace-pre-line">
                  {formData.type === 'circle' 
                    ? 'Click on the map to place the center point of your circular zone.'
                    : `• Click on the map to add new boundary corners.
• Drag any blue point to adjust its position.
• Double-click any blue point to delete it.`}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Map Area */}
        <div className="flex-1 relative">
          <MapContainer 
            center={mapCenter} 
            zoom={13} 
            className="w-full h-full z-0"
            zoomControl={false}
          >
            <TileLayer url="https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}" />
            <MapEvents />
            
            {formData.type === 'circle' && formData.center_lat !== 0 && (
              <>
                <Marker position={[formData.center_lat, formData.center_lng]} />
                <Circle 
                  center={[formData.center_lat, formData.center_lng]} 
                  radius={formData.radius_m}
                  pathOptions={{ fillColor: '#3b82f6', color: '#2563eb', weight: 2 }}
                />
              </>
            )}

            {formData.type === 'polygon' && formData.polygon.length > 0 && (
              <>
                {formData.polygon.map((p, i) => (
                  <Marker 
                    key={i} 
                    position={p} 
                    draggable={true}
                    eventHandlers={{
                      dragend: (e) => {
                        const marker = e.target;
                        const position = marker.getLatLng();
                        setFormData(prev => {
                          const newPolygon = [...prev.polygon];
                          newPolygon[i] = [position.lat, position.lng];
                          return { ...prev, polygon: newPolygon };
                        });
                      },
                      dblclick: () => {
                        setFormData(prev => ({
                          ...prev,
                          polygon: prev.polygon.filter((_, idx) => idx !== i)
                        }));
                      }
                    }}
                    icon={L.divIcon({ 
                      className: 'custom-polygon-vertex',
                      html: `<div class="w-4 h-4 bg-blue-600 hover:bg-blue-500 rounded-full border-2 border-white shadow-lg cursor-move transition-transform hover:scale-125"></div>`,
                      iconSize: [16, 16],
                      iconAnchor: [8, 8]
                    })} 
                  />
                ))}
                {formData.polygon.length > 2 && (
                  <Polygon 
                    positions={formData.polygon}
                    pathOptions={{ fillColor: '#3b82f6', color: '#2563eb', weight: 2 }}
                  />
                )}
              </>
            )}
          </MapContainer>

          {/* Floating Warning */}
          {(formData.type === 'circle' && formData.center_lat === 0) || (formData.type === 'polygon' && formData.polygon.length < 3) ? (
            <div className="absolute top-6 left-1/2 -translate-x-1/2 z-10 bg-white dark:bg-black/80 px-6 py-3 rounded-2xl border border-gray-100 dark:border-white/10 shadow-2xl backdrop-blur flex items-center gap-3">
              <AlertTriangle className="text-amber-500" size={20} />
              <p className="text-xs font-black uppercase tracking-tighter text-gray-700 dark:text-white">
                {formData.type === 'circle' ? 'Please click on map to set center' : 'Add at least 3 points to the map'}
              </p>
            </div>
          ) : null}

          {/* Location Info Overlay */}
          <div className="absolute bottom-6 left-6 z-10 bg-white dark:bg-black/60 px-4 py-3 rounded-2xl border border-gray-100 dark:border-white/10 shadow-2xl backdrop-blur">
            <div className="flex flex-col gap-1">
              <span className="text-[10px] font-black uppercase tracking-widest text-gray-400">Current Definition</span>
              {formData.type === 'circle' ? (
                <span className="text-xs font-mono font-bold text-gray-700 dark:text-white">
                  LAT: {formData.center_lat.toFixed(6)} LNG: {formData.center_lng.toFixed(6)}
                </span>
              ) : (
                <span className="text-xs font-mono font-bold text-gray-700 dark:text-white">
                  {formData.polygon.length} VERTICES RECORDED
                </span>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AddGeofence;
