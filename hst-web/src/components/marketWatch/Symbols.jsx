import { useMemo, useState } from "react";
import { Checkbox, Dropdown, Input } from "antd";
import {
  LuPanelTop,
  LuCirclePlus,
  LuChartLine,
  LuEye,
  LuEyeOff,
  LuInfo,
  LuPlus,
  LuSearch,
  LuSettings2,
} from "react-icons/lu";
import { MdAdsClick } from "react-icons/md";
import useSymbolStore from "../../store/useSymbolStore";
import useGlobalStore from "../../store/useGlobalStore";
import useWatchlistStore from "../../store/useWatchlistStore";
import useDebounce from "../../hooks/useDebounce";
import columns from "../../columns/symbols";
import VirtualTable from "../table/VirtualTable";
import Icon from "../common/Icon";
import ContextMenuTable from "../table/ContextMenuTable";
import MultiOrderScreen from "../order/MultiOrderScreen";
import SymbolInfoModal from "./components/SymbolInfoModal";
import SymbolPicker from "./components/SymbolPicker";

const defaultColumns = ["symbol", "bid", "ask", "spread"];

const Symbols = () => {
  const symbols = useSymbolStore((state) => state.symbols);
  const setGlobalStore = useGlobalStore((state) => state.setGlobalStore);
  const togglePanel = useGlobalStore((state) => state.togglePanel);
  const visibleColumns = useGlobalStore((state) => state.visibleColumns);
  const setChartSymbol = useGlobalStore((state) => state.setChartSymbol);
  const symbolOneClick = useGlobalStore((state) => state.symbolOneClick);
  const openOrderModal = useGlobalStore((state) => state.openOrderModal);
  const watchlistId = useWatchlistStore((state) => state.watchlistId);
  const wishlist = useWatchlistStore((state) => state.wishlist);
  const removeSymbol = useWatchlistStore((state) => state.removeSymbol);
  const addSymbol = useWatchlistStore((state) => state.addSymbol);
  const tab = useGlobalStore((state) => state.marketWatchTab) || "all";
  const onWatchlistTab = tab === "watchlist";
  // hiding, and the empty-list message, only make sense while the account's own list shows
  const showsWatchlist = onWatchlistTab && watchlistId != null;
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedRecord, setSelectedRecord] = useState(null);
  const [isSymbolInfoOpen, setIsSymbolInfoOpen] = useState(false);
  const [isPickerOpen, setIsPickerOpen] = useState(false);

  const debouncedSearch = useDebounce(searchTerm);

  // the All Symbols tab carries everything the account's group grants; the Watchlist tab
  // carries only the account's own list, in its own order
  const baseSymbols = useMemo(() => {
    if (!onWatchlistTab) return Object.values(symbols);
    return wishlist.map((name) => symbols[name]).filter(Boolean);
  }, [symbols, onWatchlistTab, wishlist]);

  const filteredSymbols = useMemo(() => {
    if (!debouncedSearch.trim()) return baseSymbols;
    return baseSymbols.filter((symbol) =>
      symbol?.symbol?.toLowerCase().includes(debouncedSearch.toLowerCase())
    );
  }, [baseSymbols, debouncedSearch]);

  const filteredColumns = useMemo(() => {
    return columns.filter((col) => visibleColumns.includes(col.key));
  }, [visibleColumns]);

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
      // hiding only exists while the account's own list is displayed; on All Symbols the same
      // slot offers the opposite: putting the instrument on the list. A symbol with an open
      // position may leave the list: it stays priced and reachable on the All Symbols tab.
      ...(showsWatchlist
        ? [
            {
              key: 4,
              icon: <Icon Icon={LuEyeOff} size={16} />,
              label: "Hide Symbol",
              onClick: () => removeSymbol(selectedRecord?.id),
            },
          ]
        : []),
      ...(!onWatchlistTab && !wishlist.includes(selectedRecord?.id)
        ? [
            {
              key: 5,
              icon: <Icon Icon={LuEye} size={16} />,
              label: "Add to Watchlist",
              onClick: () => addSymbol(selectedRecord?.id),
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

  const tabClass = (active) =>
    `flex-1 py-2.5 text-xs font-semibold uppercase tracking-wider border-b-2 transition-colors ${
      active ? "text-primary border-primary" : "text-gray border-transparent hover:text-theme-text"
    }`;

  return (
    <>
      <div className="flex border-b border-theme-border">
        <button
          className={tabClass(!onWatchlistTab)}
          onClick={() => setGlobalStore({ marketWatchTab: "all" })}
        >
          All Symbols
        </button>
        <button
          className={tabClass(onWatchlistTab)}
          onClick={() => setGlobalStore({ marketWatchTab: "watchlist" })}
        >
          Watchlist
        </button>
      </div>

      <div className="flex items-center justify-between gap-2">
        <Input
          placeholder="Search..."
          className="!border-none"
          prefix={<Icon Icon={LuSearch} size={16} />}
          onChange={(e) => setSearchTerm(e?.target?.value)}
        />

        <div className="flex items-center gap-2">
          {onWatchlistTab && (
            <button className="p-2" aria-label="Add Symbol" onClick={() => setIsPickerOpen(true)}>
              <Icon Icon={LuPlus} size={18} />
            </button>
          )}

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

      {onWatchlistTab && !baseSymbols.length ? (
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
