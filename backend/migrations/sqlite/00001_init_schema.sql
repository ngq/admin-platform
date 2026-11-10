-- +goose Up
CREATE TABLE IF NOT EXISTS `sys_departments` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `parent_id` INTEGER DEFAULT 0,
    `name` TEXT NOT NULL,
    `leader_user_id` INTEGER DEFAULT 0,
    `sort` INTEGER DEFAULT 0,
    `status` INTEGER DEFAULT 1,
    `created_at` DATETIME NULL DEFAULT NULL,
    `updated_at` DATETIME NULL DEFAULT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS `idx_sys_departments_parent_id` ON `sys_departments` (`parent_id`);
CREATE INDEX IF NOT EXISTS `idx_sys_departments_leader_user_id` ON `sys_departments` (`leader_user_id`);
CREATE INDEX IF NOT EXISTS `idx_sys_departments_deleted_at` ON `sys_departments` (`deleted_at`);

CREATE TABLE IF NOT EXISTS `sys_users` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `username` TEXT NOT NULL,
    `password` TEXT NOT NULL,
    `nickname` TEXT DEFAULT NULL,
    `dept_id` INTEGER DEFAULT 0,
    `email` TEXT DEFAULT NULL,
    `phone` TEXT DEFAULT NULL,
    `avatar` TEXT DEFAULT NULL,
    `status` INTEGER DEFAULT 1,
    `last_login_at` DATETIME DEFAULT NULL,
    `last_login_ip` TEXT DEFAULT NULL,
    `created_at` DATETIME NULL DEFAULT NULL,
    `updated_at` DATETIME NULL DEFAULT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS `idx_sys_users_username` ON `sys_users` (`username`);
CREATE INDEX IF NOT EXISTS `idx_sys_users_dept_id` ON `sys_users` (`dept_id`);
CREATE INDEX IF NOT EXISTS `idx_sys_users_deleted_at` ON `sys_users` (`deleted_at`);

