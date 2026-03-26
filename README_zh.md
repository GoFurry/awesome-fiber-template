# awesome-go-template

![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)

[English](./README.md)

`awesome-go-template` 是一个用于快速搭建 Go 后端项目的模板仓库，按不同 Web 框架维护对应脚手架，方便直接启动新项目或对比不同技术栈的组织方式。

## 仓库目标

这个仓库主要用于：

- 提供不同框架下的后端脚手架模板
- 统一沉淀常见工程化能力
- 帮助快速开始新服务开发
- 便于横向比较不同框架的项目结构

## 目录说明

- `fiber/v3/basic`：当前最完整的模板，包含应用启动、配置加载、路由、中间件、数据库集成、Redis、定时任务、监控指标，以及可选的前端静态资源嵌入能力
- `chi`：预留的框架模板目录
- `echo`：预留的框架模板目录
- `gin`：预留的框架模板目录
- `net`：基于 Go 标准库的模板目录

## 快速开始

当前可直接作为完整示例使用的是 `fiber/v3/basic`：

```bash
cd fiber/v3/basic
go run . serve
```

相关配置文件：

- 服务配置：`fiber/v3/basic/config/server.yaml`
- 数据库配置：支持 `sqlite`、`postgres`、`mysql`

## 说明

- 这是一个模板集合仓库，不是根目录直接运行的单体应用。
- 不同框架目录当前完成度可能不同，后续会逐步补齐。

## 许可证

本项目使用 MIT License，详见 `LICENSE`。
