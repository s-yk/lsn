package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	version = "v0.0.1"
)

func main() {

	var r bool
	flag.BoolVar(&r, "r", false, "recursive.")
	var d int
	flag.IntVar(&d, "d", 0, "recurse depth.")
	var tf bool
	flag.BoolVar(&tf, "tf", false, "only file.")
	var td bool
	flag.BoolVar(&td, "td", false, "only directory")
	var fi string
	flag.StringVar(&fi, "f", "", "filter.")
	var ex string
	flag.StringVar(&ex, "e", "", "exclusion.")
	var v bool
	flag.BoolVar(&v, "v", false, "print version.")
	flag.Usage = func() {
		printVersion()
		fmt.Println("Print files, directories.")
		flag.PrintDefaults()
	}
	flag.Parse()

	if v {
		printVersion()
		os.Exit(0)
	}

	root := flag.Arg(0)
	if root == "" {
		root = "."
	}

	if err := run(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	os.Exit(0)
}

func printVersion() {
	fmt.Printf("lsn %s\n", version)
}

func run(root string) error {
	fmt.Println(root)
	return nil
}
