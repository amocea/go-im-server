package service

import (
	"errors"
	"github.com/amocea/go-im-chat/defs"
	"github.com/amocea/go-im-chat/util"
	"time"
	"xorm.io/xorm"
)

/*
关于 User 的服务
*/

type Service struct {
	DB *xorm.Engine
	util.Validator
}

// UserService 用户服务
type UserService struct {
	Service
}

var (
	ErrUserExist    = errors.New("该手机号码已注册")
	ErrUserNotExist = errors.New("该手机号码尚未注册")
	ErrPassword     = errors.New("密码错误")
)

// Register 用户注册行为
func (u *UserService) Register(mobile, passwd, nickname, avatar, sex string) (user defs.User, err error) {
	// 检测手机号码是否存在
	exist, _ := u.DB.Exist(&defs.User{Mobile: mobile})
	if exist {
		// 表示存在手机
		err = ErrUserExist
		return
	}
	// 否则拼接插入数据库
	salt := util.GenerateSalt()
	user = defs.User{
		Mobile:   mobile,
		Nickname: nickname,
		Avatar:   avatar,
		Sex:      sex,
		Passwd:   util.EncryptPassword(passwd, salt),
		CreateAt: time.Now(),
		Salt:     salt,
	}

	// 插入创建数据
	_, err = u.DB.InsertOne(&user)
	// 最后返回新用户信息
	user.Passwd = "" // 将密码隐藏 置零
	user.Salt = ""
	return user, err
}

// Login 用户登录行为
func (u *UserService) Login(mobile, passwd string) (defs.User, error) {
	// 查找用户，进行比对
	var tmp defs.User

	_, e := u.DB.Where("mobile = ?", mobile).Get(&tmp)
	if e != nil {
		return tmp, e
	}
	if tmp.Id == 0 {
		return tmp, ErrUserNotExist
	}

	// 表示找到了数据
	if !util.ValidatePassword(passwd, tmp.Salt, tmp.Passwd) {
		// 说明密码错误
		return defs.User{}, ErrPassword
	}

	// 密码正确 刷新 token，确保 token 的合法性
	tmp.Token = util.Md5Encode(util.GenerateSalt())
	// 更新
	_, e = u.DB.ID(tmp.Id).Cols("token").Update(&tmp)
	if e != nil {
		return defs.User{}, e
	}
	// 消除敏感数据
	tmp.Salt = ""
	tmp.Passwd = ""
	return tmp, nil
}

func (u *UserService) FindById(id int64) (*defs.User, error) {
	var r = new(defs.User)

	_, err := u.DB.ID(id).Get(r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
