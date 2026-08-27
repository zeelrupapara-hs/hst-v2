export const SOCKET_EVENTS = {
  START_MARKET_FEED: "start_market_feed",

  ORDER_CREATE: "order_create",
  ORDER_UPDATE: "order_update",
  ORDER_CANCEL: "order_cancel",

  POSITION_CREATE: "position_create",
  POSITION_UPDATE: "position_update",
  POSITION_CLOSE: "position_close",
  POSITION_HEDGE: "position_close_by",

  DEAL_CREATE: "deal_create",

  JOURNAL_CREATE: "journal",

  ALERT_CREATE: "alert_create",
  ALERT_UPDATE: "alert_update",
  ALERT_DELETE: "alert_delete",
  ALERT_TRIGGER: "alert_triggered",

  MAIL_INBOX: "email_inbox",
  MAIL_OUTBOX: "email_outbox",
  MAIL_DRAFT: "email_draft",
  MAIL_TRASH: "email_bin",
  MAIL_DELETE: "email_delete",

  REPORT_CREATE: "report_create",
  REPORT_STATUS: "report_status",
  REPORT_DELETE: "report_delete",

  GROUP_UPDATED: "group_updated",

  SYMBOL_UPDATED: "symbol_updated",
};
