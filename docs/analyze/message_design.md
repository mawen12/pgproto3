# Message Design

客户端通过 `BackendMessage` 和 `FrontendMessage` 来实现与 PostgreSQL 后端的通信。

> 具体请查阅 [pgproto3.go]

## Message

顶层通用接口。提供 `Decode` 和 `Encode` 方法来解码和编码消息。

## FrontendMessage

Frontend 消息顶层接口，基于 `Message` 接口，提供一个空方法 `Frontend()` 来标识其为 Frontend 消息。

## BackendMessage

Backend 消息顶层接口，基于 `Message` 接口，提供一个空方法 `Backend()` 来标识其为 Backend 消息。

## AuthenticationResponseMessage

认证响应消息接口，基于 `BackendMessage` 接口，提供一个空方法 `AuthenticationResponse()` 来标识其为认证响应消息。

## 具体实现

**FrontendMessage**

|消息 | 说明 | 源文件 |
| --- | --- | --- |
|`Bind` | | |
|`CancelRequest` | | |
|`Close` | | |
|`CopyData` | | |
|`CopyDone` | | |
|`CopyFail` | | |
|`Describe` | | |
|`Execute` | | |
|`Flush` | | |
|`FunctionCall` | | |
|`GSSEncRequest` | | |
|`GSSResponse` | | |
|`Parse` | | |
|`PasswordMessage` | | |
|`SASInitialResponse` |  | |
|`SASLResponse` | | |
|`SSLRequest` | | [ssl_request.go](../../ssl_request.go) |
|`StartupMessage` | | [startup_message.go](../../startup_message.go) |
|`Sync` | | [sync.go](../../sync.go) |
|`Terminate` | | [terminate.go](../../terminate.go) |

**BackendMessage**

|消息 | 说明 | 源文件 |
| --- | --- | --- |
| `AuthenticationCleartextPassword` | | |
| `AuthenticationGSSContinue` | | |
| `AuthenticationGSS` | | |
| `AuthenticationMD5Password` | | |
| `AuthenticationOK` | | |
| `AuthenticationSASLContinue` | | |
| `AuthenticationSASLFinal` | | |
| `AuthenticationSASL` | | |
| `BackendKeyData` | | |
| `BindComplete` | | |
| `CloseComplete` | | |
| `CommandComplete` | | |
| `CopyBothResponse` | | |
| `CopyData` | | |
| `CopyDone` | | |
| `CopyInResponse` | | |
| `CopyOutResponse` | | |
| `DataRow` | | |
| `EmptyQueryResponse` | | |
| `ErrorResponse` |  | [error_response.go](../../error_response.go) |
| `FunctionCallResponse` | | |
| `NoData` | | |
| `NoticeResponse` | | |
| `NotificationResponse` | | |
| `ParameterDescription` | | |
| `ParameterStatus` | | |
| `ParseComplete` | | |
| `PortalSuspended` | | |
| `ReadyForQuery` | | |
| `RowDescription` | | |