CREATE TABLE IF NOT EXISTS `sys_roles` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` TEXT NOT NULL,
    `code` TEXT NOT NULL,
    `parent_id` INTEGER DEFAULT 0,
    `sort` INTEGER DEFAULT 0,
    `data_scope` INTEGER DEFAULT 1,
    `status` INTEGER DEFAULT 1,
    `description` TEXT DEFAULT NULL,
    `created_at` DATETIME NULL DEFAULT NULL,
    `updated_at` DATETIME NULL DEFAULT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS `idx_sys_roles_code` ON `sys_roles` (`code`);
CREATE INDEX IF NOT EXISTS `idx_sys_roles_deleted_at` ON `sys_roles` (`deleted_at`);

CREATE TABLE IF NOT EXISTS `sys_user_roles` (
    `user_id` INTEGER NOT NULL,
    `role_id` INTEGER NOT NULL,
    PRIMARY KEY (`user_id`, `role_id`)
);

CREATE TABLE IF NOT EXISTS `sys_menus` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `parent_id` INTEGER DEFAULT 0,
    `name` TEXT NOT NULL,
    `path` TEXT DEFAULT NULL,
    `component` TEXT DEFAULT NULL,
    `redirect` TEXT DEFAULT NULL,
    `icon` TEXT DEFAULT NULL,
    `sort` INTEGER DEFAULT 0,
    `type` INTEGER NOT NULL,
    `visible` INTEGER DEFAULT 1,
    `status` INTEGER DEFAULT 1,
    `created_at` DATETIME NULL DEFAULT NULL,
    `updated_at` DATETIME NULL DEFAULT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS `idx_sys_menus_deleted_at` ON `sys_menus` (`deleted_at`);

CREATE TABLE IF NOT EXISTS `sys_buttons` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `menu_id` INTEGER NOT NULL,
    `name` TEXT NOT NULL,
    `code` TEXT NOT NULL,
    `description` TEXT DEFAULT NULL,
    `created_at` DATETIME NULL DEFAULT NULL,
    `updated_at` DATETIME NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS `idx_menu_button_code` ON `sys_buttons` (`menu_id`, `code`);

CREATE TABLE IF NOT EXISTS `sys_apis` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `group_name` TEXT NOT NULL,
    `path` TEXT NOT NULL,
    `method` TEXT NOT NULL,
    `description` TEXT DEFAULT NULL,
    `menu_id` INTEGER NOT NULL DEFAULT 0,
    `action_type` TEXT NOT NULL DEFAULT 'other',
    `permission_mode` TEXT NOT NULL DEFAULT 'casbin',
    `ignored` INTEGER DEFAULT 0,
    `created_at` DATETIME NULL DEFAULT NULL,
    `updated_at` DATETIME NULL DEFAULT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS `idx_api_path_method` ON `sys_apis` (`path`, `method`);
CREATE INDEX IF NOT EXISTS `idx_sys_apis_menu_id` ON `sys_apis` (`menu_id`);
CREATE INDEX IF NOT EXISTS `idx_sys_apis_permission_mode` ON `sys_apis` (`permission_mode`);
CREATE INDEX IF NOT EXISTS `idx_sys_apis_deleted_at` ON `sys_apis` (`deleted_at`);

CREATE TABLE IF NOT EXISTS `sys_role_menus` (
    `role_id` INTEGER NOT NULL,
    `menu_id` INTEGER NOT NULL,
    PRIMARY KEY (`role_id`, `menu_id`)
);

CREATE TABLE IF NOT EXISTS `sys_role_apis` (
    `role_id` INTEGER NOT NULL,
    `api_id` INTEGER NOT NULL,
    PRIMARY KEY (`role_id`, `api_id`)
);

CREATE TABLE IF NOT EXISTS `sys_role_buttons` (
    `role_id` INTEGER NOT NULL,
    `button_id` INTEGER NOT NULL,
    PRIMARY KEY (`role_id`, `button_id`)
);

CREATE TABLE IF NOT EXISTS `sys_params` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `key` TEXT NOT NULL,
    `value` TEXT,
    `category` TEXT DEFAULT NULL,
    `description` TEXT DEFAULT NULL,
    `status` INTEGER DEFAULT 1,
    `created_at` DATETIME NULL DEFAULT NULL,
    `updated_at` DATETIME NULL DEFAULT NULL,
    `deleted_at` DATETIME NULL DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS `idx_sys_params_key` ON `sys_params` (`key`);
CREATE INDEX IF NOT EXISTS `idx_sys_params_category` ON `sys_params` (`category`);
CREATE INDEX IF NOT EXISTS `idx_sys_params_deleted_at` ON `sys_params` (`deleted_at`);

CREATE TABLE IF NOT EXISTS `sys_operation_logs` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `user_id` INTEGER DEFAULT NULL,
    `username` TEXT DEFAULT NULL,
    `method` TEXT NOT NULL,
    `path` TEXT NOT NULL,
    `ip` TEXT DEFAULT NULL,
    `user_agent` TEXT DEFAULT NULL,
    `request_body` TEXT,
    `response_status` INTEGER DEFAULT NULL,
    `latency_ms` INTEGER DEFAULT NULL,
    `created_at` DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS `sys_login_logs` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `user_id` INTEGER DEFAULT NULL,
    `username` TEXT DEFAULT NULL,
    `ip` TEXT DEFAULT NULL,
    `location` TEXT DEFAULT NULL,
    `user_agent` TEXT DEFAULT NULL,
    `status` INTEGER NOT NULL,
    `message` TEXT DEFAULT NULL,
    `created_at` DATETIME NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `sys_error_logs` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `user_id` INTEGER DEFAULT NULL,
    `username` TEXT DEFAULT NULL,
    `method` TEXT DEFAULT NULL,
    `path` TEXT DEFAULT NULL,
    `error_message` TEXT,
    `stack` TEXT,
    `status` INTEGER DEFAULT 1,
    `created_at` DATETIME NULL DEFAULT NULL,
    `updated_at` DATETIME NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `casbin_rule` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `ptype` TEXT DEFAULT NULL,
    `v0` TEXT DEFAULT NULL,
    `v1` TEXT DEFAULT NULL,
    `v2` TEXT DEFAULT NULL,
    `v3` TEXT DEFAULT NULL,
    `v4` TEXT DEFAULT NULL,
    `v5` TEXT DEFAULT NULL
);

-- +goose Down
DROP TABLE IF EXISTS `casbin_rule`;
DROP TABLE IF EXISTS `sys_error_logs`;
DROP TABLE IF EXISTS `sys_login_logs`;
DROP TABLE IF EXISTS `sys_operation_logs`;
DROP TABLE IF EXISTS `sys_params`;
DROP TABLE IF EXISTS `sys_role_buttons`;
DROP TABLE IF EXISTS `sys_role_apis`;
DROP TABLE IF EXISTS `sys_role_menus`;
DROP TABLE IF EXISTS `sys_apis`;
DROP TABLE IF EXISTS `sys_buttons`;
DROP TABLE IF EXISTS `sys_menus`;
DROP TABLE IF EXISTS `sys_user_roles`;
DROP TABLE IF EXISTS `sys_roles`;
DROP TABLE IF EXISTS `sys_users`;
DROP TABLE IF EXISTS `sys_departments`;
