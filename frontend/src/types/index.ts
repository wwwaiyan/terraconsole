export interface User {
    id: string;
    email: string;
    username: string;
    full_name: string;
    avatar_url: string;
    mfa_enabled: boolean;
    is_active: boolean;
    last_login_at: string | null;
    created_at: string;
    updated_at: string;
}

export interface AuthResponse {
    token: string;
    expires_at: number;
    user: User;
}

export interface Organization {
    id: string;
    name: string;
    display_name: string;
    email: string;
    description: string;
    owner_id: string;
    created_at: string;
    updated_at: string;
}

export interface OrgMember {
    id: string;
    organization_id: string;
    user_id: string;
    user?: User;
    role: 'owner' | 'admin' | 'member' | 'viewer';
    created_at: string;
}

export interface Project {
    id: string;
    name: string;
    description: string;
    organization_id: string;
    created_at: string;
    updated_at: string;
}

export interface Workspace {
    id: string;
    name: string;
    description: string;
    project_id: string;
    terraform_version: string;
    working_directory: string;
    auto_apply: boolean;
    execution_mode: string;
    locked: boolean;
    locked_by: string | null;
    locked_at: string | null;
    vcs_repo_url: string;
    vcs_branch: string;
    current_state_id: string | null;
    created_at: string;
    updated_at: string;
}

export interface Variable {
    id: string;
    workspace_id: string;
    key: string;
    value: string;
    description: string;
    category: 'terraform' | 'env';
    hcl: boolean;
    sensitive: boolean;
    created_at: string;
    updated_at: string;
}

export type RunStatus =
    | 'pending'
    | 'planning'
    | 'planned'
    | 'needs_confirmation'
    | 'applying'
    | 'applied'
    | 'errored'
    | 'cancelled'
    | 'discarded'
    | 'planned_and_finished';

export type RunOperation = 'plan' | 'plan_and_apply' | 'destroy' | 'refresh';

export interface Run {
    id: string;
    workspace_id: string;
    status: RunStatus;
    operation: RunOperation;
    message: string;
    is_destroy: boolean;
    auto_apply: boolean;
    terraform_version: string;
    created_by: string;
    creator?: User;
    resources_added: number;
    resources_changed: number;
    resources_deleted: number;
    started_at: string | null;
    plan_completed_at: string | null;
    applied_at: string | null;
    completed_at: string | null;
    created_at: string;
    updated_at: string;
}

export interface StateVersion {
    id: string;
    workspace_id: string;
    run_id: string | null;
    serial: number;
    lineage: string;
    state_hash: string;
    outputs: string;
    resource_count: number;
    created_at: string;
    created_by: string;
}

export interface TFVersion {
    version: string;
    installed: boolean;
    path?: string;
}
