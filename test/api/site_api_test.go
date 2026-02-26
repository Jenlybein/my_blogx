package api_test

import (
	site_api "myblogx/api/site_api"
	"myblogx/conf"
	siteconf "myblogx/conf/site"
	"myblogx/test/testutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateSiteNoChange(t *testing.T) {
	err := site_api.UpdateSite(conf.Site{})
	if err != nil {
		t.Fatalf("空配置不应报错: %v", err)
	}
}

func TestUpdateSiteRequireWebPath(t *testing.T) {
	err := site_api.UpdateSite(conf.Site{
		Project: siteconf.Project{
			Title:   "NewTitle",
			Icon:    "/favicon.ico",
			WebPath: "",
		},
	})
	if err == nil {
		t.Fatal("缺少 web_path 应报错")
	}
}

func TestUpdateSiteWriteHTML(t *testing.T) {
	testutil.InitGlobals()
	dir := t.TempDir()
	htmlPath := filepath.Join(dir, "index.html")
	html := "<html><head><title>Old</title></head><body>Hi</body></html>"
	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		t.Fatalf("写入 html 失败: %v", err)
	}

	err := site_api.UpdateSite(conf.Site{
		Project: siteconf.Project{
			Title:   "NewTitle",
			Icon:    "/favicon.ico",
			WebPath: htmlPath,
		},
		Seo: siteconf.Seo{
			Keywords:    "k1,k2",
			Description: "desc",
		},
	})
	if err != nil {
		t.Fatalf("UpdateSite 失败: %v", err)
	}

	out, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("读取 html 失败: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "<title>NewTitle</title>") {
		t.Fatalf("title 未更新: %s", s)
	}
	if !strings.Contains(s, "favicon.ico") || !strings.Contains(s, "keywords") || !strings.Contains(s, "description") {
		t.Fatalf("head 元素未更新: %s", s)
	}
}
