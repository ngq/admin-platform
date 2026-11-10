-- +goose Up
DROP PROCEDURE IF EXISTS `goose_drop_legacy_sys_apis_group`;

-- +goose StatementBegin
CREATE PROCEDURE `goose_drop_legacy_sys_apis_group`()
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = DATABASE()
          AND table_name = 'sys_apis'
          AND column_name = 'group'
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = DATABASE()
              AND table_name = 'sys_apis'
              AND column_name = 'group_name'
        ) THEN
            ALTER TABLE `sys_apis` ADD COLUMN `group_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'API分组' AFTER `id`;
        END IF;
        SET @sql_copy = 'UPDATE `sys_apis` SET `group_name` = `group` WHERE (`group_name` IS NULL OR `group_name` = '''') AND `group` <> ''''';
        PREPARE stmt_copy FROM @sql_copy;
        EXECUTE stmt_copy;
        DEALLOCATE PREPARE stmt_copy;
        ALTER TABLE `sys_apis` DROP COLUMN `group`;
    END IF;
END;
-- +goose StatementEnd

CALL `goose_drop_legacy_sys_apis_group`();

DROP PROCEDURE IF EXISTS `goose_drop_legacy_sys_apis_group`;

-- +goose Down
SELECT 1;
