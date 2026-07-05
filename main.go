// Command QYKit 按照拳游 Kratos 微服务约定，一键生成新的项目骨架。
//
// 用法（在 QYKit 项目根目录执行）：
//
//	go run . -module gl.quanyougame.net/backend/kbfs_demo -name kbfs_demo -out ../kbfs_demo
//
// 参数：
//
//	-module  新项目的 Go module 路径（必填），例如 gl.quanyougame.net/backend/kbfs_demo
//	-name    服务短名（选填，默认取 module 最后一段），用于二进制名、容器名等
//	-server  Kratos 注册的服务名（选填，默认 qy.kbfs.<name>）
//	-out     生成目录（选填，默认 ./<name>）
//	-force   目标目录已存在文件时是否覆盖（默认 false）
package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:templates
var templatesFS embed.FS

// templateData 提供给所有模板的渲染上下文。
type templateData struct {
	Module     string // go module 路径
	AppName    string // 服务短名
	ServerName string // Kratos 服务注册名
}

func main() {
	var (
		module = flag.String("module", "", "新项目的 Go module 路径，例如 gl.quanyougame.net/backend/kbfs_demo（必填）")
		name   = flag.String("name", "", "服务短名，默认取 module 最后一段")
		server = flag.String("server", "", "Kratos 服务注册名，默认 qy.kbfs.<name>")
		out    = flag.String("out", "", "生成目录，默认 ./<name>")
		force  = flag.Bool("force", false, "目标目录存在同名文件时是否覆盖")
	)
	flag.Parse()

	if *module == "" {
		fmt.Fprintln(os.Stderr, "错误：必须通过 -module 指定 Go module 路径")
		flag.Usage()
		os.Exit(1)
	}

	appName := *name
	if appName == "" {
		appName = lastSegment(*module)
	}
	serverName := *server
	if serverName == "" {
		serverName = "qy.kbfs." + strings.TrimPrefix(appName, "kbfs_")
	}
	outDir := *out
	if outDir == "" {
		outDir = "./" + appName
	}

	data := templateData{
		Module:     *module,
		AppName:    appName,
		ServerName: serverName,
	}

	if err := generate(outDir, data, *force); err != nil {
		fmt.Fprintf(os.Stderr, "生成失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 项目骨架已生成到 %s\n", outDir)
	fmt.Println("下一步：")
	fmt.Printf("  cd %s\n", outDir)
	fmt.Println("  go mod tidy")
	fmt.Println("  make config     # 修改 conf.proto 后重新生成 conf.pb.go")
	fmt.Println("  make gorm_gen   # 可选：从数据库表生成 model/query")
	fmt.Println("  make run-api")
}

// generate 遍历内嵌模板，渲染并写入目标目录。
func generate(outDir string, data templateData, force bool) error {
	const root = "templates"
	return fs.WalkDir(templatesFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		// 去掉 .tmpl 后缀；路径中的 __dot__ 还原为前导点（如 __dot__gitignore -> .gitignore）。
		rel = strings.TrimSuffix(rel, ".tmpl")
		rel = strings.ReplaceAll(rel, "__dot__", ".")

		dst := filepath.Join(outDir, rel)
		if !force {
			if _, statErr := os.Stat(dst); statErr == nil {
				return fmt.Errorf("文件已存在：%s（使用 -force 覆盖）", dst)
			}
		}

		raw, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}

		tmpl, err := template.New(rel).Delims("[[", "]]").Parse(string(raw))
		if err != nil {
			return fmt.Errorf("解析模板 %s 失败：%w", path, err)
		}

		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer f.Close()

		if err := tmpl.Execute(f, data); err != nil {
			return fmt.Errorf("渲染模板 %s 失败：%w", path, err)
		}
		fmt.Printf("  + %s\n", rel)
		return nil
	})
}

func lastSegment(module string) string {
	parts := strings.Split(strings.TrimRight(module, "/"), "/")
	return parts[len(parts)-1]
}
