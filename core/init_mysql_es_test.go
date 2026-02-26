package core_test

import (
	"myblogx/conf"
	"myblogx/core"
	"myblogx/global"
	"myblogx/test/testutil"
	"testing"
)

func TestInitMySQLESDisabled(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		River: conf.River{
			Enabled: false,
		},
	}

	core.InitMySQLES()
}
