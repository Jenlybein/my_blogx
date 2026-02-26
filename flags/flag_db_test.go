package flags_test

import (
	"myblogx/flags"
	"myblogx/models"
	"myblogx/test/testutil"
	"testing"
)

func TestFlagDB(t *testing.T) {
	db := testutil.SetupSQLite(t)
	flags.FlagDB(db)

	if !db.Migrator().HasTable(&models.UserModel{}) {
		t.Fatal("UserModel 表未迁移")
	}
	if !db.Migrator().HasTable(&models.ArticleModel{}) {
		t.Fatal("ArticleModel 表未迁移")
	}
	if !db.Migrator().HasTable(&models.LogModel{}) {
		t.Fatal("LogModel 表未迁移")
	}
}
