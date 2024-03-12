package service

import (
	"errors"
	"github.com/amocea/go-im-chat/defs"
	"github.com/amocea/go-im-chat/util"
	"log"
	"time"
	"xorm.io/xorm"
)

/*
联系人服务
*/
var (
	ErrAddOneself        = errors.New("不能自己添加自己")
	ErrContactExist      = errors.New("请勿频繁重复添加")
	ErrCreateUponMaxSize = errors.New("一个用户最多创建 5 个群")
)

// ContactService 联系人服务对象
type ContactService struct {
	DB *xorm.Engine
	util.Validator
}

// AddFriend 添加好友
// 操作原则：自己插入一条，另外也插入一条
func (c *ContactService) AddFriend(ownerId, dstId int64) (err error) {
	if ownerId == dstId {
		// 表示加自己了
		return ErrAddOneself
	}
	// 判断是否已经添加了
	if exist, e := c.DB.Exist(&defs.Contact{OwnerId: ownerId, Dstobj: dstId, Cate: defs.ConcatCateUser}); e != nil {
		return e
	} else if exist {
		return ErrContactExist
	}

	// 开启事务，进行添加，如果有一条添加失败，那么另外一条也自动失败
	sess := c.DB.NewSession()
	_ = sess.Begin() // 开启事务

	n := time.Now()
	_, err = sess.Insert(&defs.Contact{OwnerId: ownerId, Dstobj: dstId, Cate: defs.ConcatCateUser, Createat: n},
		&defs.Contact{OwnerId: dstId, Dstobj: ownerId, Cate: defs.ConcatCateUser, Createat: n})
	if err != nil {
		_ = sess.Rollback()
	}
	_ = sess.Commit()
	return err
}

// LoadFriends 加载数据集
func (c *ContactService) LoadFriends(oid int64) ([]*defs.User, error) {
	var fs []*defs.User

	err := c.DB.
		Cols("u.id, u.nickname, u.avatar").
		Table("user u").
		Join("", "contact c", "c.dstobj = u.id").Where("c.owner_id = ? AND c.cate = ?", oid, defs.ConcatCateUser).
		Find(&fs)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return fs, nil
}

// FindAllCommunitiesId 根据 userId 找到所有的群
func (c *ContactService) FindAllCommunitiesId(id int64) []int64 {
	var r = make([]int64, 0)
	var tmp = make([]defs.Contact, 0)

	err := c.DB.Where("owner_id = ? AND cate = ?", id, defs.ConcatCateCommunity).Find(&tmp)
	if err != nil {
		log.Println(err)
		return r
	}

	// 进行返回
	for _, t := range tmp {
		r = append(r, t.Dstobj)
	}
	return r
}

// FindAllCommunities 找到所有的
func (c *ContactService) FindAllCommunities(oid int64) ([]defs.Community, error) {
	var fs []defs.Community

	err := c.DB.
		Cols("co.*").
		Table("contact c").
		Join("", "community co", "c.dstobj = co.id").Where("c.owner_id = ? AND c.cate = ?", oid, defs.ConcatCateCommunity).
		Find(&fs)
	if err != nil {
		log.Println(err)
	}

	return fs, err
}

func (c *ContactService) JoinCommunity(id int64, dstobj int64) error {
	// 判断是否已经添加过了
	var cond = defs.Contact{
		OwnerId: id,
		Dstobj:  dstobj,
		Cate:    defs.ConcatCateCommunity,
	}
	exist, err := c.DB.Exist(&cond)
	if err != nil {
		return err
	}
	if exist {
		return ErrContactExist
	}
	cond.Createat = time.Now()
	_, err = c.DB.InsertOne(&cond)
	if err != nil {
		return err
	}
	return nil
}

func (c *ContactService) AddCommunity(t defs.Community) (*defs.Community, error) {
	var cond = defs.Community{
		OwnerId: t.OwnerId,
	}
	num, err := c.DB.Count(&cond)
	if err != nil {
		return nil, err
	}
	if num > 5 {
		return nil, ErrCreateUponMaxSize
	}
	// 由于是创建群，所以需要往两个表添加数据
	sess := c.DB.NewSession()
	_ = sess.Begin()
	cond.Createat = time.Now()
	cond.Name = t.Name
	cond.Icon = t.Icon
	cond.Memo = t.Memo
	_, err = sess.InsertOne(&cond)
	if err != nil {
		_ = sess.Rollback()
		return nil, err
	}
	_, err = sess.Insert(&defs.Contact{
		OwnerId:  cond.OwnerId,
		Dstobj:   cond.Id,
		Cate:     defs.ConcatCateCommunity,
		Createat: time.Now(),
	})
	if err != nil {
		_ = sess.Rollback()
		return nil, err
	}

	_ = sess.Commit()
	return &cond, nil
}
