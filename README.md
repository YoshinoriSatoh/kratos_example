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
1. Registration flow初期化APIが呼び出される
2. Registration flow送信APIが呼び出される(method: password)
3. 2.によってVerification flowが実行され、メールアドレス検証メールが送信される
4. メールアドレス検証メールを確認し、プロンプトに6桁の検証コードを入力する
5. Verification flow(mothod: code)送信APIが呼び出される

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

uiで返却された項目は、UIのレンダリングに使用します。

[ドキュメントによると](https://www.ory.sh/docs/kratos/self-service#form-rendering-1)、SPAの場合も[サーバサイドレンダリングの場合と同様に](https://www.ory.sh/docs/kratos/self-service#form-rendering)レンダリングする必要があるとのことです。

**フォームレンダリング**
```html
<form action="http://localhost:4533/self-service/registration?flow=601015e9-e5e2-46fa-83f8-db5fb332535e" method="POST">
  <input
    name="csrf_token"
    type="hidden"
    value="rKZzgTZOs5AqAbjrXZJmngsV60aTDZztIvejmqO8h4bh1ir5xSaZGkkMUE2JuldooHyV3xHReNUQCTw8OFEpyQ=="
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
  "id": "601015e9-e5e2-46fa-83f8-db5fb332535e",
  "type": "browser",
  "expires_at": "2024-01-31T02:38:58.256733884Z",
  "issued_at": "2024-01-31T01:38:58.256733884Z",
  "request_url": "http://localhost:4533/self-service/registration/browser",
  "ui": {
    "action": "http://localhost:4533/self-service/registration?flow=601015e9-e5e2-46fa-83f8-db5fb332535e",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "csrf_token",
          "type": "hidden",
          "value": "rKZzgTZOs5AqAbjrXZJmngsV60aTDZztIvejmqO8h4bh1ir5xSaZGkkMUE2JuldooHyV3xHReNUQCTw8OFEpyQ==",
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

**レスポンス例**
```json
{
  "session": {
    "id": "c4a0ea06-aa8e-432a-8134-0c7962b366a4",
    "active": true,
    "expires_at": "2024-02-01T01:38:58.687117426Z",
    "authenticated_at": "2024-01-31T01:38:58.699166843Z",
    "authenticator_assurance_level": "aal1",
    "authentication_methods": [
      {
        "method": "password",
        "aal": "aal1",
        "completed_at": "2024-01-31T01:38:58.687116551Z"
      }
    ],
    "issued_at": "2024-01-31T01:38:58.687117426Z",
    "identity": {
      "id": "41fecaf9-5d31-4b55-9304-4dec3635f199",
      "schema_id": "user_v1",
      "schema_url": "http://localhost:4533/schemas/dXNlcl92MQ",
      "state": "active",
      "state_changed_at": "2024-01-31T01:38:58.668492843Z",
      "traits": {
        "email": "1@local"
      },
      "verifiable_addresses": [
        {
          "id": "24a98598-ec75-438a-a81b-8615ff63fd3f",
          "value": "1@local",
          "verified": false,
          "via": "email",
          "status": "sent",
          "created_at": "2024-01-31T01:38:58.673568Z",
          "updated_at": "2024-01-31T01:38:58.673568Z"
        }
      ],
      "recovery_addresses": [
        {
          "id": "ebe87084-18cc-43f4-b2ed-127f04c94a4f",
          "value": "1@local",
          "via": "email",
          "created_at": "2024-01-31T01:38:58.676839Z",
          "updated_at": "2024-01-31T01:38:58.676839Z"
        }
      ],
      "metadata_public": null,
      "created_at": "2024-01-31T01:38:58.671179Z",
      "updated_at": "2024-01-31T01:38:58.671179Z"
    },
    "devices": [
      {
        "id": "71d91744-50e6-411a-bad6-3002205eedba",
        "ip_address": "192.168.65.1:37958",
        "user_agent": "curl/7.87.0",
        "location": ""
      }
    ]
  },
  "identity": {
    "id": "41fecaf9-5d31-4b55-9304-4dec3635f199",
    "schema_id": "user_v1",
    "schema_url": "http://localhost:4533/schemas/dXNlcl92MQ",
    "state": "active",
    "state_changed_at": "2024-01-31T01:38:58.668492843Z",
    "traits": {
      "email": "1@local"
    },
    "verifiable_addresses": [
      {
        "id": "24a98598-ec75-438a-a81b-8615ff63fd3f",
        "value": "1@local",
        "verified": false,
        "via": "email",
        "status": "sent",
        "created_at": "2024-01-31T01:38:58.673568Z",
        "updated_at": "2024-01-31T01:38:58.673568Z"
      }
    ],
    "recovery_addresses": [
      {
        "id": "ebe87084-18cc-43f4-b2ed-127f04c94a4f",
        "value": "1@local",
        "via": "email",
        "created_at": "2024-01-31T01:38:58.676839Z",
        "updated_at": "2024-01-31T01:38:58.676839Z"
      }
    ],
    "metadata_public": null,
    "created_at": "2024-01-31T01:38:58.671179Z",
    "updated_at": "2024-01-31T01:38:58.671179Z"
  },
  "continue_with": [
    {
      "action": "show_verification_ui",
      "flow": {
        "id": "d229d11d-8273-4b7e-b05e-57490c0310f0",
        "verifiable_address": "1@local",
        "url": "https://www.ory.sh/kratos/docs/fallback/verification?flow=d229d11d-8273-4b7e-b05e-57490c0310f0"
      }
    }
  ]
}
```

#### 3. 2.によってVerification flowが実行され、メールアドレス検証メールが送信される
Identity schemaで、emailをcredentialsに指定している場合、Registration flowの実行API(method: password)を実行時に、メールアドレスを検証するためのVerification flowが実行されます。

メールアドレス検証用のメールアドレスが送信され、メール本文中には6桁の検証コードが記載されています。

[mailslurper console](http://localhost:4436)へアクセスすることで、ローカルで受信メールを確認できます。

**メールアドレス検証メール例**
```
Hi, please verify your account by entering the following code: 312996 or clicking the following link: http://localhost:4533/self-service/verification?code=312996&flow=d229d11d-8273-4b7e-b05e-57490c0310f0
```

#### 4. メールアドレス検証メールを確認し、プロンプトに6桁の検証コードを入力する
メール本文中に記載されている6桁の検証コードを以下のプロンプトに入力し、Enterキーを押下すると、5. Verification flow(mothod: code)送信APIが実行されます。

```
please input code emailed to you:
```

#### 5. Verification flow(mothod: code)送信APIが呼び出される
Verification flow(mothod: code)送信APIが呼び出し、メールアドレスが検証された状態となります。