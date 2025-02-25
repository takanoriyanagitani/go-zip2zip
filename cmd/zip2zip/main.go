package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	zs "github.com/takanoriyanagitani/go-zip2zip/std"
	. "github.com/takanoriyanagitani/go-zip2zip/util"
)

var envValByKey func(string) IO[string] = Lift(
	func(key string) (string, error) {
		val, found := os.LookupEnv(key)
		switch found {
		case true:
			return val, nil
		default:
			return "", fmt.Errorf("env var %s missing", key)
		}
	},
)

var maxItemSizeInt IO[int] = Bind(
	envValByKey("ENV_MAX_ITEM_SIZE"),
	Lift(strconv.Atoi),
).Or(Of(int(zs.MaxItemSizeDefault)))

var includeFound IO[bool] = Bind(
	envValByKey("ENV_INCLUDE_FOUND"),
	Lift(strconv.ParseBool),
).Or(Of(true))

var pattern IO[string] = envValByKey("ENV_NAME_PATTERN").
	Or(Of("."))

var simpleFilter IO[zs.SimpleNameFilter] = Bind(
	All(
		pattern.ToAny(),
		includeFound.ToAny(),
	),
	Lift(func(a []any) (zs.SimpleNameFilter, error) {
		return zs.SimpleNameFilterDefault.
			WithPatternString(a[0].(string)).
			WithIncludeFound(a[1].(bool)), nil
	}),
)

var nameFilter IO[zs.NameFilter] = Bind(
	simpleFilter,
	Lift(func(s zs.SimpleNameFilter) (zs.NameFilter, error) {
		return s.ToNameFilter(), nil
	}),
)

var filter IO[zs.Filter] = Bind(
	nameFilter,
	Lift(func(n zs.NameFilter) (zs.Filter, error) {
		return n.ToFilter(), nil
	}),
)

var convCfg IO[zs.ConvertConfig] = Bind(
	All(
		filter.ToAny(),
		maxItemSizeInt.ToAny(),
	),
	Lift(func(a []any) (zs.ConvertConfig, error) {
		return zs.ConvertConfigDefault.
			WithFilter(a[0].(zs.Filter)).
			WithMaxItemSize(int64(a[1].(int))), nil
	}),
)

var zfilename IO[string] = envValByKey("ENV_INPUT_ZIP_FILENAME")

func zfilename2filtered2zip2stdout(zfilename string) IO[Void] {
	return Bind(
		convCfg,
		func(cfg zs.ConvertConfig) IO[Void] {
			return func(ctx context.Context) (Void, error) {
				return Empty, cfg.ZipFileToStdout(
					ctx,
					zfilename,
				)
			}
		},
	)
}

var zipfile2filtered2stdout IO[Void] = Bind(
	zfilename,
	zfilename2filtered2zip2stdout,
)

var sub IO[Void] = func(ctx context.Context) (Void, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return zipfile2filtered2stdout(ctx)
}

func main() {
	_, e := sub(context.Background())
	if nil != e {
		log.Printf("%v\n", e)
	}
}
