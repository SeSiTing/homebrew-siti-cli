package main

// version 承担两个职责：
//  1. CI 版本检测信号：push main 后 GitHub Actions 用正则提取此值，
//     若与最新 git tag 不同则触发 goreleaser 发版。
//  2. dev build 默认值：release build 由 goreleaser ldflags 注入真实版本号覆盖此值。
//
// 发版前将此值改为目标版本号，例如：
//
//	var version = "2.1.0"
var version = "2.0.15"
