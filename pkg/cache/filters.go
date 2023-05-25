package cache

import "strings"

type TableFilter func(item *Table) bool

func NewTableNameFilter(name string) TableFilter {
	return func(item *Table) bool {
		return strings.Contains(item.name, name)
	}
}
