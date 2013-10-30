package gofis

import (
	"bytes"
	"fmt"
	"github.com/astaxie/beego"
	"html/template"
	"strings"
)

var ConfigDir string

func SetConfigDir(path string) {
	ConfigDir = path
}

/**
 * regist user's func to template system
 */
func Register(settings map[string]string) {

	SetConfigDir(settings["config_dir"])

	beego.AddFuncMap("pageStart", PageStart)
	beego.AddFuncMap("placefolder", Placefolder)
	beego.AddFuncMap("require", Require)
	beego.AddFuncMap("widget", Widget)
	beego.AddFuncMap("pageEnd", PageEnd)
}

func AfterProcess(buffer []byte) ([]byte, error) {
	jsBlock := bytes.NewBufferString("<!--FIS_JS-->")
	cssBlock := bytes.NewBufferString("<!--FIS_CSS-->")

	var script string
	var link string

	if js, ok := gofis.StaticArr["js"]; ok {
		for _, v := range js {
			script += "<script type=\"text/javascript\" src=\"" + v + "\"></script>\n"
		}
	}

	if css, ok := gofis.StaticArr["css"]; ok {
		for _, v := range css {
			link += "<link type=\"text/css\" rel=\"stylesheet\" href=\"" + v + "\" />"
		}
	}

	buffer = bytes.Replace(buffer, jsBlock.Bytes(), bytes.NewBufferString(script).Bytes(), 1)
	buffer = bytes.Replace(buffer, cssBlock.Bytes(), bytes.NewBufferString(link).Bytes(), 1)

	return buffer, nil
}

func Require(args ...interface{}) string {
	var async bool
	async = false
	if len(args) == 2 {
		id := args[0].(string)
		strAsync := args[1].(string)
		if strAsync == "true" {
			async = true
		}
		gofis.Load(id, ConfigDir, async)
	} else if len(args) == 1 {
		id := args[0].(string)
		gofis.Load(id, ConfigDir, async)
	} else {
		beego.Error("require id [async]")
	}
	return ""
}

/**
 * Page Start
 * init resource api
 */
func PageStart(args ...interface{}) string {
	return ""
}

func Placefolder(s string) template.HTML {
	return template.HTML("<!--FIS_" + strings.ToUpper(s) + "-->")
}

/**
 * Part of full page, it's very important!
 * 1. collect Widget's static
 * 2. display widget content
 */
func Widget(args ...interface{}) template.HTML {
	ok := false
	var s string
	if len(args) == 1 {
		s, ok = args[0].(string)
	}
	if !ok {
		s = fmt.Sprint(args...)
	}

	tplPath := gofis.GetUri(s, ConfigDir)

	tpl, _ok := beego.BeeTemplates[tplPath]
	if !_ok {
		beego.Error("cont't found template: ", s)
		return ""
	} else {
		var out bytes.Buffer
		tpl.Execute(&out, nil)
		gofis.Load(s, ConfigDir, false)
		return template.HTML(out.String())
	}
}

/**
 * insert link of CSS file
 * insert script of JavaScript file
 * others ops about static file
 * Page End
 */
func PageEnd(args ...interface{}) string {
	return ""
}
