CREATE TABLE IF NOT EXISTS hst.managers (
    login                 BIGINT       PRIMARY KEY REFERENCES hst.users (login) ON DELETE CASCADE,
    name                  VARCHAR(128) NOT NULL DEFAULT '',
    mailbox               VARCHAR(128) NOT NULL DEFAULT '',
    server                INTEGER      NOT NULL DEFAULT 0,
    request_limit_logs    SMALLINT     NOT NULL DEFAULT 0,
    request_limit_reports SMALLINT     NOT NULL DEFAULT 0,
    groups                TEXT[]       NOT NULL DEFAULT '{}',

    right_admin                    SMALLINT NOT NULL DEFAULT 0,
    right_manager                  SMALLINT NOT NULL DEFAULT 0,
    right_cfg_time                 SMALLINT NOT NULL DEFAULT 0,
    right_cfg_holidays             SMALLINT NOT NULL DEFAULT 0,
    right_cfg_groups               SMALLINT NOT NULL DEFAULT 0,
    right_cfg_managers             SMALLINT NOT NULL DEFAULT 0,
    right_cfg_requests             SMALLINT NOT NULL DEFAULT 0,
    right_cfg_gateways             SMALLINT NOT NULL DEFAULT 0,
    right_cfg_datafeeds            SMALLINT NOT NULL DEFAULT 0,
    right_cfg_reports              SMALLINT NOT NULL DEFAULT 0,
    right_cfg_symbols              SMALLINT NOT NULL DEFAULT 0,
    right_cfg_web_services         SMALLINT NOT NULL DEFAULT 0,
    right_cfg_messengers           SMALLINT NOT NULL DEFAULT 0,
    right_cfg_kyc                  SMALLINT NOT NULL DEFAULT 0,
    right_cfg_automations          SMALLINT NOT NULL DEFAULT 0,
    right_cfg_allocations          SMALLINT NOT NULL DEFAULT 0,
    right_cfg_corporate            SMALLINT NOT NULL DEFAULT 0,
    right_cfg_payments             SMALLINT NOT NULL DEFAULT 0,
    right_cfg_mails                SMALLINT NOT NULL DEFAULT 0,
    right_cfg_streaming            SMALLINT NOT NULL DEFAULT 0,
    right_srv_journals             SMALLINT NOT NULL DEFAULT 0,
    right_srv_reports              SMALLINT NOT NULL DEFAULT 0,
    right_charts                   SMALLINT NOT NULL DEFAULT 0,
    right_email                    SMALLINT NOT NULL DEFAULT 0,
    right_news                     SMALLINT NOT NULL DEFAULT 0,
    right_export                   SMALLINT NOT NULL DEFAULT 0,
    right_techsupport              SMALLINT NOT NULL DEFAULT 0,
    right_market                   SMALLINT NOT NULL DEFAULT 0,
    right_accountant               SMALLINT NOT NULL DEFAULT 0,
    right_acc_read                 SMALLINT NOT NULL DEFAULT 0,
    right_acc_details_name         SMALLINT NOT NULL DEFAULT 0,
    right_acc_details_location     SMALLINT NOT NULL DEFAULT 0,
    right_acc_details_address      SMALLINT NOT NULL DEFAULT 0,
    right_acc_details_id           SMALLINT NOT NULL DEFAULT 0,
    right_acc_details_email        SMALLINT NOT NULL DEFAULT 0,
    right_acc_details_phone        SMALLINT NOT NULL DEFAULT 0,
    right_acc_details_general      SMALLINT NOT NULL DEFAULT 0,
    right_acc_technical            SMALLINT NOT NULL DEFAULT 0,
    right_acc_tech_modify          SMALLINT NOT NULL DEFAULT 0,
    right_acc_manager              SMALLINT NOT NULL DEFAULT 0,
    right_acc_delete               SMALLINT NOT NULL DEFAULT 0,
    right_acc_online               SMALLINT NOT NULL DEFAULT 0,
    right_confirm_actions          SMALLINT NOT NULL DEFAULT 0,
    right_notifications            SMALLINT NOT NULL DEFAULT 0,
    right_trades_read              SMALLINT NOT NULL DEFAULT 0,
    right_trades_manager           SMALLINT NOT NULL DEFAULT 0,
    right_trades_delete            SMALLINT NOT NULL DEFAULT 0,
    right_trades_dealer            SMALLINT NOT NULL DEFAULT 0,
    right_trades_supervisor        SMALLINT NOT NULL DEFAULT 0,
    right_quotes_raw               SMALLINT NOT NULL DEFAULT 0,
    right_quotes                   SMALLINT NOT NULL DEFAULT 0,
    right_symbol_details           SMALLINT NOT NULL DEFAULT 0,
    right_risk_manager             SMALLINT NOT NULL DEFAULT 0,
    right_group_margin             SMALLINT NOT NULL DEFAULT 0,
    right_group_commission         SMALLINT NOT NULL DEFAULT 0,
    right_reports                  SMALLINT NOT NULL DEFAULT 0,
    right_clients_access           SMALLINT NOT NULL DEFAULT 0,
    right_clients_create           SMALLINT NOT NULL DEFAULT 0,
    right_clients_edit             SMALLINT NOT NULL DEFAULT 0,
    right_clients_delete           SMALLINT NOT NULL DEFAULT 0,
    right_clients_kyc              SMALLINT NOT NULL DEFAULT 0,
    right_clients_details_name     SMALLINT NOT NULL DEFAULT 0,
    right_clients_details_location SMALLINT NOT NULL DEFAULT 0,
    right_clients_details_address  SMALLINT NOT NULL DEFAULT 0,
    right_clients_details_id       SMALLINT NOT NULL DEFAULT 0,
    right_clients_details_email    SMALLINT NOT NULL DEFAULT 0,
    right_clients_details_phone    SMALLINT NOT NULL DEFAULT 0,
    right_clients_details_general  SMALLINT NOT NULL DEFAULT 0,
    right_documents_access         SMALLINT NOT NULL DEFAULT 0,
    right_documents_create         SMALLINT NOT NULL DEFAULT 0,
    right_documents_edit           SMALLINT NOT NULL DEFAULT 0,
    right_documents_delete         SMALLINT NOT NULL DEFAULT 0,
    right_documents_files_add      SMALLINT NOT NULL DEFAULT 0,
    right_documents_files_delete   SMALLINT NOT NULL DEFAULT 0,
    right_comments_access          SMALLINT NOT NULL DEFAULT 0,
    right_comments_create          SMALLINT NOT NULL DEFAULT 0,
    right_comments_delete          SMALLINT NOT NULL DEFAULT 0,
    updated_at              BIGINT  NOT NULL DEFAULT 0
);

ALTER TABLE hst.clients
    ADD CONSTRAINT clients_assigned_manager_fkey FOREIGN KEY (assigned_manager)
    REFERENCES hst.managers (login) ON DELETE SET NULL;

ALTER TABLE hst.clients
    ADD CONSTRAINT clients_compliance_approved_by_fkey FOREIGN KEY (compliance_approved_by)
    REFERENCES hst.managers (login) ON DELETE SET NULL;

COMMENT ON TABLE hst.managers IS
    'Back-office logins. Every right_* column: 1 = granted, 0 = not granted.';
COMMENT ON COLUMN hst.managers.request_limit_logs IS
    'ManagerLimit: 0=all 1=1_month 2=3_months 3=6_months 4=1_year 5=2_years 6=3_years';
COMMENT ON COLUMN hst.managers.request_limit_reports IS
    'ManagerLimit: 0=all 1=1_month 2=3_months 3=6_months 4=1_year 5=2_years 6=3_years';

COMMENT ON TABLE hst.managers IS 'all time columns are unix nanoseconds, 0 means unset';
