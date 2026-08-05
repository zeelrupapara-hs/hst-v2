import { formatDayRange, TIME_DAY_NAMES } from "../../../lib/timeFormat.js";

/** @typedef {'section'|'timezone'|'yesno'|'text'|'day'} RowEditor */

/**
 * @param {import('../../hooks/useTimeSettings.js').TimeSettings} settings
 */
export function buildTimePropertyRows(settings) {
  if (!settings) return [];

  return [
    { type: "section", id: "sec-common", label: "Common settings" },
    {
      id: "time_zone",
      label: "Time zone",
      editor: "timezone",
      icon: "time",
      value: settings.time_zone,
    },
    {
      id: "daylight_saving",
      label: "Daylight Saving Time correction",
      editor: "yesno",
      icon: "time",
      value: settings.daylight_saving,
    },
    {
      id: "sync_servers",
      label: "Time synchronization with",
      editor: "text",
      icon: "time",
      value: settings.sync_servers,
    },
    { type: "section", id: "sec-days", label: "Day settings" },
    ...TIME_DAY_NAMES.map((name, dayIndex) => ({
      id: `day_${dayIndex}`,
      label: name,
      editor: "day",
      icon: "time",
      dayIndex,
      value: formatDayRange(settings.schedule[dayIndex]),
    })),
  ];
}

export function isDayRow(row) {
  return row?.editor === "day";
}

export function isSectionRow(row) {
  return row?.type === "section";
}
