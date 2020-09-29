package baserouter

type Param struct {
	Key   string
	Value string
}

type Params []Param

const maxParams = 255

func getParam() Params {

	return make(Params, 0, 255)
}

func (p *Params) appendKey(key string) {
	if *p == nil {
		*p = getParam()
	}

	*p = append(*p, Param{Key: key})
}

func (p *Params) setVal(val string) {
	if *p == nil {
		*p = getParam()
	}

	(*p)[len(*p)-1].Value = val
}
