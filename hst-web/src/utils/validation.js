import dayjs from "dayjs";

const MIN_PRICE = 0;
const MAX_PRICE = 1000000;

export const validateOrderField = (key, value, range = {}, data = {}) => {
  const { min, max } = range;
  let error = false;

  if (["volume", "order_price"].includes(key) && !value) return true;
  else error = false;

  switch (key) {
    case "volume":
    case "order_price":
    case "stop_loss":
    case "take_profit":
      if (value && (min || max) && (value < min || value > max)) error = true;
      else error = false;
      break;

    case "expiry_at":
      if (
        [2, 3].includes(data?.expiration_policy) &&
        dayjs.unix(value).isBefore(dayjs())
      )
        error = true;
      else error = false;
      break;

    default:
      break;
  }

  return error;
};

export const getOrderPrice = (data) => {
  const { mode, type, side, last_ask, last_bid, order_price } = data;

  if (mode === "add") return order_price;
  if (mode === "edit" && !type) return side === 0 ? last_bid : last_ask;
  return order_price;
};

export const getPriceRange = (data) => {
  const { name, type, side, last_ask, last_bid } = data;
  const orderPrice = getOrderPrice(data);

  if (name === "order_price" && (last_ask || last_bid)) {
    if (type === 1)
      return { min: MIN_PRICE, max: last_ask, range: `< ${last_ask}` }; // Buy Limit

    if (type === 2)
      return { min: last_ask, max: MAX_PRICE, range: `> ${last_ask}` }; // Buy Stop

    if (type === 3)
      return { min: last_bid, max: MAX_PRICE, range: `> ${last_bid}` }; // Sell Limit

    if (type === 4)
      return { min: MIN_PRICE, max: last_bid, range: `< ${last_bid}` }; // Sell Stop
  }

  if (name === "stop_loss" && orderPrice) {
    if (type) {
      if ([1, 2].includes(type))
        return { min: MIN_PRICE, max: orderPrice, range: `< ${orderPrice}` }; // Buy Limit / Buy Stop Order

      if ([3, 4].includes(type))
        return { min: orderPrice, max: MAX_PRICE, range: `> ${orderPrice}` }; // Sell Limit / Sell Stop Order
    } else {
      if (side === 0) {
        return { min: MIN_PRICE, max: orderPrice, range: `< ${orderPrice}` }; // Buy Position
      } else {
        return { min: orderPrice, max: MAX_PRICE, range: `> ${orderPrice}` }; // Sell Position
      }
    }
  }

  if (name === "take_profit" && orderPrice) {
    if (type) {
      if ([1, 2].includes(type))
        return { min: orderPrice, max: MAX_PRICE, range: `> ${orderPrice}` }; // Buy Limit / Buy Stop Order

      if ([3, 4].includes(type))
        return { min: MIN_PRICE, max: orderPrice, range: `< ${orderPrice}` }; // Sell Limit / Sell Stop Order
    } else {
      if (side === 0) {
        return { min: orderPrice, max: MAX_PRICE, range: `> ${orderPrice}` }; // Buy Position
      } else {
        return { min: MIN_PRICE, max: orderPrice, range: `< ${orderPrice}` }; // Sell Position
      }
    }
  }

  return { min: MIN_PRICE, max: MAX_PRICE, range: "" };
};

export const getSide = (type, side) => {
  if (!type) return side; // For position
  return [1, 2].includes(type) ? 0 : 1; // For order
};
