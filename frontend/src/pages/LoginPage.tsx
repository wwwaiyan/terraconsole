import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import toast from 'react-hot-toast';

export default function LoginPage() {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [totpCode, setTotpCode] = useState('');
    const [mfaRequired, setMfaRequired] = useState(false);
    const [loading, setLoading] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        try {
            await login(email, password, mfaRequired ? totpCode : undefined);
            toast.success('Welcome back!');
            navigate('/');
        } catch (err: any) {
            if (err.message?.includes('MFA code required')) {
                setMfaRequired(true);
                toast('Please enter your MFA code', { icon: '🔐' });
            } else {
                toast.error(err.message || 'Login failed');
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="auth-page">
            <div className="auth-card">
                <div className="auth-logo">
                    <div style={{ fontSize: '2.5rem', marginBottom: '8px' }}>🏗️</div>
                    <h1>TerraConsole</h1>
                    <p>Sign in to manage your infrastructure</p>
                </div>

                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <label className="form-label" htmlFor="email">Email</label>
                        <input
                            id="email"
                            className="form-input"
                            type="email"
                            placeholder="you@company.com"
                            value={email}
                            onChange={e => setEmail(e.target.value)}
                            required
                            autoFocus
                        />
                    </div>

                    <div className="form-group">
                        <label className="form-label" htmlFor="password">Password</label>
                        <input
                            id="password"
                            className="form-input"
                            type="password"
                            placeholder="••••••••"
                            value={password}
                            onChange={e => setPassword(e.target.value)}
                            required
                        />
                    </div>

                    {mfaRequired && (
                        <div className="form-group" style={{ animation: 'slideUp 0.3s ease-out' }}>
                            <label className="form-label" htmlFor="totp">
                                🔐 Two-Factor Authentication Code
                            </label>
                            <input
                                id="totp"
                                className="form-input form-input-mono"
                                type="text"
                                placeholder="000000"
                                value={totpCode}
                                onChange={e => setTotpCode(e.target.value)}
                                maxLength={6}
                                autoFocus
                                style={{ textAlign: 'center', fontSize: '1.25rem', letterSpacing: '0.3em' }}
                            />
                        </div>
                    )}

                    <button
                        type="submit"
                        className="btn btn-primary btn-lg w-full"
                        disabled={loading}
                        style={{ marginTop: '8px' }}
                    >
                        {loading ? (
                            <><span className="loading-spinner"></span> Signing in...</>
                        ) : (
                            'Sign In'
                        )}
                    </button>
                </form>

                <div className="auth-footer">
                    Don't have an account?{' '}
                    <Link to="/signup">Create one</Link>
                </div>
            </div>
        </div>
    );
}
