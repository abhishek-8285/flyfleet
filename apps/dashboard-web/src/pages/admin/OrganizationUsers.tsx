import React from 'react';
import { useParams } from 'react-router-dom';

export const OrganizationUsers: React.FC = () => {
  const { id } = useParams();
  
  return (
    <div className="p-8">
      <h1 className="text-white text-2xl font-bold mb-4">Organization Users</h1>
      <p className="text-text-muted">Manage users for organization ID: {id}</p>
      <div className="mt-8 bg-surface-dark rounded-xl border border-border-color p-6">
        <p className="text-text-muted text-center">User management interface coming soon...</p>
      </div>
    </div>
  );
};