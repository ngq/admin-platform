-- +goose Up
CREATE TABLE IF NOT EXISTS `sys_departments` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `parent_id` BIGINT DEFAULT 0,
    `name` VARCHAR(64) NOT NULL,
    `leader_user_id` BIGINT DEFAULT 0,
    `sort` BIGINT DEFAULT 0,
    `status` TINYINT DEFAULT 1 COMMENT '1-正常 2-禁用',
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    `updated_at` DATETIME(3) NULL DEFAULT NULL,
    `deleted_at` DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_sys_departments_parent_id` (`parent_id`),
    KEY `idx_sys_departments_leader_user_id` (`leader_user_id`),
    KEY `idx_sys_departments_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_users` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(64) NOT NULL,
    `password` VARCHAR(128) NOT NULL,
    `nickname` VARCHAR(64) DEFAULT NULL,
    `dept_id` BIGINT DEFAULT 0,
    `email` VARCHAR(128) DEFAULT NULL,
    `phone` VARCHAR(20) DEFAULT NULL,
    `avatar` VARCHAR(256) DEFAULT NULL,
    `status` TINYINT DEFAULT 1 COMMENT '1-正常 2-禁用 3-锁定',
    `last_login_at` DATETIME DEFAULT NULL,
    `last_login_ip` VARCHAR(45) DEFAULT NULL,
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    `updated_at` DATETIME(3) NULL DEFAULT NULL,
    `deleted_at` DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_sys_users_username` (`username`),
    KEY `idx_sys_users_dept_id` (`dept_id`),
    KEY `idx_sys_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_roles` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(64) NOT NULL,
    `code` VARCHAR(64) NOT NULL,
    `parent_id` BIGINT DEFAULT 0,
    `sort` BIGINT DEFAULT 0,
    `data_scope` TINYINT DEFAULT 1 COMMENT '1-全部 2-本部门 3-本部门及下级 4-仅本人',
    `status` TINYINT DEFAULT 1 COMMENT '1-正常 2-禁用',
    `description` VARCHAR(256) DEFAULT NULL,
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    `updated_at` DATETIME(3) NULL DEFAULT NULL,
    `deleted_at` DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_sys_roles_code` (`code`),
    KEY `idx_sys_roles_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_user_roles` (
    `user_id` BIGINT NOT NULL,
    `role_id` BIGINT NOT NULL,
    PRIMARY KEY (`user_id`, `role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_menus` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `parent_id` BIGINT DEFAULT 0 COMMENT '父级ID, 0为顶级',
    `name` VARCHAR(64) NOT NULL,
    `path` VARCHAR(256) DEFAULT NULL,
    `component` VARCHAR(256) DEFAULT NULL,
    `redirect` VARCHAR(256) DEFAULT NULL,
    `icon` VARCHAR(64) DEFAULT NULL,
    `sort` BIGINT DEFAULT 0,
    `type` TINYINT NOT NULL COMMENT '1-目录 2-菜单 3-按钮',
    `visible` TINYINT DEFAULT 1 COMMENT '1-显示 2-隐藏',
    `status` TINYINT DEFAULT 1 COMMENT '1-正常 2-禁用',
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    `updated_at` DATETIME(3) NULL DEFAULT NULL,
    `deleted_at` DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_sys_menus_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_buttons` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `menu_id` BIGINT NOT NULL,
    `name` VARCHAR(64) NOT NULL,
    `code` VARCHAR(64) NOT NULL,
    `description` VARCHAR(256) DEFAULT NULL,
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    `updated_at` DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_menu_button_code` (`menu_id`, `code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_apis` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `group_name` VARCHAR(64) NOT NULL COMMENT 'API分组',
    `path` VARCHAR(256) NOT NULL,
    `method` VARCHAR(16) NOT NULL,
    `description` VARCHAR(256) DEFAULT NULL,
    `menu_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '所属菜单ID',
    `action_type` VARCHAR(32) NOT NULL DEFAULT 'other' COMMENT '功能类型',
    `permission_mode` VARCHAR(16) NOT NULL DEFAULT 'casbin' COMMENT '权限模式',
    `ignored` TINYINT DEFAULT 0 COMMENT '1-忽略权限校验',
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    `updated_at` DATETIME(3) NULL DEFAULT NULL,
    `deleted_at` DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_api_path_method` (`path`, `method`),
    KEY `idx_sys_apis_menu_id` (`menu_id`),
    KEY `idx_sys_apis_permission_mode` (`permission_mode`),
    KEY `idx_sys_apis_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_role_menus` (
    `role_id` BIGINT NOT NULL,
    `menu_id` BIGINT NOT NULL,
    PRIMARY KEY (`role_id`, `menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_role_apis` (
    `role_id` BIGINT NOT NULL,
    `api_id` BIGINT NOT NULL,
    PRIMARY KEY (`role_id`, `api_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_role_buttons` (
    `role_id` BIGINT NOT NULL,
    `button_id` BIGINT NOT NULL,
    PRIMARY KEY (`role_id`, `button_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_params` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `key` VARCHAR(128) NOT NULL,
    `value` TEXT,
    `category` VARCHAR(64) DEFAULT NULL,
    `description` VARCHAR(256) DEFAULT NULL,
    `status` TINYINT DEFAULT 1 COMMENT '1-正常 2-禁用',
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    `updated_at` DATETIME(3) NULL DEFAULT NULL,
    `deleted_at` DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_sys_params_key` (`key`),
    KEY `idx_sys_params_category` (`category`),
    KEY `idx_sys_params_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_operation_logs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT DEFAULT NULL COMMENT '操作用户ID',
    `username` VARCHAR(64) DEFAULT NULL,
    `method` VARCHAR(16) NOT NULL,
    `path` VARCHAR(256) NOT NULL,
    `ip` VARCHAR(45) DEFAULT NULL,
    `user_agent` VARCHAR(256) DEFAULT NULL,
    `request_body` TEXT,
    `response_status` BIGINT DEFAULT NULL COMMENT 'HTTP响应状态码',
    `latency_ms` BIGINT DEFAULT NULL COMMENT '请求耗时(毫秒)',
    `created_at` DATETIME(3) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_login_logs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT DEFAULT NULL COMMENT '登录用户ID',
    `username` VARCHAR(64) DEFAULT NULL,
    `ip` VARCHAR(45) DEFAULT NULL,
    `location` VARCHAR(128) DEFAULT NULL,
    `user_agent` VARCHAR(256) DEFAULT NULL,
    `status` TINYINT NOT NULL COMMENT '1-成功 2-失败',
    `message` VARCHAR(256) DEFAULT NULL,
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sys_error_logs` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT DEFAULT NULL COMMENT '关联用户ID',
    `username` VARCHAR(64) DEFAULT NULL,
    `method` VARCHAR(16) DEFAULT NULL,
    `path` VARCHAR(256) DEFAULT NULL,
    `error_message` TEXT,
    `stack` TEXT,
    `status` TINYINT DEFAULT 1 COMMENT '1-未处理 2-已处理',
    `created_at` DATETIME(3) NULL DEFAULT NULL,
    `updated_at` DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `casbin_rule` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `ptype` VARCHAR(12) DEFAULT NULL,
    `v0` VARCHAR(100) DEFAULT NULL,
    `v1` VARCHAR(100) DEFAULT NULL,
    `v2` VARCHAR(100) DEFAULT NULL,
    `v3` VARCHAR(100) DEFAULT NULL,
    `v4` VARCHAR(100) DEFAULT NULL,
    `v5` VARCHAR(100) DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
