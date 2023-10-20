# 一个（最初计划自用的）Webhook处理程序

目的：接受来自直播录制程序的Webhook请求，然后给用户设置的目标（现在是Bark和企业微信）推送消息。

注意：我的开发环境为 Windows 10 x64，尽管Golang理论上是跨平台的，但是不同平台的表现终究不一样。所以不保证到你的设备上能正常跑。

## 用法：
1. 在[release页面](https://github.com/Janet-Baker/webhookGo/releases)，
找当前系统环境可以运行的软件包(大多数都是 webhookGo_windows_amd64.exe
~~_用其它系统的相信你已有自理能力了所以自己选包吧_~~)，下载下来。
2. 运行一次，如果看到文件夹里多出来一个`secrets.yml`，就说明下载的软件包是可以用的。
3. 在`secrets.yml`里，根据你的推送需求
(目前写了 企业微信应用消息 和 iOS Bark这两种)，填写相关的信息。
4. 在相关的可以发送Webhook的应用程序中，填写Webhook地址。
建议配置完成之后再进行这一步。
   - 在[mikufans录播姬](rec.danmuji.org/)
   的设置页面 Webhook V2 中，填写`http://127.0.0.1:14000/bililiverecoder`
   - 在[blrec](https://github.com/acgnhiki/blrec/)的设置页面最下方Webhooks中，添加服务器
   `http://127.0.0.1:14000/blrec`
   - 在[DDTV](https://ddtv.pro/)的配置文件`DDTV_Config.ini`中，找到`WebHookUrl=`，填写`http://127.0.0.1:14000/ddtv`
5. 如果想要自己定制，其实不难……

## 自定义读取配置文件的位置
> webhookGo.exe -s secrets.yml

## 示例配置文件
### 只需要Bark推送的
只留下你的推送目标就可以了。
```yaml
Bark:
  - url: "https://api.day.app/"
    secrets: "ABcDeFg1hIjkLmNOPQrstu"
```

### 只需要企业微信应用消息的
同样的，只留下企业微信的目标就可以了。
```yaml
WeWorkApp:
  - corpId: "ww123456789a01b2c3"
    appSecret: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"
    agentId: "1000002"
```

### 有好多好多要推的
加啊，都可以加。有几个就写几个。
```yaml
Bark:
  - url: "https://api.day.app/"
    secrets: "asdffghjklQWERTYUIIOPz"
  - url: "https://api.day.app/"
    secrets: "ABcDeFg1hIjkLmNOPQrstu"
  - url: "https://api.day.app/"
    secrets: "xCVBNMASDFGHJKLERTYUIO"
WeWorkApp:
  - corpId: "wx54654654adsadad1"
    appSecret: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"
    agentId: "1000002"
  - corpId: "wz0123456789abcfef"
    appSecret: "acbdef0123456789876543210fedcbabcdef0123409"
    agentId: "1000003"
  - corpId: "wy123456789a01b2c3"
    appSecret: "111111111111111111111222222222222222ABCDEFG"
    agentId: "1000004"
```

## 还有一件事
> 哎呀你都做了读取密钥文件了，能不能把监听端口和地址也做进去啊

答：最近很忙，为了将私有仓库转为公开仓库，做了不小的改造。
读取yaml配置文件的功能还是一下午抄完的……要不是某些原因我应该不会放出来吧。