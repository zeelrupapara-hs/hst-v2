/** |spread_diff| for slider range (always non-negative). */
export function groupSpreadMagnitude(spreadDiff) {
  const n = Number(spreadDiff) || 0;
  return Math.abs(n);
}

/** Wire balance shift allowed for spread_diff (see sql_mt5_groups_symbols). */
export function clampGroupDiffBalance(spreadDiff, balance) {
  const s = groupSpreadMagnitude(spreadDiff);
  if (s <= 0) return 0;
  const lo = Math.floor(s / 2);
  const b = Number(balance) || 0;
  return Math.min(lo, Math.max(lo - s, b));
}

/** Left-to-right balance order on the slider when spread_diff is negative. */
export function negativeBalanceSequence(spreadDiff) {
  const abs = groupSpreadMagnitude(spreadDiff);
  const lo = Math.floor(abs / 2);
  const minB = lo - abs;
  const seq = [0];
  for (let b = 1; b <= lo; b += 1) seq.push(b);
  for (let b = minB; b <= -1; b += 1) seq.push(b);
  return seq;
}

/**
 * MT5 Difference balance caption points for group spread_diff.
 * Positive spread_diff: unsigned split counts (e.g. spread 3 → "1 bid / 2 ask" at balance 0).
 * Negative spread_diff: signed values along a fixed slider order (e.g. spread -2, balance 0 → "-2 bid / 0 ask").
 */
export function groupDiffBalancePoints(spreadDiff, balance) {
  const s = Number(spreadDiff) || 0;
  const abs = groupSpreadMagnitude(s);
  const b = clampGroupDiffBalance(spreadDiff, balance);
  if (abs === 0) return { bid: 0, ask: 0 };
  const lo = Math.floor(abs / 2);

  if (s > 0) {
    const bid = lo - b;
    return { bid, ask: abs - bid };
  }

  const seq = negativeBalanceSequence(s);
  const idx = seq.indexOf(b);
  const last = seq.length - 1;
  if (idx <= 0) return { bid: s, ask: 0 };
  if (idx >= last) return { bid: 0, ask: s };
  const bid = Math.round(s * (1 - idx / last));
  return { bid, ask: s - bid };
}

export function groupDiffBalanceCaption(spreadDiff, balance) {
  const { bid, ask } = groupDiffBalancePoints(spreadDiff, balance);
  return `${bid} bid / ${ask} ask`;
}

/** Default wire balance after spread_diff changes (MT5 resets to 0). */
export function defaultGroupDiffBalance(_spreadDiff) {
  return 0;
}

/**
 * Slider position 0..|spread|.
 * MT5: thumb right = ask-heavy (0 bid / N ask); thumb left = bid-heavy (N bid / 0 ask).
 * Positive spread_diff uses inverted bid mapping; negative uses balance sequence index.
 */
export function groupDiffBalanceSliderValue(spreadDiff, balance) {
  const s = Number(spreadDiff) || 0;
  const abs = groupSpreadMagnitude(s);
  if (abs === 0) return 0;
  if (s > 0) {
    const { bid } = groupDiffBalancePoints(spreadDiff, balance);
    return abs - bid;
  }
  const seq = negativeBalanceSequence(s);
  const b = clampGroupDiffBalance(spreadDiff, balance);
  const idx = seq.indexOf(b);
  return idx >= 0 ? idx : 0;
}

export function groupDiffBalanceFromSlider(spreadDiff, sliderValue) {
  const s = Number(spreadDiff) || 0;
  const abs = groupSpreadMagnitude(s);
  if (abs === 0) return 0;
  const pos = Math.min(abs, Math.max(0, Number(sliderValue) || 0));
  if (s > 0) {
    const lo = Math.floor(abs / 2);
    const bid = abs - pos;
    return lo - bid;
  }
  const seq = negativeBalanceSequence(s);
  return seq[Math.min(seq.length - 1, pos)] ?? 0;
}
