// The classes leave out the characters that read as each other: l/I, O/0, 1.
const CLASSES = ["abcdefghijkmnopqrstuvwxyz", "ABCDEFGHJKLMNPQRSTUVWXYZ", "23456789", "!@#$%^&*-_=+"];

/** One character from each class, then filler, then shuffled — 4 classes guaranteed. */
export function generatePassword() {
  const rand = (s) => s[Math.floor(Math.random() * s.length)];
  const all = CLASSES.join("");
  const chars = CLASSES.map(rand);
  while (chars.length < 12) chars.push(rand(all));
  for (let i = chars.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [chars[i], chars[j]] = [chars[j], chars[i]];
  }
  return chars.join("");
}
