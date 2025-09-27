package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new entry to KeePass database",
	Long: `Add a new entry to KeePass database with specified credentials and metadata.
	
Examples:
  # Add an entry with all fields
  alkeepass add -u "john_doe" -p "secret123" -t "Google Account" -l "https://google.com" -n "Personal account" -g "Web/Google"
  
  # Add an entry with minimal fields
  alkeepass add -u "jane_doe" -p "password123" -t "Email Account"
  
  # Add to specific group
  alkeepass add -u "admin" -p "admin123" -g "Servers/Production"`,
}

type AddEntry struct {
	Username string
	Password string
	URL      string
	Title    string
	Notes    string
	Group    string
}

func init() {

	var (
		username string
		password string
		url      string
		title    string
		notes    string
		group    string
	)

	// 添加命令行参数
	addCmd.Flags().StringVarP(&username, "username", "u", "", "Username (required)")
	addCmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	addCmd.Flags().StringVarP(&title, "title", "t", "", "Entry title (optional)")
	addCmd.Flags().StringVarP(&url, "url", "l", "", "URL (optional)")
	addCmd.Flags().StringVarP(&notes, "notes", "n", "", "Notes (optional)")
	addCmd.Flags().StringVarP(&group, "group", "g", "temp/General", "Group path (optional)")

	// 标记必需参数
	addCmd.MarkFlagRequired("username")
	addCmd.MarkFlagRequired("password")

	addCmd.Run = func(cmd *cobra.Command, args []string) {
		kbdxpath := os.Getenv("keepassxc_db_path")
		keyfilepath := os.Getenv("keepassxc_keyfile_path")
		masterPasswd := strings.TrimSpace(os.Getenv("keepassxc_master_password"))

		// 如果没有提供标题，使用用户名作为默认标题
		if title == "" {
			title = username
		}

		entry := &AddEntry{
			Username: username,
			Password: password,
			URL:      url,
			Title:    title,
			Notes:    notes,
			Group:    group,
		}

		if err := saveEntryToFile(kbdxpath, keyfilepath, masterPasswd, entry); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	}
}

func saveEntryToFile(kbdxpath, keyfilepath, passwd string, entry *AddEntry) error {
	// 验证必需的环境变量
	if kbdxpath == "" {
		return fmt.Errorf("keepassxc_db_path environment variable is required")
	}

	// 打开数据库文件
	file, err := os.OpenFile(kbdxpath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open database file: %w", err)
	}
	defer file.Close()

	// 获取凭据并打开数据库
	cred := getCred(passwd, keyfilepath)
	db, err := openKbdx(file, cred)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 查找或创建组
	rootGroup := &db.Content.Root.Groups[0]    // 根组
	currentGroup := &db.Content.Root.Groups[0] // 根组
	groupPath := strings.Split(entry.Group, "/")

	for _, groupName := range groupPath {
		if groupName == "" {
			continue
		}

		found := false
		for i := range currentGroup.Groups {
			if currentGroup.Groups[i].Name == groupName {
				currentGroup = &currentGroup.Groups[i]
				found = true
				break
			}
		}

		if !found {
			// 创建新组
			newGroup := gokeepasslib.Group{
				Name: groupName,
			}
			currentGroup.Groups = append(currentGroup.Groups, newGroup)
			currentGroup = &currentGroup.Groups[len(currentGroup.Groups)-1]
		}
	}

	// 创建新条目
	now := wrappers.Now()
	newEntry := gokeepasslib.Entry{
		Values: []gokeepasslib.ValueData{
			{Key: "Title", Value: gokeepasslib.V{Content: entry.Title}},
			{Key: "UserName", Value: gokeepasslib.V{Content: entry.Username}},
			{Key: "Password", Value: gokeepasslib.V{Content: entry.Password, Protected: wrappers.NewBoolWrapper(false)}},
			{Key: "URL", Value: gokeepasslib.V{Content: entry.URL}},
			{Key: "Notes", Value: gokeepasslib.V{Content: entry.Notes}},
		},
		Times: gokeepasslib.TimeData{
			CreationTime:         &now,
			LastModificationTime: &now,
		},
	}

	// 添加到组
	currentGroup.Entries = append(currentGroup.Entries, newEntry)

	// 锁定受保护的条目
	db.LockProtectedEntries()

	// 重新打开文件用于写入（需要截断）
	file.Close()
	file, err = os.OpenFile(kbdxpath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to reopen file for writing: %w", err)
	}
	defer file.Close()

	// 编码并保存到文件
	db.Content.Root.Groups[0] = *rootGroup
	encoder := gokeepasslib.NewEncoder(file)
	if err := encoder.Encode(db); err != nil {
		return fmt.Errorf("failed to encode database: %w", err)
	}

	fmt.Printf("Successfully added entry '%s' to group '%s'\n", entry.Title, entry.Group)
	return nil
}
