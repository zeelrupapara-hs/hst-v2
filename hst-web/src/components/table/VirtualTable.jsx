import { useEffect, useRef, useState } from "react";
import { Table } from "antd";

const VirtualTable = ({ columns, data, loading, onRow }) => {
  const containerRef = useRef(null);
  const [tableHeight, setTableHeight] = useState(400);

  useEffect(() => {
    const resize = () => {
      if (containerRef.current) {
        const height = containerRef.current.clientHeight;
        setTableHeight(height - 40); // header height is 40
      }
    };

    resize();
    window.addEventListener("resize", resize);
    return () => window.removeEventListener("resize", resize);
  }, []);

  return (
    <div ref={containerRef} className="h-full overflow-hidden">
      <Table
        columns={columns}
        dataSource={data}
        loading={loading}
        rowKey={(record) => record?.id}
        onRow={onRow}
        pagination={false}
        scroll={{ y: tableHeight }}
        virtual
      />
    </div>
  );
};

export default VirtualTable;
