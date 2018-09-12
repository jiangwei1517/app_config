package goconfig

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

const (
	ERR_SECTION_NOT_FOUND ParseError = iota + 1
	ERR_KEY_NOT_FOUND
	ERR_BLANK_SECTION_NAME
	ERR_COULD_NOT_PARSE
)

const (
	// Default section name.
	DEFAULT_SECTION = "DEFAULT"
	// Maximum allowed depth when recursively substituing variable names.
	_DEPTH_VALUES = 200
)

var (
	BREAK_LINE = "\n"
	varPattern = regexp.MustCompile(`%\(([^\)]+)\)s`)
)

func init() {
	if runtime.GOOS == "windows" {
		BREAK_LINE = "\r\n"
	}
}

type ConfigFile struct {
	lock      sync.RWMutex                 // Go map is not safe.
	FileNames []string                     // Support mutil-files.
	Data      map[string]map[string]string // Section -> key : value

	// Lists can keep sections and keys in order.
	SectionList []string            // Section name list.
	KeyList     map[string][]string // Section -> Key name list

	SectionComments map[string]string            // Sections comments.
	KeyComments     map[string]map[string]string // Keys comments.
	BlockMode       bool                         // Indicates whether use lock or not.
}

func NewConfigFile(fileNames []string) (c *ConfigFile) {
	c = &ConfigFile{}
	c.FileNames = fileNames
	c.Data = make(map[string]map[string]string)
	c.SectionList = make([]string, 0)
	c.KeyList = make(map[string][]string, 0)
	c.SectionComments = make(map[string]string)
	c.KeyComments = make(map[string]map[string]string)
	c.BlockMode = true
	return
}

func (c *ConfigFile) SetSectionComments(section, comments string) {
	if len(section) == 0 {
		section = DEFAULT_SECTION
	}
	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	if len(comments) == 0 {
		if _, ok := c.SectionComments[section]; ok {
			delete(c.SectionComments, section)
		}
	} else {
		if comments[0] != '#' && comments[0] != ';' {
			comments = ";" + comments
		}
		c.SectionComments[section] = comments
	}
}

// 初始化操作+赋值
func (c *ConfigFile) SetValue(section, key, value string) {
	if len(section) == 0 {
		section = DEFAULT_SECTION
	}
	if c.BlockMode {
		c.lock.Lock()
		c.lock.Unlock()
	}
	if len(key) == 0 {
		return
	}
	if _, ok := c.Data[section]; !ok {
		c.Data[section] = make(map[string]string)
		c.SectionList = append(c.SectionList, section)
	}
	if _, ok := c.Data[section][key]; !ok {
		c.KeyList[section] = append(c.KeyList[section], key)
	}
	c.Data[section][key] = value
}

func (c *ConfigFile) setKeyComment(section, key, comment string) {
	//keyComments     map[string]map[string]string
	if len(section) == 0 {
		section = DEFAULT_SECTION
	}
	if len(key) == 0 {
		return
	}
	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	if _, ok := c.KeyComments[section]; !ok {
		c.KeyComments[section] = make(map[string]string)
	}
	if len(comment) == 0 {
		if _, ok := c.KeyComments[section][key]; ok {
			delete(c.KeyComments[section], key)
		}
	} else {
		if comment[0] != '#' && comment[0] != ';' {
			comment = ";" + comment
		}
		c.KeyComments[section][key] = comment
	}
}

func (c *ConfigFile) GetValue(section, key string) (result string, err error) {
	if len(section) == 0 {
		section = DEFAULT_SECTION
	}
	if c.BlockMode {
		c.lock.RLock()
		defer c.lock.RUnlock()
	}
	if _, ok := c.Data[section]; ok {
		if _, ok := c.Data[section][" "]; ok {
			delete(c.Data[section], " ")
		}
		if value, ok := c.Data[section][key]; ok {
			if varPattern.MatchString(value) {
				for i := 0; i < 200; i++ {
					vr := varPattern.FindString(value)
					if len(vr) == 0 {
						break
					}
					vr = strings.TrimLeft(vr, "%(")
					vr = strings.TrimRight(vr, ")s")
					r, e := c.GetValue(DEFAULT_SECTION, vr)
					if e != nil && section != DEFAULT_SECTION {
						r, e := c.GetValue(section, vr)
						if e != nil {
							err = e
							result = ""
							return
						}
						result = r
						return
					} else {
						result = r
						return
					}
				}
			} else {
				return value, nil
			}
		} else {
			return "", fmt.Errorf("value not found exception")
		}
	} else {
		return "", fmt.Errorf("value not found exception")
	}
	return
}
