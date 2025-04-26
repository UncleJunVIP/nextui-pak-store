package models

type MenuItems struct {
	Items []string
}

func (m MenuItems) Values() []string {
	return m.Items
}
