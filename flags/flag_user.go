package flags

import (
	"fmt"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/pwd"
	"os"

	"golang.org/x/crypto/ssh/terminal"
	"gorm.io/gorm"
)

type FlagUser struct{}

func (u *FlagUser) Create(db *gorm.DB) {
	var role enum.RoleType
	fmt.Println("请输入数字选择用户角色: ")
	for r := 1; r <= enum.RoleTypeCount; r++ {
		fmt.Printf("%d. %s\n", r, enum.RoleType(r).String())
	}
	if _, err := fmt.Scanln(&role); err != nil {
		fmt.Printf("输入类型错误 %s\n", err.Error())
		return
	}
	if int(role) < int(enum.RoleAdmin) || int(role) > enum.RoleTypeCount {
		fmt.Printf("输入角色类型错误 %d\n", role)
		return
	}

	var username string
	fmt.Println("请输入用户名: ")
	fmt.Scanln(&username)

	var model models.UserModel
	if db.Take(&model, "username = ?", username).Error == nil {
		fmt.Printf("用户名 %s 已存在\n", username)
		return
	}

	fmt.Println("请输入密码: ")
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("读取密码错误 %s\n", err.Error())
		return
	}
	fmt.Println("请确认密码: ")
	rePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("读取密码错误 %s\n", err.Error())
		return
	}
	if string(password) != string(rePassword) {
		fmt.Println("两次密码输入不一致")
		return
	}

	// 密码加密
	hashedPassword, err := pwd.GenerateFromPassword(string(password))
	if err != nil {
		fmt.Printf("密码加密错误 %s\n", err.Error())
		return
	}

	// 创建用户
	if err := db.Create(&models.UserModel{
		Username:       username,
		Nickname:       "命令用户",
		Password:       hashedPassword,
		Role:           role,
		RegisterSource: enum.RegisterTerminalSourceType,
	}).Error; err != nil {
		fmt.Printf("创建用户错误 %s\n", err.Error())
		return
	}
	msg := fmt.Sprintf("用户 %s 创建成功\n", username)
	global.Logger.Info(msg)
}
