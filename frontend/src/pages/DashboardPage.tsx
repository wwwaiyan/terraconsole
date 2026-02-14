import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { orgs } from '../api/client';
import { Organization } from '../types';

export default function DashboardPage() {
    const { user } = useAuth();
    const navigate = useNavigate();
    const [organizations, setOrganizations] = useState<Organization[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadData();
    }, []);

    const loadData = async () => {
        try {
            const data = await orgs.list() as Organization[];
            setOrganizations(data || []);
        } catch { } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="loading-page">
                <div className="loading-spinner loading-spinner-lg"></div>
            </div>
        );
    }

    return (
        <div>
            <div className="page-header">
                <div>
                    <h1 className="page-title">
                        Welcome back, {user?.full_name?.split(' ')[0] || user?.username} 👋
                    </h1>
                    <p className="page-subtitle">Here's an overview of your infrastructure</p>
                </div>
            </div>

            {/* Stats */}
            <div className="grid grid-4" style={{ marginBottom: '32px' }}>
                <div className="stat-card">
                    <div className="stat-value">{organizations.length}</div>
                    <div className="stat-label">Organizations</div>
                </div>
                <div className="stat-card">
                    <div className="stat-value">—</div>
                    <div className="stat-label">Projects</div>
                </div>
                <div className="stat-card">
                    <div className="stat-value">—</div>
                    <div className="stat-label">Workspaces</div>
                </div>
                <div className="stat-card">
                    <div className="stat-value">—</div>
                    <div className="stat-label">Active Runs</div>
                </div>
            </div>

            {/* Quick Actions */}
            <div className="card" style={{ marginBottom: '24px' }}>
                <div className="card-header">
                    <h3 className="card-title">Quick Actions</h3>
                </div>
                <div className="flex gap-md" style={{ flexWrap: 'wrap' }}>
                    <button
                        className="btn btn-primary"
                        onClick={() => navigate('/organizations')}
                    >
                        🏢 Manage Organizations
                    </button>
                    <button
                        className="btn btn-secondary"
                        onClick={() => navigate('/settings')}
                    >
                        ⚙️ Settings
                    </button>
                </div>
            </div>

            {/* Recent Organizations */}
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Your Organizations</h3>
                    <button
                        className="btn btn-primary btn-sm"
                        onClick={() => navigate('/organizations')}
                    >
                        + New Organization
                    </button>
                </div>

                {organizations.length === 0 ? (
                    <div className="empty-state">
                        <div className="empty-state-icon">🏢</div>
                        <h3 className="empty-state-title">No organizations yet</h3>
                        <p className="empty-state-message">
                            Create your first organization to start managing your Terraform infrastructure.
                        </p>
                        <button
                            className="btn btn-primary"
                            onClick={() => navigate('/organizations')}
                        >
                            Create Organization
                        </button>
                    </div>
                ) : (
                    <div className="grid grid-3">
                        {organizations.map(org => (
                            <div
                                key={org.id}
                                className="card card-clickable"
                                onClick={() => navigate(`/organizations/${org.id}/projects`)}
                            >
                                <div className="flex items-center gap-md mb-md">
                                    <div
                                        style={{
                                            width: 40, height: 40, borderRadius: 10,
                                            background: 'linear-gradient(135deg, var(--color-accent), var(--color-teal))',
                                            display: 'flex', alignItems: 'center', justifyContent: 'center',
                                            fontSize: '1.125rem', color: 'white', fontWeight: 700,
                                        }}
                                    >
                                        {(org.display_name || org.name).charAt(0).toUpperCase()}
                                    </div>
                                    <div>
                                        <div className="font-semibold">{org.display_name || org.name}</div>
                                        <div className="text-xs text-muted">{org.name}</div>
                                    </div>
                                </div>
                                {org.description && (
                                    <p className="card-description">{org.description}</p>
                                )}
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}
