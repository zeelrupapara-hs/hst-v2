import { createContext, useCallback, useContext, useEffect, useMemo, useRef } from "react";
import useWebSocket from "react-use-websocket";
import useAuthStore from "../store/useAuthStore";
import useSymbolStore from "../store/useSymbolStore";
import useLiveSymbolStore from "../store/useLiveSymbolStore";
import { blobToUtf8, getLocalItem, qs } from "../utils/utils";
import { calculateSpread } from "../utils/calculations";
import { SOCKET_EVENTS } from "./events";
import { toServerPayload } from "./adapt";
import { handleSocketMessage } from "./handler";
import usePositionStore from "../store/usePositionStore";
import { publishNewTick } from "../components/chart/components/ChartComponent";

const SocketContext = createContext();

export const SocketProvider = ({ children }) => {
  const user = useAuthStore((state) => state.user);
  const symbols = useSymbolStore((state) => state.symbols);
  const liveSymbols = useLiveSymbolStore((state) => state.liveSymbols);
  const updateSymbolPrice = useLiveSymbolStore(
    (state) => state.updateSymbolPrice
  );
  const updatePositionPL = usePositionStore((state) => state.updatePositionPL);
  const updateSummary = usePositionStore((state) => state.updateSummary);

  const lastTickAt = useRef(0);

  // Read the token at connect time, not at render time. An access token lives two hours while
  // a session lives a week, so a socket that captured the token once would reconnect for the
  // rest of the week with a credential the server has already stopped accepting.
  const socketUrl = useCallback(() => {
    const params = {
      session_id: useAuthStore.getState().user?.session_id,
      access_token: getLocalItem("token"),
    };
    const wsBase =
      import.meta.env.VITE_SOCKET_URL ||
      `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/ws`;
    const url = `${wsBase}?${qs(params)}`;

    console.info("[socket] connecting to", wsBase);

    return url;
  }, []);

  const { sendJsonMessage, lastJsonMessage, lastMessage, readyState } =
    useWebSocket(socketUrl, { share: true, shouldReconnect: () => true });

  useEffect(() => {
    if (readyState === 1)
      sendJsonMessage({ type: SOCKET_EVENTS.START_MARKET_FEED });
  }, [readyState]);

  const handleBlob = async (blob) => {
    const utf8Text = await blobToUtf8(blob);
    const type = utf8Text.split(",")[0];

    if (type !== "summary") {
      // symbol,bid,ask,last,volume,ts,open,high,low,close,change,change_percent
      const [symbolId, last_bid, last_ask, , , , , high_bid, low_bid, tickClose] =
        utf8Text.split(",");
      const { digits } = symbols?.[symbolId] || {};

      // the tick carries the session close, so the change is right even before the rest loads
      const close = Number(tickClose) || symbols?.[symbolId]?.close;

      const { newBid, newAsk, newSpread } = calculateSpread({
        ...symbols?.[symbolId],
        last_bid,
        last_ask,
      });

      const updatedData = {
        high_bid,
        low_bid,
        newSpread,
        last_bid: newBid,
        last_ask: newAsk,
        netChange: (newBid - close).toFixed(digits),
        percentChange: (((newBid - close) / close) * 100).toFixed(2),
        ...(liveSymbols?.[symbolId]?.last_bid > newBid
          ? { bidColor: "red" }
          : { bidColor: "green" }),
        ...(liveSymbols?.[symbolId]?.last_ask < newAsk
          ? { askColor: "red" }
          : { askColor: "green" }),
      };
      lastTickAt.current = Date.now();
      updateSymbolPrice(symbolId, updatedData);
      publishNewTick({ ...updatedData, ts: Date.now() }, symbolId);
    }
  };

  const handleSummary = (data) => {
    const dataArray = data?.split(",");

    const keys = [
      "ignored1",
      "ignored2",
      "balance",
      "credit",
      "equity",
      "used_margin",
      "free_margin",
      "margin_level",
      "profit",
    ];

    const summary = keys.reduce((acc, key, index) => {
      if (!key.startsWith("ignored")) {
        acc[key] = dataArray[index];
      }
      return acc;
    }, {});

    updateSummary(summary);

    const positionData = dataArray.slice(9);

    for (let i = 0; i < positionData.length; i += 2) {
      const positionId = positionData[i];
      const profitLoss = Number(positionData[i + 1]);

      if (positionId && !isNaN(profitLoss))
        updatePositionPL(positionId, profitLoss);
    }
  };

  useEffect(() => {
    if (lastMessage?.data instanceof Blob) {
      handleBlob(lastMessage?.data);
    } else {
      const type = lastMessage?.data?.split?.(",")?.[0];
      if (type === "summary") handleSummary(lastMessage?.data);
    }
  }, [lastMessage]);

  useEffect(() => {
    if (lastJsonMessage) handleSocketMessage(lastJsonMessage);
  }, [lastJsonMessage]);

  const sendEvent = (type, payload) => {
    sendJsonMessage({ type, payload: toServerPayload(type, payload) });
  };

  const socketValue = useMemo(() => {
    return {
      sendJsonMessage,
      sendEvent,
      readyState,
      lastTickAt,
    };
  }, [sendJsonMessage, readyState]);

  return (
    <SocketContext.Provider value={socketValue}>
      {children}
    </SocketContext.Provider>
  );
};

export const useSocket = () => useContext(SocketContext);
