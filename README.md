# portv2ssh
portv2ssh是一款端口激活结合端口转发的防护性工具，旨在构建对暴露于互联网的ssh端口（22端口）的防护性保护
#### [主要功能]
通过创建config.yml配置文件，其中的DragonBallPort字段设置为数组结构，从上到下依次为需要激活的端口（具有顺序性），只有依次激活了这些端口后，才可访问SshAgentPort字段对应的端口，最终该端口转发SshIP:SshPort的内容（为达到目的型，SshPort端口的源访问白名单需设置为本机访问）
#### [示例]
比如默认config.yml中DragonBallPort字段的元素为3333和9999，因此为访问目标主机的SshAgentPort端口，
首先需要在命令访问3333端口：

`telnet ip 3333`

接着访问9999端口：
`telnet ip 9999`

最终ssh连接7000端口即可：
`ssh -p 7000 username@ip`