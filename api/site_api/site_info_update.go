package site_api

import (
	"errors"
	"fmt"
	"myblogx/common/res"
	"myblogx/conf"
	"myblogx/core"
	"myblogx/global"
	"myblogx/middleware"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

// 数据结构映射表
var confMap = map[string]any{
	"site":  &conf.Site{},
	"email": &conf.Email{},
	"qq":    &conf.QQ{},
	"qiniu": &conf.QiNiu{},
	"ai":    &conf.AI{},
}

// 更新站点信息
func (s SiteApi) SiteUpdateView(c *gin.Context) {
	// 匹配 Uri
	cr := middleware.GetBindUri[SiteInfoRequest](c)

	// 获取请求体数据
	targetStruct, ok := confMap[cr.Name]
	if !ok {
		res.FailWithMsg("站点信息不存在", c)
		return
	}
	err := c.ShouldBindJSON(targetStruct)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	rep := targetStruct

	switch s := rep.(type) {
	case *conf.Site:
		err = UpdateSite(*s)
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		global.Config.Site = *s

	case *conf.Email:
		if s.AuthCode == sensitive_place_holder {
			s.AuthCode = global.Config.Email.AuthCode
		}
		global.Config.Email = *s

	case *conf.QQ:
		if s.AppKey == sensitive_place_holder {
			s.AppKey = global.Config.QQ.AppKey
		}
		global.Config.QQ = *s

	case *conf.QiNiu:
		if s.SecretKey == sensitive_place_holder {
			s.SecretKey = global.Config.QiNiu.SecretKey
		}
		global.Config.QiNiu = *s

	case *conf.AI:
		if s.SecretKey == sensitive_place_holder {
			s.SecretKey = global.Config.AI.SecretKey
		}
		global.Config.AI = *s
	}

	core.SetCfg(global.Config, &global.Flags.File)
	// global.Config = core.ReadCfg(global.FlagOptions.File)

	res.OkWithMsg("站点配置更新成功", c)
}

// TODO：此处原项目直接修改了前端页面的index.html来更新[标题]和[icon]等，后续需要考虑通过API接口来更新这些配置。
func UpdateSite(site conf.Site) error {
	if site.Project.Icon == "" && site.Project.Title == "" &&
		site.Seo.Keywords == "" && site.Seo.Description == "" &&
		site.Project.WebPath == "" {
		return nil
	}

	if site.Project.WebPath == "" {
		return errors.New("请配置前端地址")
	}

	file, err := os.Open(site.Project.WebPath)
	if err != nil {
		return err
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		global.Logger.Errorf("goquery解析失败: %v", err)
		return err
	}

	if site.Project.Title != "" {
		selection := doc.Find("title")
		selection.SetText(site.Project.Title)
	}

	if site.Project.Icon != "" {
		if doc.Is("link[rel='icon']") {
			selection := doc.Find("link[rel='icon']")
			selection.SetAttr("href", site.Project.Icon)
		} else {
			selection := doc.Find("head")
			selection.AppendHtml(fmt.Sprintf("<link rel='icon' href='%s' />", site.Project.Icon))
		}
	}

	if site.Seo.Keywords != "" {
		if doc.Is("meta[name='keywords']") {
			selection := doc.Find("meta[name='keywords']")
			selection.SetAttr("content", site.Seo.Keywords)
		} else {
			selection := doc.Find("head")
			selection.AppendHtml(fmt.Sprintf("<meta name='keywords' content='%s' />", site.Seo.Keywords))
		}
	}

	if site.Seo.Description != "" {
		if doc.Is("meta[name='description']") {
			selection := doc.Find("meta[name='description']")
			selection.SetAttr("content", site.Seo.Description)
		} else {
			selection := doc.Find("head")
			selection.AppendHtml(fmt.Sprintf("<meta name='description' content='%s' />", site.Seo.Description))
		}
	}

	html, err := doc.Html()
	if err != nil {
		global.Logger.Errorf("生成 html 失败: %v", err)
		return err
	}

	err = os.WriteFile(site.Project.WebPath, []byte(html), 0666)
	if err != nil {
		global.Logger.Errorf("写入 html 失败: %v", err)
		return err
	}

	return nil
}
