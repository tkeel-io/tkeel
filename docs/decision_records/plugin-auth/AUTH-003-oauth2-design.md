# AUTH-003-oauth2-design

## Status
Proposed

## Context
用户授权和认证系统为了以后更好的去对接第三方平台或第三方用户管理系统，在协议规范上应尽可能的遵循现有的主流认证规范。

## Decision
将**Oauth2**规范进行部分实现，加入到现有的**auth**插件中，结合平台的设计和特有的租户管理系统，对规范中的流程选择性的实现。


### Oauth介绍和平台实现

OAuth是一个开放标准，它允许用户在不提供用户名密码给第三方应用的情况下给第三方应用授权。

### 角色

OAuth2.0定义了四种角色，分别如下：

1.  **ResourceOwner** 

    >    资源所有者，一般指用户

2.  **Client**

    >   客户端，通过申请ResourceOwner 的授权，从而实现访问受保护资源的第三方软件或者服务

3.  **AuthorizationServer**

    >   授权服务，在ResourceOwner授权完毕后，负责颁发access token 的服务

4.  **ResourceServer** 

    >   资源服务，**存储**着ResourceOwner的受保护的**资源**的**服务**，可通过验证access token来开放对ResourceServer的数据的访问

### 协议流程

#### access token

![image-20211010235103384](C:\Users\96331\AppData\Roaming\Typora\typora-user-images\image-20211010235103384.png)

A: 客户端向从资源所有者请求授权。

B:资源所有者同意授权。

C:客户端获得了资源所有者的授权之后，向授权服务器申请授权令牌

D:授权服务器验证客户端无误后发放授权令牌。。

E:客户端拿到授权令牌之后请求资源服务器发送用户信息。

F:资源服务器验证令牌无误后将用户信息发放给客户端。



#### token & refresh token

![image-20211011000444955](C:\Users\96331\AppData\Roaming\Typora\typora-user-images\image-20211011000444955.png)

A:客户端通过与授权服务器进行身份验证并出示授权许可请求访问令牌

B:权服务器对客户端进行身份验证并验证授权许可，若有效则颁发访问令牌和刷新令牌。

C:端通过出示访问令牌向资源服务器发起受保护资源的请求。

D:服务器验证访问令牌，若有效则满足该要求。

E: C和D重复进行，直到访问令牌到期。如果客户端知道访问令牌已过期，跳到步骤G，否 则它将继续发起另一个对受保护资源的请求。

F:由于访问令牌是无效的，资源服务器返回无效令牌错误。

G:客户端通过与授权服务器进行身份验证并出示刷新令牌，请求一个新的访问令牌。客户端身份验证要求基于客户端的类型和授权服务器的策略。

H: 授权服务器对客户端进行身份验证并验证刷新令牌，若有效则颁发一个新的访问令牌



### 目前授权类型分为四种

四种分别时授权码模式、简化模式、密码模式、客户端模式，流程具体如下：

#### 授权码模式

![image-20211011001318826](C:\Users\96331\AppData\Roaming\Typora\typora-user-images\image-20211011001318826.png)

A、B ：**用户使用第三方应用请求授权**：第三方应用将资源所有者导向一个特定的地址，附带以下请求信息：

-   response_type：必选，请求类型。这里固定为"code"。
-   client_id：必选，标识第三方应用的id。很多地方也用apppid来代替。
-   redirect_uri：可选，授权完成后重定向的地址。当取得用户授权后，服务将重定向到这个地址，并在地址里附带上**授权码**。
-   scope：可选，第三方请求的资源范围。比如是想获取基本信息、敏感信息等。
-   state：推荐，用于状态保持，可以是任意字符串。授权服务会原封不动地返回。

C：**第三方应用拿到授权请求的response**：资源所有者同意授权第三方应用访问受限资源后，请求返回，跳转到 redirect_uri 指定的地址。返回body有：

-   code：必选，授权码。后续步骤中，用来交换access token。
-   state：必选（如果授权请求中，带上了state），这里原封不动地回传。

D：**第三方应用继续请求access token**：第三方应用向授权服务请求获取access token。请求参数包括：

-   grant_type：必选，许可类型，这里固定为“authorization_code”。
-   code：必选，授权码。在用户授权步骤中，授权服务返回的。
-   redirect_uri：必选，如果在授权请求步骤中，带上了redirect_uri，那么这里也必须带上，且值相同。
-   client_id：必选，第三方应用id。

E：**授权服务返回给第三方应用access token**：请求合法且授权验证通过，那么授权服务将access token返回给第三方应用。返回body有：

