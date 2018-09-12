package main

import (
	"app_config/goconfig"
	"fmt"
)

func main() {
	fileName := "./conf.ini"
	c, err := goconfig.LoadConfigFiles(fileName)
	if err != nil {
		fmt.Println(err)
	}
	c.SetValue("jiangwei", `"qwe=ueu"`, "123")
	err = goconfig.SaveConfigFile(c, "./conf2.ini")
	if err != nil {
		fmt.Println("goconfig.SaveConfigFile failed", err)
		return
	}
	fmt.Println(c.Data)
}
