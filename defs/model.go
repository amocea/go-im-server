package defs

import "time"

const (
	SexWomen   = "W"
	SexMan     = "M"
	SexUnknown = "U"
)

const (
	ConcatCateUser      = 0x01
	ConcatCateCommunity = 0x02
	CommunityCateCom    = 0x03
)

var (
	DefaultAvatar = ""
)

// User /* 用户实体类
type User struct {
	//用户ID
	Id       int64  `xorm:"pk autoincr bigint(20)" form:"id" json:"id"`
	Mobile   string `xorm:"varchar(20)" form:"mobile" json:"mobile"`
	Passwd   string `xorm:"varchar(40)" form:"passwd" json:"-"` // 什么角色
	Avatar   string `xorm:"varchar(150)" form:"avatar" json:"avatar"`
	Sex      string `xorm:"varchar(2)" form:"sex" json:"sex"`            // 什么角色
	Nickname string `xorm:"varchar(20)" form:"nickname" json:"nickname"` // 什么角色
	//加盐随机字符串6
	Salt   string `xorm:"varchar(20)" form:"salt" json:"-"`    // 什么角色
	Online int    `xorm:"int(10)" form:"online" json:"online"` //是否在线
	//前端鉴权因子, chat?id=1&token=x
	Token    string    `xorm:"varchar(40)" form:"token" json:"token"`    // 什么角色
	Memo     string    `xorm:"varchar(140)" form:"memo" json:"memo"`     // 什么角色
	CreateAt time.Time `xorm:"datetime" form:"createat" json:"createat"` // 什么角色
}

// Contact 添加好友
type Contact struct {
	Id int64 `xorm:"pk autoincr bigint(20)" form:"id" json:"id"`
	//谁的10000
	OwnerId int64 `xorm:"bigint(20)" form:"ownerid" json:"ownerid"` // 记录是谁的
	//对端,10001
	Dstobj int64 `xorm:"bigint(20)" form:"dstobj" json:"dstobj"` // 对端信息
	//
	Cate int    `xorm:"int(11)" form:"cate" json:"cate"`      // 什么类型
	Memo string `xorm:"varchar(120)" form:"memo" json:"memo"` // 备注
	//
	Createat time.Time `xorm:"datetime" form:"createat" json:"createat"` // 创建时间
}

// Community 群
type Community struct {
	Id int64 `xorm:"pk autoincr bigint(20)" form:"id" json:"id"`
	//名称
	Name string `xorm:"varchar(30)" form:"name" json:"name"`
	//群主ID
	OwnerId int64 `xorm:"bigint(20)" form:"ownerid" json:"ownerid"` // 什么角色
	//群logo
	Icon string `xorm:"varchar(250)" form:"icon" json:"icon"`
	//como
	Cate int `xorm:"int(11)" form:"cate" json:"cate"` // 什么角色
	//描述
	Memo string `xorm:"varchar(120)" form:"memo" json:"memo"` // 什么角色
	//
	Createat time.Time `xorm:"datetime" form:"createat" json:"createat"` // 什么角色
}

type ChatDto struct {
	Id    int64  `json:"id"`
	Token string `json:"string"`
}

const (
	CMD_SINGLE_MSG int = 10
	CMD_ROOM_MSG       = 11
	CMD_HEART          = 0
)

type MediaType int

const (
	TEXT MediaType = iota
	NEWS
	VOICE
	IMG
	REDPACKAGR
	EMOJ
	LINK
	VIDEO
	CONTACT
)

type Message struct {
	Id      int64  `json:"id,omitempty" form:"id"`           //消息ID
	Userid  int64  `json:"userid,omitempty" form:"userid"`   //谁发的
	Cmd     int    `json:"cmd,omitempty" form:"cmd"`         //群聊还是私聊
	Dstid   int64  `json:"dstid,omitempty" form:"dstid"`     //对端用户ID/群ID
	Media   int    `json:"media,omitempty" form:"media"`     //消息按照什么样式展示
	Content string `json:"content,omitempty" form:"content"` //消息的内容
	Pic     string `json:"pic,omitempty" form:"pic"`         //预览图片
	Url     string `json:"url,omitempty" form:"url"`         //服务的URL
	Memo    string `json:"memo,omitempty" form:"memo"`       //简单描述
	Amount  int    `json:"amount,omitempty" form:"amount"`   //其他和数字相关的
}
