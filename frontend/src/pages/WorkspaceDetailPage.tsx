import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { workspaces, runs, variables, state } from '../api/client';
import { Workspace, Run, Variable, StateVersion, RunStatus } from '../types';
import toast from 'react-hot-toast';

const STATUS_CONFIG: Record<RunStatus, { label: string; class: string; icon: string }> = {
    pending: { label: 'Pending', class: 'badge-pending', icon: '⏳' },
    planning: { label: 'Planning', class: 'badge-info badge-pulse', icon: '🔄' },
    planned: { label: 'Planned', class: 'badge-info', icon: '📋' },
    needs_confirmation: { label: 'Needs Confirm', class: 'badge-warning', icon: '⚠️' },
    applying: { label: 'Applying', class: 'badge-info badge-pulse', icon: '🔄' },
    applied: { label: 'Applied', class: 'badge-success', icon: '✅' },
    errored: { label: 'Errored', class: 'badge-error', icon: '❌' },
    cancelled: { label: 'Cancelled', class: 'badge-neutral', icon: '🚫' },
    discarded: { label: 'Discarded', class: 'badge-neutral', icon: '🗑️' },
    planned_and_finished: { label: 'Plan Only', class: 'badge-success', icon: '📋' },
};

export default function WorkspaceDetailPage() {
    const { workspaceId } = useParams<{ workspaceId: string }>();
    const navigate = useNavigate();
    const [workspace, setWorkspace] = useState<Workspace | null>(null);
    const [runList, setRunList] = useState<Run[]>([]);
    const [varList, setVarList] = useState<Variable[]>([]);
    const [stateVersions, setStateVersions] = useState<StateVersion[]>([]);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState('runs');
    const [showRunModal, setShowRunModal] = useState(false);
    const [showVarModal, setShowVarModal] = useState(false);
    const [newRun, setNewRun] = useState({ operation: 'plan' as string, message: '' });
    const [newVar, setNewVar] = useState({ key: '', value: '', category: 'terraform' as string, sensitive: false, hcl: false, description: '' });

    useEffect(() => { loadAll(); }, [workspaceId]);

    const loadAll = async () => {
        try {
            const [wsData, runData, varData, stateData] = await Promise.all([
                workspaces.get(workspaceId!),
                runs.list(workspaceId!),
                variables.list(workspaceId!),
                state.listVersions(workspaceId!),
            ]) as [Workspace, Run[], Variable[], StateVersion[]];
            setWorkspace(wsData);
            setRunList(runData || []);
            setVarList(varData || []);
            setStateVersions(stateData || []);
        } catch { } finally { setLoading(false); }
    };

    const handleCreateRun = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await runs.create(workspaceId!, { operation: newRun.operation, message: newRun.message });
            toast.success('Run created!');
            setShowRunModal(false);
            setNewRun({ operation: 'plan', message: '' });
            loadAll();
        } catch (err: any) { toast.error(err.message); }
    };

    const handleCreateVar = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await variables.create(workspaceId!, newVar);
            toast.success('Variable created!');
            setShowVarModal(false);
            setNewVar({ key: '', value: '', category: 'terraform', sensitive: false, hcl: false, description: '' });
            loadAll();
        } catch (err: any) { toast.error(err.message); }
    };

    const handleDeleteVar = async (varId: string) => {
        if (!confirm('Delete this variable?')) return;
        try {
            await variables.delete(workspaceId!, varId);
            toast.success('Variable deleted');
            loadAll();
        } catch (err: any) { toast.error(err.message); }
    };

    const handleLockToggle = async () => {
        try {
            if (workspace?.locked) {
                await workspaces.unlock(workspaceId!);
                toast.success('Workspace unlocked');
            } else {
                await workspaces.lock(workspaceId!);
                toast.success('Workspace locked');
            }
            loadAll();
        } catch (err: any) { toast.error(err.message); }
    };

    if (loading || !workspace) {
        return <div className="loading-page"><div className="loading-spinner loading-spinner-lg"></div></div>;
    }

    return (
        <div>
            <div className="page-header">
                <div>
                    <div className="text-xs text-muted" style={{ marginBottom: 4, cursor: 'pointer' }}
                        onClick={() => navigate(`/projects/${workspace.project_id}/workspaces`)}>
                        ← Back to Workspaces
                    </div>
                    <h1 className="page-title flex items-center gap-md">
                        {workspace.name}
                        {workspace.locked && <span className="badge badge-warning">🔒 Locked</span>}
                    </h1>
                    <p className="page-subtitle">{workspace.description || 'No description'}</p>
                </div>
                <div className="flex gap-md">
                    <button className="btn btn-secondary" onClick={handleLockToggle}>
                        {workspace.locked ? '🔓 Unlock' : '🔒 Lock'}
                    </button>
                    <button className="btn btn-primary" onClick={() => setShowRunModal(true)} disabled={workspace.locked}>
                        ▶ Start Run
                    </button>
                </div>
            </div>

            {/* Info Bar */}
            <div className="grid grid-4" style={{ marginBottom: '24px' }}>
                <div className="stat-card">
                    <div className="stat-value text-mono" style={{ fontSize: '1.25rem' }}>{workspace.terraform_version}</div>
                    <div className="stat-label">Terraform Version</div>
                </div>
                <div className="stat-card">
                    <div className="stat-value">{runList.length}</div>
                    <div className="stat-label">Total Runs</div>
                </div>
                <div className="stat-card">
                    <div className="stat-value">{varList.length}</div>
                    <div className="stat-label">Variables</div>
                </div>
                <div className="stat-card">
                    <div className="stat-value">{stateVersions.length}</div>
                    <div className="stat-label">State Versions</div>
                </div>
            </div>

            {/* Tabs */}
            <div className="tabs">
                {['runs', 'variables', 'state', 'settings'].map(tab => (
                    <button
                        key={tab}
                        className={`tab ${activeTab === tab ? 'active' : ''}`}
                        onClick={() => setActiveTab(tab)}
                    >
                        {tab.charAt(0).toUpperCase() + tab.slice(1)}
                    </button>
                ))}
            </div>

            {/* Runs Tab */}
            {activeTab === 'runs' && (
                <div>
                    {runList.length === 0 ? (
                        <div className="card">
                            <div className="empty-state">
                                <div className="empty-state-icon">🚀</div>
                                <h3 className="empty-state-title">No runs yet</h3>
                                <p className="empty-state-message">Start a run to plan or apply your infrastructure changes.</p>
                                <button className="btn btn-primary" onClick={() => setShowRunModal(true)}>Start First Run</button>
                            </div>
                        </div>
                    ) : (
                        <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
                            {runList.map(run => {
                                const cfg = STATUS_CONFIG[run.status] || STATUS_CONFIG.pending;
                                return (
                                    <div key={run.id} className="run-item" onClick={() => navigate(`/runs/${run.id}`)}>
                                        <div className="run-status-icon" style={{
                                            background: `var(--color-${run.status === 'applied' ? 'success' : run.status === 'errored' ? 'error' : 'info'}-bg, rgba(59,130,246,0.12))`,
                                        }}>
                                            {cfg.icon}
                                        </div>
                                        <div className="run-info">
                                            <div className="run-info-title">
                                                {run.operation === 'destroy' ? '🗑️ Destroy' :
                                                    run.operation === 'refresh' ? '🔄 Refresh' :
                                                        run.operation === 'plan' ? '📋 Plan Only' : '🚀 Plan & Apply'}
                                            </div>
                                            <div className="run-info-meta">
                                                {run.message || 'No message'} · {run.creator?.username || 'Unknown'} · {new Date(run.created_at).toLocaleString()}
                                            </div>
                                        </div>
                                        <div className="run-changes">
                                            {run.resources_added > 0 && <span className="run-change-add">+{run.resources_added}</span>}
                                            {run.resources_changed > 0 && <span className="run-change-modify">~{run.resources_changed}</span>}
                                            {run.resources_deleted > 0 && <span className="run-change-destroy">-{run.resources_deleted}</span>}
                                        </div>
                                        <span className={`badge ${cfg.class}`}>{cfg.label}</span>
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>
            )}

            {/* Variables Tab */}
            {activeTab === 'variables' && (
                <div>
                    <div className="flex justify-between items-center mb-lg">
                        <h3 className="font-semibold">Workspace Variables</h3>
                        <button className="btn btn-primary btn-sm" onClick={() => setShowVarModal(true)}>
                            + Add Variable
                        </button>
                    </div>
                    {varList.length === 0 ? (
                        <div className="card">
                            <div className="empty-state">
                                <div className="empty-state-icon">📦</div>
                                <h3 className="empty-state-title">No variables</h3>
                                <p className="empty-state-message">Add Terraform and environment variables for this workspace.</p>
                            </div>
                        </div>
                    ) : (
                        <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
                            {varList.map(v => (
                                <div key={v.id} className="var-row">
                                    <div className="var-key">{v.key}</div>
                                    <div className={v.sensitive ? 'var-sensitive' : 'var-value'}>
                                        {v.sensitive ? '***SENSITIVE***' : v.value || '(empty)'}
                                    </div>
                                    <div className="flex gap-sm">
                                        <span className="var-badge">{v.category}</span>
                                        {v.sensitive && <span className="var-badge" style={{ color: 'var(--color-warning)' }}>🔒 secret</span>}
                                        {v.hcl && <span className="var-badge">HCL</span>}
                                    </div>
                                    <button className="btn btn-ghost btn-sm" onClick={() => handleDeleteVar(v.id)}>🗑️</button>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}

            {/* State Tab */}
            {activeTab === 'state' && (
                <div>
                    {stateVersions.length === 0 ? (
                        <div className="card">
                            <div className="empty-state">
                                <div className="empty-state-icon">📄</div>
                                <h3 className="empty-state-title">No state versions</h3>
                                <p className="empty-state-message">State will appear here after a successful apply.</p>
                            </div>
                        </div>
                    ) : (
                        <div className="table-container">
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>Serial</th>
                                        <th>Lineage</th>
                                        <th>Resources</th>
                                        <th>Created</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {stateVersions.map(sv => (
                                        <tr key={sv.id}>
                                            <td className="font-semibold text-mono">#{sv.serial}</td>
                                            <td className="text-mono text-xs">{sv.lineage || '—'}</td>
                                            <td>{sv.resource_count}</td>
                                            <td className="text-xs text-muted">{new Date(sv.created_at).toLocaleString()}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            )}

            {/* Settings Tab */}
            {activeTab === 'settings' && (
                <div>
                    <div className="card mb-lg">
                        <h3 className="card-title mb-lg">Workspace Settings</h3>
                        <div className="grid grid-2">
                            <div className="form-group">
                                <label className="form-label">Workspace Name</label>
                                <input className="form-input" value={workspace.name} readOnly />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Terraform Version</label>
                                <input className="form-input form-input-mono" value={workspace.terraform_version} readOnly />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Working Directory</label>
                                <input className="form-input form-input-mono" value={workspace.working_directory} readOnly />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Execution Mode</label>
                                <input className="form-input" value={workspace.execution_mode} readOnly />
                            </div>
                        </div>
                    </div>
                    <div className="card" style={{ borderColor: 'var(--color-error)' }}>
                        <h3 className="card-title" style={{ color: 'var(--color-error)' }}>Danger Zone</h3>
                        <p className="card-description mb-lg">
                            Once you delete a workspace, there is no going back. This will destroy all runs, variables, and state.
                        </p>
                        <button className="btn btn-danger" onClick={async () => {
                            if (confirm('Are you sure you want to delete this workspace?')) {
                                await workspaces.delete(workspaceId!);
                                toast.success('Workspace deleted');
                                navigate(`/projects/${workspace.project_id}/workspaces`);
                            }
                        }}>
                            Delete Workspace
                        </button>
                    </div>
                </div>
            )}

            {/* Create Run Modal */}
            {showRunModal && (
                <div className="modal-overlay" onClick={() => setShowRunModal(false)}>
                    <div className="modal" onClick={e => e.stopPropagation()}>
                        <h2 className="modal-title">Start a New Run</h2>
                        <form onSubmit={handleCreateRun}>
                            <div className="form-group">
                                <label className="form-label">Operation</label>
                                <select
                                    className="form-input form-select"
                                    value={newRun.operation}
                                    onChange={e => setNewRun({ ...newRun, operation: e.target.value })}
                                >
                                    <option value="plan">📋 Plan Only</option>
                                    <option value="plan_and_apply">🚀 Plan & Apply</option>
                                    <option value="destroy">🗑️ Destroy</option>
                                    <option value="refresh">🔄 Refresh State</option>
                                </select>
                            </div>
                            <div className="form-group">
                                <label className="form-label">Message (optional)</label>
                                <input
                                    className="form-input"
                                    placeholder="Describe this run..."
                                    value={newRun.message}
                                    onChange={e => setNewRun({ ...newRun, message: e.target.value })}
                                />
                            </div>
                            <div className="modal-actions">
                                <button type="button" className="btn btn-secondary" onClick={() => setShowRunModal(false)}>Cancel</button>
                                <button type="submit" className={`btn ${newRun.operation === 'destroy' ? 'btn-danger' : 'btn-primary'}`}>
                                    {newRun.operation === 'destroy' ? '🗑️ Start Destroy' : '▶ Start Run'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {/* Create Variable Modal */}
            {showVarModal && (
                <div className="modal-overlay" onClick={() => setShowVarModal(false)}>
                    <div className="modal" onClick={e => e.stopPropagation()}>
                        <h2 className="modal-title">Add Variable</h2>
                        <form onSubmit={handleCreateVar}>
                            <div className="form-group">
                                <label className="form-label">Key</label>
                                <input
                                    className="form-input form-input-mono"
                                    placeholder="variable_name"
                                    value={newVar.key}
                                    onChange={e => setNewVar({ ...newVar, key: e.target.value })}
                                    required
                                />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Value</label>
                                <textarea
                                    className="form-input form-input-mono"
                                    placeholder="variable value"
                                    value={newVar.value}
                                    onChange={e => setNewVar({ ...newVar, value: e.target.value })}
                                    rows={3}
                                />
                            </div>
                            <div className="form-group">
                                <label className="form-label">Category</label>
                                <select
                                    className="form-input form-select"
                                    value={newVar.category}
                                    onChange={e => setNewVar({ ...newVar, category: e.target.value })}
                                >
                                    <option value="terraform">Terraform Variable</option>
                                    <option value="env">Environment Variable</option>
                                </select>
                            </div>
                            <div className="form-group">
                                <label className="form-label">Description</label>
                                <input
                                    className="form-input"
                                    placeholder="Optional description"
                                    value={newVar.description}
                                    onChange={e => setNewVar({ ...newVar, description: e.target.value })}
                                />
                            </div>
                            <div className="flex gap-lg">
                                <label className="form-checkbox">
                                    <input
                                        type="checkbox"
                                        checked={newVar.sensitive}
                                        onChange={e => setNewVar({ ...newVar, sensitive: e.target.checked })}
                                    />
                                    Sensitive
                                </label>
                                <label className="form-checkbox">
                                    <input
                                        type="checkbox"
                                        checked={newVar.hcl}
                                        onChange={e => setNewVar({ ...newVar, hcl: e.target.checked })}
                                    />
                                    HCL
                                </label>
                            </div>
                            <div className="modal-actions">
                                <button type="button" className="btn btn-secondary" onClick={() => setShowVarModal(false)}>Cancel</button>
                                <button type="submit" className="btn btn-primary">Add Variable</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}
