# Windows Security Events for SIEM Collection

Windows 安全事件日志是 SIEM 系统的核心数据源。本文档详细列出需要采集的 Windows 事件，包括事件 ID、说明、采集原因，以及 Microsoft 官方文档出处。

> 文档版本: 2026-03-03
> 主要来源: Microsoft Learn - Security Auditing

---

## 目录

1. [核心采集要求](#核心采集要求)
2. [高危事件 (Critical)](#高危事件-critical)
3. [账户管理事件](#账户管理事件)
4. [登录认证事件](#登录认证事件)
5. [进程创建/终止](#进程创建终止)
6. [特殊权限事件](#特殊权限事件)
7. [Kerberos 认证事件](#kerberos-认证事件)
8. [计划任务事件](#计划任务事件)
9. [审计策略变更](#审计策略变更)
10. [Sysmon 推荐采集](#sysmon-推荐采集)
11. [其他推荐事件](#其他推荐事件)
12. [采集清单汇总](#采集清单汇总)
13. [官方文档链接](#官方文档链接)

---

## 核心采集要求

### 日志来源

| 日志名称 | 路径 | 说明 |
|----------|------|------|
| Security | `Windows Logs/Security` | 安全审计主日志 |
| System | `Windows Logs/System` | 系统组件事件 |
| Application | `Windows Logs/Application` | 应用程序事件 |
| PowerShell | `Applications and Services Logs/Microsoft/Windows/PowerShell/Operational` | PowerShell 脚本执行 |
| Sysmon | `Applications and Services Logs/Microsoft/Windows/Sysmon/Operational` | Sysmon 扩展监控 |
| Windows Defender | `Applications and Services Logs/Microsoft/Windows/Windows Defender/Operational` | 防病毒事件 |

### 启用审计策略

使用 `auditpol` 命令启用所有安全审计类别：

```cmd
# 启用所有账户登录审计
auditpol /set /subcategory:"Credential Validation" /success:enable /failure:enable
auditpol /set /subcategory:"Kerberos Authentication Service" /success:enable /failure:enable
auditpol /set /subcategory:"Kerberos Service Ticket Operations" /success:enable /failure:enable

# 启用账户管理审计
auditpol /set /subcategory:"User Account Management" /success:enable /failure:enable
auditpol /set /subcategory:"Computer Account Management" /success:enable /failure:enable
auditpol /set /subcategory:"Security Group Management" /success:enable /failure:enable
auditpol /set /subcategory:"Distribution Group Management" /success:enable /failure:enable

# 启用登录/注销审计
auditpol /set /subcategory:"Logon" /success:enable /failure:enable
auditpol /set /subcategory:"Logoff" /success:enable /failure:enable
auditpol /set /subcategory:"Special Logon" /success:enable /failure:enable

# 启用进程审计
auditpol /set /subcategory:"Process Creation" /success:enable /failure:enable
auditpol /set /subcategory:"Process Termination" /success:enable

# 启用策略变更审计
auditpol /set /subcategory:"Audit Policy Change" /success:enable /failure:enable
auditpol /set /subcategory:"Authentication Policy Change" /success:enable /failure:enable
auditpol /set /subcategory:"Authorization Policy Change" /success:enable /failure:enable

# 启用特权使用审计
auditpol /set /subcategory:"Sensitive Privilege Use" /success:enable /failure:enable
```

---

## 高危事件 (Critical)

这些事件表明可能的攻击行为，需要立即关注。

### 1102 - 审计日志被清除

| 属性 | 说明 |
|------|------|
| **事件ID** | 1102 |
| **子类** | Other Events |
| **日志通道** | Security |
| **严重性** | 🔴 Critical |
| **采集原因** | 攻击者清除审计日志以掩盖行踪，这是典型的攻击链末端行为 |

**官方说明:**

> This event generates every time Windows Security audit log was cleared.
> Typically you should not see this event. There is no need to manually clear the Security event log in most cases. We recommend monitoring this event and investigating why this action was performed.

**来源:** [Microsoft Learn - Event 1102](https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-1102)

---

### 4719 - 审计策略被更改

| 属性 | 说明 |
|------|------|
| **事件ID** | 4719 |
| **子类** | Audit Policy Change |
| **日志通道** | Security |
| **严重性** | 🔴 Critical |
| **采集原因** | 攻击者可能禁用某些审计策略以减少检测机会 |

**官方说明:**

> This event generates when the computer's audit policy changes.
> Monitor for all events of this type, especially on high value assets or computers, because any change in local audit policy should be planned. If this action was not planned, investigate the reason for the change.

**来源:** [Microsoft Learn - Event 4719](https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4719)

---

### 4765/4766 - SID History 操作

| 属性 | 说明 |
|------|------|
| **事件ID** | 4765, 4766 |
| **子类** | User Account Management |
| **日志通道** | Security |
| **严重性** | 🔴 Critical |
| **采集原因** | SID History 攻击，允许攻击者获取额外权限或跨域提权 |

**采集配置:** 启用 "User Account Management" 审计

---

### 4794 - 尝试设置 DSRM

| 属性 | 说明 |
|------|------|
| **事件ID** | 4794 |
| **子类** | User Account Management |
| **日志通道** | Security |
| **严重性** | 🔴 Critical |
| **采集原因** | 尝试设置目录服务恢复模式密码，可能是攻击前期准备 |

---

### 4649 - 检测到重放攻击

| 属性 | 说明 |
|------|------|
| **事件ID** | 4649 |
| **子类** | Credential Validation |
| **日志通道** | Security |
| **严重性** | 🔴 High |
| **采集原因** | 可能存在凭证重放攻击 |

---

## 账户管理事件

### 4720 - 用户账户创建

| 属性 | 说明 |
|------|------|
| **事件ID** | 4720 |
| **子类** | Audit User Account Management |
| **日志通道** | Security |
| **严重性** | 🟠 High |
| **采集原因** | 检测恶意账户创建或持久化 |

**官方说明:**

> This event generates every time a new user object is created.
> This event generates on domain controllers, member servers, and workstations.

**建议监控:**

- 监控非预期的账户创建
- 监控 SAM Account Name 为空的情况
- 监控 Password Last Set 为 "never"（永不过期密码）

**来源:** [Microsoft Learn - Event 4720](https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4720)

---

### 4722 - 用户账户启用

| 属性 | 说明 |
|------|------|
| **事件ID** | 4722 |
| **子类** | Audit User Account Management |
| **日志通道** | Security |
| **严重性** | 🟠 High |
| **采集原因** | 之前禁用的账户被启用，可能用于持久化 |

---

### 4723/4724 - 密码修改尝试

| 属性 | 说明 |
|------|------|
| **事件ID** | 4723, 4724 |
| **子类** | Audit User Account Management |
| **日志通道** | Security |
| **严重性** | 🟡 Medium |
| **采集原因** | 密码修改尝试，可能是凭证重置攻击 |

**官方说明:**

> 4723: An attempt was made to change an account's password
> 4724: An attempt was made to reset an account's password

---

### 4725 - 用户账户禁用

| 属性 | 说明 |
|------|------|
| **事件ID** | 4725 |
| **子类** | Audit User Account Management |
| **日志通道** | Security |
| **严重性** | 🟡 Medium |
| **采集原因** | 账户被禁用，可能导致拒绝服务或攻击后清理 |

---

### 4726 - 用户账户删除

| 属性 | 说明 |
|------|------|
| **事件ID** | 4726 |
| **子类** | Audit User Account Management |
| **日志通道** | Security |
| **严重性** | 🟠 High |
| **采集原因** | 账户删除，可能是攻击者清理痕迹 |

---

### 4732 - 成员添加到安全组

| 属性 | 说明 |
|------|------|
| **事件ID** | 4732 |
| **子类** | Audit Security Group Management |
| **日志通道** | Security |
| **严重性** | 🟠 High |
| **采集原因** | 用户被添加到特权组（如 Administrators, Domain Admins） |

---

### 4733 - 成员从安全组移除

| 属性 | 说明 |
|------|------|
| **事件ID** | 4733 |
| **子类** | Audit Security Group Management |
| **日志通道** | Security |
| **严重性** | 🟡 Medium |
| **采集原因** | 从特权组移除成员 |

---

### 4735/4737/4739 - 安全组更改

| 属性 | 说明 |
|------|------|
| **事件ID** | 4735, 4737, 4739 |
| **子类** | Audit Security Group Management |
| **日志通道** | Security |
| **严重性** | 🟡 Medium |
| **采集原因** | 安全组属性被修改 |

---

### 4740 - 账户锁定

| 属性 | 说明 |
|------|------|
| **事件ID** | 4740 |
| **子类** | Audit Account Lockout |
| **日志通道** | Security |
| **严重性** | 🟡 Medium |
| **采集原因** | 账户被锁定，可能是暴力破解攻击导致 |

---

## 登录认证事件

### 4624 - 登录成功

| 属性 | 说明 |
|------|------|
| **事件ID** | 4624 |
| **子类** | Audit Logon |
| **日志通道** | Security |
| **严重性** | 🟢 Low (需结合分析) |
| **采集原因** | 记录所有成功登录，用于溯源和异常检测 |

**官方说明:**

> This event generates when a logon session is created (on destination machine). It generates on the computer that was accessed, where the session was created.

**Logon Type 说明:**

| 类型 | 值 | 含义 | 安全关注 |
|------|-----|------|----------|
| Interactive | 2 | 本地登录 | 检查非工作时间 |
| Network | 3 | 网络共享访问 | 检查异常来源 |
| Batch | 4 | 计划任务 | 检查服务账户使用 |
| Service | 5 | 服务启动 | 检查服务账户权限 |
| Unlock | 7 | 解锁工作站 | 检查异常解锁 |
| RemoteInteractive | 10 | RDP 登录 | 🔴 高风险！ |
| CachedInteractive | 11 | 缓存登录 | 检查本地凭证使用 |

**监控建议:**

- 监控高价值账户的登录
- 监控非工作时间的登录
- 监控 RDP 登录 (Logon Type = 10)
- 监控管理员账户的登录
- 监控 Elevated Token = Yes 的登录

**Source:** [Microsoft Learn - Event 4624](https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4624)

---

### 4625 - 登录失败

| 属性 | 说明 |
|------|------|
| **事件ID** | 4625 |
| **子类** | Audit Logon, Audit Account Lockout |
| **日志通道** | Security |
| **严重性** | 🟡 Medium (需结合分析) |
| **采集原因** | 检测暴力破解、凭证猜测攻击 |

**官方说明:**

> This event is logged for any logon failure. It generates on the computer where logon attempt was made.

**失败状态码 (Status/Sub Status):**

| 状态码 | 含义 | 安全意义 |
|--------|------|----------|
| 0xC0000064 | 用户名错误 | 可能是用户枚举攻击 |
| 0xC000006A | 密码错误 | 暴力破解尝试 |
| 0xC000006D | 凭据错误 | 认证失败 |
| 0xC000006F | 登录时间限制 | 非工作时间尝试 |
| 0xC0000070 | 工作站限制 | 从未授权工作站登录 |
| 0xC0000072 | 账户被管理员禁用 | 尝试使用禁用账户 |
| 0xC0000193 | 账户已过期 | 尝试使用过期账户 |
| 0xC000015B | 登录权限不足 | 尝试获取未授权登录类型 |
| 0xC0000234 | 账户已锁定 | 暴力破解导致锁定 |

**Source:** [Microsoft Learn - Event 4625](https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4625)

---

### 4634 - 账户注销

| 属性 | 说明 |
|------|------|
| **事件ID** | 4634 |
| **子类** | Audit Logoff |
| **日志通道** | Security |
| **严重性** | 🟢 Low |
| **采集原因** | 与登录事件配合计算会话时长 |

---

### 4648 - 使用显式凭据登录

| 属性 | 说明 |
|------|------|
| **事件ID** | 4648 |
| **子类** | Audit Logon |
| **日志通道** | Security |
| **严重性** | 🟡 Medium |
| **采集原因** | 使用 RunAs 或其他方式以不同账户登录执行程序 |

**官方说明:**

> This event is generated when a process attempts to log on an account by explicitly specifying that account's credentials. This most commonly occurs in batch configurations such as scheduled tasks, or when using the RunAs command.

**监控建议:**

- 监控非预期的 RunAs 使用
- 监控从异常进程执行的管理员操作

---

### 4672 - 特殊权限分配

| 属性 | 说明 |
|------|------|
| **事件ID** | 4672 |
| **子类** | Audit Special Logon |
| **日志通道** | Security |
| **严重性** | 🟠 High |
| **采集原因** | 敏感权限分配给登录账户，高权限账户活动的指示器 |

**官方说明:**

> This event generates for new account logons if any of the following sensitive privileges are assigned to the new logon session

**敏感权限列表:**

| 权限名称 | 中文说明 | 风险 |
|----------|----------|------|
| SeTcbPrivilege | 作为操作系统的一部分 | 🔴 极高 |
| SeDebugPrivilege | 调试程序 | 🔴 极高 |
| SeSecurityPrivilege | 管理审计和安全日志 | 🟠 高 |
| SeBackupPrivilege | 备份文件和目录 | 🟠 高 |
| SeRestorePrivilege | 还原文件和目录 | 🟠 高 |
| SeTakeOwnershipPrivilege | 取得文件所有权 | 🟠 高 |
| SeImpersonatePrivilege | 模拟客户端 | 🔴 极高 |
| SeLoadDriverPrivilege | 加载/卸载驱动 | 🟠 高 |
| SeCreateTokenPrivilege | 创建令牌对象 | 🔴 极高 |
| SeEnableDelegationPrivilege | 启用委派 | 🟠 高 |

**Source:** [Microsoft Learn - Event 4672](https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4672)

---

## 进程创建/终止

### 4688 - 新进程创建

| 属性 | 说明 |
|------|------|
| **事件ID** | 4688 |
| **子类** | Audit Process Creation |
| **日志通道** | Security |
| **严重性** | 🟢 Low (需结合分析) |
| **采集原因** | 检测恶意软件执行、横向移动工具 |

**官方说明:**

> This event generates every time a new process starts.

**关键字段:**

- New Process Name - 新进程路径
- Command Line - 完整命令行
- Parent Process Name - 父进程
- Token Elevation Type - UAC 提升状态
- Mandatory Label - 完整性级别

**Token Elevation Type:**

| 值 | 含义 | 说明 |
|----|------|------|
| %%1936 | Full | 无提升 (UAC 禁用或 SYSTEM) |
| %%1937 | Elevated | 已提升 (管理员权限) |
| %%1938 | Limited | 受限权限 |

**Mandatory Label (完整性级别):**

| SID | 值 | 含义 |
|-----|-----|------|
| S-1-16-0 | 0x0000 | Untrusted |
| S-1-16-4096 | 0x1000 | Low |
| S-1-16-8192 | 0x2000 | Medium |
| S-1-16-12288 | 0x3000 | High |
| S-1-16-16384 | 0x4000 | System |

**监控建议:**

- 监控非 System32/Program Files 目录的进程
- 监控已知恶意工具 (mimikatz, cain.exe, psexec.exe 等)
- 监控异常的父进程（如 Word 启动 PowerShell）
- 监控不常见的高权限进程

**Source:** [Microsoft Learn - Event 4688](https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4688)

---

### 4689 - 进程退出

| 属性 | 说明 |
|------|------|
| **事件ID** | 4689 |
| **子类** | Audit Process Termination |
| **日志通道** | Security |
| **严重性** | 🟢 Low |
| **采集原因** | 与 4688 配合分析进程执行时长 |

---

## Kerberos 认证事件

### 4768 - Kerberos TGT 请求

| 属性 | 说明 |
|------|------|
| **事件ID** | 4768 |
| **子类** | Audit Kerberos Authentication Service |
| **日志通道** | Security |
| **严重性** | 🟢 Low (需分析) |
| **采集原因** | 域账户认证请求，检测异常票据请求 |

**监控建议:**

- 监控失败请求 (0x6 - 客户端密钥错误/密码错误)
- 监控异常时间段请求

---

### 4769 - Kerberos 服务票据请求

| 属性 | 说明 |
|------|------|
| **事件ID** | 4769 |
| **子类** | Audit Kerberos Service Ticket Operations |
| **日志通道** | Security |
| **严重性** | 🟢 Low (需分析) |
| **采集原因** | 服务票据请求，检测横向移动 (如金票攻击) |

**异常指标:**

- 请求的服务票据没有对应的服务账户
- 异常的加密类型 (使用 DES 或 RC4 而非 AES)
- 来源 IP 异常

---

### 4771 - Kerberos 预认证失败

| 属性 | 说明 |
|------|------|
| **事件ID** | 4771 |
| **子类** | Audit Kerberos Authentication Service |
| **日志通道** | Security |
| **严重性** | 🟡 Medium |
| **采集原因** | Kerberos 认证失败，可能暴力破解或 Golden Ticket 攻击 |

---

### 4776 - 凭据验证尝试

| 属性 | 说明 |
|------|------|
| **事件ID** | 4776 |
| **子类** | Audit Credential Validation |
| **日志通道** | Security |
| **严重性** | 🟡 Medium |
| **采集原因** | NTLM 认证验证尝试，检测传递哈希攻击 |

---

## 计划任务事件

### 4698-4702 - 计划任务操作

| 事件ID | 操作类型 |
|--------|----------|
| 4698 | 计划任务创建 |
| 4699 | 计划任务删除 |
| 4700 | 计划任务启用 |
| 4701 | 计划任务禁用 |
| 4702 | 计划任务更新 |

**采集原因:** 攻击者常使用计划任务实现持久化 (如 schtasks 创建后门)

---

## 审计策略变更

### 4719 - 审计策略更改 (见上文高危部分)

### 4730 - 安全组删除

| 属性 | 说明 |
|------|------|
| **事件ID** | 4730 |
| **子类** | Audit Security Group Management |
| **日志通道** | Security |
| **严重性** | 🟡 Medium |
| **采集原因** | 安全组被删除，攻击后清理 |

---

## Sysmon 推荐采集

Sysmon (System Monitor) 是 Microsoft Sysinternals 工具，提供更细粒度的系统活动监控。

### 安装

```cmd
# 安装并接受 EULA
sysmon64 -accepteula -i config.xml

# 安装同时读取配置文件
sysmon64 -accepteula -i c:\windows\config.xml
```

### 推荐采集的事件

| Event ID | 名称 | 说明 | 采集原因 |
|----------|------|------|----------|
| **1** | Process Create | 进程创建 | 替代 4688，提供更详细信息 (完整命令行、哈希等) |
| **2** | File Create Time | 文件创建时间修改 | 检测伪装文件时间戳的后门 |
| **3** | Network Connect | 网络连接 | 检测恶意网络通信 |
| **4** | Sysmon State | 服务状态变更 | 检测 Sysmon 被停止 |
| **5** | Process Terminate | 进程终止 | 分析进程生命周期 |
| **6** | Driver Load | 驱动加载 | 检测恶意内核驱动 |
| **7** | Image Load | 镜像加载 | 检测 DLL 注入 |
| **8** | Create Remote Thread | 远程线程创建 | 检测进程注入 (如 Reflective PE Injection) |
| **9** | Raw Access Read | 原始磁盘读取 | 检测数据窃取 (如使用 \\.\ 访问磁盘) |
| **10** | Process Access | 进程访问 | 检测凭证窃取 (如读取 Lsass.exe) |
| **11** | File Create | 文件创建 | 检测恶意文件写入 |
| **12/13/14** | Registry Event | 注册表修改 | 检测注册表持久化 |
| **15** | File Stream Create | 文件流创建 | 检测 Zone.Identifier 流 |
| **17/18** | Pipe Event | 命名管道 | 检测恶意进程间通信 |
| **19/20/21** | WMI Event | WMI 活动 | 检测 WMI 持久化技术 |
| **22** | DNS Event | DNS 查询 | 🔴 **强烈推荐** - 检测 DNS 隧道/恶意域名 |
| **23** | File Delete | 文件删除 | 检测文件清理/攻击痕迹 |
| **25** | Process Tampering | 进程篡改 | 检测 Process Hollowing/Herpaderp |

**Source:** [Microsoft Sysinternals - Sysmon](https://learn.microsoft.com/en-us/sysinternals/downloads/sysmon)

---

## 其他推荐事件

### Windows Defender (需要 Windows 10+)

| 事件ID | 说明 |
|--------|------|
| 1116 | 检测到恶意软件 |
| 1117 | 恶意软件操作被阻止 |
| 1118 | 恶意软件操作被部分阻止 |
| 1119 | 恶意软件操作失败但需要操作 |
| 1006-1009 | 恶意软件扫描事件 |
| 1015 | 行为检测事件 |

### PowerShell 审计

| 事件ID | 说明 |
|--------|------|
| 4103 | 模块日志记录 |
| 4104 | 脚本块日志 (需启用) |
| 4105 | 开始使用 Runspace |
| 4106 | 关闭 Runspace |

### SMB 审计 (详细文件共享)

| 事件ID | 说明 |
|--------|------|
| 5140 | 网络共享访问 |
| 5142 | 网络共享对象创建 |
| 5143 | 网络共享修改 |
| 5144 | 网络共享对象删除 |
| 5145 | 检查文件/对象访问权限 |

---

## 采集清单汇总

### Domain Controller (必须采集)

| 事件ID | 描述 | 原因 |
|--------|------|------|
| 1102 | 审计日志清除 | 🔴 检测攻击痕迹 |
| 4719 | 审计策略更改 | 🔴 检测审计关闭 |
| 4624 | 登录成功 | 活动监控 |
| 4625 | 登录失败 | 暴力破解检测 |
| 4672 | 特殊权限分配 | 权限提升检测 |
| 4720 | 用户创建 | 账户创建检测 |
| 4726 | 用户删除 | 账户删除检测 |
| 4732 | 加入安全组 | 权限提升检测 |
| 4740 | 账户锁定 | 暴力破解检测 |
| 4768 | TGT 请求 | 认证监控 |
| 4769 | 服务票据请求 | 横向移动检测 |
| 4765/4766 | SID History | 提权攻击检测 |
| 4794 | DSRM 设置尝试 | 攻击准备检测 |

### Member Server (必须采集)

| 事件ID | 描述 | 原因 |
|--------|------|------|
| 1102 | 审计日志清除 | 🔴 检测攻击痕迹 |
| 4624 | 登录成功 | 活动监控 |
| 4625 | 登录失败 | 暴力破解检测 |
| 4672 | 特殊权限分配 | 权限提升检测 |
| 4688/4689 | 进程创建/终止 | 恶意软件检测 |
| 4697 | 服务安装 | 恶意服务检测 |
| 4698-4702 | 计划任务 | 持久化检测 |

### Workstation (推荐采集)

| 事件ID | 描述 | 原因 |
|--------|------|------|
| 4624 | 登录成功 | 活动监控 |
| 4625 | 登录失败 | 暴力破解检测 |
| 4672 | 特殊权限分配 | 权限提升检测 |
| 4688 | 进程创建 | 恶意软件检测 |
| 4648 | 显式凭据登录 | RunAs 检测 |
| 4702 | 计划任务更改 | 持久化检测 |

### Sysmon (推荐安装在所有服务器)

| Event ID | 描述 |
|----------|------|
| 1 | 进程创建 |
| 3 | 网络连接 |
| 6 | 驱动加载 |
| 8 | 远程线程创建 |
| 10 | 进程访问 |
| 11 | 文件创建 |
| 22 | DNS 查询 (建议启用) |

---

## 官方文档链接

### Microsoft 官方文档

1. **Security Auditing Overview**
   https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/security-auditing-overview

2. **Events to Monitor (Appendix L)**
   https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/plan/appendix-l--events-to-monitor

3. **Event 4624 - Logon Success**
   https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4624

4. **Event 4625 - Logon Failure**
   https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4625

5. **Event 4688 - Process Creation**
   https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4688

6. **Event 4672 - Special Privileges**
   https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4672

7. **Event 4720 - User Account Created**
   https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4720

8. **Event 1102 - Audit Log Cleared**
   https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-1102

9. **Event 4719 - Audit Policy Changed**
   https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4719

10. **Sysmon Documentation**
    https://learn.microsoft.com/en-us/sysinternals/downloads/sysmon

### 审计策略配置

- **Basic Security Audit Policies**
  https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/plan/security-best-practices/basic-security-audit-policies

- **Advanced Security Audit Policy Settings**
  https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/component-updates/command-line-process-security

---

## 附录: 状态码参考 (Event 4625)

| 状态码 (Hex) | NTSTATUS 名称 | 说明 |
|-------------|---------------|------|
| 0xC0000064 | STATUS_NO_SUCH_USER | 用户不存在 |
| 0xC000006A | STATUS_WRONG_PASSWORD | 密码错误 |
| 0xC000006D | STATUS_BAD_VALIDATION_CLASS | 凭据验证失败 |
| 0xC000006F | STATUS_INVALID_LOGON_HOURS | 登录时间限制 |
| 0xC0000070 | STATUS_INVALID_WORKSTATION | 工作站限制 |
| 0xC0000072 | STATUS_ACCOUNT_DISABLED | 账户被禁用 |
| 0xC00000DC | STATUS_SERVER_UNAVAILABLE | 服务器不可用 |
| 0xC0000133 | STATUS_CLOCK_SKEW | 时钟不同步 |
| 0xC0000193 | STATUS_ACCOUNT_EXPIRED | 账户过期 |
| 0xC0000224 | STATUS_PASSWORD_MUST_CHANGE | 密码必须更改 |
| 0xC0000234 | STATUS_ACCOUNT_LOCKED_OUT | 账户已锁定 |

---

*本文档基于 Microsoft 官方文档和行业最佳实践编写。*