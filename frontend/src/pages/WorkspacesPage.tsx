import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { workspaces, projects } from '../api/client';
import { Workspace, Project } from '../types';
import toast from 'react-hot-toast';

export default function WorkspacesPage() {
    const { projectId } = useParams<{ projectId: string }>();
    const navigate = useNavigate();
    const [project, setProject] = useState<Project | null>(null);
    const [wsList, setWsList] = useState<Workspace[]>([]);
    const [loading, setLoading] = useState(true);
    const [showModal, setShowModal] = useState(false);
    const [newWs, setNewWs] = useState({
        name: '', description: '', terraform_version: 'latest', auto_apply: false,
    });

    useEffect(() => { loadData(); }, [projectId]);

    const loadData = async () => {
        try {
            const [projData, wsData] = await Promise.all([
                projects.get(projectId!) as Promise<Project>,
                workspaces.list(projectId!) as Promise<Workspace[]>,
            ]);
            setProject(projData);
            setWsList(wsData || []);
        } catch { } finally { setLoading(false); }
    };

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await workspaces.create(projectId!, newWs);
            toast.success('Workspace created!');
            setShowModal(false);
            setNewWs({ name: '', description: '', terraform_version: 'latest', auto_apply: false });
            loadData();
        } catch (err: any) { toast.error(err.message); }
    };

    if (loading) {
        return <div className="loading-page"><div className="loading-spinner loading-spinner-lg"></div></div>;
    }

    return (
        <div>
            <div className="page-header">
                <div>
                    <div className="text-xs text-muted" style={{ marginBottom: 4 }}>
                        <span style={{ cursor: 'pointer' }} onClick={() => navigate('/organizations')}>Organizations</span>
                        <span> / </span>
                        <span style={{ cursor: 'pointer' }} onClick={() => navigate(`/organizations/${project?.organization_id}/projects`)}>Projects</span>
                        <span> / {project?.name}</span>
                    </div>
                    <h1 className="page-title">Workspaces</h1>
                    <p className="page-subtitle">Manage infrastructure workspaces</p>
                </div>
                <button className="btn btn-primary" onClick={() => setShowModal(true)}>
                    + New Workspace
                </button>
            </div>

            {wsList.length === 0 ? (
                <div className="card">
                    <div className="empty-state">
                        <div className="empty-state-icon">🔧</div>
                        <h3 className="empty-state-title">No workspaces</h3>
                        <p className="empty-state-message">
                            Workspaces contain your Terraform configuration, variables, and state.
                        </p>
                        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
                            Create Workspace
                        </button>
                    </div>
                </div>
            ) : (
                <div className="table-container">
                    <table className="table">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Terraform Version</th>
                                <th>Auto Apply</th>
                                <th>Status</th>
                                <th>Updated</th>
                            </tr>
                        </thead>
                        <tbody>
                            {wsList.map(ws => (
                                <tr
                                    key={ws.id}
                                    onClick={() => navigate(`/workspaces/${ws.id}`)}
                                    style={{ cursor: 'pointer' }}
                                >
                                    <td>
                                        <div className="font-semibold" style={{ color: 'var(--color-text-primary)' }}>{ws.name}</div>
                                        {ws.description && <div className="text-xs text-muted mt-sm">{ws.description}</div>}
                                    </td>
                                    <td>
                                        <span className="badge badge-info">{ws.terraform_version}</span>
                                    </td>
                                    <td>
                                        <span className={`badge ${ws.auto_apply ? 'badge-success' : 'badge-neutral'}`}>
                                            {ws.auto_apply ? 'Enabled' : 'Disabled'}
                                        </span>
                                    </td>
                                    <td>
                                        {ws.locked ? (
                                            <span className="badge badge-warning">🔒 Locked</span>
                                        ) : (
                                            <span className="badge badge-success">Ready</span>
                                        )}
                                    </td>
                                    <td className="text-xs text-muted">
                                        {new Date(ws.updated_at).toLocaleDateString()}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {showModal && (
                <div className="modal-overlay" onClick={() => setShowModal(false)}>
                    <div className="modal" onClick={e => e.stopPropagation()}>
                        <h2 className="modal-title">Create Workspace</h2>
                        <form onSubmit={handleCreate}>
                            <div className="form-group">
                                <label className="form-label">Name</label>
                                <input
                                    className="form-input"
                                    placeholder="staging-infrastructure"
                                    value={newWs.name}
                                    onChange={e => setNewWs({ ...newWs, name: e.target.value })}
                                    required
                                />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Description</label>
                                <textarea
                                    className="form-input"
                                    placeholder="Optional description..."
                                    value={newWs.description}
                                    onChange={e => setNewWs({ ...newWs, description: e.target.value })}
                                    rows={2}
                                />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Terraform Version</label>
                                <input
                                    className="form-input form-input-mono"
                                    placeholder="latest"
                                    value={newWs.terraform_version}
                                    onChange={e => setNewWs({ ...newWs, terraform_version: e.target.value })}
                                />
                            </div>
                            <div className="form-group">
                                <label className="form-checkbox">
                                    <input
                                        type="checkbox"
                                        checked={newWs.auto_apply}
                                        onChange={e => setNewWs({ ...newWs, auto_apply: e.target.checked })}
                                    />
                                    Auto Apply — Automatically apply plans without confirmation
                                </label>
                            </div>
                            <div className="modal-actions">
                                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>Cancel</button>
                                <button type="submit" className="btn btn-primary">Create Workspace</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}
