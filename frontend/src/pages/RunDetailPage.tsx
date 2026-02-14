import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { runs } from '../api/client';
import { Run, RunStatus } from '../types';
import toast from 'react-hot-toast';

const STATUS_CONFIG: Record<RunStatus, { label: string; class: string; icon: string }> = {
    pending: { label: 'Pending', class: 'badge-pending', icon: '⏳' },
    planning: { label: 'Planning', class: 'badge-info badge-pulse', icon: '🔄' },
    planned: { label: 'Planned', class: 'badge-info', icon: '📋' },
    needs_confirmation: { label: 'Needs Confirmation', class: 'badge-warning', icon: '⚠️' },
    applying: { label: 'Applying', class: 'badge-info badge-pulse', icon: '🔄' },
    applied: { label: 'Applied', class: 'badge-success', icon: '✅' },
    errored: { label: 'Errored', class: 'badge-error', icon: '❌' },
    cancelled: { label: 'Cancelled', class: 'badge-neutral', icon: '🚫' },
    discarded: { label: 'Discarded', class: 'badge-neutral', icon: '🗑️' },
    planned_and_finished: { label: 'Plan Only', class: 'badge-success', icon: '📋' },
};

export default function RunDetailPage() {
    const { runId } = useParams<{ runId: string }>();
    const navigate = useNavigate();
    const [run, setRun] = useState<Run | null>(null);
    const [planLog, setPlanLog] = useState('');
    const [applyLog, setApplyLog] = useState('');
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState('plan');

    useEffect(() => { loadRun(); }, [runId]);

    const loadRun = async () => {
        try {
            const runData = await runs.get(runId!) as Run;
            setRun(runData);

            try {
                const planData = await runs.getPlanLog(runId!) as { log: string };
                setPlanLog(planData.log || '');
            } catch { }

            try {
                const applyData = await runs.getApplyLog(runId!) as { log: string };
                setApplyLog(applyData.log || '');
            } catch { }
        } catch { } finally { setLoading(false); }
    };

    const handleApprove = async () => {
        try {
            await runs.approve(runId!);
            toast.success('Run approved! Applying...');
            loadRun();
        } catch (err: any) { toast.error(err.message); }
    };

    const handleDiscard = async () => {
        if (!confirm('Are you sure you want to discard this run?')) return;
        try {
            await runs.discard(runId!);
            toast.success('Run discarded');
            loadRun();
        } catch (err: any) { toast.error(err.message); }
    };

    const handleCancel = async () => {
        if (!confirm('Cancel this run?')) return;
        try {
            await runs.cancel(runId!);
            toast.success('Run cancelled');
            loadRun();
        } catch (err: any) { toast.error(err.message); }
    };

    if (loading || !run) {
        return <div className="loading-page"><div className="loading-spinner loading-spinner-lg"></div></div>;
    }

    const cfg = STATUS_CONFIG[run.status] || STATUS_CONFIG.pending;

    return (
        <div>
            <div className="page-header">
                <div>
                    <div className="text-xs text-muted" style={{ marginBottom: 4, cursor: 'pointer' }}
                        onClick={() => navigate(`/workspaces/${run.workspace_id}`)}>
                        ← Back to Workspace
                    </div>
                    <h1 className="page-title flex items-center gap-md">
                        {run.operation === 'destroy' ? '🗑️ Destroy Run' :
                            run.operation === 'refresh' ? '🔄 Refresh Run' :
                                run.operation === 'plan' ? '📋 Plan Only' : '🚀 Plan & Apply'}
                        <span className={`badge ${cfg.class}`}>{cfg.icon} {cfg.label}</span>
                    </h1>
                    <p className="page-subtitle">{run.message || 'No message'} · by {run.creator?.username || 'Unknown'}</p>
                </div>
                <div className="flex gap-md">
                    {run.status === 'needs_confirmation' && (
                        <>
                            <button className="btn btn-success" onClick={handleApprove}>
                                ✅ Confirm & Apply
                            </button>
                            <button className="btn btn-danger" onClick={handleDiscard}>
                                🗑️ Discard
                            </button>
                        </>
                    )}
                    {(run.status === 'pending' || run.status === 'planning') && (
                        <button className="btn btn-secondary" onClick={handleCancel}>
                            Cancel Run
                        </button>
                    )}
                </div>
            </div>

            {/* Run Summary */}
            <div className="grid grid-4" style={{ marginBottom: '24px' }}>
                <div className="stat-card">
                    <div className="stat-value run-change-add">+{run.resources_added}</div>
                    <div className="stat-label">To Add</div>
                </div>
                <div className="stat-card">
                    <div className="stat-value run-change-modify">~{run.resources_changed}</div>
                    <div className="stat-label">To Change</div>
                </div>
                <div className="stat-card">
                    <div className="stat-value run-change-destroy">-{run.resources_deleted}</div>
                    <div className="stat-label">To Destroy</div>
                </div>
                <div className="stat-card">
                    <div className="stat-value text-mono" style={{ fontSize: '1.125rem' }}>{run.terraform_version}</div>
                    <div className="stat-label">TF Version</div>
                </div>
            </div>

            {/* Timeline */}
            <div className="card mb-lg">
                <h3 className="card-title mb-lg">Timeline</h3>
                <div className="flex flex-col gap-md">
                    <div className="flex items-center gap-md text-sm">
                        <span style={{ width: 80 }} className="text-muted">Created</span>
                        <span>{new Date(run.created_at).toLocaleString()}</span>
                    </div>
                    {run.started_at && (
                        <div className="flex items-center gap-md text-sm">
                            <span style={{ width: 80 }} className="text-muted">Started</span>
                            <span>{new Date(run.started_at).toLocaleString()}</span>
                        </div>
                    )}
                    {run.plan_completed_at && (
                        <div className="flex items-center gap-md text-sm">
                            <span style={{ width: 80 }} className="text-muted">Planned</span>
                            <span>{new Date(run.plan_completed_at).toLocaleString()}</span>
                        </div>
                    )}
                    {run.applied_at && (
                        <div className="flex items-center gap-md text-sm">
                            <span style={{ width: 80 }} className="text-muted">Applied</span>
                            <span>{new Date(run.applied_at).toLocaleString()}</span>
                        </div>
                    )}
                    {run.completed_at && (
                        <div className="flex items-center gap-md text-sm">
                            <span style={{ width: 80 }} className="text-muted">Completed</span>
                            <span>{new Date(run.completed_at).toLocaleString()}</span>
                        </div>
                    )}
                </div>
            </div>

            {/* Output Tabs */}
            <div className="tabs">
                <button className={`tab ${activeTab === 'plan' ? 'active' : ''}`} onClick={() => setActiveTab('plan')}>
                    Plan Output
                </button>
                <button className={`tab ${activeTab === 'apply' ? 'active' : ''}`} onClick={() => setActiveTab('apply')}>
                    Apply Output
                </button>
            </div>

            {activeTab === 'plan' && (
                <div className="terminal">
                    {planLog || (
                        <span style={{ color: 'var(--color-text-muted)' }}>
                            {run.status === 'pending' ? 'Waiting for run to start...' :
                                run.status === 'planning' ? 'Plan is running...' :
                                    'No plan output available.'}
                        </span>
                    )}
                </div>
            )}

            {activeTab === 'apply' && (
                <div className="terminal">
                    {applyLog || (
                        <span style={{ color: 'var(--color-text-muted)' }}>
                            {run.status === 'needs_confirmation' ? 'Waiting for approval...' :
                                run.status === 'applying' ? 'Apply is running...' :
                                    'No apply output available.'}
                        </span>
                    )}
                </div>
            )}
        </div>
    );
}
