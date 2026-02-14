import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import Layout from './components/Layout';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';
import DashboardPage from './pages/DashboardPage';
import OrganizationsPage from './pages/OrganizationsPage';
import ProjectsPage from './pages/ProjectsPage';
import WorkspacesPage from './pages/WorkspacesPage';
import WorkspaceDetailPage from './pages/WorkspaceDetailPage';
import RunDetailPage from './pages/RunDetailPage';
import SettingsPage from './pages/SettingsPage';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
    const { user, loading } = useAuth();

    if (loading) {
        return (
            <div className="loading-page">
                <div className="loading-spinner loading-spinner-lg"></div>
                <p className="text-muted">Loading TerraConsole...</p>
            </div>
        );
    }

    if (!user) {
        return <Navigate to="/login" replace />;
    }

    return <>{children}</>;
}

export default function App() {
    return (
        <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/signup" element={<SignupPage />} />
            <Route
                path="/*"
                element={
                    <ProtectedRoute>
                        <Layout>
                            <Routes>
                                <Route path="/" element={<DashboardPage />} />
                                <Route path="/organizations" element={<OrganizationsPage />} />
                                <Route path="/organizations/:orgId/projects" element={<ProjectsPage />} />
                                <Route path="/projects/:projectId/workspaces" element={<WorkspacesPage />} />
                                <Route path="/workspaces/:workspaceId/*" element={<WorkspaceDetailPage />} />
                                <Route path="/runs/:runId" element={<RunDetailPage />} />
                                <Route path="/settings" element={<SettingsPage />} />
                            </Routes>
                        </Layout>
                    </ProtectedRoute>
                }
            />
        </Routes>
    );
}
