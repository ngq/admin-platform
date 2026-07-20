package datainfra

import "gorm.io/gorm"

// Paginate applies offset/limit pagination with order and counts total rows.
// db should have all conditions (Where, Joins, etc.) applied before passing in.
// order is the ORDER BY clause (e.g. "id DESC"); pass "" for no ordering.
func Paginate(db *gorm.DB, page, pageSize int, order string, out any, count *int64) error {
	if err := db.Session(&gorm.Session{}).Count(count).Error; err != nil {
		return err
	}

	offset := (page - 1) * pageSize
	q := db.Session(&gorm.Session{}).Offset(offset).Limit(pageSize)
	if order != "" {
		q = q.Order(order)
	}
	return q.Find(out).Error
}
