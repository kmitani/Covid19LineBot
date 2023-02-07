# Covid-19 News LINE Bot
新型コロナ感染者数が変動しはじめたタイミングで通知してくれるLINE Botです。サイバーエージェント主催の第2期Go Academyの最終課題で開発しました。
AWS Lambda, RDS(MySQL)で動作させることを想定しています。

# 機能

### 感染状況変動の通知 (notificationフォルダ)
このLINE Botは感染者数増減がはじまったタイミングにメッセージを送ります。そのため、メッセージに煩わしさを感じずに、感染状況が変化した重要な時にニュースを受け取ることができます。

### 都道府県ごとの感染状況の返信(showフォルダ)
“神奈川県”などを含むメッセージがユーザーから送られた時に、その場所の感染状況をお知らせするメッセージを返信します。

# 感染状況変動検知のLogic
- オープンデータから新規感染者数を取得
- 週平均の感染者数を計算
- 週平均感染者数の前日比を計算
- 前日比の2週平均を計算
- 平均前日比が閾値（1.01 or 0.99）を超えた場合に通知メッセージを送信
