import { useEffect, useRef, useState, useMemo } from "react";
import { Table, Dropdown, Checkbox } from "antd";

const SimpleTable = ({
  columns = [],
  data = [],
  loading = false,
  pagination = false,
  rowSelection = null,
  rowClassName = "",
  onRow = null,
}) => {
  const containerRef = useRef(null);
  const [tableHeight, setTableHeight] = useState(0);
  const [visibleColumns, setVisibleColumns] = useState([]);
  const [visible, setVisible] = useState(false);
  const [position, setPosition] = useState({ x: 0, y: 0 });

  const filteredColumns = useMemo(() => {
    return columns.map((col) => ({
      ...col,
      width: col?.width || 100,
      hidden: !visibleColumns.includes(col?.key),
    }));
  }, [columns, visibleColumns]);

  useEffect(() => {
    setVisibleColumns(columns.map((col) => col?.key));
  }, [columns]);

  useEffect(() => {
    const observer = new ResizeObserver((entries) => {
      for (let entry of entries) {
        const height = entry.contentRect.height;
        setTableHeight(height - 40); // header height is 40
      }
    });

    if (containerRef.current) observer.observe(containerRef.current);

    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const tableEl = containerRef.current?.querySelector(".ant-table-thead");
    if (!tableEl) return;

    const handleContextMenu = (e) => {
      e.preventDefault();
      setVisible(true);
      setPosition({ x: e.clientX, y: e.clientY });
    };

    tableEl.addEventListener("contextmenu", handleContextMenu);
    return () => tableEl.removeEventListener("contextmenu", handleContextMenu);
  }, [containerRef.current]);

  const handleToggleColumn = (key, checked) => {
    const updated = checked
      ? [...visibleColumns, key]
      : visibleColumns.filter((k) => k !== key);

    setVisibleColumns(updated);
  };

  const columnsMenu = {
    items: columns.map((col, index) => ({
      key: index + 1,
      label: (
        <div onClick={(e) => e.stopPropagation()}>
          <Checkbox
            checked={visibleColumns.includes(col?.key)}
            onChange={(e) => handleToggleColumn(col?.key, e.target.checked)}
          >
            {col?.title}
          </Checkbox>
        </div>
      ),
    })),
  };

  return (
    <div ref={containerRef} className="h-full overflow-hidden relative">
      <Dropdown
        menu={columnsMenu}
        open={visible}
        onOpenChange={setVisible}
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

      <Table
        columns={filteredColumns}
        dataSource={data}
        loading={loading}
        rowKey={(record) => record?.id}
        rowSelection={rowSelection}
        rowClassName={rowClassName}
        onRow={onRow}
        pagination={pagination}
        scroll={{ y: tableHeight, x: "max-content" }}
        virtual={data?.length ? true : false}
      />
    </div>
  );
};

export default SimpleTable;
