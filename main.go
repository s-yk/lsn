package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/saracen/walker"
)

const (
	version = "0.0.4"
	name    = "lsn"
)

type cli struct {
	in   *os.File
	out  *os.File
	err  *os.File
	args []string
}
type context struct {
	depth     int
	fullPath  bool
	onlyFile  bool
	onlyDir   bool
	filter    string
	exclusion string
	all       bool
	version   bool
}
type filter func(pathname string, fi os.FileInfo) (filterStatus, error)
type filterStatus int

const (
	filterIncluded filterStatus = iota
	filterExluded
	filterError
)

var errLimitDepth = errors.New("limit of depth")

func main() {
	os.Exit(run(&cli{in: os.Stdin, out: os.Stdout, err: os.Stderr, args: os.Args}))
}

func run(cli *cli) int {
	flg := flag.NewFlagSet(name, flag.ExitOnError)
	flg.SetOutput(cli.err)

	cxt := &context{}
	flg.IntVar(&cxt.depth, "d", 0, "recurse depth.")
	flg.BoolVar(&cxt.fullPath, "f", false, "print full path.")
	flg.BoolVar(&cxt.onlyFile, "of", false, "only file.")
	flg.BoolVar(&cxt.onlyDir, "od", false, "only directory")
	flg.StringVar(&cxt.filter, "fi", "", "filter.")
	flg.StringVar(&cxt.exclusion, "ex", "", "exclusion.")
	flg.BoolVar(&cxt.all, "a", false, "include hidden file, directory.")
	var v bool
	flg.BoolVar(&v, "v", false, "print version.")
	flg.Usage = func() {
		printVersion(flg.Output())
		fmt.Fprintln(flg.Output(), "Print files, directories.")
		flg.PrintDefaults()
	}
	flg.Parse(cli.args[1:])

	if v {
		printVersion(cli.out)
		return 0
	}

	root := flg.Arg(0)
	if root == "" {
		root = "."
	}

	if err := proc(root, cxt, cli); err != nil {
		fmt.Fprintln(cli.err, err)
		return 1
	}

	return 0
}

func printVersion(writer io.Writer) {
	fmt.Fprintf(writer, "%s v%s\n", name, version)
}

func proc(root string, ctx *context, cli *cli) error {
	fs := filters(ctx)
	err := walker.Walk(root, func(pathname string, fi os.FileInfo) error {
		s, err := doFilter(pathname, fi, fs)
		if err != nil {
			return err
		}
		if s == filterIncluded {
			var op string
			if ctx.fullPath && !filepath.IsAbs(pathname) {
				op, err = filepath.Abs(pathname)
				if err != nil {
					return err
				}
			} else {
				op = filepath.Clean(pathname)
			}
			fmt.Fprintf(cli.out, "%s\n", op)
		}
		return nil
	}, walker.WithErrorCallback(func(pathname string, err error) error {
		if os.IsPermission(err) {
			return nil
		}
		return err
	}))

	if err == errLimitDepth {
		return nil
	}

	return err
}

func filters(ctx *context) []filter {
	var fs []filter

	if !ctx.all {
		fs = append(fs, func(pathname string, fi os.FileInfo) (filterStatus, error) {
			if "." != fi.Name() && strings.HasPrefix(fi.Name(), ".") {
				if fi.IsDir() {
					return filterExluded, filepath.SkipDir
				}
				return filterExluded, nil
			}
			return filterIncluded, nil
		})
	}

	if ctx.depth > 0 {
		fs = append(fs, func(pathname string, fi os.FileInfo) (filterStatus, error) {
			if len(strings.Split(filepath.Clean(pathname), string(filepath.Separator))) > ctx.depth {
				return filterExluded, errLimitDepth
			}
			return filterIncluded, nil
		})
	}

	if !(ctx.onlyDir && ctx.onlyFile) {
		if ctx.onlyDir {
			fs = append(fs, func(pathname string, fi os.FileInfo) (filterStatus, error) {
				if fi.IsDir() {
					return filterIncluded, nil
				}
				return filterExluded, nil
			})
		}
		if ctx.onlyFile {
			fs = append(fs, func(pathname string, fi os.FileInfo) (filterStatus, error) {
				if fi.IsDir() {
					return filterExluded, nil
				}
				return filterIncluded, nil
			})
		}
	}

	if ctx.filter != "" {
		// case insensitive
		cis := ctx.filter == strings.ToLower(ctx.filter) || ctx.filter == strings.ToUpper(ctx.filter)
		for _, cfl := range strings.Split(ctx.filter, " ") {
			fl := cfl
			if cis {
				fl = strings.ToLower(fl)
			}

			fs = append(fs, func(pathname string, fi os.FileInfo) (filterStatus, error) {
				var m bool
				if cis {
					m = strings.Contains(strings.ToLower(pathname), fl)
				} else {
					m = strings.Contains(pathname, fl)
				}

				if m {
					return filterIncluded, nil
				}
				return filterExluded, nil
			})
		}
	}

	if ctx.exclusion != "" {
		fs = append(fs, func(pathname string, fi os.FileInfo) (filterStatus, error) {
			m := strings.Contains(pathname, ctx.exclusion)
			if !m {
				return filterIncluded, nil
			}
			return filterExluded, nil
		})
	}

	return fs
}

func doFilter(pathname string, fi os.FileInfo, filters []filter) (filterStatus, error) {
	for _, f := range filters {
		s, err := f(pathname, fi)

		if err != nil {
			return s, err
		}

		if s == filterExluded {
			return s, nil
		}
	}

	return filterIncluded, nil
}
