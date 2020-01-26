package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Everlag/slippery-policy/pob"
)

func main() {
	flag.Usage = func() {
		fmt.Println(`
pob-decode decodes a provided Path of Building code to the contained XML

This DOES NOT deserialize the code to any internal structure. Rather,
base64 decoding and zlib compression are applied and that output is shown.

Usage:
	pob-decode $POB_CODE`)
		flag.PrintDefaults()
	}
	flag.Parse()

	code := flag.Arg(0)
	if len(code) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	decoded, err := pob.XMLDecoder(bytes.NewReader([]byte(code)))
	if err != nil {
		fmt.Println("failed initializing PoB decode:\n", err)
		os.Exit(1)
	}
	defer decoded.Close()

	buf, err := ioutil.ReadAll(decoded)
	if err != nil {
		fmt.Println("failed decoding PoB:\n", err)
		os.Exit(1)
	}

	fmt.Println(string(buf))
}
