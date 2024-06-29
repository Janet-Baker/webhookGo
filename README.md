# 一个Webhook处理程序

目的：接受来自直播录制程序的Webhook请求，然后给用户设置的目标（现在是Bark和企业微信）推送消息。

## 用法：
1. 在[release页面](https://github.com/Janet-Baker/webhookGo/releases)，
找当前系统环境可以运行的软件包，下载下来。
2. 运行一次，如果看到文件夹里多出来一个`config.yml`，就说明下载的软件包是可以用的。
3. 在`config.yml`里，根据你的监听和推送需求
(目前写了 企业微信应用消息 和 iOS Bark 这两种)，填写相关的信息。
4. 启动程序。
5. 在相关的可以发送Webhook的应用程序中，填写Webhook地址。
建议配置完成之后再进行这一步。
   - 在 [mikufans录播姬](https://rec.danmuji.org/)
   的设置页面 **Webhook V2** 中，填写`http://127.0.0.1:14000/bililiverecorder`
   - 在 [blrec](https://github.com/acgnhiki/blrec/) 网页控制台的设置页面最下方**Webhooks**中，添加服务器
   `http://127.0.0.1:14000/blrec`
   - 在 [DDTV3](https://ddtv.pro/) 的配置文件`DDTV_Config.ini`中，找到`WebHookUrl=`，填写`http://127.0.0.1:14000/ddtv3`
   - 在 [DDTV5](https://ddtv.pro/) 的Desktop版的`设置`-`DDTV基础设置`-`Webhook`中，点击右边箭头展开，在文本框中填写`http://127.0.0.1:14000/ddtv5`

## 自定义读取配置文件的位置
> webhookGo.exe -c config.yml

## 配置文件

根据你的需要，增删模块即可。

### 侦听地址和端口

address 监听地址，默认`127.0.0.1:14000`

```yaml
address: '127.0.0.1:14000'
```

### 允许连接至Bilibili

为了能够在部分没有传递主播头像的程序中取得主播头像，以及在直播结束时检查直播间封禁状态，需要连接至哔哩哔哩服务器。

```yaml
# contact_bilibili 允许访问Bilibili服务器，获取主播头像，下播时检查直播间封禁状态。
contact_bilibili: true
```

### 需要Bark推送的

```yaml
Bark:
  - url: "https://api.day.app/"
    secrets: "asdffghjklQWERTYUIIOPz"
  - url: "https://api.day.app/"
    secrets: "ABcDeFg1hIjkLmNOPQrstu"
  - url: "https://api.day.app/"
    secrets: "xCVBNMASDFGHJKLERTYUIO"
```

### 需要企业微信应用消息的

```yaml
WeWorkApp:
  - corpId: "ww123456789a01b2c3"
    appSecret: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"
    agentId: "1000002"
  - corpId: "wz0123456789abcfef"
    appSecret: "acbdef0123456789876543210fedcbabcdef0123409"
    agentId: "1000003"
  - corpId: "wy123456789a01b2c3"
    appSecret: "111111111111111111111222222222222222ABCDEFG"
    agentId: "1000004"
```

### 提供webhook响应
```yaml
Receivers:
   - type: "BililiveRecorder"
      # enable 是否启用该服务
     enable: true
      # path 该服务的访问路径
     path: '/bililiverecorder'
      # events 该服务监听的事件，事件种类见 https://rec.danmuji.org/reference/webhook/#webhook-v2
     events:
        SessionStarted:
           # care 是否在控制台提示收到了该事件
           care: false
           # notify 是否推送该事件
           notify: false
           # have_command 是否执行命令
           have_command: false
           # exec_command 执行的命令
           exec_command: ""
   - type: "Blrec"
     enable: true
     path: '/blrec'
      # 事件种类及可提取的信息见 https://github.com/acgnhiki/blrec/wiki/Webhook
     events:
        LiveBeganEvent:
           care: true
           notify: true
           have_command: false
           exec_command: ""
#...
```