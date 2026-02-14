import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { projects, orgs } from '../api/client';
import { Project, Organization } from '../types';
import toast from 'react-hot-toast';

export default function ProjectsPage() {
    const { orgId } = useParams<{ orgId: string }>();
    const navigate = useNavigate();
    const [org, setOrg] = useState<Organization | null>(null);
    const [projectList, setProjectList] = useState<Project[]>([]);
    const [loading, setLoading] = useState(true);
    const [showModal, setShowModal] = useState(false);
    const [newProject, setNewProject] = useState({ name: '', description: '' });

    useEffect(() => { loadData(); }, [orgId]);

    const loadData = async () => {
        try {
            const [orgData, projData] = await Promise.all([
                orgs.get(orgId!) as Promise<Organization>,
                projects.list(orgId!) as Promise<Project[]>,
            ]);
            setOrg(orgData);
            setProjectList(projData || []);
        } catch { } finally { setLoading(false); }
    };

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await projects.create(orgId!, newProject);
            toast.success('Project created!');
            setShowModal(false);
            setNewProject({ name: '', description: '' });
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
                    <div className="text-xs text-muted mb-md" style={{ marginBottom: 4 }}>
                        <span style={{ cursor: 'pointer' }} onClick={() => navigate('/organizations')}>Organizations</span>
                        <span> / {org?.display_name || org?.name}</span>
                    </div>
                    <h1 className="page-title">Projects</h1>
                    <p className="page-subtitle">Group your workspaces by project</p>
                </div>
                <button className="btn btn-primary" onClick={() => setShowModal(true)}>
                    + New Project
                </button>
            </div>

            {projectList.length === 0 ? (
                <div className="card">
                    <div className="empty-state">
                        <div className="empty-state-icon">📁</div>
                        <h3 className="empty-state-title">No projects yet</h3>
                        <p className="empty-state-message">
                            Projects help organize your workspaces. Create one to start managing infrastructure.
                        </p>
                        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
                            Create First Project
                        </button>
                    </div>
                </div>
            ) : (
                <div className="grid grid-3">
                    {projectList.map(project => (
                        <div
                            key={project.id}
                            className="card card-clickable"
                            onClick={() => navigate(`/projects/${project.id}/workspaces`)}
                        >
                            <div className="flex items-center gap-md mb-md">
                                <div style={{
                                    width: 40, height: 40, borderRadius: 10,
                                    background: 'linear-gradient(135deg, var(--color-teal), var(--color-cyan))',
                                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                                    fontSize: '1rem', color: 'white', fontWeight: 600,
                                }}>
                                    📁
                                </div>
                                <div className="font-semibold">{project.name}</div>
                            </div>
                            {project.description && <p className="card-description">{project.description}</p>}
                        </div>
                    ))}
                </div>
            )}

            {showModal && (
                <div className="modal-overlay" onClick={() => setShowModal(false)}>
                    <div className="modal" onClick={e => e.stopPropagation()}>
                        <h2 className="modal-title">Create Project</h2>
                        <form onSubmit={handleCreate}>
                            <div className="form-group">
                                <label className="form-label">Name</label>
                                <input
                                    className="form-input"
                                    placeholder="my-project"
                                    value={newProject.name}
                                    onChange={e => setNewProject({ ...newProject, name: e.target.value })}
                                    required
                                />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Description</label>
                                <textarea
                                    className="form-input"
                                    placeholder="Optional description..."
                                    value={newProject.description}
                                    onChange={e => setNewProject({ ...newProject, description: e.target.value })}
                                    rows={3}
                                />
                            </div>
                            <div className="modal-actions">
                                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>Cancel</button>
                                <button type="submit" className="btn btn-primary">Create Project</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}
