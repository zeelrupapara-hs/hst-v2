import { useEffect, useRef, useState } from "react";
import useSymbolStore from "../../../store/useSymbolStore";
import useGlobalStore from "../../../store/useGlobalStore";
import useChartLayoutStore from "../../../store/useChartLayoutStore";
import { useTheme } from "../../../context/ThemeContext";
import { widget } from "../../../assets/charting_library";
import { marketHistory } from "../../../api/request/symbol";
import { applyChartInactiveState } from "../../../utils/chartInactive";
import ChartInactiveOverlay from "./ChartInactiveOverlay";

export const handlerMap = new Map();
export const channelToSubscription = new Map();

export const publishNewTick = (tick, symbolId) => {
  try {
    // subscribeBars keys these maps by symbolInfo.id, which is the symbol itself, so the id
    // is used as it arrives; coercing it to a number turned every lookup into NaN
    const subscriptionItem = channelToSubscription.get(symbolId);
    if (subscriptionItem === undefined) return;

    const bar = {
      time: tick.ts,
      open: tick.last_bid,
      high: tick.last_bid,
      low: tick.last_bid,
      close: tick.last_bid,
    };

    const handler = handlerMap.get(symbolId);
    handler?.callback(bar);
  } catch (error) {
    console.error("❌ Error publishing new tick", error);
  }
};

const themeOverrides = (darkMode) => ({
  "paneProperties.background": darkMode ? "#101013" : "#ffffff",
  "paneProperties.backgroundType": "solid",
  "scalesProperties.textColor": darkMode ? "#89898b" : "#151924",
});

