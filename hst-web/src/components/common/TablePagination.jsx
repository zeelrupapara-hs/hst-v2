import React, { useState } from "react";
import { Pagination } from "antd";

const TablePagination = ({ total, onChange }) => {
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const startIndex = (currentPage - 1) * pageSize + 1;
  const endIndex = Math.min(currentPage * pageSize, total);

  if (!total) return null;

  return (
    <div className="flex items-center justify-between gap-4 my-4">
      <p className="text-gray max-md:text-sm">
        {`Showing ${startIndex}-${endIndex} of ${total} items`}
      </p>

      <Pagination
        total={total}
        pageSize={pageSize}
        current={currentPage}
        showSizeChanger
        onChange={(current, pageSize) => {
          setCurrentPage(current);
          setPageSize(pageSize);
          onChange({ current, pageSize });
        }}
      />
    </div>
  );
};

export default TablePagination;
