FROM golang:1.11

MAINTAINER lilianewang@tencent.com
WORKDIR /go/src

COPY ./ /go/src

RUN set -ex && \
go build -v -o /usr/bin/component-gobuild \
-gcflags '-N -l' \
./*.go

RUN apk add --update git

RUN go get -u github.com/golang/build/

CMD ["component-gobuild"]

LABEL TencentHubComponent='{\
  "description": "Golang编译组件, 用以对Golang编写的程序进行编译，输出可执行文件.",\
  "input": [\
    {"name": "GIT_CLONE_URL", "desc": "必填，源代码地址，如为私有仓库需要授权; 如需使用系统关联的git仓库, 可以从系统提供的全局环境变量中获取: ${_WORKFLOW_GIT_CLONE_URL}"},\
    {"name": "GIT_REF", "desc": "非必填, 源代码git目标引用，可以是一个git branch, git tag 或者git commit ID, 默认值master"},\
    {"name": "LINT_PACKAGE", "desc": "非必填, 待分析的代码包, 通过路径的形式给出, 默认检索所有的代码包"}\
  ],\
  "output": [\
	{"name": "OUTPUT", "desc": "构建产物结果目录"}\
	]\
}'