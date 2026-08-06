export const calculateSpread = (symbol) => {
  const { last_bid, last_ask } = symbol;
  const digits = symbol?.digits || 0;
  const power = Math.pow(10, digits);
  const spread = Number(symbol?.spread || 0);
  const spreadBalance = Number(symbol?.spread_balance || 0);

  const rawBid = Number(last_bid) + spreadBalance * (1 / power);
  const rawAsk = Number(last_ask) + (spread + spreadBalance) * (1 / power);

  const newBid = Number(rawBid.toFixed(digits));
  const newAsk = Number(rawAsk.toFixed(digits));

  const diffSpread = newAsk * power - newBid * power;
  const newSpread = Math.max(0, Math.round(diffSpread));

  return { newBid, newAsk, newSpread };
};

export const calculateNetValue = (data) => {
  const { volume, contract_size, price, side } = data;
  
  const multiplier = side === 0 ? 1 : -1;
  const netValue = volume * contract_size * price * multiplier;

  return Number(netValue.toFixed(2)) || 0;
};
