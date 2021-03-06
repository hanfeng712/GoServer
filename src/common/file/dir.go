package file

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//dir, name := filepath.Dir(path), filepath.Base(path) //不包含尾分隔符，且会转换为对应平台的分隔符
//dir, name := filepath.Split(path) 				   //dir包含尾分隔符，同参数的分隔符

func GetExeDir() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return filepath.ToSlash(dir)
}
func IsExist(path string) bool { //file or folder
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

//获取指定目录下(不递归子目录)的所有文件 --- names, err := filepath.Glob("csv/*.csv")
//获取指定目录及子目录下的所有文件，可以匹配后缀过滤 --- names, err := WalkDir("csv/", ".csv")
func WalkDir(dir, suffix string) ([]string, error) {
	ret := make([]string, 0, 16)
	//filepath.Walk【对软连接无效】
	//err := filepath.Walk(dir, func(filename string, fi os.FileInfo, err error) error {
	//	if err != nil {
	//		return err
	//	}
	//	if fi.IsDir() {
	//		return nil
	//	}
	//	if strings.HasSuffix(fi.Name(), suffix) {
	//		ret = append(ret, filepath.ToSlash(filename))
	//	}
	//	return nil
	//})
	err := _walkDir(dir, suffix, &ret)
	return ret, err
}
func _walkDir(dir, suffix string, names *[]string) error {
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	if f, err := os.Open(dir); err == nil {
		list, err := f.Readdir(-1)
		f.Close()
		if err == nil {
			for _, fi := range list {
				name := dir + fi.Name()
				if fi.Mode()&(os.ModeDir|os.ModeSymlink) != 0 {
					if err = _walkDir(name, suffix, names); err != nil {
						return err
					}
				} else if strings.HasSuffix(name, suffix) {
					*names = append(*names, name)
				}
			}
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func ReadLine(filename string, cb func(string)) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		cb(strings.TrimSpace(line))
	}
	return nil
}

func CreateFile(dir, name string, flag int) (*os.File, error) {
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	if file, err := os.OpenFile(dir+name, flag|os.O_CREATE, 0666); err != nil {
		return nil, err
	} else {
		return file, nil
	}
}
func CreateTemplate(data interface{}, outDir, filename, tempText string) (bf bytes.Buffer) {
	tpl, err := template.New(filename).Parse(tempText)
	if err != nil {
		panic(err.Error())
		return bf
	}
	if err = tpl.Execute(&bf, data); err != nil {
		panic(err.Error())
		return bf
	}
	f, err := CreateFile(outDir, filename, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		panic(err.Error())
		return bf
	}
	f.Write(bf.Bytes())
	f.Close()
	return bf
}

// ------------------------------------------------------------
// 计算文件md5
func CalcMd5(name string) string {
	f, err := os.Open(name)
	if err != nil {
		return ""
	}
	defer f.Close()

	md5hash := md5.New()
	io.Copy(md5hash, f)
	return fmt.Sprintf("%x", md5hash.Sum(nil))
}
