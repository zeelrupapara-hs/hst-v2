const inferDigits = (bid, ask, digits) => {
  const d = Number(digits);
  if (Number.isFinite(d) && d > 0) return d;
  const sample = String(bid ?? ask ?? "");
  const dot = sample.indexOf(".");
  return dot >= 0 ? sample.length - dot - 1 : 0;
};

export const calculateSpread = ({ last_bid, last_ask, digits = 0 }) => {
  const bid = Number(last_bid);
  const ask = Number(last_ask);
  if (!Number.isFinite(bid) || !Number.isFinite(ask)) {
    return { newBid: bid, newAsk: ask, newSpread: null };
  }
  const power = 10 ** inferDigits(bid, ask, digits);
  // MT5 spread in points: distance between bid and ask (always non-negative).
  const newSpread = Math.abs(Math.round(ask * power - bid * power));
  return { newBid: bid, newAsk: ask, newSpread };
};

export const calculateNetValue = (data) => {
  const { volume, contract_size, price, side } = data;
  
  const multiplier = side === 0 ? 1 : -1;
  const netValue = volume * contract_size * price * multiplier;

  return Number(netValue.toFixed(2)) || 0;
};
