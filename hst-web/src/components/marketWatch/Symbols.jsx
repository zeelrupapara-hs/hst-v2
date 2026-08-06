import { useEffect, useMemo, useRef, useState } from "react";
import { Checkbox, Dropdown, Input } from "antd";
import {
  LuPanelTop,
  LuCirclePlus,
  LuChartLine,
  LuInfo,
  LuSearch,
  LuSettings2,
} from "react-icons/lu";
import { MdAdsClick } from "react-icons/md";
import useSymbolStore from "../../store/useSymbolStore";
import useGlobalStore from "../../store/useGlobalStore";
import useDebounce from "../../hooks/useDebounce";
import columns from "../../columns/symbols";
import VirtualTable from "../table/VirtualTable";
import Icon from "../common/Icon";
import ContextMenuTable from "../table/ContextMenuTable";
import MultiOrderScreen from "../order/MultiOrderScreen";
import SymbolInfoModal from "./components/SymbolInfoModal";

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
  const [selectedGroup, setSelectedGroup] = useState("all");
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedRecord, setSelectedRecord] = useState(null);
  const [isSymbolInfoOpen, setIsSymbolInfoOpen] = useState(false);

  const debouncedSearch = useDebounce(searchTerm);

  const filteredSymbols = useMemo(() => {
    let results = Object.values(symbols);

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
  }, [symbols, symbolGroups, selectedGroup, debouncedSearch]);

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
        {["all", ...Object.keys(symbolGroups)].map((key) => (
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
