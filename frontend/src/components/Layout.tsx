import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { orgs } from '../api/client';
import { Organization } from '../types';

export default function Layout({ children }: { children: React.ReactNode }) {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const [organizations, setOrganizations] = useState<Organization[]>([]);
    const [sidebarOpen, setSidebarOpen] = useState(false);

    useEffect(() => {
        loadOrgs();
    }, []);

    // Auto-close sidebar on route change (mobile)
    useEffect(() => {
        setSidebarOpen(false);
    }, [location.pathname]);

    // Close sidebar on Escape key
    useEffect(() => {
        const handler = (e: KeyboardEvent) => {
            if (e.key === 'Escape') setSidebarOpen(false);
        };
        window.addEventListener('keydown', handler);
        return () => window.removeEventListener('keydown', handler);
    }, []);

    const loadOrgs = async () => {
        try {
            const data = await orgs.list() as Organization[];
            setOrganizations(data || []);
        } catch { }
    };

    const isActive = (path: string) => location.pathname.startsWith(path);

    return (
        <div className="app-layout">
            {/* Mobile overlay with blur */}
            {sidebarOpen && (
                <div
                    style={{
                        position: 'fixed',
                        inset: 0,
                        background: 'rgba(0,0,0,0.6)',
                        backdropFilter: 'blur(4px)',
                        WebkitBackdropFilter: 'blur(4px)',
                        zIndex: 99,
                        animation: 'fadeIn 0.2s ease-out',
                    }}
                    onClick={() => setSidebarOpen(false)}
                />
            )}

            {/* Sidebar */}
            <aside className={`sidebar ${sidebarOpen ? 'open' : ''}`}>
                <div className="sidebar-logo" onClick={() => navigate('/')} style={{ cursor: 'pointer' }}>
                    <div className="logo-icon">TC</div>
                    <h1>TerraConsole</h1>
                </div>

                <nav className="sidebar-nav">
                    <div className="sidebar-section">
                        <p className="sidebar-section-title">Navigation</p>
                        <button
                            className={`sidebar-link ${location.pathname === '/' ? 'active' : ''}`}
                            onClick={() => navigate('/')}
                        >
                            <span className="icon">📊</span> Dashboard
                        </button>
                        <button
                            className={`sidebar-link ${isActive('/organizations') ? 'active' : ''}`}
                            onClick={() => navigate('/organizations')}
                        >
                            <span className="icon">🏢</span> Organizations
                        </button>
                        <button
                            className={`sidebar-link ${isActive('/settings') ? 'active' : ''}`}
                            onClick={() => navigate('/settings')}
                        >
                            <span className="icon">⚙️</span> Settings
                        </button>
                    </div>

                    {organizations.length > 0 && (
                        <div className="sidebar-section">
                            <p className="sidebar-section-title">Organizations</p>
                            {organizations.map(org => (
                                <button
                                    key={org.id}
                                    className={`sidebar-link ${isActive(`/organizations/${org.id}`) ? 'active' : ''}`}
                                    onClick={() => navigate(`/organizations/${org.id}/projects`)}
                                >
                                    <span className="icon">📁</span> {org.display_name || org.name}
                                </button>
                            ))}
                        </div>
                    )}
                </nav>

                <div className="sidebar-user">
                    <div className="sidebar-user-avatar">
                        {user?.full_name?.charAt(0) || user?.username?.charAt(0) || '?'}
                    </div>
                    <div className="sidebar-user-info">
                        <div className="sidebar-user-name">{user?.full_name || user?.username}</div>
                        <div className="sidebar-user-email">{user?.email}</div>
                    </div>
                    <button className="btn btn-ghost btn-sm" onClick={logout} title="Logout">
                        🚪
                    </button>
                </div>
            </aside>

            {/* Main Content */}
            <main className="main-content">
                <header className="topbar">
                    <div className="topbar-left">
                        <button
                            className="btn btn-ghost btn-sm mobile-menu-btn"
                            onClick={() => setSidebarOpen(!sidebarOpen)}
                            aria-label="Toggle menu"
                        >
                            {sidebarOpen ? '✕' : '☰'}
                        </button>
                        <div className="topbar-breadcrumb">
                            <span>TerraConsole</span>
                        </div>
                    </div>
                    <div className="topbar-right">
                        <span className="text-xs text-muted">
                            {user?.mfa_enabled ? '🔒 MFA Enabled' : ''}
                        </span>
                    </div>
                </header>

                <div className="page-content">
                    {children}
                </div>
            </main>

            <style>{`
                .mobile-menu-btn { display: none !important; }
                @media (max-width: 768px) {
                    .mobile-menu-btn { display: flex !important; font-size: 1.25rem; }
                }
            `}</style>
        </div>
    );
}