const ChartComponent = ({ chartId = 1, symbolId = "25100016", isLive = true }) => {
  const lastBarsCacheRef = useRef(new Map());
  const chartContainerRef = useRef(null);
  const chartWidgetRef = useRef(null);
  const isLiveRef = useRef(isLive);
  const readyRef = useRef(false);
  const [useExternalOverlay, setUseExternalOverlay] = useState(false);
  const symbols = useSymbolStore((state) => state?.symbols);
  const { theme } = useTheme();

  const darkMode = theme === "dark";

  // The widget lives for the whole mount while props keep changing, so the
  // datafeed and the ready callbacks read these refs instead of stale closures.
  const symbolsRef = useRef(symbols);
  const darkModeRef = useRef(darkMode);
  isLiveRef.current = isLive;
  symbolsRef.current = symbols;
  darkModeRef.current = darkMode;

  const syncInactiveState = () => {
    const applied = applyChartInactiveState(chartContainerRef.current, !isLiveRef.current, {
      darkMode: darkModeRef.current,
    });

    setUseExternalOverlay(!applied && !isLiveRef.current);
  };

  const configurationData = {
    supported_resolutions: [
      "1",
      "5",
      "15",
      "30",
      "60",
      "240",
      "1D",
      "1W",
      "1M",
    ],
    symbols_types: [{ name: "All", value: "all" }],
  };

  const resolutionMapper = (resolution) => {
    // check if resolution is a number
    const isNumber = /^\d+$/.test(resolution);
    if (isNumber) {
      return resolution + "m";
    } else if (resolution.endsWith("M")) {
      return resolution.slice(0, -1) + "mo";
    } else {
      return resolution.toLowerCase();
    }
  };

  useEffect(() => {
    if (!window.TradingView) {
      console.error("❌ TradingView library not loaded!");
      return;
    }

    const accessList = {
      type: "black",
      tools: [
        { name: "VWAP", grayed: false },
        { name: "52 Week High/Low", grayed: false },
        { name: "Heikin Ashi", grayed: false },
      ],
    };

    const dataFeed = {
      onReady: (callback) => {
        const symbolsTypeMap = new Map();

        Object.values(symbolsRef.current ?? {})?.forEach((symbol) => {
          if (symbolsTypeMap.has(symbol?.symbol_class?.desc)) return;

          symbolsTypeMap.set(symbol?.symbol_class?.desc, true);

          const symbolType = {
            name: symbol?.symbol_class?.desc,
            value: symbol?.symbol_class?.desc?.toLowerCase(),
          };

          configurationData.symbols_types?.push(symbolType);
        });

        setTimeout(() => callback(configurationData), 0);
      },

      resolveSymbol: async (symbolName, onSymbolResolvedCallback) => {
        const symbol = Object.values(symbolsRef.current ?? {}).find(
          (item) => item?.symbol === symbolName
        );

        const symbolInfo = {
          id: symbol?.id,
          symbol: symbol?.symbol,
          name: symbol?.symbol,
          full_name: symbol?.symbol,
          description: symbol?.symbol,
          type: symbol?.symbol_class?.desc,
          ticker: symbol?.symbol,
          timezone: "UTC",
          session: "24x7",
          minmov: 1,
          pricescale: 100000,
          has_intraday: true,
          has_daily: true,
          has_weekly_and_monthly: true,
          supported_resolutions: configurationData.supported_resolutions,
          data_status: "streaming",
          exchange: symbol?.symbol_class?.desc || "",
          format: "price",
        };

        setTimeout(() => onSymbolResolvedCallback(symbolInfo), 0);
      },

      searchSymbols: (
        userInput,
        exchange,
        symbolType,
        onResultReadyCallback
      ) => {
        if (!Object.values(symbolsRef.current ?? {}).length) return;

        const lowerInput = userInput?.toLowerCase() || "";
        const lowerType = symbolType?.toLowerCase() || "";

        const filteredSymbols = Object.values(symbolsRef.current ?? {})
          .filter(({ symbol, symbol_class }) => {
            const matchesInput = symbol?.toLowerCase().includes(lowerInput);
            if (lowerType && lowerType !== "all") {
              return (
                symbol_class?.desc?.toLowerCase() === lowerType && matchesInput
              );
            }
            return matchesInput;
          })
          .map(({ symbol, desc, symbol_class }) => ({
            symbol,
            full_name: symbol,
            description: desc,
            type: symbol_class?.desc,
            ticker: symbol,
            exchange: symbol_class?.desc,
          }));

        onResultReadyCallback(filteredSymbols);
      },

      getBars: async (
        symbolInfo,
        resolution,
        periodParams,
        onHistoryCallback
      ) => {
        if (!isLiveRef.current) {
          onHistoryCallback([], { noData: true });
          return;
        }

        const { from, to, firstDataRequest, countBack } = periodParams;

        const symbol = Object.values(symbolsRef.current ?? {}).find(
          (symbol) => symbol?.symbol === symbolInfo?.name
        );

        if (symbol) {
          const { id } = symbol;
          const params = {
            symbol_id: id,
            resolution: resolutionMapper(resolution),
            from,
            to,
            count_back: countBack,
          };

          const { data: response } = await marketHistory(params);
          const data = response?.data;

          if (
            data?.length === 0 &&
            (resolution === "1" || resolution === "5")
          ) {
            let hasData = false;

            for (let i = 0; i <= 4; i++) {
              let fromDate = from;

              if (resolution === "1") {
                fromDate = from - 21600 * i * i;
              } else {
                fromDate = from - 21600 * i * 4 * i;
              }

              const params = {
                symbol_id: id,
                resolution: resolutionMapper(resolution),
                from: fromDate,
                to,
                count_back: countBack,
              };

              const { data: historyData } = await marketHistory(params);

              if (historyData?.length) {
                hasData = true;

                if (firstDataRequest) {
                  lastBarsCacheRef.current.set(symbolInfo?.id, {
                    ...historyData[historyData.length - 1],
                  });
                }

                onHistoryCallback(historyData, { noData: false });
                break;
              }
            }

            if (!hasData) onHistoryCallback([], { noData: true });
          } else {
            if (data?.length === 0) {
              onHistoryCallback([], { noData: true });
            } else {
              if (firstDataRequest) {
                lastBarsCacheRef.current.set(symbolInfo?.id, {
                  ...data[data.length - 1],
                });
              }
              onHistoryCallback(data, { noData: false });
            }
          }
        } else {
          onHistoryCallback([], { noData: true });
        }
      },

      subscribeBars: (
        symbolInfo,
        resolution,
        onRealtimeCallback,
        subscriberUID
      ) => {
        if (!isLiveRef.current) return;

        const handler = { id: subscriberUID, callback: onRealtimeCallback };
        handlerMap.set(symbolInfo?.id, handler);

        let subscriptionItem = channelToSubscription.get(symbolInfo?.id);
        if (subscriptionItem) {
          subscriptionItem.handlers.push(handler);
          return;
        }

        subscriptionItem = {
          subscriberUID,
          resolution,
          lastDailyBar: lastBarsCacheRef.current.get(symbolInfo?.id),
          handlers: [handler],
        };
        channelToSubscription.set(symbolInfo?.id, subscriptionItem);
      },

      unsubscribeBars: (subscriberUID) => {
        const currencyPair = subscriberUID.split("_#_")[0];
        handlerMap.delete(currencyPair);
      },
    };

    // The trader's saved state, fetched before any chart renders; its symbol,
    // interval and studies win over the defaults below.
    const savedState = useChartLayoutStore.getState().layouts[chartId];
    const constructedDark = darkModeRef.current;

    const chartWidget = new widget({
      studies_access: accessList,
      autosize: true,
      symbol: symbolsRef.current?.[symbolId]?.symbol,
      interval: "1",
      timezone: "UTC",
      locale: "en",
      range: "12M",
      hide_top_toolbar: false,
      allow_symbol_change: true,
      enable_publishing: false,
      style: "1",
      theme: constructedDark ? "dark" : "light",
      container: chartContainerRef.current,
      datafeed: dataFeed,
      library_path: "/charting_library_cloned_data/charting_library/",
      auto_save_delay: 5,
      ...(savedState ? { saved_data: savedState } : { studies: ["Moving Average"] }),
      settings_adapter: {
        initialSettings: useChartLayoutStore.getState().settings,
        setValue: (key, value) =>
          useChartLayoutStore.getState().setSetting(key, value),
        removeValue: (key) => useChartLayoutStore.getState().removeSetting(key),
      },
      withdateranges: true,
      hide_side_toolbar: false,
      volumePaneSize: "hide",
      enabled_features: [],
      // rtl: direction === 'rtl',
      custom_css_url: "/tradingview-theme.css",
      overrides: {
        "scalesProperties.showSymbolLabels": false,
        "mainSeriesProperties.highLowAvgPrice.highLowPriceLabelsVisible": true,
        ...themeOverrides(constructedDark),
      },
      disabled_features: [
        "use_localstorage_for_settings",
        "popup_hints",
        "volume_force_overlay",
        "create_volume_indicator_by_default",
      ],
    });

    chartWidgetRef.current = chartWidget;

    const saveNow = () => {
      chartWidget.save((state) =>
        useChartLayoutStore.getState().saveLayoutDebounced(chartId, state)
      );
    };

    chartWidget.onChartReady(() => {
      readyRef.current = true;

      chartWidget.applyOverrides(themeOverrides(darkModeRef.current));
      // the theme may have been toggled between construction and ready
      if (darkModeRef.current !== constructedDark) {
        chartWidget
          .changeTheme(darkModeRef.current ? "dark" : "light", { disableUndo: true })
          .then(() => chartWidget.applyOverrides(themeOverrides(darkModeRef.current)));
      }

      // silent MT5-style persistence: the library raises this on every change
      // worth keeping, auto_save_delay seconds after the trader's last touch
      chartWidget.subscribe("onAutoSaveNeeded", saveNow);

      // keep the slot assignment in step with the symbol the chart actually
      // shows, whether restored from the blob or picked in the chart's own
      // toolbar; direct setGlobalStore since setChartSymbol's duplicate guard
      // would block this sync
      const syncStoreSymbol = () => {
        const name = chartWidget.activeChart().symbol();
        const { chartSymbols, setGlobalStore } = useGlobalStore.getState();
        if (name && chartSymbols?.[chartId] !== name) {
          setGlobalStore({
            chartSymbols: { ...(chartSymbols ?? {}), [chartId]: name },
          });
        }
      };
      syncStoreSymbol();
      chartWidget
        .activeChart()
        .onSymbolChanged()
        .subscribe(null, () => {
          syncStoreSymbol();
          saveNow();
        });

      syncInactiveState();
      setTimeout(syncInactiveState, 300);
    });

    return () => {
      const w = chartWidgetRef.current;
      const wasReady = readyRef.current;
      chartWidgetRef.current = null;
      readyRef.current = false;
      if (!w) return;

      try {
        if (wasReady) {
          let saved = false;
          w.save((state) => {
            saved = true;
            useChartLayoutStore.getState().flushLayout(chartId, state);
          });
          // if the save could not complete before teardown, push the last
          // autosaved copy instead
          if (!saved) useChartLayoutStore.getState().flushLayout(chartId);
        }
      } catch {
        // widget already torn down
      }
      try {
        w.remove();
      } catch {
        // iframe already gone
      }
    };
    // mount-once: symbol and theme changes are applied to the live widget below
    // instead of rebuilding it, which would wipe the trader's session state
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chartId]);

  // symbol changes (drag-drop, market watch) retarget the live chart
  useEffect(() => {
    const w = chartWidgetRef.current;
    if (!w || !readyRef.current) return;

    const name = symbolsRef.current?.[symbolId]?.symbol;
    if (!name) return;

    try {
      if (w.activeChart().symbol() === name) return;
      // capture the outgoing symbol's state before the switch
      w.save((state) =>
        useChartLayoutStore.getState().saveLayoutDebounced(chartId, state)
      );
      w.activeChart().setSymbol(name);
    } catch {
      // chart not ready yet; the constructor symbol covers first paint
    }
  }, [symbolId, chartId]);

  // theme changes restyle the live chart instead of rebuilding it
  useEffect(() => {
    const w = chartWidgetRef.current;
    if (!w || !readyRef.current) return;

    w.changeTheme(darkMode ? "dark" : "light", { disableUndo: true }).then(() => {
      chartWidgetRef.current?.applyOverrides(themeOverrides(darkMode));
    });
  }, [darkMode]);

  useEffect(() => {
    syncInactiveState();

    if (isLive && chartWidgetRef.current) {
      try {
        chartWidgetRef.current.activeChart()?.resetData();
      } catch {
        // Chart may not be ready yet.
      }
    }
  }, [isLive, darkMode]);

  return (
    <div className="relative h-full w-full">
      <div ref={chartContainerRef} className="h-full w-full" />
      {!isLive && useExternalOverlay ? <ChartInactiveOverlay /> : null}
    </div>
  );
};

export default ChartComponent;
