package service

func ParsePagination(limit, offset int) (int, int, error) {
	if limit == 0 {
		limit = 20
	}

	if limit < 1 || limit > 100 {
		return 0, 0, BadRequest("Limit must be between 1 and 100")
	}

	if offset < 0 {
		return 0, 0, BadRequest("Offset must be non-negative")
	}

	return limit, offset, nil
}
