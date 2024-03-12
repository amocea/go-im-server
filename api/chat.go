package api

import (
	"encoding/json"
	"github.com/amocea/go-im-chat/config"
	"github.com/amocea/go-im-chat/defs"
	"github.com/amocea/go-im-chat/service"
	"github.com/amocea/go-im-chat/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"log"
	"net/http"
	"sync"
)

const (
	EvictAll = iota // 销毁所有的 websocket 连接
	EvictOne        // 销毁单个连接
)

// EvictionListener 淘汰监听器 监听哪些被销毁的连接
type EvictionListener struct {
	Queue     chan []int64  // 使用数组作为参数，以确定其知接受两个数
	Interrupt chan struct{} // 中断信号
}

// InitialAndListen 开启监听
func (e *EvictionListener) InitialAndListen() chan struct{} {
	e.Queue = make(chan []int64, 10)
	e.Interrupt = make(chan struct{})

	go func() {
	EXIT:
		for {
			select {
			case f := <-e.Queue:
				// 说明有需要淘汰的长连接
				switch f[0] {
				case EvictAll:
					// 淘汰所有
					log.Println("EvictListen: [ALL] 销毁所有的 websocket 长连接")
					rw.Lock() // 进行清理的时候不允许所有的读取请求
					for k, n := range clientMap {
						_ = n.Conn.Close() // 执行关闭行为
						close(n.DataQueue) // 执行通道关闭
						delete(clientMap, k)
					}
					rw.Unlock()
				case EvictOne:
					// 淘汰他了
					uid := f[1]
					if n, ok := clientMap[uid]; ok {
						rw.Lock()
						delete(clientMap, uid)
						// 表示存在
						rw.Unlock()
						log.Printf("EvictListen: [One] flag: [%s] --> [uid: %d] \n", n.Id, uid)
						_ = n.Conn.Close()
						close(n.DataQueue)
					}
				}
			case <-e.Interrupt:
				// 表示被中断了
				break EXIT
			}
		}
		log.Println("监听者程序关闭...")
	}()

	return e.Interrupt
}

// GetEvictInterrupt 获取监听者中断通道
func GetEvictInterrupt() chan struct{} {
	return el.Interrupt
}

type Node struct {
	Id        string          // 节点的长标识
	Conn      *websocket.Conn // 长连接
	DataQueue chan []byte     // 进行数据传输的通道 为了使消息具备顺序性，避免消息的无序传输
	GroupSets set.Interface
	UserId    int64
}

var (
	css       *service.ChatService
	clientMap = make(map[int64]*Node)
	rw        sync.RWMutex
	el        *EvictionListener
)

func init() {
	css = new(service.ChatService)
	css.DB = config.DB() // 执行 db 的赋值

	el = new(EvictionListener)
	el.InitialAndListen() // 执行操作
}

var RegisterChatHandlers = func() {
	http.HandleFunc("/chat", Chat)
}

// Chat 进行聊天的 ws://127.0.0.1/chat?id=1&token=xxxx
func Chat(w http.ResponseWriter, r *http.Request) {
	c := new(defs.ChatDto)
	if err := util.BindUrlArg(r, c); err != nil {
		css.ErrOutput(w, err)
	}
	// 检查用户的 token 是否合法
	iv := checkToken(c.Id, c.Token)
	// 将是否合法传回 前端
	// 将普通的 http 请求升级为是 ws 请求
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return iv
		},
	}).Upgrade(w, r, nil)
	if err != nil {
		// 表示无法创建长连接
		log.Println(err)
		return
	}

	ui := uuid.New()
	// 新建节点
	node := &Node{
		Id:        ui.String(),
		Conn:      conn,
		DataQueue: make(chan []byte, 50), // 允许一次性传输 50 条消息，当网络不通畅时，可能会在本地存储较久时间的消息
		GroupSets: set.New(set.ThreadSafe),
		UserId:    c.Id,
	}
	comIds := cs.FindAllCommunitiesId(c.Id)
	// 刷新 conn.GroupSets
	for _, v := range comIds {
		node.GroupSets.Add(v)
	}
	// 注意，这里有并发安全问题，锁起来，以确保其他协程也可以更新 最新的clientMap，
	rw.Lock()
	clientMap[c.Id] = node
	rw.Unlock()

	// 开启协程完成发送逻辑
	go sendproc(node)
	// 开启协程完成接受逻辑
	go recvproc(node)

	// 输出欢迎消息
	sendMsg(c.Id, []byte("hello world"))
}

func recvproc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			// 一旦无法从长连接当中读取消息，那么就表示 长连接 失效了，进行销毁即可
			log.Println(err)
			el.Queue <- EvictAdvice(EvictOne, node.UserId)
			return
		}
		dispatch(data)
		// 对 data 数据做进一步的处理
		log.Printf("rec: %s\n", data)
	}
}

// 调度分发消息
func dispatch(data []byte) {
	msg := new(defs.Message)
	if err := json.Unmarshal(data, msg); err != nil {
		log.Println(err)
		return
	}
	// 根据 cmd 对其进行逻辑处理
	switch msg.Cmd {
	case defs.CMD_SINGLE_MSG:
		// 表示是单发消息
		sendMsg(msg.Dstid, data)
	case defs.CMD_ROOM_MSG:
		// 表示是群发消息
		// 判断当前处于连接的用户是否带有群聊 id
		for _, n := range clientMap {
			if n.GroupSets.Has(msg.Dstid) {
				// 即其他用户是否处于这一个群 如果是
				n.DataQueue <- data
			}
		}
	case defs.CMD_HEART:
		// 心跳消息 一般啥都不会做
	}
}

func EvictAdvice(flag, uid int64) []int64 {
	return []int64{flag, uid}
}

func sendproc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				// 一旦发现无法写入数据，进行协程终止，然后需要对数据进行销毁
				log.Println(err)
				el.Queue <- EvictAdvice(EvictOne, node.UserId)
				return // 将协程终止，即该次失效
			}
		}
	}
}

// 用以发送表情包和文字信息
func sendMsg(userid int64, msg []byte) {
	rw.RLock()
	defer rw.RUnlock()
	if node, ok := clientMap[userid]; ok {
		node.DataQueue <- msg
	}
}

func checkToken(id int64, token string) bool {
	// 从数据库当中查询并比对
	return css.CheckToken(id, token)
}
