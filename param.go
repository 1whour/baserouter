package baserouter

type Param struct {
	Key   string
	Value string
}

type Params []Param

const maxParams = 255

func getParam(p Params) Params {
	if p != nil {
		return p
	}

	return make(Params, 255)
}
