package baserouter

import "errors"

type method struct {
	method [7]*datrie
}

var ErrMethod = errors.New("error method")

func methodIndex(method string) (int, error) {
	if len(method) == 0 {
		return 0, ErrMethod
	}

	switch method[0] {
	case 'G':
		return 0, nil
	case 'P':
		if len(method) <= 1 {
			return 0, ErrMethod
		}

		switch method[1] {
		case 'O':
			return 1, nil
		case 'U':
			return 2, nil
		case 'A':
			return 3, nil
		default:
			return 0, ErrMethod
		}
	case 'D':
		return 4, nil
	case 'H':
		return 5, nil
	default:
		return 0, ErrMethod
	}

	return 0, ErrMethod
}

func (m *method) save(method, path string, handle handle) {
}
