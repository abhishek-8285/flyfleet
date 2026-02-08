import {
    ClockIcon,
    MapPinIcon,
    TruckIcon, UserIcon,
    XMarkIcon
} from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import api from '../../../services/api';

export default function AddTrip() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [formData, setFormData] = useState({
    vehicle_id: '',
    driver_id: '',
    start_location: '',
    end_location: '',
    start_time: '',
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await api.createTrip({
        ...formData,
        start_time: new Date(formData.start_time).toISOString(),
      });
      navigate('/admin/trips');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create trip');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
            <Link to="/admin/trips" className="hover:text-white">Trips</Link>
            <span>/</span>
            <span className="text-primary">Add</span>
          </div>
          <h1 className="text-2xl font-bold text-white">Add New Trip</h1>
        </div>
        <div className="flex items-center gap-3">
          <Link to="/admin/trips" className="px-4 py-2 bg-surface-dark border border-border-color text-text-muted hover:text-white text-sm rounded transition-colors">
            Cancel
          </Link>
          <button
            onClick={handleSubmit}
            disabled={loading}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded transition-colors disabled:opacity-50"
          >
            {loading ? 'Creating...' : 'Create Trip'}
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-danger/10 border border-danger/20 rounded-lg flex items-center gap-2 text-danger">
          <XMarkIcon className="w-5 h-5" />
          {error}
        </div>
      )}

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
              <TruckIcon className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-white font-semibold">Trip Essentials</h3>
              <p className="text-xs text-text-muted">Basic identification for the trip</p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase">Vehicle ID <span className="text-danger">*</span></label>
              <div className="relative">
                <TruckIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                <input
                  type="text"
                  name="vehicle_id"
                  value={formData.vehicle_id}
                  onChange={handleChange}
                  required
                  placeholder="Enter vehicle UUID"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase">Driver ID</label>
              <div className="relative">
                <UserIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                <input
                  type="text"
                  name="driver_id"
                  value={formData.driver_id}
                  onChange={handleChange}
                  placeholder="Enter driver UUID (optional)"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
                />
              </div>
            </div>
          </div>
        </div>

        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-success/10 flex items-center justify-center text-success">
              <MapPinIcon className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-white font-semibold">Route Details</h3>
              <p className="text-xs text-text-muted">Origin and destination</p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase">Start Location <span className="text-danger">*</span></label>
              <input
                type="text"
                name="start_location"
                value={formData.start_location}
                onChange={handleChange}
                required
                placeholder="e.g. New York, NY"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase">End Location <span className="text-danger">*</span></label>
              <input
                type="text"
                name="end_location"
                value={formData.end_location}
                onChange={handleChange}
                required
                placeholder="e.g. Boston, MA"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
              />
            </div>
          </div>
        </div>

        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-warning/10 flex items-center justify-center text-warning">
              <ClockIcon className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-white font-semibold">Schedule</h3>
              <p className="text-xs text-text-muted">Planned departure time</p>
            </div>
          </div>
          <div className="p-6">
            <div className="space-y-1.5 max-w-xs">
              <label className="text-xs font-medium text-text-muted uppercase">Start Time <span className="text-danger">*</span></label>
              <input
                type="datetime-local"
                name="start_time"
                value={formData.start_time}
                onChange={handleChange}
                required
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
              />
            </div>
          </div>
        </div>
      </form>
    </div>
  );
}
