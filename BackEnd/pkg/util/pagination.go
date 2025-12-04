package util

// Pagination 分页参数结构
type Pagination struct {
	Page   int // 当前页码
	Count  int // 每页数量
	Offset int // 偏移量
}

// NormalizePagination 规范化分页参数
// 如果 page < 1，设置为 1
// 如果 count < 1，设置为 10
func NormalizePagination(page, count int) Pagination {
	if page < 1 {
		page = 1
	}
	if count < 1 {
		count = 10
	}
	return Pagination{
		Page:   page,
		Count:  count,
		Offset: (page - 1) * count,
	}
}

