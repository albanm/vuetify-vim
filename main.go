// original programm here: https://gist.github.com/suyash/76ce40081f99a42c3eb1926e9986f7aa

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/suyash/algo"
)

type Snippet struct {
	Prefix string      `json:"prefix"`
	Body   interface{} `json:"body"`
}

type SnippetData struct {
	Name string
	Snip Snippet
}

func createVimSnippet(name string, snip Snippet) string {
	var b string
	sl, isSlice := snip.Body.([]interface{})
	if !isSlice {
		b = snip.Body.(string)
	} else {
		ss := make([]string, len(sl))
		for i := 0; i < len(sl); i++ {
			ss[i] = sl[i].(string)
		}
		b = strings.Join(ss, "\n")
	}

	return fmt.Sprintf(
		`snippet %v "%v" 
%v
endsnippet`,
		snip.Prefix, name, b)
}

func convert(file string) {
	ext := filepath.Ext(file)
	basefilename := string(file[:len(file)-len(ext)])

	outfilename := basefilename + ".snippets"

	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	snips := make(map[string]Snippet)

	// TODO: figure this out
	// json.Unmarshal fails with \t
	data = bytes.Replace(data, []byte("\t"), []byte("        "), -1)

	err = json.Unmarshal(data, &snips)
	if err != nil {
		panic(err)
	}

	sortedsnips := make([]SnippetData, len(snips))

	i := 0
	for name, snip := range snips {
		sortedsnips[i] = SnippetData{name, snip}
		i++
	}

	algo.Sort(len(sortedsnips), func(i, j int) {
		sortedsnips[i], sortedsnips[j] = sortedsnips[j], sortedsnips[i]
	}, func(i, j int) bool {
		return sortedsnips[i].Name < sortedsnips[j].Name
	})

	out := "priority 1\n\n"

	for _, sn := range sortedsnips {
		out += createVimSnippet(sn.Name, sn.Snip) + "\n\n"
	}

	outfile, err := os.Create(outfilename)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	out += "# vim:ft=snippets:\n"

	// TODO: remove this when previous one isn't needed
	out = strings.Replace(out, "        ", "\t", -1)

	outfile.Write([]byte(out))
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: main.go vscodesnippets.json")
		os.Exit(1)
	}

	filepath.Walk(flag.Arg(0), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			convert(path)
		}
		return err
	})
}
