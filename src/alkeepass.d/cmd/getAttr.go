package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tobischo/gokeepasslib/v3"
)

var getAttrCmd = &cobra.Command{
	Use:   "getAttr",
	Short: "Get Attr",
	Run:   getAttrMain,
}

func getAttrMain(cmd *cobra.Command, args []string) {
	kbdxpath := os.Getenv("keepassxc_db_path")
	keyfilepath := os.Getenv("keepassxc_keyfile_path")
	passwd := strings.TrimSpace(os.Getenv("keepassxc_master_password"))
	cred := getCred(passwd, keyfilepath)

	path := args[0]
	ATTR := args[1]

	res := getAttr(kbdxpath, cred, path, ATTR)
	if res != "" {
		fmt.Println(res)
	} else {
		print("can not find pwd")
	}
}

func getAttr(kbdxpath string, cred *gokeepasslib.DBCredentials, path string, ATTR string) string {
	file, _ := os.Open(kbdxpath)
	defer file.Close()

	db, err := openKbdx(file, cred)
	if err != nil {
		panic(err)
	}
	root := db.Content.Root
	result := []KPEntry{}
	args := []string{path}
	scan(&root.Groups, []string{}, args, &result) // start recursive scan
	// alf := AlfredJSON{}                           // search

	entry := getEntry(result, path)
	if entry == nil {
		panic("entry is nil")
	}
	switch ATTR {
	case "username":
		return entry.Entry.GetContent("UserName")
	case "password":
		return entry.Entry.GetContent("Password")
	case "url":
		return entry.Entry.GetContent("URL")
	case "notes":
		return entry.Entry.GetContent("Notes")
	default:
		return entry.Entry.GetContent(ATTR)
	}
	return ""
}
