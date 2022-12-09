package values

import "encoding/json"

type Attach string

var attachNoneBytes = []byte("\"\"")

func (a *Attach) UnmarshalJSON(b2 []byte) error {
	*a = Attach(b2)
	return nil
}

func (a *Attach) MarshalJSON() ([]byte, error) {
	if a == nil || len(*a) == 0 {
		return attachNoneBytes, nil
	}
	r := []byte(*a)
	return r, nil
}

// Marshal 将一个对象放入Attach TODO len(*a) == 0
func (a *Attach) Marshal(v interface{}) error {
	d, err := json.Marshal(v)
	if err == nil {
		*a = Attach(d)
	}
	return err
}

// Unmarshal 使用i解析Attach
func (a *Attach) Unmarshal(i interface{}) error {
	if len(*a) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(*a), i)
}
