package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Sku struct {
	Name  string  `info:"name" doc:"name of the sku" json:"name"`
	Price float32 `info:"price" doc:"price of the sku" json:"price"`
}

func FindTag(arg any) {
	t := reflect.TypeOf(arg).Elem()
	for i := 0; i < t.NumField(); i++ {
		fmt.Println("info: ", t.Field(i).Tag.Get("info"))
		fmt.Println("doc: ", t.Field(i).Tag.Get("doc"))
	}
}

func main() {
	var s Sku
	FindTag(&s)

	s = Sku{
		Name:  "test",
		Price: 10.0,
	}
	jsonStr, err := json.Marshal(s)
	if err != nil {
		fmt.Println("json marshal error: ", err)
		return
	}
	fmt.Println("json: ", string(jsonStr))
	err = json.Unmarshal(jsonStr, &s)
	if err != nil {
		fmt.Println("json unmarshal error: ", err)
		return
	}
	fmt.Println("json: ", s)
}
