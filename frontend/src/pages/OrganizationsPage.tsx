import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { orgs } from '../api/client';
import { Organization } from '../types';
import toast from 'react-hot-toast';

export default function OrganizationsPage() {
    const navigate = useNavigate();
    const [organizations, setOrganizations] = useState<Organization[]>([]);
    const [loading, setLoading] = useState(true);
    const [showModal, setShowModal] = useState(false);
    const [newOrg, setNewOrg] = useState({ name: '', display_name: '', description: '' });

    useEffect(() => { loadOrgs(); }, []);

    const loadOrgs = async () => {
        try {
            const data = await orgs.list() as Organization[];
            setOrganizations(data || []);
        } catch { } finally { setLoading(false); }
    };

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await orgs.create(newOrg);
            toast.success('Organization created!');
            setShowModal(false);
            setNewOrg({ name: '', display_name: '', description: '' });
            loadOrgs();
        } catch (err: any) {
            toast.error(err.message);
        }
    };

    if (loading) {
        return <div className="loading-page"><div className="loading-spinner loading-spinner-lg"></div></div>;
    }

    return (
        <div>
            <div className="page-header">
                <div>
                    <h1 className="page-title">Organizations</h1>
                    <p className="page-subtitle">Manage your teams and infrastructure</p>
                </div>
                <button className="btn btn-primary" onClick={() => setShowModal(true)}>
                    + New Organization
                </button>
            </div>

            {organizations.length === 0 ? (
                <div className="card">
                    <div className="empty-state">
                        <div className="empty-state-icon">🏢</div>
                        <h3 className="empty-state-title">No organizations</h3>
                        <p className="empty-state-message">
                            Organizations let you group projects and manage team members with role-based access control.
                        </p>
                        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
                            Create Your First Organization
                        </button>
                    </div>
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
                                <div style={{
                                    width: 48, height: 48, borderRadius: 12,
                                    background: 'linear-gradient(135deg, var(--color-accent), var(--color-indigo))',
                                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                                    fontSize: '1.25rem', color: 'white', fontWeight: 700,
                                    boxShadow: '0 0 20px rgba(124, 58, 237, 0.2)',
                                }}>
                                    {(org.display_name || org.name).charAt(0).toUpperCase()}
                                </div>
                                <div>
                                    <div className="font-semibold" style={{ fontSize: '1.0625rem' }}>
                                        {org.display_name || org.name}
                                    </div>
                                    <div className="text-xs text-muted text-mono">{org.name}</div>
                                </div>
                            </div>
                            {org.description && <p className="card-description">{org.description}</p>}
                            <div className="flex items-center gap-md mt-md text-xs text-muted">
                                <span>Created {new Date(org.created_at).toLocaleDateString()}</span>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {/* Create Modal */}
            {showModal && (
                <div className="modal-overlay" onClick={() => setShowModal(false)}>
                    <div className="modal" onClick={e => e.stopPropagation()}>
                        <h2 className="modal-title">Create Organization</h2>
                        <form onSubmit={handleCreate}>
                            <div className="form-group">
                                <label className="form-label">Name (slug)</label>
                                <input
                                    className="form-input form-input-mono"
                                    placeholder="my-organization"
                                    value={newOrg.name}
                                    onChange={e => setNewOrg({ ...newOrg, name: e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '-') })}
                                    required
                                />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Display Name</label>
                                <input
                                    className="form-input"
                                    placeholder="My Organization"
                                    value={newOrg.display_name}
                                    onChange={e => setNewOrg({ ...newOrg, display_name: e.target.value })}
                                />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Description</label>
                                <textarea
                                    className="form-input"
                                    placeholder="Describe your organization..."
                                    value={newOrg.description}
                                    onChange={e => setNewOrg({ ...newOrg, description: e.target.value })}
                                    rows={3}
                                />
                            </div>
                            <div className="modal-actions">
                                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>Cancel</button>
                                <button type="submit" className="btn btn-primary">Create Organization</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}
