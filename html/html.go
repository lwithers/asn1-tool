package html

import (
	"bufio"
	"bytes"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/lwithers/asn1-tool/util"
)

func Html(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return &util.UsageError{
			Problem: "expecting one or more input files",
		}
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

	err = HtmlAux(c, out, c.Args())
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

func HtmlAux(c *cli.Context, rawOut io.Writer, fnames []string) error {
	out := bufio.NewWriter(rawOut)

	hdrDat := &tppHeaderFooter{
		Filenames:   fnames,
		ProgName:    c.App.Name,
		ProgVersion: c.App.Version,
	}

	err := HtmlHeader(out, hdrDat)
	if err != nil {
		return err
	}

	for _, fname := range fnames {
		err = HtmlFile(out, fname)
		if err != nil {
			return err
		}
	}

	err = HtmlFooter(out, hdrDat)
	if err != nil {
		return err
	}

	return out.Flush()
}

func HtmlHeader(out *bufio.Writer, hdrDat *tppHeaderFooter) error {
	return tplHeader.Execute(out, hdrDat)
}

func HtmlFooter(out *bufio.Writer, hdrDat *tppHeaderFooter) error {
	return tplFooter.Execute(out, hdrDat)
}

func HtmlFile(out *bufio.Writer, fname string) error {
	raw, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	isPem := bytes.HasPrefix(raw, []byte("-----BEGIN "))

	err = tplFileHeading.Execute(out, tppFileHeading{
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

		tplPemBlock.Execute(out, block)
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
	var err error

	_, err = fmt.Fprintf(out, `<div class="der">
<p>DER: %d bytes</p>
`, len(raw))
	if err != nil {
		return err
	}

	err = HtmlASN1(out, raw)
	if err != nil {
		return err
	}

	_, err = out.WriteString("</div>\n")
	return err
}

func HtmlASN1(out *bufio.Writer, raw []byte) error {
	var err, parseErr error

	for len(raw) > 0 {
		var rv asn1.RawValue
		raw, parseErr = asn1.Unmarshal(raw, &rv)
		if parseErr != nil {
			badHdr := tppAsn1BadHeading{
				Error:  parseErr.Error(),
				Length: len(raw),
			}
			err = tplAsn1BadHeading.Execute(out, &badHdr)
			if err != nil {
				return err
			}
			err = HtmlHexdump(out, raw)
			if err != nil {
				return err
			}
			_, err = out.WriteString("</div>\n")
			return err
		}

		rvHdr := tppAsn1Heading{
			Class:      className[rv.Class],
			IsCompound: rv.IsCompound,
			Tag:        rv.Tag,
			HdrSize:    len(rv.FullBytes) - len(rv.Bytes),
			BodySize:   len(rv.Bytes),
		}
		err = tplAsn1Heading.Execute(out, rvHdr)
		if err != nil {
			return err
		}

		/*
			err = HtmlHexdump(out, rv.FullBytes[:rvHdr.HdrSize])
			if err != nil {
				return err
			}
		*/

		if rv.IsCompound {
			err = HtmlASN1(out, rv.Bytes)
		} else {
			err = HtmlASN1Scalar(out, &rv)
		}
		if err != nil {
			return err
		}

		_, err = out.WriteString("</div>")
		if err != nil {
			return err
		}
	}
	return nil
}

func HtmlASN1Scalar(out *bufio.Writer, rv *asn1.RawValue) error {
	switch rv.Tag {
	case asn1.TagBoolean:
		var b bool
		asn1.Unmarshal(rv.FullBytes, &b)
		fmt.Fprintf(out, "<div>BOOLEAN %v</div>", b)

	case asn1.TagInteger:
		var q *big.Int
		asn1.Unmarshal(rv.FullBytes, &q)
		fmt.Fprintf(out, "<div>INTEGER %d</div>", q)

	case asn1.TagBitString:
		var b asn1.BitString
		asn1.Unmarshal(rv.FullBytes, &b)
		HtmlHexdump(out, b.RightAlign())

	case asn1.TagOctetString:
		var b []byte
		asn1.Unmarshal(rv.FullBytes, &b)
		HtmlHexdump(out, b)

	case asn1.TagOID:
		var oid asn1.ObjectIdentifier
		asn1.Unmarshal(rv.FullBytes, &oid)
		fmt.Fprintf(out, "<div>OID %v</div>", oid)

	case asn1.TagUTF8String, asn1.TagPrintableString, asn1.TagT61String,
		asn1.TagIA5String, asn1.TagGeneralString:
		var s string
		asn1.Unmarshal(rv.FullBytes, &s)
		fmt.Fprintf(out, "<div>STRING %q</div>", s)
	}
	return nil
}

func HtmlBadAsn1(out *bufio.Writer, raw []byte, errToShow error) error {
	badHdr := tppAsn1BadHeading{
		Error:  errToShow.Error(),
		Length: len(raw),
	}
	err := tplAsn1BadHeading.Execute(out, &badHdr)
	if err != nil {
		return err
	}
	err = HtmlHexdump(out, raw)
	if err != nil {
		return err
	}
	_, err = out.WriteString("</div>\n")
	return err
}

var className = [4]string{
	"universal",
	"application",
	"context specific",
	"private",
}

func HtmlHexdump(out *bufio.Writer, raw []byte) error {
	var (
		pos int
		err error
	)
	_, err = out.WriteString("<pre class='hexdump pre-scrollable'>")
	for ; pos < len(raw)-15; pos += 16 {
		err = HtmlHexdumpLine(out, raw[pos:pos+16])
		if err != nil {
			return err
		}
	}
	if pos < len(raw) {
		err = HtmlHexdumpLine(out, raw[pos:])
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(out, "</pre>\n")
	return err
}

func HtmlHexdumpLine(out *bufio.Writer, ln []byte) error {
	var err error

	for i := 0; i < 16; i++ {
		if i == 8 {
			err = out.WriteByte(' ')
		}
		if i < len(ln) {
			_, err = fmt.Fprintf(out, "%02X ", ln[i])
		} else {
			_, err = out.WriteString("   ")
		}
		if err != nil {
			return err
		}
	}

	_, err = out.WriteString("   ")
	if err != nil {
		return err
	}

	for i := 0; i < len(ln); i++ {
		switch ln[i] {
		case 0:
			_, err = fmt.Fprint(out, "<span class='hexdumpZero'>⌁</span>")
		case '\n':
			_, err = fmt.Fprint(out, "<span class='hexdumpNewline'>↲</span>")
		case 0xFF:
			_, err = fmt.Fprint(out, "<span class='hexdumpFF'>^</span>")
		case '<':
			_, err = fmt.Fprint(out, "<span class='hexdumpPrintable>&lt;</span>")
		case '&':
			_, err = fmt.Fprint(out, "<span class='hexdumpPrintable>&amp;</span>")
		default:
			if ln[i] >= ' ' && ln[i] <= '~' {
				_, err = fmt.Fprintf(out, "<span class='hexdumpPrintable'>%c</span>", ln[i])
			} else {
				_, err = fmt.Fprint(out, "<span class='hexdumpUnprintable'>.</span>")
			}
		}
		if err != nil {
			return err
		}
	}
	return out.WriteByte('\n')
}
