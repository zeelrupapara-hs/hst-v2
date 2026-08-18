import { useEffect, useMemo, useRef, useState } from "react";
import { Checkbox, Dropdown, Input } from "antd";
import {
  LuPanelTop,
  LuCirclePlus,
  LuChartLine,
  LuEyeOff,
  LuInfo,
  LuPlus,
  LuSearch,
  LuSettings2,
} from "react-icons/lu";
import { MdAdsClick } from "react-icons/md";
import useSymbolStore from "../../store/useSymbolStore";
import useGlobalStore from "../../store/useGlobalStore";
import usePositionStore from "../../store/usePositionStore";
import useWatchlistStore from "../../store/useWatchlistStore";
import useDebounce from "../../hooks/useDebounce";
import columns from "../../columns/symbols";
import VirtualTable from "../table/VirtualTable";
import Icon from "../common/Icon";
import ContextMenuTable from "../table/ContextMenuTable";
import MultiOrderScreen from "../order/MultiOrderScreen";
import SymbolInfoModal from "./components/SymbolInfoModal";
import SymbolPicker from "./components/SymbolPicker";
import WatchlistSwitcher from "./components/WatchlistSwitcher";
import { errorToast } from "../../utils/utils";

const defaultColumns = ["symbol", "bid", "ask", "spread"];

