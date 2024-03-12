package api

import (
	"errors"
	"github.com/amocea/go-im-chat/config"
	"github.com/amocea/go-im-chat/defs"
	"github.com/amocea/go-im-chat/service"
	"github.com/amocea/go-im-chat/util"
	"net/http"
)

var cs *service.ContactService

/*
添加联系人
*/
func init() {
	cs = new(service.ContactService)
	cs.DB = config.DB() // 执行 db 的赋值
}

var RegisterContactHandlers = func() {
	r := Group("/contact")
	{
		r.Register(http.MethodPost, "/addfriend", AddFriend)
		r.Register(http.MethodPost, "/loadfriend", LoadFriends)
		r.Register(http.MethodPost, "/joincommunity", JoinCommunity)
		r.Register(http.MethodPost, "/createcommunity", AddCommunity)
		r.Register(http.MethodPost, "/community", FindAllCommunities)
	}
}

// AddFriend 添加好友 该操作发送的是 异步请求，所以发送的是 JSON 数据
func AddFriend(w http.ResponseWriter, r *http.Request) {
	var c defs.Contact
	// 参数校验是否通过
	if err := cs.Validate(w, r, &c, "ownerid", "dstobj"); err != nil {
		cs.ErrOutput(w, err)
		return
	}
	_ = util.Bind(r, &c)
	if c.OwnerId == 0 {
		defs.RespJSON(w, http.StatusBadRequest, -1, "非法参数", nil)
		return
	}
	// 通过了之后 进行判断行为
	err := cs.AddFriend(c.OwnerId, c.Dstobj)
	if err != nil {
		if errors.Is(err, service.ErrAddOneself) {
			defs.RespJSON(w, http.StatusBadRequest, -1, err.Error(), nil)
		} else if errors.Is(err, service.ErrContactExist) {
			defs.RespJSON(w, http.StatusBadRequest, -1, err.Error(), nil)
		} else {
			defs.RespJSON(w, http.StatusInternalServerError, -1, "内部错误", nil)
		}
		return
	}
	defs.RespJSON(w, http.StatusCreated, 0, "已添加", nil)
}

// LoadFriends 加载朋友数据
func LoadFriends(w http.ResponseWriter, r *http.Request) {
	var c defs.Contact
	if err := util.Bind(r, &c); err != nil {
		cs.ErrOutput(w, err)
		return
	}
	if c.OwnerId == 0 {
		defs.RespJSON(w, http.StatusBadRequest, -1, "非法参数", nil)
		return
	}

	m, err := cs.LoadFriends(c.OwnerId)
	if err != nil {
		defs.RespJSON(w, http.StatusInternalServerError, -1, "查询失败", nil)
		return
	}

	defs.RespPageJSON(w, http.StatusOK, 0, len(m), m)
}

// JoinCommunity 加入群当中 ownerId
func JoinCommunity(w http.ResponseWriter, r *http.Request) {
	var c defs.Contact
	if err := util.Bind(r, &c); err != nil {
		cs.ErrOutput(w, err)
		return
	}
	// 通过了之后 进行判断行为
	err := cs.JoinCommunity(c.OwnerId, c.Dstobj)
	if err != nil {
		if errors.Is(err, service.ErrContactExist) {
			defs.RespJSON(w, http.StatusBadRequest, -1, "请勿重复频繁添加", nil)
		} else {
			defs.RespJSON(w, http.StatusInternalServerError, -1, "内部错误", nil)
		}
		return
	}
	defs.RespJSON(w, http.StatusCreated, 0, "已添加", nil)
}

// AddCommunity 添加新的群
func AddCommunity(w http.ResponseWriter, r *http.Request) {
	var cond defs.Community
	if err := util.Bind(r, &cond); err != nil {
		cs.ErrOutput(w, err)
		return
	}
	// 通过了之后 进行判断行为
	com, err := cs.AddCommunity(cond)
	if err != nil {
		if errors.Is(err, service.ErrCreateUponMaxSize) {
			defs.RespJSON(w, http.StatusBadRequest, -1, err.Error(), nil)
		} else {
			defs.RespJSON(w, http.StatusInternalServerError, -1, "内部错误", nil)
		}
		return
	}
	defs.RespJSON(w, http.StatusCreated, 0, "创建成功", com)
}

func FindAllCommunities(w http.ResponseWriter, r *http.Request) {
	var cond defs.Contact
	if err := util.Bind(r, &cond); err != nil {
		cs.ErrOutput(w, err)
		return
	}
	// 通过了之后 进行判断行为
	coms, err := cs.FindAllCommunities(cond.OwnerId)
	if err != nil {
		defs.RespJSON(w, http.StatusInternalServerError, -1, "内部错误", nil)
		return
	}
	defs.RespPageJSON(w, http.StatusOK, 0, len(coms), coms)
}
