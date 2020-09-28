package baserouter

type Param struct {
	Key   string
	Value string
}

type Params []Param

const maxParams = 255

func getParam() Params {

	return make(Params, 255)
}

func (p *Params) setKey(index int, key string) {
	if *p == nil {
		*p = getParam()
	}

	(*p)[index].Key = key
}

func (p *Params) setVal(index int, val string) {
	if *p == nil {
		*p = getParam()
	}

	(*p)[index].Value = val
}