const Symbols = () => {
  const scrollRef = useRef(null);
  const symbols = useSymbolStore((state) => state.symbols);
  const symbolGroups = useSymbolStore((state) => state.symbolGroups);
  const setGlobalStore = useGlobalStore((state) => state.setGlobalStore);
  const togglePanel = useGlobalStore((state) => state.togglePanel);
  const visibleColumns = useGlobalStore((state) => state.visibleColumns);
  const setChartSymbol = useGlobalStore((state) => state.setChartSymbol);
  const symbolOneClick = useGlobalStore((state) => state.symbolOneClick);
  const openOrderModal = useGlobalStore((state) => state.openOrderModal);
  const hasLists = useWatchlistStore((state) => state.watchlists.length > 0);
  const viewAll = useWatchlistStore((state) => state.viewAll);
  const wishlist = useWatchlistStore((state) => state.wishlist);
  const removeSymbol = useWatchlistStore((state) => state.removeSymbol);
  const showsWatchlist = hasLists && !viewAll;
  const positions = usePositionStore((state) => state.positions);
  const orders = usePositionStore((state) => state.orders);
  const [selectedGroup, setSelectedGroup] = useState("all");
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedRecord, setSelectedRecord] = useState(null);
  const [isSymbolInfoOpen, setIsSymbolInfoOpen] = useState(false);
  const [isPickerOpen, setIsPickerOpen] = useState(false);

  const debouncedSearch = useDebounce(searchTerm);

  // the active watchlist decides what the Market Watch carries; "All Symbols" and an account
  // that owns no list yet see everything its group grants
  const baseSymbols = useMemo(() => {
    if (!showsWatchlist) return Object.values(symbols);
    return wishlist.map((name) => symbols[name]).filter(Boolean);
  }, [symbols, showsWatchlist, wishlist]);

  // pills for folders the current list still reaches, so no pill filters to nothing
  const visibleGroups = useMemo(() => {
    const ids = new Set();
    for (const symbol of baseSymbols) {
      if (symbol?.symbol_class?.id != null) ids.add(symbol.symbol_class.id);
      if (symbol?.symbol_class?.parent_id != null) ids.add(symbol.symbol_class.parent_id);
    }
    return Object.keys(symbolGroups).filter((key) => ids.has(Number(key)));
  }, [baseSymbols, symbolGroups]);

  useEffect(() => {
    if (selectedGroup !== "all" && !visibleGroups.includes(selectedGroup)) {
      setSelectedGroup("all");
    }
  }, [selectedGroup, visibleGroups]);

  const filteredSymbols = useMemo(() => {
    let results = baseSymbols;

    if (selectedGroup !== "all") {
      results = results.filter(
        (symbol) =>
          symbol?.symbol_class?.id === Number(selectedGroup) ||
          symbol?.symbol_class?.parent_id === Number(selectedGroup)
      );
    }

    if (debouncedSearch.trim()) {
      results = results.filter((symbol) =>
        symbol?.symbol?.toLowerCase().includes(debouncedSearch.toLowerCase())
      );
    }

    return results;
  }, [baseSymbols, selectedGroup, debouncedSearch]);

  const filteredColumns = useMemo(() => {
    return columns.filter((col) => visibleColumns.includes(col.key));
  }, [visibleColumns]);

  useEffect(() => {
    const scrollContainer = scrollRef.current;
    if (!scrollContainer) return;

    const handleWheel = (e) => {
      if (e.deltaY === 0) return;
      e.preventDefault();
      scrollContainer.scrollLeft += e.deltaY;
    };

    scrollContainer.addEventListener("wheel", handleWheel, { passive: false });
    return () => scrollContainer.removeEventListener("wheel", handleWheel);
  }, []);

  const handleColumnChange = (key, checked) => {
    if (defaultColumns.includes(key)) return;

    const updated = checked
      ? [...visibleColumns, key]
      : visibleColumns.filter((k) => k !== key);
    setGlobalStore({ visibleColumns: updated });
  };

  const columnsMenu = {
    items: columns
      .filter((col) => col.key !== "oneClick")
      .map((col, index) => ({
        key: index + 1,
        label: (
          <div onClick={(e) => e.stopPropagation()}>
            <Checkbox
              checked={visibleColumns.includes(col.key)}
              onChange={(e) => handleColumnChange(col.key, e.target.checked)}
              disabled={defaultColumns.includes(col.key)}
            >
              {col.title}
            </Checkbox>
          </div>
        ),
      })),
  };

  const symbolMenu = {
    items: [
      {
        key: 0,
        icon: <Icon Icon={LuPanelTop} size={16} />,
        label: "Popup New Order",
        onClick: () => openOrderModal(selectedRecord?.id),
      },
      {
        key: 1,
        icon: <Icon Icon={LuCirclePlus} size={16} />,
        label: "New Order",
        onClick: () => {
          setGlobalStore({ symbol: selectedRecord?.id });
          togglePanel("order", true);
        },
      },
      {
        key: 2,
        icon: <Icon Icon={LuChartLine} size={16} />,
        label: "Chart",
        onClick: () => setChartSymbol(1, selectedRecord?.id),
      },
      {
        key: 3,
        icon: <Icon Icon={LuInfo} size={16} />,
        label: "Details",
        onClick: () => setIsSymbolInfoOpen(true),
      },
      // hiding only exists while a watchlist is displayed; hiding from "show everything" would
      // silently materialize the whole instrument set as a list
      ...(showsWatchlist
        ? [
            {
              key: 4,
              icon: <Icon Icon={LuEyeOff} size={16} />,
              label: "Hide Symbol",
              onClick: () =>
                [...positions, ...orders].some((x) => x?.symbol_id === selectedRecord?.id)
                  ? errorToast("Cannot hide: symbol has an open position or order")
                  : removeSymbol(selectedRecord?.id),
            },
          ]
        : []),
    ],
  };

  const toggleOneClick = () => {
    setGlobalStore({ symbolOneClick: !symbolOneClick });

    if (!symbolOneClick) {
      visibleColumns?.splice(1, 2, "oneClick");
      handleColumnChange("oneClick", true);
    } else {
      visibleColumns?.splice(1, 1, "bid", "ask");
      handleColumnChange("oneClick", false);
    }
  };

  // the picker takes the whole panel over in place, the way the MT5 terminal does
  if (isPickerOpen) {
    return <SymbolPicker onClose={() => setIsPickerOpen(false)} />;
  }

  return (
    <>
      <div className="flex items-center justify-between gap-2">
        <Input
          placeholder="Search..."
          className="!border-none"
          prefix={<Icon Icon={LuSearch} size={16} />}
          onChange={(e) => setSearchTerm(e?.target?.value)}
        />

        <div className="flex items-center gap-2">
          <button className="p-2" aria-label="Add Symbol" onClick={() => setIsPickerOpen(true)}>
            <Icon Icon={LuPlus} size={18} />
          </button>

          <button onClick={toggleOneClick}>
            <Icon Icon={MdAdsClick} isActive={symbolOneClick} />
          </button>

          <Dropdown
            menu={columnsMenu}
            trigger={["click"]}
            placement="bottomRight"
          >
            <button className="p-2">
              <Icon Icon={LuSettings2} size={18} />
            </button>
          </Dropdown>
        </div>
      </div>

      <div ref={scrollRef} className="flex gap-2 p-2 overflow-auto scrollbar-hide scroll-smooth">
        <WatchlistSwitcher />
        {["all", ...visibleGroups].map((key) => (
          <button
            key={key}
            className={`text-xs font-medium capitalize px-[10px] py-[2px] border rounded-full whitespace-nowrap ${
              key === selectedGroup
                ? "text-primary border-primary dark:text-white dark:border-white"
                : "border-theme-border"
            }`}
            onClick={() => setSelectedGroup(key)}
          >
            {key === "all" ? "All" : symbolGroups?.[key]?.desc}
          </button>
        ))}
      </div>

      {showsWatchlist && !baseSymbols.length ? (
        <div className="flex h-1/2 items-center justify-center p-4 text-center text-sm text-gray">
          Your Market Watch is empty — tap + to add symbols
        </div>
      ) : (
        <ContextMenuTable
          menu={symbolMenu}
          setSelectedRecord={setSelectedRecord}
          onRowDoubleClick={(record) => {
            setGlobalStore({ symbol: record?.id });
            setChartSymbol(1, record?.id);
            togglePanel("order", true);
          }}
          isDraggable={true}
        >
          <VirtualTable columns={filteredColumns} data={filteredSymbols} />
        </ContextMenuTable>
      )}

      <SymbolInfoModal
        isOpen={isSymbolInfoOpen}
        setIsOpen={setIsSymbolInfoOpen}
        data={selectedRecord}
      />

      <MultiOrderScreen />
    </>
  );
};

export default Symbols;
