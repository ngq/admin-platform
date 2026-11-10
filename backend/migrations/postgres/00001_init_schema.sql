-- +goose Up
CREATE TABLE IF NOT EXISTS sys_departments (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT DEFAULT 0,
    name VARCHAR(64) NOT NULL,
    leader_user_id BIGINT DEFAULT 0,
    sort BIGINT DEFAULT 0,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ NULL DEFAULT NULL,
    updated_at TIMESTAMPTZ NULL DEFAULT NULL,
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_sys_departments_parent_id ON sys_departments (parent_id);
CREATE INDEX IF NOT EXISTS idx_sys_departments_leader_user_id ON sys_departments (leader_user_id);
CREATE INDEX IF NOT EXISTS idx_sys_departments_deleted_at ON sys_departments (deleted_at);

CREATE TABLE IF NOT EXISTS sys_users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    password VARCHAR(128) NOT NULL,
    nickname VARCHAR(64) DEFAULT NULL,
    dept_id BIGINT DEFAULT 0,
    email VARCHAR(128) DEFAULT NULL,
    phone VARCHAR(20) DEFAULT NULL,
    avatar VARCHAR(256) DEFAULT NULL,
    status SMALLINT DEFAULT 1,
    last_login_at TIMESTAMPTZ DEFAULT NULL,
    last_login_ip VARCHAR(45) DEFAULT NULL,
    created_at TIMESTAMPTZ NULL DEFAULT NULL,
    updated_at TIMESTAMPTZ NULL DEFAULT NULL,
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_users_username ON sys_users (username);
CREATE INDEX IF NOT EXISTS idx_sys_users_dept_id ON sys_users (dept_id);
CREATE INDEX IF NOT EXISTS idx_sys_users_deleted_at ON sys_users (deleted_at);

CREATE TABLE IF NOT EXISTS sys_roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    parent_id BIGINT DEFAULT 0,
    sort BIGINT DEFAULT 0,
    data_scope SMALLINT DEFAULT 1,
    status SMALLINT DEFAULT 1,
    description VARCHAR(256) DEFAULT NULL,
    created_at TIMESTAMPTZ NULL DEFAULT NULL,
    updated_at TIMESTAMPTZ NULL DEFAULT NULL,
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_roles_code ON sys_roles (code);
CREATE INDEX IF NOT EXISTS idx_sys_roles_deleted_at ON sys_roles (deleted_at);

CREATE TABLE IF NOT EXISTS sys_user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS sys_menus (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT DEFAULT 0,
    name VARCHAR(64) NOT NULL,
    path VARCHAR(256) DEFAULT NULL,
    component VARCHAR(256) DEFAULT NULL,
    redirect VARCHAR(256) DEFAULT NULL,
    icon VARCHAR(64) DEFAULT NULL,
    sort BIGINT DEFAULT 0,
    type SMALLINT NOT NULL,
    visible SMALLINT DEFAULT 1,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ NULL DEFAULT NULL,
    updated_at TIMESTAMPTZ NULL DEFAULT NULL,
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_sys_menus_deleted_at ON sys_menus (deleted_at);

CREATE TABLE IF NOT EXISTS sys_buttons (
    id BIGSERIAL PRIMARY KEY,
    menu_id BIGINT NOT NULL,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    description VARCHAR(256) DEFAULT NULL,
    created_at TIMESTAMPTZ NULL DEFAULT NULL,
    updated_at TIMESTAMPTZ NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_button_code ON sys_buttons (menu_id, code);

CREATE TABLE IF NOT EXISTS sys_apis (
    id BIGSERIAL PRIMARY KEY,
    group_name VARCHAR(64) NOT NULL,
    path VARCHAR(256) NOT NULL,
    method VARCHAR(16) NOT NULL,
    description VARCHAR(256) DEFAULT NULL,
    menu_id BIGINT NOT NULL DEFAULT 0,
    action_type VARCHAR(32) NOT NULL DEFAULT 'other',
    permission_mode VARCHAR(16) NOT NULL DEFAULT 'casbin',
    ignored SMALLINT DEFAULT 0,
    created_at TIMESTAMPTZ NULL DEFAULT NULL,
    updated_at TIMESTAMPTZ NULL DEFAULT NULL,
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_path_method ON sys_apis (path, method);
CREATE INDEX IF NOT EXISTS idx_sys_apis_menu_id ON sys_apis (menu_id);
CREATE INDEX IF NOT EXISTS idx_sys_apis_permission_mode ON sys_apis (permission_mode);
CREATE INDEX IF NOT EXISTS idx_sys_apis_deleted_at ON sys_apis (deleted_at);

CREATE TABLE IF NOT EXISTS sys_role_menus (
    role_id BIGINT NOT NULL,
    menu_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, menu_id)
);

CREATE TABLE IF NOT EXISTS sys_role_apis (
    role_id BIGINT NOT NULL,
    api_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, api_id)
);

CREATE TABLE IF NOT EXISTS sys_role_buttons (
    role_id BIGINT NOT NULL,
    button_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, button_id)
);

CREATE TABLE IF NOT EXISTS sys_params (
    id BIGSERIAL PRIMARY KEY,
    "key" VARCHAR(128) NOT NULL,
    value TEXT,
    category VARCHAR(64) DEFAULT NULL,
    description VARCHAR(256) DEFAULT NULL,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ NULL DEFAULT NULL,
    updated_at TIMESTAMPTZ NULL DEFAULT NULL,
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_params_key ON sys_params ("key");
CREATE INDEX IF NOT EXISTS idx_sys_params_category ON sys_params (category);
CREATE INDEX IF NOT EXISTS idx_sys_params_deleted_at ON sys_params (deleted_at);

CREATE TABLE IF NOT EXISTS sys_operation_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT DEFAULT NULL,
    username VARCHAR(64) DEFAULT NULL,
    method VARCHAR(16) NOT NULL,
    path VARCHAR(256) NOT NULL,
    ip VARCHAR(45) DEFAULT NULL,
    user_agent VARCHAR(256) DEFAULT NULL,
    request_body TEXT,
    response_status BIGINT DEFAULT NULL,
    latency_ms BIGINT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS sys_login_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT DEFAULT NULL,
    username VARCHAR(64) DEFAULT NULL,
    ip VARCHAR(45) DEFAULT NULL,
    location VARCHAR(128) DEFAULT NULL,
    user_agent VARCHAR(256) DEFAULT NULL,
    status SMALLINT NOT NULL,
    message VARCHAR(256) DEFAULT NULL,
    created_at TIMESTAMPTZ NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS sys_error_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT DEFAULT NULL,
    username VARCHAR(64) DEFAULT NULL,
    method VARCHAR(16) DEFAULT NULL,
    path VARCHAR(256) DEFAULT NULL,
    error_message TEXT,
    stack TEXT,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ NULL DEFAULT NULL,
    updated_at TIMESTAMPTZ NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(12) DEFAULT NULL,
    v0 VARCHAR(100) DEFAULT NULL,
    v1 VARCHAR(100) DEFAULT NULL,
    v2 VARCHAR(100) DEFAULT NULL,
    v3 VARCHAR(100) DEFAULT NULL,
    v4 VARCHAR(100) DEFAULT NULL,
    v5 VARCHAR(100) DEFAULT NULL
);

-- +goose Down
DROP TABLE IF EXISTS casbin_rule;
DROP TABLE IF EXISTS sys_error_logs;
DROP TABLE IF EXISTS sys_login_logs;
DROP TABLE IF EXISTS sys_operation_logs;
DROP TABLE IF EXISTS sys_params;
DROP TABLE IF EXISTS sys_role_buttons;
DROP TABLE IF EXISTS sys_role_apis;
DROP TABLE IF EXISTS sys_role_menus;
DROP TABLE IF EXISTS sys_apis;
DROP TABLE IF EXISTS sys_buttons;
DROP TABLE IF EXISTS sys_menus;
DROP TABLE IF EXISTS sys_user_roles;
DROP TABLE IF EXISTS sys_roles;
DROP TABLE IF EXISTS sys_users;
DROP TABLE IF EXISTS sys_departments;
