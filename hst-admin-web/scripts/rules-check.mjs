// The two rules no linter ships: no vendor naming, and one-way imports.
// Run by `npm run lint`; a violation fails the build the moment it is written.
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";

const SRC = new URL("../src", import.meta.url).pathname;
const failures = [];

function walk(dir) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) walk(p);
    else if (/\.(jsx?|css)$/.test(name)) check(p);
  }
}

function check(path) {
  const rel = relative(SRC, path);
  const text = readFileSync(path, "utf8");

  // rule 3: no vendor naming anywhere
  const vendor = text.match(/mt5|metatrader|metaquotes|mql5/i);
  if (vendor) failures.push(`${rel}: vendor naming "${vendor[0]}"`);

  // rule 2: a module may not import from a sibling module
  const mod = rel.match(/^modules\/([^/]+)\//);
  if (mod) {
    for (const m of text.matchAll(/from\s+["']([^"']+)["']/g)) {
      const hit = m[1].match(/modules\/([^/]+)/);
      if (hit && hit[1] !== mod[1])
        failures.push(`${rel}: imports sibling module "${hit[1]}"`);
    }
  }
}

walk(SRC);

if (failures.length) {
  console.error("rules-check failed:");
  for (const f of failures) console.error("  " + f);
  process.exit(1);
}
console.log("rules-check ok");
