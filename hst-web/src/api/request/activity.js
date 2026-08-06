import api from "../api";
import { qs } from "../../utils/utils";
import {
  adaptClosedPosition,
  adaptDeal,
  adaptJournal,
  adaptOrder,
  adaptPosition,
  adaptAlert,
  adaptSummary,
  mapList,
  mapOne,
} from "../adapt";

// Accounts
export const getAccountDetails = () =>
  api.get("/account").then((r) => mapOne(r, adaptSummary));

// Positions. Ours keeps no closed positions on this route; the history tab reads them back
// from the deals that closed them, which is where MT5 keeps them too.
export const getAllPositions = () =>
  api.get("/positions").then((r) => mapList(r, adaptPosition));

export const getAllClosedPositions = (params) =>
  api
    .get(`/history/positions?${qs(params)}`)
    .then((r) => mapList(r, adaptClosedPosition));

// History
export const getAllOrders = (params) =>
  api.get(`/orders?${qs(params)}`).then((r) => mapList(r, adaptOrder));

export const getAllDeals = (params) =>
  api.get(`/deals?${qs(params)}`).then((r) => mapList(r, adaptDeal));

// Alerts
export const getAllAlerts = () =>
  api.get("/alerts").then((r) => mapList(r, adaptAlert));

// Mails: one route, the folder is a parameter
const mailFolder = (type) =>
  api.get(`/mails?type=${type}`).then((r) =>
    mapList(r, (m) => ({
      ...m,
      id: m.tracking_id,
      common_id: m.tracking_id,
      to_account_id: m.recipient_login,
    }))
  );

export const getAllInboxMails = () => mailFolder("inbox");
export const getAllOutboxMails = () => mailFolder("outbox");
export const getAllDraftMails = () => mailFolder("draft");
export const getAllTrashMails = () => mailFolder("bin");

// Reports are not built yet, so the panel gets an empty list rather than a failed request
export const getAllReports = () =>
  Promise.resolve({ data: { success: true, data: [] } });

// Journals
export const getAllJournals = () =>
  api.get("/journal").then((r) => mapList(r, adaptJournal));

// News
export const getAllNews = () =>
  api.get("/news").then((r) =>
    mapList(r, (n) => ({
      ...n,
      id: n.news_id,
      link: n.url,
      pub_date: n.published_at
        ? new Date(n.published_at / 1e6).toLocaleString()
        : "",
    }))
  );
