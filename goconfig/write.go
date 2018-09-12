package goconfig

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func SaveConfigFile(c *ConfigFile, fileName string) (err error) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("SaveConfigFile failed", err)
		return
	}
	err = saveConfigData(c, file)
	if err != nil {
		fmt.Println("saveConfigData failed", err)
		return
	}
	defer file.Close()
	return
}

func saveConfigData(c *ConfigFile, file *os.File) (err error) {
	buf := bytes.NewBuffer(nil)
	for _, section := range c.SectionList {
		if comment, ok := c.SectionComments[section]; ok {
			_, err := buf.WriteString(comment + BREAK_LINE)
			if err != nil {
				return err
			}
		}

		if section != DEFAULT_SECTION {
			_, err := buf.WriteString("[" + section + "]" + BREAK_LINE)
			if err != nil {
				return err
			}
		}

		for _, key := range c.KeyList[section] {
			if key != " " {
				keyName := key
				if keyName[0] == '#' {
					keyName = "-"
				}
				if strings.Contains(keyName, `=`) || strings.Contains(keyName, `:`) {
					if strings.Contains(keyName, "`") {
						if strings.Contains(keyName, `"`) {
							keyName = `"""` + keyName + `"""`
						} else {
							keyName = `"` + keyName + `"`
						}
					} else {
						keyName = "`" + keyName + "`"
					}
				}
				if _, ok := c.KeyComments[section][key]; ok {
					_, err = buf.WriteString(c.KeyComments[section][key] + BREAK_LINE)
					if err != nil {
						return
					}
				}
				if _, ok := c.Data[section][key]; ok {
					_, err = buf.WriteString(keyName + " = " + c.Data[section][key] + BREAK_LINE)
				}
			}
		}

	}
	_, err = buf.WriteTo(file)
	if err != nil {
		fmt.Println(" buf.WriteTo(file) failed", err)
		return
	}
	return
}
