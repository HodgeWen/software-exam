interface PaginationProps {
  page: number;
  pageSize: number;
  total: number;
  onChange: (page: number) => void;
}

// 分页条：配合列表接口的 page/page_size/total 契约，只负责翻页与页码信息展示
export function Pagination({ page, pageSize, total, onChange }: PaginationProps) {
  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  const button =
    "rounded border border-gray-300 px-3 py-1.5 text-sm hover:border-indigo-400 disabled:opacity-50";

  return (
    <div className="flex items-center gap-3 text-sm">
      <button
        type="button"
        disabled={page <= 1}
        onClick={() => onChange(page - 1)}
        className={button}
      >
        上一页
      </button>
      <span className="text-gray-600">
        第 {page} / {pageCount} 页 · 共 {total} 条
      </span>
      <button
        type="button"
        disabled={page >= pageCount}
        onClick={() => onChange(page + 1)}
        className={button}
      >
        下一页
      </button>
    </div>
  );
}
