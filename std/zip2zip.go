package zip2zip

import (
	"archive/zip"
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"regexp"

	z2 "github.com/takanoriyanagitani/go-zip2zip"
)

type Filter func(*zip.FileHeader) z2.FilterResult

var NopFilter Filter = func(_ *zip.FileHeader) z2.FilterResult {
	return z2.FilterResultKeep
}

type NameFilter func(filename string) z2.FilterResult

func (n NameFilter) ToFilter() Filter {
	return func(h *zip.FileHeader) z2.FilterResult {
		var name string = h.Name
		return n(name)
	}
}

const MaxItemSizeDefault int64 = 16777216

type ConvertConfig struct {
	Filter
	MaxItemSize int64
}

func (c ConvertConfig) WithFilter(f Filter) ConvertConfig {
	c.Filter = f
	return c
}

func (c ConvertConfig) WithMaxItemSize(i int64) ConvertConfig {
	c.MaxItemSize = i
	return c
}

func (c ConvertConfig) ZipToFiltered(
	ctx context.Context,
	izip io.ReaderAt,
	isize int64,
	ozip io.Writer,
) error {
	zrdr, e := zip.NewReader(izip, isize)
	if nil != e {
		return e
	}

	var zwtr *zip.Writer = zip.NewWriter(ozip)

	var werr error = func() error {
		var files []*zip.File = zrdr.File
		for _, zitem := range files {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var hdr zip.FileHeader = zitem.FileHeader
			var csz uint64 = hdr.CompressedSize64
			if c.MaxItemSize < int64(csz) {
				log.Printf("too big file(%v): %s\n", csz, hdr.Name)
				log.Printf("make ENV_MAX_ITEM_SIZE bigger to process it.\n")
				continue
			}

			var fres z2.FilterResult = c.Filter(&hdr)
			if z2.FilterResultKeep != fres {
				continue
			}

			cerr := zwtr.Copy(zitem)
			if nil != cerr {
				return cerr
			}
		}

		return nil
	}()

	return errors.Join(werr, zwtr.Close())
}

func (c ConvertConfig) ZipFileToStdout(
	ctx context.Context,
	zfilename string,
) error {
	zfile, e := os.Open(zfilename)
	if nil != e {
		return e
	}
	defer zfile.Close()

	stat, e := zfile.Stat()
	if nil != e {
		return e
	}
	var fsize int64 = stat.Size()

	var bw *bufio.Writer = bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	return c.ZipToFiltered(
		ctx,
		zfile,
		fsize,
		bw,
	)
}

type SimpleNameFilter struct {
	z2.Pattern
	IncludeFound bool
}

// Creates new [NameFilter].
//
//	| Include    | Found   | Keep | Include ^ Found |
//	|:----------:|:-------:|:----:|:---------------:|
//	| Include    | Found   | Keep | 0               |
//	| Include    | Missing | Skip | 1               |
//	| Exclude    | Found   | Skip | 1               |
//	| Exclude    | Missing | Keep | 0               |
func (s SimpleNameFilter) ToNameFilter() NameFilter {
	return func(filename string) z2.FilterResult {
		var found bool = s.Pattern.Regexp.MatchString(filename)
		var skip bool = found != s.IncludeFound
		var keep bool = !skip
		switch keep {
		case true:
			return z2.FilterResultKeep
		default:
			return z2.FilterResultSkip
		}
	}
}

func (s SimpleNameFilter) WithPattern(p z2.Pattern) SimpleNameFilter {
	s.Pattern = p
	return s
}

func (s SimpleNameFilter) WithPatternString(pattern string) SimpleNameFilter {
	parsed, e := regexp.Compile(pattern)
	if nil != e {
		log.Printf("ignoring invalid pattern %s: %v\n", pattern, e)
		return s
	}
	s.Pattern = z2.Pattern{Regexp: parsed}
	return s
}

func (s SimpleNameFilter) WithIncludeFound(include bool) SimpleNameFilter {
	s.IncludeFound = include
	return s
}

var SimpleNameFilterDefault SimpleNameFilter = SimpleNameFilter{
	Pattern:      z2.PatternDefault,
	IncludeFound: true,
}

var ConvertConfigDefault ConvertConfig = ConvertConfig{
	Filter:      SimpleNameFilterDefault.ToNameFilter().ToFilter(),
	MaxItemSize: MaxItemSizeDefault,
}
