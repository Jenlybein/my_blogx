package flags_test

import (
	"flag"
	"myblogx/flags"
	"myblogx/test/testutil"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	oldCommandLine := flag.CommandLine
	oldArgs := os.Args
	defer func() {
		flag.CommandLine = oldCommandLine
		os.Args = oldArgs
	}()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"cmd", "-f", "custom.yaml", "-db"}

	op := flags.Parse()
	if op.File != "custom.yaml" {
		t.Fatalf("File 解析错误: %s", op.File)
	}
	if !op.DB {
		t.Fatal("DB 标志解析错误")
	}
}

func TestRunNoOp(t *testing.T) {
	testutil.InitGlobals()
	flags.Run(&flags.FlagOptions{}, nil)
}
