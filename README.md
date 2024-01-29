# kratos_selfservice_example

[kartosのSelfService flow](https://www.ory.sh/docs/kratos/self-service)のサンプルです。

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


## browser flow(SPA)実行例

### ユーザー登録
```
./scripts/registration_browser.sh
```
