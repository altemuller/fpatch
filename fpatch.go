package main

import (
	"fmt"
	"os"
	"gopkg.in/ini.v1"
	"strconv"
	"io/ioutil"
	"bytes"
)

func main() {
	const MAXSTRINGLENGTH = 0x64
	if len(os.Args) < 2 {
		fmt.Print("fpatch v1.2 Dmitry Mikhaltsov\nUsage:\nfpatch <config.ini>")
		return
	}

	config, err := ini.Load(os.Args[1])

	if err != nil {
		fmt.Print(err)
		return
	}

	filename := config.Section("DEFAULT").Key("filename").String()

	if len(filename) == 0 {
		fmt.Print("Error: filename is empty.")
		return
	}

	file, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Print(err)
		return
	}

	sections := config.Sections()[1:]
	sectionsString := config.SectionStrings()[1:]

	for i, section := range sections {
		source := section.Key("src").String()
		addressString := section.Key("address").String()
		line := section.Key("line").MustInt()
		address, err := strconv.ParseInt(addressString, 0, 32)

		if err != nil {
			fmt.Printf("Error: unknown address: %v", addressString)
			return
		}

		text, err := ioutil.ReadFile(source)

		if err != nil {
			fmt.Printf("Error: file not found: %v", source)
		}

		space := []byte{'\n'}
		text = bytes.Split(text, space)[line]

		fmt.Printf("Patch section: %v\nSource: %v\nLine: %v [%v]\nAddress: %v\nLength: %v\n\n", sectionsString[i], source, line, string(text), addressString, len(text))

		buff1 := file[:address]
		start := file[address:]
		lengthStr := bytes.Index(start, []byte{0x0})
		endPoint := address + int64(lengthStr)
		buff2 := file[endPoint:]

		var temp [][]byte
		var final []byte

		if len(text) <= lengthStr {
			temp = [][]byte{buff1, text, bytes.Repeat([]byte{0x20}, lengthStr - len(text)), buff2}
			final = bytes.Join(temp, []byte(""))
		} else {
			fmt.Printf("Error: [%v] newString (%v) > originalString (%v)", addressString, len(text), lengthStr)
			return
		}

		ioutil.WriteFile(filename, final, os.FileMode(0077))

		fmt.Printf("Section [%v] <%v> successful patched.", sectionsString[i], addressString)
	}	
}