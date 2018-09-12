package goconfig

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

func (c *ConfigFile) read(r io.Reader) (err error) {
	reader := bufio.NewReader(r)
	headerytes, err := reader.Peek(3)
	if err != nil {
		return &ReadError{ERR_COULD_NOT_PARSE, "read io failed"}
	}
	// Bom头
	if len(headerytes) >= 3 && headerytes[0] == 0xef && headerytes[1] == 0xbb && headerytes[2] == 0xbf {
		_, err = reader.Read(headerytes)
		if err != nil {
			return &ReadError{ERR_COULD_NOT_PARSE, "reader.Read(headerytes) failed"}
		}
	}
	var section string = DEFAULT_SECTION
	var count int
	var comments string
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return &ReadError{ERR_COULD_NOT_PARSE, "read line failed"}
		}
		if err == io.EOF {
			if len(line) == 0 {
				break
			}
		}
		line = strings.TrimSpace(line)
		var length = len(line)
		switch {
		// empty line
		case line == "":
			continue
		//  注释
		case line[0] == '#' || line[0] == ';':
			if len(comments) == 0 {
				comments = line
			} else {
				comments = comments + BREAK_LINE + line
			}
			continue
		// section
		case line[0] == '[' && line[length-1] == ']':
			section = line[1 : length-1]
			if len(comments) > 0 {
				c.SetSectionComments(section, comments)
				comments = ""
			}
			// 初始化data
			c.SetValue(section, " ", " ")
			count = 1
		case section == "":
			return &ReadError{ERR_COULD_NOT_PARSE, "section is null"}
		default:
			var (
				i          int
				keyQuote   string
				key        string
				valueQuote string
				value      string
			)
			line = strings.TrimSpace(line)
			if line[0] == '"' {
				if length >= 6 && line[0:3] == `"""` {
					keyQuote = `"""`
				} else {
					keyQuote = `"`
				}
			} else if line[0] == '`' {
				keyQuote = "`"
			} else if line[0] == '\'' {
				keyQuote = "'"
			}
			if keyQuote != "" {
				qLen := len(keyQuote)
				pos := strings.Index(line[qLen:], keyQuote)
				if pos == -1 {
					return &ReadError{ERR_COULD_NOT_PARSE, "keyQuote found error"}
				}
				pos = pos + qLen
				i = strings.IndexAny(line[pos:], "=:")
				if i < 0 {
					return &ReadError{ERR_COULD_NOT_PARSE, "=: 1found error"}
				}
				i = i + pos
				key = line[qLen:pos]

			} else {
				i = strings.IndexAny(line, "=:")

				if i < 0 {
					return &ReadError{ERR_COULD_NOT_PARSE, "=: 2found error"}
				}
				key = line[0:i]
				key = strings.TrimSpace(key)

			}

			if key == "-" {
				key = "#" + fmt.Sprint(count)
				count++
			}

			lineRight := strings.TrimSpace(line[i+1:])
			lineRightLength := len(lineRight)
			firstchar := ""
			if lineRightLength >= 2 {
				firstchar = lineRight[0:1]
			}
			if firstchar == "'" {
				valueQuote = "'"
			} else if lineRightLength >= 6 && lineRight[0:3] == `"""` {
				valueQuote = `"""`
			} else if firstchar == `"` {
				valueQuote = `"`
			}

			if valueQuote != "" {
				rlen := len(valueQuote)
				if rlen == -1 {
					return &ReadError{ERR_COULD_NOT_PARSE, "valueQuote found error"}
				}
				pos := strings.LastIndex(lineRight[rlen:], valueQuote)
				if pos == -1 {
					return &ReadError{ERR_COULD_NOT_PARSE, "valueQuote found error"}
				}
				pos = pos + rlen
				value = lineRight[rlen:pos]

			} else {
				value = strings.TrimSpace(lineRight[0:])

			}
			c.SetValue(section, key, value)
			if len(comments) > 0 {
				c.setKeyComment(section, key, comments)
				comments = ""
			}
		}
		if err == io.EOF {
			println("break")
			break
		}
	}
	return
}

func LoadConfigFiles(fileName string, fileNames ...string) (c *ConfigFile, err error) {
	fileArr := make([]string, 1, len(fileNames)+1)
	fileArr[0] = fileName
	if len(fileNames) > 0 {
		fileArr = append(fileArr, fileName)
	}
	c = NewConfigFile(fileArr)
	for _, file := range fileArr {
		if err = c.loadFile(file); err != nil {
			fmt.Println(&ReadError{
				Reason:  ERR_COULD_NOT_PARSE,
				Content: fileName,
			})
		}
	}
	return
}

func (c *ConfigFile) loadFile(fileName string) (err error) {
	if len(fileName) != 0 {
		f, err := os.Open(fileName)
		if err != nil {
			return &ReadError{ERR_COULD_NOT_PARSE, err.Error()}
		}
		defer f.Close()
		err = c.read(f)
		if err != nil {
			return &ReadError{ERR_COULD_NOT_PARSE, err.Error()}
		}
	}
	return
}

func LoadFromReader(in io.Reader) (c *ConfigFile, err error) {
	c = NewConfigFile([]string{""})
	err = c.read(in)
	return
}

func LoadFromData(date []byte) (c *ConfigFile, err error) {
	var tmpFile = path.Join(os.TempDir(), "appconfig", fmt.Sprintf("%d", time.Now().Nanosecond()))
	err = os.MkdirAll(tmpFile, os.ModePerm)
	if err != nil {
		fmt.Println("os.MkdirAll failed", err)
		return
	}
	f, err := os.Create(tmpFile)
	if err != nil {
		fmt.Println("os.Create failed", err)
		return
	}
	f.Write(date)
	defer f.Close()
	c = NewConfigFile([]string{tmpFile})
	c.read(f)
	return
}

func (c *ConfigFile) Reload() (err error) {
	if len(c.FileNames) == 0 && c.FileNames[0] == "" {
		return fmt.Errorf("file opened from in-memory data, use ReloadData to reload")
	}
	var cfg *ConfigFile
	if len(c.FileNames) == 1 {
		cfg, err = LoadConfigFiles(c.FileNames[0])
	} else {
		cfg, err = LoadConfigFiles(c.FileNames[0], c.FileNames[1:]...)
	}
	if err != nil {
		return
	}
	c = cfg
	return
}

func (c *ConfigFile) ReloadData(in io.Reader) (err error) {
	if c.FileNames[0] != "" {
		return fmt.Errorf("file opened from multi file, use Reload to reload")
	}
	var cfg *ConfigFile
	cfg, err = LoadFromReader(in)
	if err != nil {
		return
	}
	c = cfg
	return
}

func (c *ConfigFile) Append(fileName string) (err error) {
	if len(c.FileNames) == 1 && c.FileNames[0] == "" {
		return fmt.Errorf("file opened from multi file, use Reload to reload")
	}
	c.FileNames = append(c.FileNames, fileName)
	return c.Reload()
}

type ParseError int

type ReadError struct {
	Reason  ParseError
	Content string
}

func (myErr *ReadError) Error() string {
	switch myErr.Reason {
	case ERR_BLANK_SECTION_NAME:
		return "empty section name not allowed"
	case ERR_COULD_NOT_PARSE:
		return fmt.Sprintf("could not parse line: %s", string(myErr.Content))
	}
	return "invalid read error"
}
