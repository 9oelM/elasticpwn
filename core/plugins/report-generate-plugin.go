package EPPlugins

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	EPUtils "github.com/9oelm/elasticpwn/core/util"
)

// Keep the behavior of this plugin as idempotent as possible
// to avoid any weird errors
type ReportGeneratePlugin struct {
	MongoUrl       string
	CollectionName string
	ServerRootUrl  string
}

func (rp *ReportGeneratePlugin) cmdInDir(cwd string, rootCommand string, arg ...string) {
	cmd := exec.Command(rootCommand, arg...)
	cmd.Dir = filepath.FromSlash(cwd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(string(output))
		log.Fatal(err)
	}
	log.Println(string(output))
}

func (rp *ReportGeneratePlugin) Run() {
	unixHomePath, unixHomePathOk := os.LookupEnv("HOME")
	windowsHomePath, windowsHomePathOk := os.LookupEnv("HOMEPATH")

	finalHomePath := ""
	if unixHomePathOk {
		finalHomePath = unixHomePath
	} else if windowsHomePathOk {
		finalHomePath = windowsHomePath
	} else {
		log.Fatal("Either of environment variables \"HOME\" or \"HOMEPATH\" is not defined. Define your ~ directory as that and try again.")
	}

	elasticpwnRootDir := filepath.FromSlash(fmt.Sprintf("%v/%v", finalHomePath, ".elasticpwn"))
	elasticpwnRepoCloneDir := filepath.FromSlash(fmt.Sprintf("%v/%v", elasticpwnRootDir, "elasticpwn"))
	elasticpwnReportFrontendDir := filepath.FromSlash(fmt.Sprintf("%v/report/frontend", elasticpwnRepoCloneDir))

	log.Println(fmt.Sprintf("Trying to create a directory at %v", elasticpwnRootDir))
	err := os.Mkdir(elasticpwnRootDir, 0755)

	if !errors.Is(os.ErrExist, err) {
		log.Println(fmt.Sprintf("%v already exists. Continuing.", elasticpwnRootDir))
	} else if err != nil {
		log.Fatal(err)
	}

	err = os.RemoveAll(elasticpwnRepoCloneDir)
	if err != nil {
		log.Fatal(err)
	}
	output, err := exec.Command("git", "clone", "https://github.com/9oelM/elasticpwn.git", elasticpwnRepoCloneDir).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(output))

	nvmrcContent := string(EPUtils.ReadFile(filepath.FromSlash(fmt.Sprintf("%v/.nvmrc", elasticpwnReportFrontendDir))))
	nodeVOutput, err := exec.Command("node", "-v").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	currentNodeVersion := strings.ReplaceAll(string(nodeVOutput), "\n", "")

	if strings.TrimSpace(nvmrcContent) != strings.TrimSpace(currentNodeVersion) {
		log.Println(fmt.Sprintf("elasticpwn recommends node version of %v, but the current node version is %v. Continuing regardless. If you are seeing unexpected things, switch to node version of %v and try again.", nvmrcContent, currentNodeVersion, nvmrcContent))
	}

	rp.cmdInDir(elasticpwnReportFrontendDir, "npm", "i")

	dotEnvLocalContent := fmt.Sprintf(`MONGODB_URI=%v
COLLECTION_NAME=%v
DB_NAME=ep
SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH=%v
`, rp.MongoUrl, rp.CollectionName, rp.ServerRootUrl)
	EPUtils.OverwriteFile(filepath.FromSlash(fmt.Sprintf("%v/.env.local", elasticpwnReportFrontendDir)), dotEnvLocalContent)

	nextConfigJsContent := `
	/** @type {import('next').NextConfig} */
	module.exports = {
		reactStrictMode: true,
		distDir: 'build',
		// https://stackoverflow.com/questions/66137368/next-js-environment-variables-are-undefined-next-js-10-0-5
		env: {
			SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH: process.env.SERVER_ROOT_URL_WITHOUT_TRAILING_SLASH
		},
		// https://github.com/vercel/next.js/discussions/13578
		// this needs to be uncommented when npm run build
		assetPrefix: '.',
	}	
	`
	EPUtils.OverwriteFile(filepath.FromSlash(fmt.Sprintf("%v/next.config.js", elasticpwnReportFrontendDir)), nextConfigJsContent)

	rp.cmdInDir(elasticpwnReportFrontendDir, "npm", "run", "build")
	rp.cmdInDir(elasticpwnReportFrontendDir, "npm", "run", "export")
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cwd)
	err = os.Rename(filepath.FromSlash(fmt.Sprintf("%v/out", elasticpwnReportFrontendDir)), filepath.FromSlash(fmt.Sprintf("%v/report", cwd)))
	if err != nil {
		log.Fatal(err)
	}
}
