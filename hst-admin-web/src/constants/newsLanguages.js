/** Windows LANGID values for group news filtering (mt5_groups.NewsLangs). */

const RAW = [
  { id: 0x0436, label: "Afrikaans" },
  { id: 0x041c, label: "Albanian" },
  { id: 0x0401, label: "Arabic" },
  { id: 0x042b, label: "Armenian" },
  { id: 0x042c, label: "Azeri" },
  { id: 0x042d, label: "Basque" },
  { id: 0x0423, label: "Belarusian" },
  { id: 0x0445, label: "Bengali" },
  { id: 0x0402, label: "Bulgarian" },
  { id: 0x0455, label: "Burmese" },
  { id: 0x0403, label: "Catalan" },
  { id: 0x0804, label: "Chinese" },
  { id: 0x0404, label: "Chinese (Traditional)" },
  { id: 0x041a, label: "Croatian" },
  { id: 0x0405, label: "Czech" },
  { id: 0x0406, label: "Danish" },
  { id: 0x0413, label: "Dutch" },
  { id: 0x0409, label: "English" },
  { id: 0x0425, label: "Estonian" },
  { id: 0x0429, label: "Farsi" },
  { id: 0x040b, label: "Finnish" },
  { id: 0x040c, label: "French" },
  { id: 0x0437, label: "Georgian" },
  { id: 0x0407, label: "German" },
  { id: 0x0408, label: "Greek" },
  { id: 0x0447, label: "Gujarati" },
  { id: 0x040d, label: "Hebrew" },
  { id: 0x040e, label: "Hungarian" },
  { id: 0x0421, label: "Indonesian" },
  { id: 0x0410, label: "Italian" },
  { id: 0x0411, label: "Japanese" },
  { id: 0x043f, label: "Kazakh" },
  { id: 0x0412, label: "Korean" },
  { id: 0x0426, label: "Latvian" },
  { id: 0x0427, label: "Lithuanian" },
  { id: 0x042f, label: "Macedonian" },
  { id: 0x043e, label: "Malay" },
  { id: 0x044c, label: "Malayalam" },
  { id: 0x044e, label: "Marathi" },
  { id: 0x0414, label: "Norwegian" },
  { id: 0x0415, label: "Polish" },
  { id: 0x0416, label: "Portuguese" },
  { id: 0x0816, label: "Portuguese (Portugal)" },
  { id: 0x0418, label: "Romanian" },
  { id: 0x0419, label: "Russian" },
  { id: 0x081a, label: "Serbian" },
  { id: 0x041b, label: "Slovak" },
  { id: 0x0424, label: "Slovenian" },
  { id: 0x040a, label: "Spanish" },
  { id: 0x041d, label: "Swedish" },
  { id: 0x0441, label: "Tamil" },
  { id: 0x0449, label: "Telugu" },
  { id: 0x041e, label: "Thai" },
  { id: 0x041f, label: "Turkish" },
  { id: 0x0422, label: "Ukrainian" },
  { id: 0x0420, label: "Urdu" },
  { id: 0x042a, label: "Vietnamese" },
];

export const NEWS_LANGUAGES = [...RAW].sort((a, b) => a.label.localeCompare(b.label));

export const NEWS_LANGUAGE_BY_ID = Object.fromEntries(NEWS_LANGUAGES.map((l) => [l.id, l.label]));

export function newsLangLabel(id) {
  return NEWS_LANGUAGE_BY_ID[id] ?? `Language ${id}`;
}

/** Empty array = MT5 "Auto select" (all languages). */
export function newsLangSummary(ids) {
  if (!ids?.length) return "Auto select";
  return ids.map(newsLangLabel).join(", ");
}
