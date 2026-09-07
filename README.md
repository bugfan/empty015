## ang admin

### vue
pnpm install
pnpm dev
pnpm build

### admin
go run main.go

## 项目介绍
ang-admin是ang全协议代理网关的管理端，包含管理后台接口和管理端web页面，主要是用来管理用户配置数据，代理数据，与ang引擎交互，发送配置到ang，ang基于给定配置运行
采用restful风格
基于github.com/bugfan/rest路由框架
基于xorm数据库框架
暂时基于sqlite3本地数据库
管理端web页面要始终支持中英双语
管理端web页面要始终支持手机/PC双端响应式布局
管理端web原地址仓库 https://github.com/pure-admin/vue-pure-admin 

## ai agent注意事项
改完代码测试运行之后一定要自己停到程序，不要影响我启动测试代码

## todo
- [x] 按增量下发配置
- [x] tunnel client链接时候暂时没有限制，谁都能链接，这个应该在tunnel server添加/编辑时候搞一个配置项，控制是否允许未注册的tunnel client token能链接上来
tunnel菜单里的server/client改成tunnel和node是否更好
- [x] server.json tunnel部分，不再需要指定sni，直接使用内置证书与默认 DefaultSNI ("ang")，配置轻量化
- [x] 需要把tunnel/client的配置项也下发到ang，后续要对链接上来的client做校验，因为已经用了统一的sni和cert.校验包含配置好的client
tunnel里配置一些js代码动态的根据token到认证源认证
准备支持各种认证方式，认证源
    - sso
    - cas
    - ldap
    - http basicauth
    - radius
    - oauth
    - 自定义js代码，能读取用户传入的参数，以及到指定的上游认证服务器认证校验
    - token认证
第三方的这种认证的用户是否倒入自己的系统内
本地的用户token/本地账户存本地json还是其他的服务器，还是仅仅在admin如果在admin就得支持反向查询功能，因为admin有可能在内网

日志
规则
认证

方向三：统一认证门户的“应用大厅 /          
  工作台导航页”,这个应该跟portal_action是一个页面或逻辑，登录后访问该域名应该显示导航页
方向四：在线会话与安全审计（会话管理 /     
  一键踢人），这个坐在用户管理下搞一个字菜单“在线用户”
webvpn里面的规则是直接封装到webvpn配置项里，跟http隔离开，尽量
webvpn里选择的base泛域代理http,如果选中了该http站点，那么该站点应该是不需要配置原本http里面的上游部分了
