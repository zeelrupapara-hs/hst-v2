import { useState } from "react";
import { Dropdown, Input } from "antd";
import { LuCheck, LuChevronDown, LuPencil, LuPlus, LuTrash2, LuX } from "react-icons/lu";
import useWatchlistStore from "../../../store/useWatchlistStore";
import ModalComponent from "../../modal/ModalComponent";
import ConfirmationModal from "../../modal/ConfirmationModal";
import Icon from "../../common/Icon";

// The name dialog serves create and rename; mode is told apart by whether a list is passed.
const NameModal = ({ isOpen, setIsOpen, list, onSubmit }) => {
  const [name, setName] = useState(list?.name ?? "");

  const submit = async () => {
    const trimmed = name.trim();
    if (!trimmed) return;
    if (await onSubmit(trimmed)) setIsOpen(false);
  };

  return (
    <ModalComponent isOpen={isOpen} width={360}>
      <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
        <span>{list ? "Rename Watchlist" : "New Watchlist"}</span>
        <button onClick={() => setIsOpen(false)}>
          <Icon Icon={LuX} size={18} className="!text-white" />
        </button>
      </div>

      <div className="p-4">
        <Input
          placeholder="Watchlist name"
          defaultValue={list?.name}
          maxLength={64}
          autoFocus
          onChange={(e) => setName(e?.target?.value)}
          onPressEnter={submit}
        />

        <div className="flex justify-end items-center gap-3 mt-5">
          <button className="btn-outline" onClick={() => setIsOpen(false)}>
            Cancel
          </button>
          <button className="btn-primary" onClick={submit}>
            Save
          </button>
        </div>
      </div>
    </ModalComponent>
  );
};

const WatchlistSwitcher = () => {
  const watchlists = useWatchlistStore((state) => state.watchlists);
  const activeId = useWatchlistStore((state) => state.activeId);
  const viewAll = useWatchlistStore((state) => state.viewAll);
  const setActive = useWatchlistStore((state) => state.setActive);
  const createList = useWatchlistStore((state) => state.createList);
  const renameList = useWatchlistStore((state) => state.renameList);
  const deleteList = useWatchlistStore((state) => state.deleteList);

  const [modal, setModal] = useState(null); // "create" | "rename" | "delete"

  const active = watchlists.find((w) => w.id === activeId);
  const showingAll = viewAll || !watchlists.length;
  const label = showingAll ? "All Symbols" : active?.name;

  const check = <Icon Icon={LuCheck} size={16} isActive />;

  const menu = {
    items: [
      {
        key: "all",
        label: "All Symbols",
        icon: showingAll ? check : <span className="w-4" />,
        onClick: () => setActive("all"),
      },
      ...watchlists.map((w) => ({
        key: w.id,
        label: w.name,
        icon: !showingAll && w.id === activeId ? check : <span className="w-4" />,
        onClick: () => setActive(w.id),
      })),
      { type: "divider" },
      {
        key: "create",
        label: "New Watchlist",
        icon: <Icon Icon={LuPlus} size={16} />,
        onClick: () => setModal("create"),
      },
      ...(active && !showingAll
        ? [
            {
              key: "rename",
              label: `Rename "${active.name}"`,
              icon: <Icon Icon={LuPencil} size={16} />,
              onClick: () => setModal("rename"),
            },
            {
              key: "delete",
              label: `Delete "${active.name}"`,
              icon: <Icon Icon={LuTrash2} size={16} />,
              onClick: () => setModal("delete"),
            },
          ]
        : []),
    ],
  };

  return (
    <>
      <Dropdown menu={menu} trigger={["click"]} placement="bottomLeft">
        <button
          className="flex items-center gap-1 text-xs font-medium px-[10px] py-[2px] border border-primary text-primary rounded-full whitespace-nowrap"
          aria-label="Switch Watchlist"
        >
          {label}
          <LuChevronDown size={12} />
        </button>
      </Dropdown>

      {(modal === "create" || modal === "rename") && (
        <NameModal
          isOpen
          setIsOpen={() => setModal(null)}
          list={modal === "rename" ? active : null}
          onSubmit={(name) =>
            modal === "rename" ? renameList(active.id, name) : createList(name)
          }
        />
      )}

      <ConfirmationModal
        isOpen={modal === "delete"}
        setIsOpen={() => setModal(null)}
        title="Delete Watchlist"
        message={`Delete "${active?.name}" and its symbols?`}
        onConfirm={() => deleteList(active.id)}
      />
    </>
  );
};

export default WatchlistSwitcher;
