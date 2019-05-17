# peicon

Read & Write PE file icons.

```
package main

import (
	"io/ioutil"

	"github.com/faceair/peicon"
)

func main() {
	f, err := peicon.Open("logo.exe")
	if err != nil {
		panic(err)
	}

	iconData, err := f.Icon()
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("logo.ico", iconData, 0644)
	if err != nil {
		panic(err)
	}
}
```
