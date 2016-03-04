package main

import (
	"bufio"
	"bytes"
	"encoding/pem"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
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

	err := HtmlHeader(out, fnames)
	if err != nil {
		return err
	}

	for _, fname := range fnames {
		err = HtmlFile(out, fname)
		if err != nil {
			return err
		}
	}

	err = HtmlFooter(out)
	if err != nil {
		return err
	}

	return out.Flush()
}

func HtmlHeader(out *bufio.Writer, fnames []string) error {
	t := template.Must(template.New("header").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>asn1-tool | {{.}}</title>
</head>
<body>
`))
	return t.Execute(out, fnames)
}

func HtmlFooter(out *bufio.Writer) error {
	_, err := out.WriteString(`</body>
</html>
`)
	return err
}

func HtmlFile(out *bufio.Writer, fname string) error {
	raw, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	isPem := bytes.HasPrefix(raw, []byte("-----BEGIN "))

	t := template.Must(template.New("file").Parse(`<h1>{{.Fname}}</h1>
<p>{{.Size}} byte {{if .IsPem}}PEM{{else}}DER{{end}} file</p>
`))
	err = t.Execute(out, struct {
		Fname string
		Size  int
		IsPem bool
	}{
		Fname: fname,
		Size:  len(raw),
		IsPem: isPem,
	})
	if err != nil {
		return err
	}

	if isPem {
		return HtmlPEMFile(out, raw)
	}

	return HtmlDERFile(out, raw)
}

func HtmlPEMFile(out *bufio.Writer, raw []byte) error {
	for len(raw) > 0 {
		block, rest := pem.Decode(raw)
		if block == nil {
			break
		}
		raw = rest

		t := template.Must(template.New("pem").Parse(`<div class="pem">
<p>PEM, type: <code>{{.Type}}</code></p>
{{if .Headers}}<table>
<thead>
<tr><th>Header</th><th>Value</th></tr>
</thead>
<tbody>
{{range $hdr, $val := .Headers}}<tr><td>{{$hdr}}</td><td>{{$val}}</td></tr>
{{end}}</tbody>
</table>
{{end}}`))
		t.Execute(out, block)
		err := HtmlDERFile(out, block.Bytes)
		if err != nil {
			return err
		}
		_, err = out.WriteString("</div>\n")
		if err != nil {
			return err
		}
	}

	if len(raw) > 0 {
		_, err := out.WriteString("<h3>Trailing data</h3>\n")
		if err != nil {
			return err
		}
		return HtmlHexdump(out, raw)
	}

	return nil
}

func HtmlDERFile(out *bufio.Writer, raw []byte) error {
	fmt.Fprintf(out, `<div class="der">
<p>DER: %d bytes</p>
</div>
`, len(raw))
	return nil
}

func HtmlHexdump(out *bufio.Writer, raw []byte) error {
	_, err := fmt.Fprintf(out, "<p>Hexdump, %d bytes:</p>\n", len(raw))
	return err
}
