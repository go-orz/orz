package orz

import "encoding/json"

func MustJsonMarshal(v any) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}

func MustJsonUnmarshal(data string, v any) {
	_ = json.Unmarshal([]byte(data), v)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
