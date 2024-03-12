# go-im-server
即时im，允许在线进行聊天的应用程序，实现了较好的并发的

使用库
1. net/http
2. gorilla/websocket
3. xorm.io/xorm
4. ...

实现特点
1. 在原有的 net/http 实现了关于请求方法的校验，确保符合 restful 风格的请求
2. 定义了一系列较为好用的组件，如参数校验组件 Validator，用于确认在 http.Request 下是否有指定的参数，如果没有，返回指定的参数
3. 提供了请求参数直接绑定结构体对象的绑定方法，支持 Content-Type = application/x-www-form-urlencoded 、multipart/form-data、 application/json 的MINE
4. 采用了 websocket 作为即时的底层依赖，为了实现对 websocket 的管理，自定义内置了一个关于 websocket （当不可用时，会自动被监听器监听到，进而直接销毁）
5. 采用了 MVC 风格的开发方式，提供了较多的 util 方法。