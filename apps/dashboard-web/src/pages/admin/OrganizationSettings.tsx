import { ArrowLeftIcon, BellAlertIcon, KeyIcon, ShieldCheckIcon, UserGroupIcon } from '@heroicons/react/24/outline';
import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';

export const OrganizationSettings: React.FC = () => {
    const { id } = useParams();
    const navigate = useNavigate();

    const sections = [
        {
            title: 'Security & Access',
            items: [
                {
                    icon: KeyIcon,
                    name: 'MFA Configuration',
                    description: 'Enforce Multi-Factor Authentication policies for this organization.',
                    action: () => navigate(`/security/policy`), // Ideally scoped to org in admin view
                    color: 'text-purple-500',
                    bgColor: 'bg-purple-50'
                },
                {
                    icon: ShieldCheckIcon,
                    name: 'Audit Log Settings',
                    description: 'Configure data retention and event logging granularity.',
                    action: () => navigate(`/security/audit/settings`), // Ideally scoped to org in admin view
                    color: 'text-blue-500',
                    bgColor: 'bg-blue-50'
                }
            ]
        },
        {
            title: 'Workspace',
            items: [
                {
                    icon: UserGroupIcon,
                    name: 'Role-Based Access Control',
                    description: 'Manage roles and permissions for organization members.',
                    action: () => {}, // Future implementation
                    color: 'text-green-500',
                    bgColor: 'bg-green-50'
                },
                {
                    icon: BellAlertIcon,
                    name: 'Notification Preferences',
                    description: 'Set up alerts for critical fleet events.',
                    action: () => {}, // Future implementation
                    color: 'text-orange-500',
                    bgColor: 'bg-orange-50'
                }
            ]
        }
    ];

    return (
        <div className="p-8 max-w-5xl mx-auto">
             <button 
                onClick={() => navigate(`/admin/organizations/${id}`)}
                className="flex items-center space-x-2 text-slate-500 hover:text-blue-600 transition-colors mb-8 group"
            >
                <ArrowLeftIcon className="h-4 w-4 group-hover:-translate-x-1 transition-transform" />
                <span>Back to Profile</span>
            </button>

            <div className="mb-10">
                <h1 className="text-3xl font-bold text-slate-900">Organization Settings</h1>
                <p className="text-slate-500 mt-2">Manage security protocols, workspace preferences, and compliance standards.</p>
            </div>

            <div className="space-y-12">
                {sections.map((section, idx) => (
                    <div key={idx} className="animate-in fade-in slide-in-from-bottom-4 duration-500" style={{ animationDelay: `${idx * 100}ms` }}>
                        <h2 className="text-lg font-bold text-slate-800 mb-4 px-1">{section.title}</h2>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {section.items.map((item, i) => (
                                <div 
                                    key={i}
                                    onClick={item.action}
                                    className="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm hover:shadow-md hover:border-blue-200 transition-all cursor-pointer group"
                                >
                                    <div className="flex items-start space-x-4">
                                        <div className={`p-3 rounded-xl ${item.bgColor}`}>
                                            <item.icon className={`h-6 w-6 ${item.color}`} />
                                        </div>
                                        <div>
                                            <h3 className="font-semibold text-slate-900 group-hover:text-blue-600 transition-colors">
                                                {item.name}
                                            </h3>
                                            <p className="text-sm text-slate-500 mt-1 leading-relaxed">
                                                {item.description}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                ))}
            </div>

            {/* Danger Zone */}
            <div className="mt-16 pt-10 border-t border-slate-200 animate-in fade-in duration-700">
                <h2 className="text-lg font-bold text-red-600 mb-4">Danger Zone</h2>
                <div className="bg-red-50 border border-red-100 rounded-2xl p-6 flex justify-between items-center">
                    <div>
                        <h3 className="font-semibold text-red-900">Archive Organization</h3>
                        <p className="text-sm text-red-700 mt-1">
                            This will restrict access for all users immediately. Organization data will be retained for 30 days.
                        </p>
                    </div>
                    <button className="bg-white border border-red-200 text-red-600 hover:bg-red-600 hover:text-white px-6 py-2 rounded-lg font-medium transition-colors shadow-sm">
                        Archive Workspace
                    </button>
                </div>
            </div>
        </div>
    );
};
