package util

import (
	"errors"
	"strconv"
)

// StringToUint 将字符串转换为 uint，失败返回错误
func StringToUint(s string) (uint, error) {
	if s == "" {
		return 0, errors.New("id cannot be empty")
	}
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, errors.New("invalid id format")
	}
	return uint(id), nil
}

// StringToUintSafe 将字符串转换为 uint，失败返回 0（不返回错误）
func StringToUintSafe(s string) uint {
	id, _ := StringToUint(s)
	return id
}

// UintToString 将 uint 转换为字符串
func UintToString(id uint) string {
	return strconv.Itoa(int(id))
}

// StringToUintSlice 批量转换字符串数组为 uint 数组，过滤无效值
func StringToUintSlice(strs []string) []uint {
	if len(strs) == 0 {
		return nil
	}
	result := make([]uint, 0, len(strs))
	for _, s := range strs {
		if id, err := StringToUint(s); err == nil && id > 0 {
			result = append(result, id)
		}
	}
	return result
}

