/** Format 24 boolean hour flags as MT5 day range string */
export function formatDayRange(hours) {
  if (!hours?.length) return "";
  if (hours.every(Boolean) && hours.length === 24) return "00:00-24:00";
  if (!hours.some(Boolean)) return "";

  const ranges = [];
  let start = null;
  for (let i = 0; i < 24; i++) {
    if (hours[i]) {
      if (start === null) start = i;
    } else if (start !== null) {
      ranges.push([start, i]);
      start = null;
    }
  }
  if (start !== null) ranges.push([start, 24]);

  const pad = (n) => String(n).padStart(2, "0");
  return ranges.map(([a, b]) => `${pad(a)}:00-${pad(b)}:00`).join(", ");
}

export const TIME_DAY_NAMES = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

export function allHoursActive() {
  return Array.from({ length: 24 }, () => true);
}
