package utilities

import "fmt"

func Must(err error)  {
	if err != nil {
		fmt.Println(err.Error())
	}
}
