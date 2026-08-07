import {
  defaultGroupDiffBalance,
  groupDiffBalanceCaption,
  groupDiffBalanceFromSlider,
  groupDiffBalanceSliderValue,
} from "../src/lib/groupSpreadBalance.js";

const cases = [
  { spread: -2, balance: 0, caption: "-2 bid / 0 ask", label: "spread=-2 balance=0" },
  { spread: 2, balance: 0, caption: "1 bid / 1 ask", label: "spread=2 balance=0" },
  { spread: 3, balance: 0, caption: "1 bid / 2 ask", label: "spread=3 balance=0" },
  { spread: -3, balance: 0, caption: "-3 bid / 0 ask", label: "spread=-3 balance=0" },
  { spread: 3, balance: -2, caption: "3 bid / 0 ask", label: "spread=3 slider left (MT5)" },
  { spread: 3, balance: 1, caption: "0 bid / 3 ask", label: "spread=3 slider right (MT5)" },
  { spread: -2, balance: -1, caption: "0 bid / -2 ask", label: "spread=-2 slider right" },
  { spread: -2, balance: 1, caption: "-1 bid / -1 ask", label: "spread=-2 center" },
];

let failed = 0;
for (const { spread, balance, caption, label } of cases) {
  const got = groupDiffBalanceCaption(spread, balance);
  if (got !== caption) {
    console.error(`FAIL ${label}: expected "${caption}", got "${got}"`);
    failed += 1;
  } else {
    console.log(`ok ${label}: ${got}`);
  }
}

for (const spread of [-2, 2, -3, 3]) {
  const def = defaultGroupDiffBalance(spread);
  const cap = groupDiffBalanceCaption(spread, def);
  console.log(`default spread=${spread} balance=${def} → ${cap}`);
}

for (const [spread, balance, slider] of [
  [-2, 0, 0],
  [-2, -1, 2],
  [-2, 1, 1],
  [2, 0, 1],
  [3, -2, 0],
  [3, 1, 3],
]) {
  const val = groupDiffBalanceSliderValue(spread, balance);
  if (val !== slider) {
    console.error(`FAIL slider spread=${spread} balance=${balance}: expected ${slider}, got ${val}`);
    failed += 1;
  }
  const roundTrip = groupDiffBalanceFromSlider(spread, val);
  if (roundTrip !== balance) {
    console.error(`FAIL round-trip spread=${spread} slider=${slider}: expected balance ${balance}, got ${roundTrip}`);
    failed += 1;
  }
}

for (const spread of [-5, -3, -2, 2, 3, 5]) {
  const abs = Math.abs(spread);
  for (let pos = 0; pos <= abs; pos += 1) {
    const bal = groupDiffBalanceFromSlider(spread, pos);
    const back = groupDiffBalanceSliderValue(spread, bal);
    if (back !== pos) {
      console.error(`FAIL drag spread=${spread} pos=${pos} balance=${bal} back=${back}`);
      failed += 1;
    }
  }
}

if (failed > 0) {
  process.exit(1);
}
console.log("all group spread balance tests passed");
