/***********************************************************************
* @ 生成RpcFunc注册代码
* @ brief
	1、用正则表达式检测各个服务器源码中的 Rpc_* 函数，提取函数名

	2、golang 部分，函数散列在各处，所以遍历了源文件

	3、c++、c# 部分，函数均在固定文件有记录，只需解析单个文件即可

* @ author zhoumf
* @ date 2017-10-17
***********************************************************************/
package main

import (
	"bytes"
	"common/file"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

const (
	K_RegistOutDir   = K_OutDir + "rpc/"
	K_RegistFileName = "generate_rpc.go"
)

type Func struct {
	Pack string //package dir
	Name string
}
type RpcInfo struct {
	Moudles       map[string]bool //package
	TcpRpc        []Func
	HttpRpc       []Func
	HttpPlayerRpc []Func
	HttpHandle    []Func
}

func generatRpcRegist(svr string) *RpcInfo {
	pinfo := &RpcInfo{Moudles: make(map[string]bool)}
	names, _ := file.WalkDir(K_SvrDir+svr, ".go")
	for _, v := range names {
		moudle := "" //package dir
		file.ReadLine(v, func(line string) {
			fname := "" //func name
			if moudle == "" {
				//moudle = getPackage(line)
				moudle = filepath.Dir(v)[len(K_SvrDir):]
				moudle = filepath.ToSlash(moudle)
			} else if fname = getTcpRpc(line); fname != "" {
				pinfo.TcpRpc = append(pinfo.TcpRpc, Func{moudle, fname})
			} else if fname = getHttpRpc(line); fname != "" {
				pinfo.HttpRpc = append(pinfo.HttpRpc, Func{moudle, fname})
			} else if fname = getHttpPlayerRpc(line); fname != "" {
				pinfo.HttpPlayerRpc = append(pinfo.HttpPlayerRpc, Func{moudle, fname})
			} else if fname = getHttpHandle(line); fname != "" {
				pinfo.HttpHandle = append(pinfo.HttpHandle, Func{moudle, fname})
			}
			if moudle != "" && fname != "" {
				pinfo.Moudles[moudle] = true
			}
		})
	}
	pinfo.makeFile(svr)
	return pinfo
}

// -------------------------------------
// -- 提取 package、RpcFunc
func getPackage(s string) string {
	if ok, _ := regexp.MatchString(`^package \w+`, s); ok {
		reg := regexp.MustCompile(`\w+`)
		return reg.FindAllString(s, -1)[1]
	}
	return ""
}
func getTcpRpc(s string) string {
	if ok, _ := regexp.MatchString(`^func Rpc_\w+\(\w+, \w+ \*common.NetPack, \w+ \*tcp.TCPConn\) \{`, s); ok {
		reg := regexp.MustCompile(`Rpc_\w+`)
		return reg.FindAllString(s, -1)[0]
	}
	return ""
}
func getHttpRpc(s string) string {
	if ok, _ := regexp.MatchString(`^func Rpc_\w+\(\w+, \w+ \*common.NetPack\) \{`, s); ok {
		reg := regexp.MustCompile(`Rpc_\w+`)
		return reg.FindAllString(s, -1)[0]
	}
	return ""
}
func getHttpPlayerRpc(s string) string {
	if ok, _ := regexp.MatchString(`^func Rpc_\w+\(\w+, \w+ \*common.NetPack, \w+ interface\{\}\) \{`, s); ok {
		reg := regexp.MustCompile(`Rpc_\w+`)
		return reg.FindAllString(s, -1)[0]
	}
	return ""
}
func getHttpHandle(s string) string {
	if ok, _ := regexp.MatchString(`^func Http_\w+\(\w+ http.ResponseWriter, \w+ \*http.Request\) \{`, s); ok {
		reg := regexp.MustCompile(`Http_\w+`)
		return reg.FindAllString(s, -1)[0][5:]
	}
	return ""
}

// -------------------------------------
// -- 填充模板
const codeRegistTemplate = `
// Generated by GoServer/src/generat
// Don't edit !
package rpc
import (
	"common/net/register"
	{{if .UsingRpcEnum}}"generate_out/rpc/enum"{{end}}
	{{range $k, $_ := .Moudles}}"{{$k}}"
	{{end}}
)
func init() {
	register.RegTcpRpc(map[uint16]register.TcpRpc{
		{{range .TcpRpc}}enum.{{.Name}}: {{GetPackage .Pack}}.{{.Name}},
		{{end}}
	})
	register.RegHttpRpc(map[uint16]register.HttpRpc{
		{{range .HttpRpc}}enum.{{.Name}}: {{GetPackage .Pack}}.{{.Name}},
		{{end}}
	})
	register.RegHttpPlayerRpc(map[uint16]register.HttpPlayerRpc{
		{{range .HttpPlayerRpc}}enum.{{.Name}}: {{GetPackage .Pack}}.{{.Name}},
		{{end}}
	})
	register.RegHttpHandler(map[string]register.HttpHandle{
		{{range .HttpHandle}}"{{.Name}}": {{GetPackage .Pack}}.Http_{{.Name}},
		{{end}}
	})
}
`

func (self *RpcInfo) makeFile(svr string) {
	filename := K_RegistFileName
	var err error
	tpl := template.New(filename).Funcs(map[string]interface{}{
		"GetPackage": GetPackage,
	})
	if tpl, err = tpl.Parse(codeRegistTemplate); err != nil {
		panic(err.Error())
		return
	}
	var bf bytes.Buffer
	if err = tpl.Execute(&bf, self); err != nil {
		panic(err.Error())
		return
	}
	if err := os.MkdirAll(K_RegistOutDir+svr, 0777); err != nil {
		panic(err.Error())
		return
	}
	f, err := os.OpenFile(K_RegistOutDir+svr+"/"+filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err.Error())
		return
	}
	defer f.Close()
	f.Write(bf.Bytes())
}
func (p *RpcInfo) UsingRpcEnum() bool { return len(p.TcpRpc)+len(p.HttpRpc)+len(p.HttpPlayerRpc) > 0 }
func GetPackage(dir string) string    { return dir[strings.LastIndex(dir, "/")+1:] }
