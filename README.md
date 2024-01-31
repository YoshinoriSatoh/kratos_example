# kratos_selfservice_example

[kartosのSelfService flow](https://www.ory.sh/docs/kratos/self-service)のサンプルです。

[こちら](https://zenn.dev/yoshinori_satoh/articles/kartos_usecase_overview)の記事で、kratosのSelfService flowについて記述しています。

ローカルでkratos一式を起動するためのdocker composeと、Browser/APIそれぞれのSelfService flowをcurlで再現するbashスクリプトを用意しています。

Browser flowには、[サーバーサイドレンダリングの場合](https://www.ory.sh/docs/kratos/self-service#browser-flows-for-server-side-apps-nodejs-php-java-)と、[クライアントサイドレンダリングの場合](https://www.ory.sh/docs/kratos/self-service#browser-flows-for-client-side-apps-single-page-apps-reactjs-angular-nextjs-)がありますが、ここではクライアントサイドレンダリングの場合を想定しています。

## 構成
docker-compose.yaml で以下のコンテナが起動します。

| container | description |
| ---- | ---- |
| kratos | kratos本体です |
| kratos-migrate | kratos DBに対してマイグレーションを行います |
| db-kratos | kratosのDBです(PostgreSQL) |
| mailslurper | メールサーバーです |

### SelfService flow実行スクリプト
./scripts配下に各種SelfService flowを実行するスクリプトを格納しています。

| flow type | operation | script file |
| ---- | ---- | ---- |
| browser | registration | registration_browser.sh |
| browser | login | login_browser.sh |
| browser | logout | logout_browser.sh |
| browser | verification | verification_browser.sh |
| browser | recover | recovery_browser.sh |
| browser | settings | setting_browser.sh |
| browser | check session | whoami_browser.sh |
| api | registration | registration_api.sh |
| api | login | login_api.sh |
| api | logout | logout_api.sh |
| api | verification | verification_api.sh |
| api | settings | setting_api.sh |
| api | check session | whoami_api.sh |

**注意点**

[API flowについて、recoverはサポートされていません](https://github.com/ory/kratos/discussions/2959)

recoveryはリダイレクトが必要となるため、ここだけはブラウザで実装する必要があるようです。

## 起動

### docker compose
```
docker compose up
```

## エンドポイントなど

| 項目 | URL |
| ---- | ---- |
| kratos public endpoint | http://localhost:4533 |
| kratos admin endpoint | http://localhost:4534 |
| kratos DB | postgres://kratos:secret@localhost:5432/kratos |
| mailslurper console | http://localhost:4436 |


## Identity schema

本サンプルでは以下のIdentity schemaを想定しています。

```json
{
  "$id": "https://schemas.ory.sh/presets/kratos/quickstart/email-password/identity.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "user",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "email": {
          "type": "string",
          "format": "email",
          "title": "E-Mail",
          "ory.sh/kratos": {
            "credentials": {
              "password": {
                "identifier": true
              }
            },
            "verification": {
              "via": "email"
            },
            "recovery": {
              "via": "email"
            }
          }
        },
        "nickname": {
          "type": "string",
          "title": "nickname"
        },
        "birthdate": {
          "type": "string",
          "title": "birthdate"
        }
      },
      "required": [
        "email"
      ],
      "additionalProperties": false
    }
  }
}
```

## browser flow(SPA)実行例

### ユーザー登録
以下が実行されます。
1. Registration flow初期化API
2. Registration flow送信API(method: password)
3. 3. 2.で実行されたVerification flowによるメールアドレス検証メール確認 ov
4. メールアドレス検証メールを確認し、プロンプトに6桁の検証コードを入力
5. Verification flow(mothod: code)送信API

#### コマンド実行手順
```
./scripts/registration_browser.sh <email> <password>
```

上記実行後に以下のプロンプトが表示されます。

```
please input code emailed to you:
```

[mailslurper console](http://localhost:4436)へアクセスすると、"Please verify your email address"というメールが届いています。

メール本文中に記載されている6桁の検証コードをプロンプトに入力し、Enterキーを押下すると、5. Verification flow(mothod: code)送信APIが実行され、メールアドレスが検証された状態となります。

#### 実行例
```
./scripts/registration_browser.sh 1@local overwatch2023
```

#### 1. Registration flowの初期化API

endpoint: `GET {{ kratos public endpoint }}/self-service/registration/browser`

[APIドキュメント](https://www.ory.sh/docs/kratos/reference/api#tag/frontend/operation/createBrowserRegistrationFlow)

Registration flowの初期化を行います。

レスポンスには、flow id等の他、uiという項目が含まれています。

uiで返却された項目は、本来はUIのレンダリングに使用します。

[ドキュメントによると](https://www.ory.sh/docs/kratos/self-service#form-rendering-1)、SPAの場合も[サーバサイドレンダリングの場合と同様に](https://www.ory.sh/docs/kratos/self-service#form-rendering)レンダリングする必要があるとのことです。

但し、本サンプルではcurlを使用していますので、UIレンダリングの過程は省いています。

**フォームレンダリング例**
```html
<form action="http://localhost:4533/self-service/registration?flow=65bcf3af-5b7d-4daa-a556-6a2443b8d52d" method="POST">
  <input
    name="csrf_token"
    type="hidden"
    value="esIfcdIQhbArLsvpsVir0pOiXWdc6FGKyZLg/S7cKN+83orzcSTk5Z7NMc6GeUUxO8tdaQocu7sYorTcRunAVQ=="
  />
  <fieldset>
    <label>
      <input name="traits.email" type="email" value="" placeholder="E-Mail" />
      <span>E-Mail</span>
    </label>
  </fieldset>
  <fieldset>
    <label>
      <input name="password" type="password" value="" placeholder="Password" />
      <span>Password</span>
    </label>
  </fieldset>
  <fieldset>
    <label>
      <input name="traits.nickname" type="text" value="" placeholder="nickname" />
      <span>nickname</span>
    </label>
  </fieldset>
  <fieldset>
    <label>
      <input name="traits.birthdat" type="text" value="" placeholder="birthdate" />
      <span>birthdate</span>
    </label>
  </fieldset> 
  <button name="method" type="submit" value="password">Sign up</button>
</form>
```


**レスポンス例**
```json
{
  "id": "65bcf3af-5b7d-4daa-a556-6a2443b8d52d",
  "type": "browser",
  "expires_at": "2024-01-31T04:53:49.281479005Z",
  "issued_at": "2024-01-31T03:53:49.281479005Z",
  "request_url": "http://localhost:4533/self-service/registration/browser",
  "ui": {
    "action": "http://localhost:4533/self-service/registration?flow=65bcf3af-5b7d-4daa-a556-6a2443b8d52d",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "csrf_token",
          "type": "hidden",
          "value": "esIfcdIQhbArLsvpsVir0pOiXWdc6FGKyZLg/S7cKN+83orzcSTk5Z7NMc6GeUUxO8tdaQocu7sYorTcRunAVQ==",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {}
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.email",
          "type": "email",
          "required": true,
          "autocomplete": "email",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "E-Mail",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "password",
          "type": "password",
          "required": true,
          "autocomplete": "new-password",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070001,
            "text": "Password",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.nickname",
          "type": "text",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "nickname",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.birthdate",
          "type": "text",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "birthdate",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "method",
          "type": "submit",
          "value": "password",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1040001,
            "text": "Sign up",
            "type": "info",
            "context": {}
          }
        }
      }
    ]
  }
}
```

#### 2. Registration flowの実行API(method: password)

endpoint: `POST {{ kratos public endpoint }}/self-service/registration/browser`

[APIドキュメント](https://www.ory.sh/docs/kratos/reference/api#tag/frontend/operation/updateRegistrationFlow)

1.で初期化したRegistration flowを実行します。

ここでは、認証にpasswordを使用しています。

(他には、oidcやwebauthnが指定可能です。)

1.のレスポンスを参照してレンダリングされたinput情報と、cookieが付与されている必要があります。

(curlの場合は明示的にcookieを付与していますが、ブラウザの場合は意識することはありません。)

レスポンスの`continue_with.flow.url`にリダイレクト先のURLが含まれています。

Identity schemaで、emailをcredentialsに指定している場合、Registration flowの実行API(method: password)を実行時に、メールアドレスを検証するためのVerification flowが実行されます。

Registration flowからVerification flowへ切り替わるため、次のflowを継続するための情報が`continue_with`に含まれています。

2024年1月現在、ドキュメントに明確な記載はないようなのですが、UI側で`continue_with.flow.url`へリダイレクトし、クエリパラメータで指定されたflow idから、以下のVerification flow取得APIを呼び出して、Verification flowを実行してほしいという意図があるのではないかと思います。

（本サンプルでは、curlを使用しているため、レンダリングの過程は省いています。）

**レスポンス例**
```json
{
  "session": {
    "id": "f5fc03b7-f923-4b80-a802-2847fd4c1796",
    "active": true,
    "expires_at": "2024-02-01T03:53:49.79546688Z",
    "authenticated_at": "2024-01-31T03:53:49.800609505Z",
    "authenticator_assurance_level": "aal1",
    "authentication_methods": [
      {
        "method": "password",
        "aal": "aal1",
        "completed_at": "2024-01-31T03:53:49.79546663Z"
      }
    ],
    "issued_at": "2024-01-31T03:53:49.79546688Z",
    "identity": {
      "id": "793126a9-3c8b-43ec-89d0-e48395235131",
      "schema_id": "user_v1",
      "schema_url": "http://localhost:4533/schemas/dXNlcl92MQ",
      "state": "active",
      "state_changed_at": "2024-01-31T03:53:49.789625713Z",
      "traits": {
        "email": "1@local"
      },
      "verifiable_addresses": [
        {
          "id": "a7d3f207-0a8d-47af-b0fb-576806a1bcde",
          "value": "1@local",
          "verified": false,
          "via": "email",
          "status": "sent",
          "created_at": "2024-01-31T03:53:49.7915Z",
          "updated_at": "2024-01-31T03:53:49.7915Z"
        }
      ],
      "recovery_addresses": [
        {
          "id": "694551fc-4074-4b92-b8e2-8cfe0a67c2e6",
          "value": "1@local",
          "via": "email",
          "created_at": "2024-01-31T03:53:49.792294Z",
          "updated_at": "2024-01-31T03:53:49.792294Z"
        }
      ],
      "metadata_public": null,
      "created_at": "2024-01-31T03:53:49.790597Z",
      "updated_at": "2024-01-31T03:53:49.790597Z"
    },
    "devices": [
      {
        "id": "27a7e329-858b-4ed0-bb81-69506551f53f",
        "ip_address": "192.168.65.1:38530",
        "user_agent": "curl/7.87.0",
        "location": ""
      }
    ]
  },
  "identity": {
    "id": "793126a9-3c8b-43ec-89d0-e48395235131",
    "schema_id": "user_v1",
    "schema_url": "http://localhost:4533/schemas/dXNlcl92MQ",
    "state": "active",
    "state_changed_at": "2024-01-31T03:53:49.789625713Z",
    "traits": {
      "email": "1@local"
    },
    "verifiable_addresses": [
      {
        "id": "a7d3f207-0a8d-47af-b0fb-576806a1bcde",
        "value": "1@local",
        "verified": false,
        "via": "email",
        "status": "sent",
        "created_at": "2024-01-31T03:53:49.7915Z",
        "updated_at": "2024-01-31T03:53:49.7915Z"
      }
    ],
    "recovery_addresses": [
      {
        "id": "694551fc-4074-4b92-b8e2-8cfe0a67c2e6",
        "value": "1@local",
        "via": "email",
        "created_at": "2024-01-31T03:53:49.792294Z",
        "updated_at": "2024-01-31T03:53:49.792294Z"
      }
    ],
    "metadata_public": null,
    "created_at": "2024-01-31T03:53:49.790597Z",
    "updated_at": "2024-01-31T03:53:49.790597Z"
  },
  "continue_with": [
    {
      "action": "show_verification_ui",
      "flow": {
        "id": "af77553e-ae12-43b9-aaaa-7c5c167eb8a6",
        "verifiable_address": "1@local",
        "url": "http://localhost:8000/auth/verification?flow=af77553e-ae12-43b9-aaaa-7c5c167eb8a6"
      }
    }
  ]
}
```

#### 3. 2.で実行されたVerification flowによるメールアドレス検証メール確認
2.で実行されたVerification flowによって、メールアドレス検証用のメールアドレスが送信されています。

メール本文中には6桁の検証コードが記載されており、[mailslurper console](http://localhost:4436)へアクセスすることで、ローカルで受信メールを確認できます。

**メールアドレス検証メール例**
```
Hi, please verify your account by entering the following code: 312996 or clicking the following link: http://localhost:4533/self-service/verification?code=312996&flow=d229d11d-8273-4b7e-b05e-57490c0310f0
```

#### 4. メールアドレス検証メールを確認し、プロンプトに6桁の検証コードを入力
メール本文中に記載されている6桁の検証コードを以下のプロンプトに入力し、Enterキーを押下すると、5. Verification flow(mothod: code)送信APIが実行されます。

```
please input code emailed to you:
```

#### 5. Verification flow(mothod: code)送信API

endpoint: `POST {{ kratos public endpoint }}/self-service/verification`

[APIドキュメント](https://www.ory.sh/docs/kratos/reference/api#tag/frontend/operation/updateVerificationFlow)

Verification flow(mothod: code)送信APIが呼び出し、メールアドレスが検証された状態となります。



### ログイン
以下が実行されます。
1. Login flow初期化API
2. Login flow送信API

#### コマンド実行手順
```
./scripts/login_browser.sh <email> <password>
```

#### 実行例
```
./scripts/login_browser.sh 1@local overwatch2023
```

#### 1. Login flowの初期化API

endpoint: `GET {{ kratos public endpoint }}/self-service/login/browser`

[APIドキュメント](https://www.ory.sh/docs/kratos/reference/api#tag/frontend/operation/createBrowserLoginFlow)

Login flowの初期化を行います。

[Registration flowと同様に](https://github.com/YoshinoriSatoh/kratos_selfservice_example?tab=readme-ov-file#1-registration-flow%E3%81%AE%E5%88%9D%E6%9C%9F%E5%8C%96api)、uiの内容に従ってUIをレンダリングします。

**レスポンス例**
```json
{
  "id": "85a83b3d-835a-4ef1-a2e3-2e7d7cf8f826",
  "type": "browser",
  "expires_at": "2024-01-31T05:08:59.442571343Z",
  "issued_at": "2024-01-31T04:08:59.442571343Z",
  "request_url": "http://localhost:4533/self-service/login/browser",
  "ui": {
    "action": "http://localhost:4533/self-service/login?flow=85a83b3d-835a-4ef1-a2e3-2e7d7cf8f826",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "csrf_token",
          "type": "hidden",
          "value": "Dx5gyPsVHX1U2OYriH8qYxLT/6P8G5TuyROFqhyNGUQI4zMPaE0vrTC3jFmNBYHuofb9KUgN4/bzdc9yunIpXg==",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {}
      },
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "identifier",
          "type": "text",
          "value": "",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070004,
            "text": "ID",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "password",
          "type": "password",
          "required": true,
          "autocomplete": "current-password",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070001,
            "text": "Password",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "method",
          "type": "submit",
          "value": "password",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1010001,
            "text": "Sign in",
            "type": "info",
            "context": {}
          }
        }
      }
    ]
  },
  "created_at": "2024-01-31T04:08:59.446151Z",
  "updated_at": "2024-01-31T04:08:59.446151Z",
  "refresh": false,
  "requested_aal": "aal1"
}
```

#### 2. Login flowの送信API

1.で初期化したLogin flowを実行します。

session情報が返却されます。

```json
{
  "session": {
    "id": "0f867e3b-7c89-432c-9368-f021f4f686d4",
    "active": true,
    "expires_at": "2024-02-01T04:08:59.829461593Z",
    "authenticated_at": "2024-01-31T04:08:59.829461593Z",
    "authenticator_assurance_level": "aal1",
    "authentication_methods": [
      {
        "method": "password",
        "aal": "aal1",
        "completed_at": "2024-01-31T04:08:59.829459176Z"
      }
    ],
    "issued_at": "2024-01-31T04:08:59.829461593Z",
    "identity": {
      "id": "793126a9-3c8b-43ec-89d0-e48395235131",
      "schema_id": "user_v1",
      "schema_url": "http://localhost:4533/schemas/dXNlcl92MQ",
      "state": "active",
      "state_changed_at": "2024-01-31T03:53:49.789625Z",
      "traits": {
        "email": "1@local"
      },
      "verifiable_addresses": [
        {
          "id": "a7d3f207-0a8d-47af-b0fb-576806a1bcde",
          "value": "1@local",
          "verified": true,
          "via": "email",
          "status": "completed",
          "verified_at": "2024-01-31T04:08:28.273878Z",
          "created_at": "2024-01-31T03:53:49.7915Z",
          "updated_at": "2024-01-31T03:53:49.7915Z"
        }
      ],
      "recovery_addresses": [
        {
          "id": "694551fc-4074-4b92-b8e2-8cfe0a67c2e6",
          "value": "1@local",
          "via": "email",
          "created_at": "2024-01-31T03:53:49.792294Z",
          "updated_at": "2024-01-31T03:53:49.792294Z"
        }
      ],
      "metadata_public": null,
      "created_at": "2024-01-31T03:53:49.790597Z",
      "updated_at": "2024-01-31T03:53:49.790597Z"
    },
    "devices": [
      {
        "id": "98e9b8cc-5ffb-446c-b76d-7383b112600d",
        "ip_address": "192.168.65.1:38548",
        "user_agent": "curl/7.87.0",
        "location": ""
      }
    ]
  }
}
```




### ログインセッション取得
以下が実行されます。
1. Login session取得API

#### コマンド実行手順
```
./scripts/whoami_browser.sh <email> <password>
```

#### 実行例
```
./scripts/whoami_browser.sh
```

#### 1. Login session取得API

endpoint: `GET {{ kratos public endpoint }}/sessions/whoami`

[APIドキュメント](https://www.ory.sh/docs/kratos/reference/api#tag/frontend/operation/toSession)

ログイン中のセッションが有効であれば、セッション情報が返却されます。

セッション情報取得の他、現在ログイン中であるかどうかを確認するエンドポイントでもあります。

**レスポンス例**
```json
{
  "id": "0f867e3b-7c89-432c-9368-f021f4f686d4",
  "active": true,
  "expires_at": "2024-02-01T04:08:59.829461Z",
  "authenticated_at": "2024-01-31T04:08:59.829461Z",
  "authenticator_assurance_level": "aal1",
  "authentication_methods": [
    {
      "method": "password",
      "aal": "aal1",
      "completed_at": "2024-01-31T04:08:59.829459176Z"
    }
  ],
  "issued_at": "2024-01-31T04:08:59.829461Z",
  "identity": {
    "id": "793126a9-3c8b-43ec-89d0-e48395235131",
    "schema_id": "user_v1",
    "schema_url": "http://localhost:4533/schemas/dXNlcl92MQ",
    "state": "active",
    "state_changed_at": "2024-01-31T03:53:49.789625Z",
    "traits": {
      "email": "1@local"
    },
    "verifiable_addresses": [
      {
        "id": "a7d3f207-0a8d-47af-b0fb-576806a1bcde",
        "value": "1@local",
        "verified": true,
        "via": "email",
        "status": "completed",
        "verified_at": "2024-01-31T04:08:28.273878Z",
        "created_at": "2024-01-31T03:53:49.7915Z",
        "updated_at": "2024-01-31T03:53:49.7915Z"
      }
    ],
    "recovery_addresses": [
      {
        "id": "694551fc-4074-4b92-b8e2-8cfe0a67c2e6",
        "value": "1@local",
        "via": "email",
        "created_at": "2024-01-31T03:53:49.792294Z",
        "updated_at": "2024-01-31T03:53:49.792294Z"
      }
    ],
    "metadata_public": null,
    "created_at": "2024-01-31T03:53:49.790597Z",
    "updated_at": "2024-01-31T03:53:49.790597Z"
  },
  "devices": [
    {
      "id": "98e9b8cc-5ffb-446c-b76d-7383b112600d",
      "ip_address": "192.168.65.1:38548",
      "user_agent": "curl/7.87.0",
      "location": ""
    }
  ]
}
```