package pagination

type Page struct {
	Limit  int
	Offset int
}

func Parse(limit, offset int) Page {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}
	return Page{limit, offset}
}
