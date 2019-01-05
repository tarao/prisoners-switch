The Prisoners and the Switch Room Problem
=========================================

The Problem
-----------

You are one of a hundred prisoners.  You all put into a game described
below, in which you will be set free if you win but will be executed
if you lose.

- When the game starts, you will be in isolated cells.
  - You cannot talk to other prisoners during the game.
- You will randomly picked one by one into a _switch room_.
  - Each prisoner will visit the room arbitrarily often if he waits enough.
  - You will never know who is in the room now except when you are in the room.
  - There are only _two switches_ (A and B) in the room and nothing else.
  - Nobody except the prisoners will enter this room.
- You can do only two things in the _switch room_:
  - see the state ("on" or "off") of the _switches_, and,
  - toggle one of / both of / none of the two _switches_.
- Every prisoner can _assert_ at any time that "everybody visited the _switch room_ at least once".
  - If it is true, then you win the game.
  - If it is false, i.e., someone has not yet visited the _switch room_, then you lose the game.
- The initial states of the _two switches_ are decided at random.

Before starting the game, you all may meet together and plan a strategy.  Provide a strategy which guarantees you can always win the game.

----

If this is too easy for you, improve the strategy by
- minimizing (the expected value of) the total number of visits to the _switch room_, or,
- using only _one switch_ throughout the game.

Giving a Solution by a Program
------------------------------

You can give a solution by forking this repository and making a pull request.  It will be automatically verified that it is a correct strategy.  Follow the instructions below to make a pull request.

1. You need [Golang](https://golang.org/) environment.
1. Fork this repository to, say, <code><var>yourname</var>/prisoners-switch</code>.
1. <code>go get github.com/tarao/prisoners-switch</code>
   - Note that it isn't <code><var>yourname</var>/prisoners-switch</code> but <code>tarao/prisoners-switch</code>.
1. <code>cd "$GOPATH"/src/github.com/tarao/prisoners-switch</code>
1. <code>git remote add <var>yourname</var> git@github.com:<var>yourname</var>/prisoners-switch</code>
1. <code>git checkout -b <var>your-answer</var></code>
1. Edit <code>strategy/my_strategy.go</code> and write your strategy, and then <code>git commit</code> it.
1. <code>git push --set-upstream <var>yourname</var> <var>your-answer</var></code>
1. Make a pull request from <code><var>your-answer</var></code> branch to the <code>master</code> of <code>tarao/prisoners-switch</code>.

There are some restrictions you need to follow.

- You can change nothing other than files under <code>strategy/</code>.
  - You may create a new file
  - The maximum number of changed files is 20.
- You can <code>import</code> nothing other than <code>github.com/tarao/prisoners-switch/rule</code>.

To verify the strategy locally, run <code>verifier/run</code>.

問題
----

あなたは100人の囚人の一人です. 全員で以下のようなゲームをして, 見事勝利できれば全員釈放, 負ければ全員死刑となります.

- ゲーム開始と同時に全員別々の独房に入ります
  - 独房内や通路で他の囚人とやりとりすることはできません
- ランダムに1人ずつ<strong>スイッチの部屋</strong>に呼ばれます
  - 十分な時間待てば同じ人が何度でも呼ばれます
  - いま誰が呼ばれているかは本人以外にはわかりません
  - 部屋には<strong>スイッチ</strong>が2つ(AとB)ある以外はなにもありません
  - ゲームに参加中の囚人以外がこの部屋に入ることはありません
- <strong>スイッチの部屋</strong>では以下のことができます
  - 2つある<strong>スイッチ</strong>のon/offの状態を確認する
  - 2つある<strong>スイッチ</strong>のいずれか, もしくは両方のon/offを切り替える
- 囚人はいつでも「全員<strong>スイッチの部屋</strong>に入った!」と<strong>宣言</strong>することができます
  - 本当であれば勝利となります
  - まだ一度も部屋に入っていない人がいたらその時点で負けです
- ゲーム開始時点での<strong>スイッチ</strong>の状態はランダムに決められます

ゲームを開始する前に, 囚人全員で集まって作戦を立てることができます. 必ず勝利できる作戦を考えてください.

----

答えがわかって物足りなければ以下についても考えてください.

- <strong>スイッチの部屋</strong>に入った総回数(の期待値)がなるべく少なくなるような作戦を立ててください
- <strong>スイッチ</strong>を1つしか使わずに勝利する方法を考えてください

プログラムによる解答方法
------------------------

このリポジトリをForkしてPull Requestすることで解答できます.  正しく解答できているかどうかは自動的にチェックされます.  以下の手順に従ってPull Requestしてください.

1. [Golang](https://golang.org/)環境を用意
1. このリポジトリを<code><var>yourname</var>/prisoners-switch</code>にFork
1. <code>go get github.com/tarao/prisoners-switch</code>
   - <code><var>yourname</var>/prisoners-switch</code>ではなく<code>tarao/prisoners-switch</code>な点に注意
1. <code>cd "$GOPATH"/src/github.com/tarao/prisoners-switch</code>
1. <code>git remote add <var>yourname</var> git@github.com:<var>yourname</var>/prisoners-switch</code>
1. <code>git checkout -b <var>your-answer</var></code>
1. <code>strategy/my_strategy.go</code>を編集して解答して<code>git commit</code>
1. <code>git push --set-upstream <var>yourname</var> <var>your-answer</var></code>
1. <code><var>your-answer</var></code>ブランチを<code>tarao/prisoners-switch</code>の<code>master</code>ブランチにPull Requestする

解答は以下の制限に合致している必要があります.

- 書き換えてよいのは<code>strategy/</code>以下のみ
  - ファイルを追加してもよい
  - 変更ファイル数の上限は20
- <code>github.com/tarao/prisoners-switch/rule</code>以外を<code>import</code>してはいけない

うまく解けているか手元で確認したい場合は<code>verifier/run</code>を実行すると確認できます.

License
-------

- MIT
