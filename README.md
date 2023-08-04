# gopkg 本质上是一个反向代理, 只用于代理Golang的mod

## 原理

Golang的模块下载时使用[go mod协议](https://go.dev/ref/mod)，实现了协议的代码仓库（如github、gitea）都在go命令发起"go-get=1"查询时都会返回一个html文本，格式如下:


```html
<!doctype html>
<html>
	<head>
		<meta name="go-import" content="172.16.18.33/gopkg/demo git http://172.16.18.33/gopkg/demo.git">
		<meta name="go-source" content="172.16.18.33/gopkg/demo _ http://172.16.18.33/gopkg/demo/src/branch/dev{/dir} http://172.16.18.33/gopkg/demo/src/branch/dev{/dir}/{file}#L{line}">
	</head>
	<body>
		go get --insecure 172.16.18.33/gopkg/demo
	</body>
</html>
```

我们通过反向代理提供"go mod"服务时, 该地址需要同步被修改为代理前端, 即本命令行工具所解决的问题，经过本服务后格式转换为:

```html
<!doctype html>
<html>
	<head>
		<meta name="go-import" content="code.local/gopkg/demo git http://code.local/gopkg/demo.git">
		<meta name="go-source" content="code.local/gopkg/demo _ http://code.local/gopkg/demo/src/branch/dev{/dir} http://code.local/gopkg/demo/src/branch/dev{/dir}/{file}#L{line}">
	</head>
	<body>
		go get --insecure code.local/gopkg/demo
	</body>
</html>
```

## Usage

```sh
go run . -backend https://172.16.18.33 -frontend http://code.local -addr :8081 -debug
```

## 说明

- 其中"http://code.local"现在是一个nginx本地代理,代理到"http://127.0.0.1:8081"
- ":8081"是本命令行监听地址
- "https://172.16.18.33"是真正的后端Git服务
- debug打印反向代理的url转换日志

### nginx 代理

```
location / {
		proxy_pass http://localhost:8081/;
		proxy_set_header Host $http_host;
	}
```

### 本机hosts

```
127.0.0.1 code.local
```