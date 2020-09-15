package cmd

import (
	"flag"
	"fmt"
	"github.com/neucn/elise"
	"github.com/neucn/elise/ics"
	"os"
)

var (
	u, p, f string
	webVPN  bool
	path    string
)

func init() {
	flag.StringVar(&u, "u", "", "一网通学号")
	flag.StringVar(&p, "p", "", "一网通密码")
	flag.StringVar(&f, "f", "ics", "输出格式，目前仅支持ics格式")
	flag.StringVar(&path, "o", "", "输出路径，默认为当前目录")
	flag.BoolVar(&webVPN, "v", false, "使用 webVPN")

	flag.Usage = usage
}

func Run() {
	if len(u) == 0 || len(p) == 0 {
		fatal("\n必须填写学号与密码，请使用 elise -help 查看使用说明")
	}

	generator, err := elise.New(u, p, webVPN)
	if err != nil {
		fatal("\n登陆时出错: " + err.Error())
	}

	var generateFunc elise.GenerateFunc
	switch f {
	case "ics":
		generateFunc = ics.Generate
	default:
		fatal("\n尚未支持导出 " + f + " 格式的课程表")
	}

	absPath, err := generator.Generate(generateFunc, path)
	if err != nil {
		fatal("\n解析课表时出错: " + err.Error())
	}

	fmt.Printf("\n课程表已导出至: %s\n", absPath)
}

func fatal(content string) {
	_, _ = fmt.Fprintln(os.Stderr, content)
	os.Exit(2)
}

func usage() {
	fmt.Println(`Elise 东北大学教务系统课程表导出工具

  -u    一网通学号
  -p    一网通密码
  -v    使用 webVPN，默认不使用
  -f    输出格式，默认为ics格式，可选值: ics
  -o    输出路径，默认为当前目录

获取帮助 https://github.com/neucn/elise/issues`)
}
