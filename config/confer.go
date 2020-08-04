package confer

import "fmt"

func ReadConf(provider string, param ConferParam, resultVal interface{}) (Confer, error) {

	var c Confer
	switch provider {
	case DEF_HCONF_PROVIDER_LOCAL:
		c = newLocalConfer(provider, param)
	case DEF_HCONF_PROVIDER_CONFD:
		c = newConfdConfer(provider, param)
	default:
		return nil, fmt.Errorf("confer unsupport provider type:%v", provider)
	}

	err := c.Verify()
	if err != nil {
		return nil, err
	}

	err = c.Read(resultVal)
	if err != nil {
		return nil, err
	}

	return c, nil
}
