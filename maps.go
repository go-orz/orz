package orz

type Map map[string]interface{}

type StringSlice []string

func (s StringSlice) Get(index int) string {
	if len(s) > index {
		return s[index]
	}
	return ""
}
