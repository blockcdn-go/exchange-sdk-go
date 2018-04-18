[![Build Status](https://travis-ci.org/gotoxu/query.svg?branch=master)](https://travis-ci.org/gotoxu/query)

# query

query is Go library for encoding structs into URL query string.



## Installation

To install query, use `go get`

```go
go get -u github.com/gotoxu/query
```



## Example usages

### Encoder

```go
package main

import (
	"fmt"
	"net/url"

	"github.com/gotoxu/query"
)

type example struct {
	Name     string  `url:"name"`
	Password string  `url:"password"`
	Age      int     `url:"age"`
	Salary   float64 `url:"salary"`
	ID       int64   `url:"id"`
}

func main() {
	e := &example{
		Name:     "XuQiaolun",
		Password: "123456",
		Age:      33,
		Salary:   20000.00,
		ID:       1523798459,
	}

	var err error
	var m url.Values
	m, err = query.NewEncoder().Encode(e)
	if err != nil {
		panic(err)
	}

	fmt.Println(m.Encode())
}
```

默认的Struct tag name是`url`，Result：

`age=33&id=1523798459&name=XuQiaolun&password=123456&salary=20000.000000`



你也可以使用函数`SetAliasTag`来自定义tag，如：

```go
encoder := query.NewEncoder().SetAliasTag("schema")
```



### Decoder

```go
package main

import (
	"net/url"

	"github.com/gotoxu/query"
)

type example struct {
	Name     string  `url:"name"`
	Password string  `url:"password"`
	Age      int     `url:"age"`
	Salary   float64 `url:"salary"`
	ID       int64   `url:"id"`
}

func main() {
	m, err := url.ParseQuery("age=33&id=1523798459&name=XuQiaolun&password=123456&salary=20000.000000")
	if err != nil {
		panic(err)
	}

	decoder := query.NewDecoder()
	var e example
	err = decoder.Decode(m, &e)
	if err != nil {
		panic(err)
	}
}
```