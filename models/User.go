package models

import (
	"EasyChat/utils/AES"
	"EasyChat/utils/StringByte"
)

type User struct {
	Id       int64  `gorm:"column:id" json:"id"`
	Username string `gorm:"column:username" json:"username"`
	Password string `gorm:"column:password" json:"password"`
}

func (User) TableName() string {
	return "users"
}

// 查询用户名是否存在
func (u *User) CheckUsername() (bool, error) {
	var count int64
	result := DB.Model(&User{}).Where("username = ?", u.Username).Count(&count)
	if result.Error != nil {
		return true, result.Error
	}
	if count != 0 {
		return true, nil
	}
	return false, nil
}

// 创建用户
func (u *User) Create() error {
	result := DB.Create(&u)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// 查询用户信息
func (u *User) Check(username string) error {
	result := DB.Where("username = ?", username).First(&u)
	if result.Error != nil {
		return result.Error
	}
	//对密文密码进行解密
	bytes, err := AES.AesDecrypt(u.Password, StringByte.String2Bytes(AES.Key))
	if err != nil {
		return err
	}
	u.Password = StringByte.Bytes2String(bytes)
	return nil
}

// 根据id查询用户名
func (u *User) GetUsernameById(id int64) (string, error) {
	result := DB.Model(&User{}).Select("username").Where("id = ?", id).First(&u)
	if result.Error != nil {
		return "", result.Error
	}
	return u.Username, nil
}
