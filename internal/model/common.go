package model

type Sort struct {
	FieldName    string
	IsDescending bool
}

type Page struct {
	ItemPerPage int64
	Number      int64
}

var usersFieldToSort map[string]struct{} = map[string]struct{}{
	"id":       {},
	"name":     {},
	"email":    {},
	"is_admin": {},
}

func IsAvailableToSortUser(fieldName string) bool {
	_, availableToSort := usersFieldToSort[fieldName]

	return availableToSort
}
