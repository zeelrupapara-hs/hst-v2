import { useMemo, useState } from "react";
import { Input } from "antd";
import {
  LuCheck,
  LuChevronLeft,
  LuChevronRight,
  LuPlus,
  LuSearch,
  LuX,
} from "react-icons/lu";
import useSymbolStore from "../../../store/useSymbolStore";
import usePositionStore from "../../../store/usePositionStore";
import useWatchlistStore from "../../../store/useWatchlistStore";
import useDebounce from "../../../hooks/useDebounce";
import Icon from "../../common/Icon";
import { errorToast } from "../../../utils/utils";

// the picker's categories are MT5's top folder, read straight off each symbol's path
const topFolder = (symbol) => String(symbol?.path || "").split("\\")[0] || "Other";

const SymbolRow = ({ symbol, inWishlist, locked, onToggle }) => (
  <div className="flex items-center justify-between gap-2 px-3 py-2 border-b border-theme-border last:border-b-0">
    <div className="min-w-0">
      <div className="text-sm font-semibold text-theme-text truncate">{symbol.symbol}</div>
      <div className="text-xs text-gray truncate">{symbol.description}</div>
    </div>
    <button
      className={`p-1 shrink-0 ${locked ? "cursor-default" : ""}`}
      aria-label={`${inWishlist ? "Remove" : "Add"} ${symbol.id}`}
      onClick={() => onToggle(symbol.id)}
    >
      {inWishlist ? (
        <span
          className={`flex h-5 w-5 items-center justify-center rounded-full ${
            locked ? "bg-gray" : "bg-primary"
          }`}
        >
          <LuCheck size={12} className="text-white" />
        </span>
      ) : (
        <Icon Icon={LuPlus} size={18} />
      )}
    </button>
  </div>
);

// The MT5-style symbol picker. It takes the Market Watch panel over in place: search on top,
// then the categories, a drilled-in folder, or the flat search matches. The X hands the panel
// back to the quotes list.
const SymbolPicker = ({ onClose }) => {
  const symbols = useSymbolStore((state) => state.symbols);
  const positions = usePositionStore((state) => state.positions);
  const orders = usePositionStore((state) => state.orders);
  const wishlist = useWatchlistStore((state) => state.wishlist);
  const addSymbol = useWatchlistStore((state) => state.addSymbol);
  const removeSymbol = useWatchlistStore((state) => state.removeSymbol);

  const [folder, setFolder] = useState(null);
  const [searchTerm, setSearchTerm] = useState("");
  const debouncedSearch = useDebounce(searchTerm);

  const wishlistSet = useMemo(() => new Set(wishlist), [wishlist]);

  // a symbol carrying an open position or a pending order cannot leave the Market Watch
  const locked = useMemo(
    () => new Set([...positions, ...orders].map((x) => x?.symbol_id)),
    [positions, orders]
  );

  const categories = useMemo(() => {
    const byFolder = {};
    for (const symbol of Object.values(symbols)) {
      (byFolder[topFolder(symbol)] ??= []).push(symbol);
    }
    return Object.keys(byFolder)
      .sort((a, b) => a.localeCompare(b))
      .map((name) => ({
        name,
        symbols: byFolder[name].sort((a, b) => a.symbol.localeCompare(b.symbol)),
        selected: byFolder[name].filter((s) => wishlistSet.has(s.id)).length,
      }));
  }, [symbols, wishlistSet]);

  const searched = useMemo(() => {
    const term = debouncedSearch.trim().toLowerCase();
    if (!term) return null;

    return Object.values(symbols)
      .filter(
        (s) =>
          s?.symbol?.toLowerCase().includes(term) ||
          s?.description?.toLowerCase().includes(term)
      )
      .sort((a, b) => a.symbol.localeCompare(b.symbol));
  }, [symbols, debouncedSearch]);

  const toggle = (name) => {
    if (!wishlistSet.has(name)) return addSymbol(name);
    if (locked.has(name)) {
      return errorToast("Cannot remove: symbol has an open position or order");
    }
    removeSymbol(name);
  };

  const current = folder && categories.find((c) => c.name === folder);

  return (
    <div className="flex h-full flex-col bg-theme-bg text-theme-text">
      <div className="flex items-center gap-1 border-b border-theme-border pr-2">
        <Input
          placeholder="Search symbol"
          className="!border-none"
          prefix={<Icon Icon={LuSearch} size={16} />}
          value={searchTerm}
          onChange={(e) => setSearchTerm(e?.target?.value)}
        />
        <button className="p-2" aria-label="Close Symbol Picker" onClick={onClose}>
          <Icon Icon={LuX} size={18} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {searched ? (
          searched.length ? (
            searched.map((symbol) => (
              <SymbolRow
                key={symbol.id}
                symbol={symbol}
                inWishlist={wishlistSet.has(symbol.id)}
                locked={locked.has(symbol.id)}
                onToggle={toggle}
              />
            ))
          ) : (
            <div className="p-4 text-center text-sm text-gray">No symbols found</div>
          )
        ) : current ? (
          <>
            <div className="flex items-center justify-between px-3 py-2 border-b border-theme-border">
              <button
                className="flex items-center gap-1 text-primary font-medium"
                onClick={() => setFolder(null)}
              >
                <LuChevronLeft size={16} className="text-primary" />
                {current.name}
              </button>
              <span className="text-sm text-gray tabular-nums">
                {current.selected}/{current.symbols.length}
              </span>
            </div>
            {current.symbols.map((symbol) => (
              <SymbolRow
                key={symbol.id}
                symbol={symbol}
                inWishlist={wishlistSet.has(symbol.id)}
                locked={locked.has(symbol.id)}
                onToggle={toggle}
              />
            ))}
          </>
        ) : (
          categories.map((category) => (
            <button
              key={category.name}
              className="flex w-full items-center justify-between gap-2 px-3 py-3 border-b border-theme-border last:border-b-0"
              onClick={() => setFolder(category.name)}
            >
              <span className="font-medium">{category.name}</span>
              <span className="flex items-center gap-1 text-sm text-gray tabular-nums">
                {category.selected}/{category.symbols.length}
                <LuChevronRight size={14} />
              </span>
            </button>
          ))
        )}
      </div>
    </div>
  );
};

export default SymbolPicker;
