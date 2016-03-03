package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codegangsta/cli"
)

func Html(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return &UsageError{Problem: "expecting one or more input files"}
	}

	var (
		out      io.Writer = os.Stdout
		f        *os.File
		fname    string
		fnameNew string
		err      error
	)
	fname = c.String("out")
	if fname != "" {
		if fname[len(fname)-1] == '/' {
			return fmt.Errorf("%q is a directory", fname)
		}
		idx := strings.LastIndexByte(fname, '/') + 1
		fnameNew = fname[:idx] + "." + fname[idx:] + ".new"
		f, err = os.OpenFile(fnameNew,
			os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			return err
		}
		out = f
	}

	err = HtmlAux(out, c.Args())
	if err != nil {
		if fname != "" {
			f.Close()
			os.Remove(fnameNew)
		}
		return err
	}

	if fname != "" {
		err = f.Close()
		if err != nil {
			return err
		}
		err = os.Rename(fnameNew, fname)
		if err != nil {
			return err
		}
	}

	return nil
}

func HtmlAux(rawOut io.Writer, fnames []string) error {
	out := bufio.NewWriter(rawOut)

	err := HtmlHeader(out)
	if err != nil {
		return err
	}

	for _, fname := range fnames {
		err = HtmlFile(out, fname)
		if err != nil {
			return err
		}
	}

	err := HtmlFooter(out)
	if err != nil {
		return err
	}

	return out.Flush()
}

func HtmlHeader(out bufio.Writer) error {
	return nil
}

func HtmlFooter(out bufio.Writer) error {
	return nil
}

func HtmlFile(out bufio.Writer, fname string) error {
	return nil
}