-   access token：访问令牌，第三方应用访问用户资源的凭证。
-   access_token_expires_in：access token的有效时长。
-   refresh token：更新access token的凭证。当access token过期，可以refresh token为凭证，获取新的access token。
-   refresh_token_expires_in：refresh token的有效时长。



#### 简化模式

![image-20211011001407542](C:\Users\96331\AppData\Roaming\Typora\typora-user-images\image-20211011001407542.png)

A：客户端通过向授权点引导资源所有者的用户代理开始流程。客户端包括它的客户端标识、请求范围、本地状态和重定向URI，一旦访问被许可（或拒绝）授权服务将传送用户代理回到该URI。

B：授权服务器验证资源拥有者的身份（通过用户代理），并确定资源所有者是否授予或拒绝客户端的访问请求。

C：假设资源所有者许可访问，授权服务使用之前（在请求时或客户端注册时）提供的重定向URI重定向用户代理回到客户端。重定向URI在URI片段中包含访问令牌。

D：用户代理顺着重定向指示向Web托管的客户端资源发起请求。用户代理在本地保留片段信息。

E：Web托管的客户端资源返回一个网页（通常是带有嵌入式脚本的HTML文档），该网页能够访问包含用户代理保留的片段的完整重定向URI并提取包含在片段中的访问令牌（和其他参数）。

F：用户代理在本地执行Web托管的客户端资源提供的提取访问令牌的脚本。

G:用户代理传送访问令牌给客户端

-   授权请求

    GET /authorize?response_type=token&client_id=s6BhdRkqt3&state=xyz&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb

#### 用户名/密码

![image-20211011001421879](C:\Users\96331\AppData\Roaming\Typora\typora-user-images\image-20211011001421879.png)

A：资源所有者提供给客户端它的用户名和密码。

B：通过包含从资源所有者处接收到的凭据，客户端从授权服务的令牌接口请求访问令牌。当发起请求时，客户端与授权服务进行身份验证。

C：授权服务对客户端进行身份验证，验证资源所有者的凭证，如果有效，颁发访问令牌。

-   授权请求

    ```
    POST /token HTTP/1.1
         Content-Type: application/x-www-form-urlencoded     
         grant_type=password&username=johndoe&password=A3ddj3w
    ```

#### 客户端凭据

![image-20211011001434071](C:\Users\96331\AppData\Roaming\Typora\typora-user-images\image-20211011001434071.png)

A:客户端与授权服务进行身份验证。

B:授权服务对客户端进行身份验证，如果有效，颁发访问令牌

-   授权请求

    ```
     POST /token HTTP/1.1
         Host: server.example.com
         Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
         Content-Type: application/x-www-form-urlencoded
         grant_type=client_credentials
    ```



#### 请求响应 规范

**success**:

``` json
{
    "access_token":"",
    "token_type":"",
    "expires_in":"",
    "refresh_token":"",
    "example_parameter":""
 }
```

**error:**

```json
{
    "error":"invalid_request"
}

invalid_request:
请求缺少必需的参数、包含不支持的参数值（除了许可类型）、重复参数、包含多个凭据、采用超过一种客户端身份验证机制或其他不规范的格式。

invalid_client:
客户端身份验证失败（例如，未知的客户端，不包含客户端身份验证，或不支持的身份验证方法）。授权服务器可以返回HTTP 401（未授权）状态码来指出支持的HTTP身份验证方案。如果客户端试图通过“Authorization”请求标头域进行身份验证，授权服务器必须响应
HTTP 401（未授权）状态码，并包含与客户端使用的身份验证方案匹配的“WWW-Authenticate”响应标头字段。

invalid_grant:
提供的授权许可（如授权码、资源所有者凭据）或刷新令牌无效、过期、吊销、与在授权请求使用的重定向URI不匹配或颁发给另一个客户端。

unauthorized_client
进行身份验证的客户端没有被授权使用这种授权许可类型。

unsupported_grant_type
授权许可类型不被授权服务器支持。

invalid_scope
请求的范围无效、未知的、格式不正确或超出资源所有者许可的范围。

error_description
可选的,用于协助客户端开发人员理解所发生的错误。

error_uri
可选的。指向带有有关错误的信息的人类可读网页的URI，用于提供客户端开发人员关于该错误的额外信息.

```

### api

-   oauth
    -   login
    -   token( Resource Owner Password Credentials Grant)
    -   authorize (Only support implicit grant flow)
    -   authenticate( authentication)