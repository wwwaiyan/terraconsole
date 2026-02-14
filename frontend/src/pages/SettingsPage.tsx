import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { auth } from '../api/client';
import toast from 'react-hot-toast';

export default function SettingsPage() {
    const { user, refreshUser } = useAuth();
    const [mfaSetup, setMfaSetup] = useState<{ secret: string; qr_code: string; url: string } | null>(null);
    const [totpCode, setTotpCode] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSetupMFA = async () => {
        setLoading(true);
        try {
            const data = await auth.setupMFA() as { secret: string; qr_code: string; url: string };
            setMfaSetup(data);
        } catch (err: any) { toast.error(err.message); }
        finally { setLoading(false); }
    };

    const handleVerifyMFA = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        try {
            await auth.verifyMFA(totpCode);
            toast.success('MFA enabled successfully!');
            setMfaSetup(null);
            setTotpCode('');
            await refreshUser();
        } catch (err: any) { toast.error(err.message); }
        finally { setLoading(false); }
    };

    const handleDisableMFA = async () => {
        if (!confirm('Are you sure you want to disable MFA?')) return;
        setLoading(true);
        try {
            await auth.disableMFA();
            toast.success('MFA disabled');
            await refreshUser();
        } catch (err: any) { toast.error(err.message); }
        finally { setLoading(false); }
    };

    return (
        <div>
            <div className="page-header">
                <div>
                    <h1 className="page-title">Settings</h1>
                    <p className="page-subtitle">Manage your account and security</p>
                </div>
            </div>

            {/* Profile */}
            <div className="card mb-lg">
                <h3 className="card-title mb-lg">Profile</h3>
                <div className="grid grid-2">
                    <div className="form-group">
                        <label className="form-label">Full Name</label>
                        <input className="form-input" value={user?.full_name || ''} readOnly />
                    </div>
                    <div className="form-group">
                        <label className="form-label">Username</label>
                        <input className="form-input form-input-mono" value={user?.username || ''} readOnly />
                    </div>
                    <div className="form-group">
                        <label className="form-label">Email</label>
                        <input className="form-input" value={user?.email || ''} readOnly />
                    </div>
                    <div className="form-group">
                        <label className="form-label">Member Since</label>
                        <input className="form-input" value={user?.created_at ? new Date(user.created_at).toLocaleDateString() : ''} readOnly />
                    </div>
                </div>
            </div>

            {/* MFA */}
            <div className="card mb-lg">
                <h3 className="card-title mb-lg">🔐 Two-Factor Authentication</h3>

                {user?.mfa_enabled ? (
                    <div>
                        <div className="flex items-center gap-md mb-lg">
                            <span className="badge badge-success">✅ MFA Enabled</span>
                            <span className="text-sm text-muted">Your account is protected with TOTP-based 2FA</span>
                        </div>
                        <button className="btn btn-danger" onClick={handleDisableMFA} disabled={loading}>
                            Disable MFA
                        </button>
                    </div>
                ) : mfaSetup ? (
                    <div>
                        <p className="text-sm text-muted mb-lg">
                            Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)
                        </p>
                        <div style={{ textAlign: 'center', marginBottom: '24px' }}>
                            <img
                                src={mfaSetup.qr_code}
                                alt="MFA QR Code"
                                style={{ width: 200, height: 200, borderRadius: 12, border: '1px solid var(--color-border)' }}
                            />
                        </div>
                        <div className="form-group">
                            <label className="form-label">Manual Entry Key</label>
                            <input
                                className="form-input form-input-mono"
                                value={mfaSetup.secret}
                                readOnly
                                onClick={e => {
                                    (e.target as HTMLInputElement).select();
                                    navigator.clipboard.writeText(mfaSetup.secret);
                                    toast.success('Copied to clipboard!');
                                }}
                                style={{ cursor: 'pointer', textAlign: 'center' }}
                            />
                        </div>
                        <form onSubmit={handleVerifyMFA}>
                            <div className="form-group">
                                <label className="form-label">Enter the 6-digit code from your app</label>
                                <input
                                    className="form-input form-input-mono"
                                    placeholder="000000"
                                    value={totpCode}
                                    onChange={e => setTotpCode(e.target.value)}
                                    maxLength={6}
                                    style={{ textAlign: 'center', fontSize: '1.25rem', letterSpacing: '0.3em', maxWidth: 200 }}
                                    required
                                />
                            </div>
                            <div className="flex gap-md">
                                <button type="submit" className="btn btn-primary" disabled={loading || totpCode.length !== 6}>
                                    {loading ? 'Verifying...' : 'Enable MFA'}
                                </button>
                                <button type="button" className="btn btn-secondary" onClick={() => setMfaSetup(null)}>
                                    Cancel
                                </button>
                            </div>
                        </form>
                    </div>
                ) : (
                    <div>
                        <p className="text-sm text-muted mb-lg">
                            Add an extra layer of security to your account using a Time-based One-Time Password (TOTP) app.
                        </p>
                        <button className="btn btn-primary" onClick={handleSetupMFA} disabled={loading}>
                            {loading ? 'Setting up...' : '🔐 Set Up MFA'}
                        </button>
                    </div>
                )}
            </div>

            {/* API Tokens */}
            <div className="card">
                <h3 className="card-title mb-lg">🔑 API Tokens</h3>
                <p className="text-sm text-muted mb-lg">
                    Generate API tokens for CLI access and CI/CD integrations. Tokens provide the same access as your user account.
                </p>
                <div className="empty-state" style={{ padding: '32px 0' }}>
                    <div className="empty-state-icon">🔑</div>
                    <h3 className="empty-state-title">No API tokens</h3>
                    <p className="empty-state-message">Create an API token to authenticate from the CLI or CI/CD pipelines.</p>
                    <button className="btn btn-primary btn-sm">Generate Token</button>
                </div>
            </div>
        </div>
    );
}
