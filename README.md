<p align="center">
    <img src="https://github.com/neucn/elise/blob/master/docs/logo.png?raw=true" alt="logo" width="200">
</p>

<h1 align="center">Elise</h1>

<p align="center">
    <img src="https://img.shields.io/github/v/tag/neucn/elise?label=version&style=flat-square" alt="">
    <img src="https://img.shields.io/github/license/neucn/elise?style=flat-square" alt="">
</p>


> 东北大学 教务处课程表导出工具

## 下载

进入本仓库 [Release 页面](https://github.com/neucn/elise/releases/latest) 下载最新压缩包后解压即可。

目前提供了 amd64 (64位)架构 Linux, OSX, Windows 使用的压缩包。

其他架构或系统请自行构建。

## 更新

下载最新版压缩包并解压覆盖原来的程序。

## 使用

下载压缩包并解压之后，在当前目录打开命令行即可使用。

也可将解压后的路径添加入`Path`环境变量中，从而可以在任何目录下使用本工具。

例如，`elise -u "学号" -p "密码" -v` 使用 webVPN 登陆教务处，并导出`ics`格式的课程表至当前路径下

```shell script
> elise -u "学号" -p "密码" -v

当前为第 2 教学周
计算得到本学期开始于 2020-09-06
官方校历 http://www.neu.edu.cn/xl/list.htm

======开始生成课程表======
...
======课程表生成成功======

课程表已导出至: /test/courses.ics
```

更多参数可以参见 `elise --help` 的输出，有默认值的参数都可以省略
```shell script
> elise --help

Elise 东北大学教务系统课程表导出工具

  -u    一网通学号
  -p    一网通密码
  -v    使用 webVPN, 默认不使用
  -f    输出格式, 默认为ics格式, 可选值: ics
  -o    输出路径, 默认为当前目录

获取帮助 https://github.com/neucn/elise/issues
```

可以通过 `elise --version` 查看当前使用的版本

## 效果

本工具导出的`ics`格式课程表效果可参见 [whoisnian/getMyCourses#效果图](https://github.com/whoisnian/getMyCourses#%E6%95%88%E6%9E%9C%E5%9B%BE)

> 更多导出格式等待支持

## 鸣谢

[whoisnian/getMyCourses](https://github.com/whoisnian/getMyCourses)

## 开源协议
MIT License.