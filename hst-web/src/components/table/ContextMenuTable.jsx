import { cloneElement, useState } from "react";
import { Dropdown } from "antd";
import useGlobalStore from "../../store/useGlobalStore";

const ContextMenuTable = ({
  menu,
  setSelectedRecord,
  onRowClick,
  onRowDoubleClick,
  rowStyle,
  isDraggable = false,
  isContextMenuAllowed = () => true,
  children,
}) => {
  const setGlobalStore = useGlobalStore((state) => state.setGlobalStore);
  const [visible, setVisible] = useState(false);
  const [position, setPosition] = useState({ x: 0, y: 0 });

  const handleContextMenu = (event, record) => {
    event.preventDefault();
    
    if (!isContextMenuAllowed(record)) return;
    
    setVisible(true);
    setPosition({ x: event.clientX, y: event.clientY });
    if (setSelectedRecord) setSelectedRecord(record);
  };

  const handleDragStart = (event, record) => {
    if (isDraggable) {
      event.dataTransfer.setData("id", record?.id);
      event.dataTransfer.effectAllowed = "move";
      setGlobalStore({ isDrag: true });

      // Create element with only symbol name
      const el = document.createElement("div");
      el.textContent = record?.symbol || "";
      el.className =
        "absolute -top-[99999px] text-xs font-semibold px-2 py-1 bg-white border border-gray-300 rounded shadow";

      document.body.appendChild(el);

      // Use the element as the drag image
      event.dataTransfer.setDragImage(el, 0, 0);

      // Cleanup after a short delay
      setTimeout(() => document.body.removeChild(el), 0);
    }
  };

  const wrappedChild = cloneElement(children, {
    onRow: (record) => ({
      className: `cursor-pointer${rowStyle?.(record) ? " row-tinted" : ""}`,
      ...(rowStyle && { style: rowStyle(record) }),
      onContextMenu: (event) => handleContextMenu(event, record),
      ...(onRowClick && { onClick: () => onRowClick(record) }),
      ...(onRowDoubleClick && {
        onDoubleClick: () => onRowDoubleClick(record),
      }),
      ...(isDraggable && {
        draggable: true,
        onDragStart: (e) => handleDragStart(e, record),
        onDragEnd: () => setGlobalStore({ isDrag: false }),
      }),
    }),
  });

  return (
    <div className="h-full" onClick={() => setVisible(false)}>
      <Dropdown
        menu={menu}
        open={visible}
        onOpenChange={(open) => setVisible(open)}
        trigger={["contextMenu"]}
      >
        <div
          style={{
            position: "fixed",
            top: position.y,
            left: position.x,
            zIndex: 1000,
            width: 1,
            height: 1,
          }}
        />
      </Dropdown>

      {wrappedChild}
    </div>
  );
};

export default ContextMenuTable;
