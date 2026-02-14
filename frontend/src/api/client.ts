const API_BASE = '/api';

async function request<T>(url: string, options: RequestInit = {}): Promise<T> {
    const token = localStorage.getItem('token');

    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...((options.headers as Record<string, string>) || {}),
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE}${url}`, {
        ...options,
        headers,
    });

    if (response.status === 401) {
        localStorage.removeItem('token');
        window.location.href = '/login';
        throw new Error('Unauthorized');
    }

    const data = await response.json();

    if (!response.ok) {
        throw new Error(data.error || 'Request failed');
    }

    return data as T;
}

// Auth
export const auth = {
    signup: (data: { email: string; username: string; password: string; full_name: string }) =>
        request('/auth/signup', { method: 'POST', body: JSON.stringify(data) }),
    login: (data: { email: string; password: string; totp_code?: string }) =>
        request('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
    me: () => request('/auth/me'),
    refresh: () => request('/auth/refresh', { method: 'POST' }),
    setupMFA: () => request('/auth/mfa/setup', { method: 'POST' }),
    verifyMFA: (code: string) =>
        request('/auth/mfa/verify', { method: 'POST', body: JSON.stringify({ code }) }),
    disableMFA: () => request('/auth/mfa/disable', { method: 'POST' }),
};

// Organizations
export const orgs = {
    list: () => request('/organizations'),
    create: (data: { name: string; display_name?: string; email?: string; description?: string }) =>
        request('/organizations', { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => request(`/organizations/${id}`),
    update: (id: string, data: any) =>
        request(`/organizations/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request(`/organizations/${id}`, { method: 'DELETE' }),
    listMembers: (id: string) => request(`/organizations/${id}/members`),
    addMember: (id: string, data: { email: string; role: string }) =>
        request(`/organizations/${id}/members`, { method: 'POST', body: JSON.stringify(data) }),
    updateMember: (orgId: string, memberId: string, data: { role: string }) =>
        request(`/organizations/${orgId}/members/${memberId}`, { method: 'PUT', body: JSON.stringify(data) }),
    removeMember: (orgId: string, memberId: string) =>
        request(`/organizations/${orgId}/members/${memberId}`, { method: 'DELETE' }),
};

// Projects
export const projects = {
    list: (orgId: string) => request(`/organizations/${orgId}/projects`),
    create: (orgId: string, data: { name: string; description?: string }) =>
        request(`/organizations/${orgId}/projects`, { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => request(`/projects/${id}`),
    update: (id: string, data: any) =>
        request(`/projects/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request(`/projects/${id}`, { method: 'DELETE' }),
};

// Workspaces
export const workspaces = {
    list: (projectId: string) => request(`/projects/${projectId}/workspaces`),
    create: (projectId: string, data: any) =>
        request(`/projects/${projectId}/workspaces`, { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => request(`/workspaces/${id}`),
    update: (id: string, data: any) =>
        request(`/workspaces/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => request(`/workspaces/${id}`, { method: 'DELETE' }),
    lock: (id: string) => request(`/workspaces/${id}/lock`, { method: 'POST' }),
    unlock: (id: string) => request(`/workspaces/${id}/unlock`, { method: 'POST' }),
};

// Variables
export const variables = {
    list: (wsId: string) => request(`/workspaces/${wsId}/variables`),
    create: (wsId: string, data: any) =>
        request(`/workspaces/${wsId}/variables`, { method: 'POST', body: JSON.stringify(data) }),
    update: (wsId: string, varId: string, data: any) =>
        request(`/workspaces/${wsId}/variables/${varId}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (wsId: string, varId: string) =>
        request(`/workspaces/${wsId}/variables/${varId}`, { method: 'DELETE' }),
};

// Runs
export const runs = {
    list: (wsId: string) => request(`/workspaces/${wsId}/runs`),
    create: (wsId: string, data: { operation: string; message?: string; auto_apply?: boolean }) =>
        request(`/workspaces/${wsId}/runs`, { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => request(`/runs/${id}`),
    getPlanLog: (id: string) => request(`/runs/${id}/plan-log`),
    getApplyLog: (id: string) => request(`/runs/${id}/apply-log`),
    approve: (id: string) => request(`/runs/${id}/approve`, { method: 'POST' }),
    discard: (id: string) => request(`/runs/${id}/discard`, { method: 'POST' }),
    cancel: (id: string) => request(`/runs/${id}/cancel`, { method: 'POST' }),
};

// State
export const state = {
    getCurrent: (wsId: string) => request(`/workspaces/${wsId}/state`),
    listVersions: (wsId: string) => request(`/workspaces/${wsId}/state-versions`),
    getVersion: (wsId: string, versionId: string) =>
        request(`/workspaces/${wsId}/state-versions/${versionId}`),
    getOutputs: (wsId: string) => request(`/workspaces/${wsId}/outputs`),
};

// Terraform Versions
export const terraform = {
    listVersions: () => request('/terraform/versions'),
    listInstalled: () => request('/terraform/versions/installed'),
    install: (version: string) =>
        request(`/terraform/versions/${version}/install`, { method: 'POST' }),
};
