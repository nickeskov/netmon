# netmon

Сервис базового мониторинга сетей блокчейна **WAVES** (mainnet, stagenet, testnet). Сервис хранит все данные в
оперативной памяти не имеет постоянного хранилища, вследствие чего после перезапуска сервис теряет все накопленные
статистики и оценки.

## Command line parameters

Далее будут описаны параметры, которые можно передать исполняемому файлу при запуске для изменения его поведения по
умолчанию.

### Basic command line parameters

Базовые настройки:

- _--log-level_ - уровень логирования. Поддерживаемые уровни:  _DEV_, _DEBUG_, _INFO_, _WARN_, _ERROR_, _FATAL_. По
  умолчанию _INFO_.
- _--bind-addr_ - адрес, на котором будет запущен сервис. По умолчанию _0.0.0.0:2048_.
- _--network_ - сеть WAVES, за которой будет наблюдать сервис. Поддерживаемые сети: _mainnet_, _testnet_, _stagenet_. По
  умолчанию _mainnet_.
- _--stats-url_ - URL, с которого будет собираться статистика по узлам сети. По
  умолчанию _https://waves-nodes-get-height.wavesnodes.com/_.
- _--stats-poll-interval_ - интервал сбора, через который будет собираться статистика по узлам сети и *обновляться
  статистики*. По умолчанию _1m_.
- _--stats-history-size_ - хранимое количество последних хранимых снимков статистик. Должен быть больше 0. По
  умолчанию _10_.
- _--network-errors-streak_ - число последовательных ошибок, после будет считаться, что сеть находится в деградированном
  состоянии. По умолчанию _5_.
- _--initial-mon-state_ - состояние мониторинга при старте. Возможные значения: _active_, _frozen_operates_stable_, _
  frozen_degraded_. По умолчанию _active_.
- _--http-auth-header_ - HTTP заголовок, в котором будет проверяться наличие токена для доступа к приватным URL. По
  умолчанию _X-Waves-Monitor-Auth_.
- _--http-auth-token_ - токен доступа к приватным URL. **ОБЯЗАТЕЛЬНЫЙ** параметр. Значение по умолчанию отсутствует.

## Monitoring criteria

Далее будут описаны опции (критерии), которые непосредственно влияют на мониторинг ошибок. Состояние сети будет
оцениваться по серии *последовательных* ошибок, которые генерирует сервис. При генерации ошибки будет увеличиваться
счётчик последовательных ошибок, однако если после серии ошибок произойдёт сбор статистик и их оценка, в результате
которой ошибка *не* будет сгенерирована, то счётчик последовательности ошибок сбросится.

#### Down nodes criterion

- _--criterion-down-total-part_ - пороговое значение при достижении которого будет генерироваться ошибка. Считается как
  отношение недоступных узлов ко всем отслеживаемым узлам. Диапазон значений: от _0.0_ не включительно до _1.0_
  включительно. По умолчанию _0.3_.

#### Nodes height criterion

- _--criterion-height-diff_ - разница высот узлов сети при достижении которой будет генерироваться ошибка. По
  умолчанию _10_ блоков.
- _--criterion-height-require-min-nodes-on-same-height_ - необходимое количество узлов сети, которые находятся на одной
  высоте. По умолчанию _2_ узла.

#### Statehash criterion

Здесь под группой понимается группа узлов сети на одной высоте, узлы которой имеют одинаковые стейтхеши.

- _--criterion-statehash-min-groups-on-same-height_ - минимальное количество групп узлов сети с разными стейтхешами на
  одной высоте. По умолчанию _2_ группы.
- _--criterion-statehash-min-valuable-groups_ - минимальное значение _значащих_ групп узлов на одной высоте. По
  умолчанию _2_ группы.
- _--criterion-statehash-min-nodes-in-valuable-group_ - минимальное количество узлов сети в группе. Если группа узлов с
  одинаковыми стейтхешами имеет данное и более количество узлов, то она считается _значащей_. По умолчанию _2_ узла.
- _--criterion-statehash-require-min-nodes-on-same-height_ - необходимое количество узлов сети, которые находятся на
  одной высоте. По умолчанию _4_ узла.

## HTTP API

### Public URLs

1) **GET** _/health_ - возвращает состояние отслеживаемой сети, максимальную высоту и время запуска последней попытки
   обновления статистик. Если высоту не удалось получить хотя бы с одного узла, который участвует в мониторинге, то
   вместо высоты будет отдано _-1_.

    - Возможные HTTP коды ответа:
        - _200 OK_
        - _409 Method Not Allowed_
        - _500 Internal Server Error_
    - Возвращаемый результат:
        - `{"updated":"2021-12-02T19:35:24.144994Z","status":true,"height":2882018}` - сеть в порядке
        - `{"updated":"2021-12-02T19:35:24.144994Z","status":false,"height":2882018}` - сеть в деградированном
          состоянии, но если хотя бы один узел доступен
        - `{"updated":"2021-12-02T19:35:24.144994Z","status":false,"height":-1}` - сеть в деградированном состоянии и
          все узлы недоступны
    - Пример запроса: `curl http://localhost:2048/health`

### Private URLs

1) **POST** _/state_ - устанавливает состояние мониторинга. В случае, если новое состояние отличается от старого, то
   установка нового состояния сбрасывает счётчик последовательности ошибок.

    - Возможные HTTP коды ответа:
        - _200 OK_
        - _400 Bad Request_
        - _403 Forbidden_
        - _409 Method Not Allowed_
        - _500 Internal Server Error_
    - Возвращаемый результат: отсутствует
    - Возможные значения тела запроса:
        - `{"state":"active"}` - установка мониторинга в обычный (активный) режим работы
        - `{"state":"frozen_operates_stable"}` - установка мониторинга в режим, при котором он всегда будет отвечать на
          запрос **GET** _/health_ ответом  `{"status":true}`
        - `{"state":"frozen_degraded"}` - установка мониторинга в режим, при котором он всегда будет отвечать на
          запрос **GET** _/health_ ответом  `{"status":false}`
    - Пример
      запроса: `curl -X POST -H "Content-Type: application/json" -d '{"state":"active"}' http://localhost:2048/state`

## Build

Требования: на машине должны быть установлены утилита `Make`, компилятор и стандартная библиотека языка `Go`. Собрать
исполняемый файл можно командой `make build`. Исполняемый файл после сборки будет внутри директории `build`.

## Docker

Сборка проекта происходит при значении переменной окружения `CGO_ENABLED=0`, соответственно исполняемый файл сервиса не
требует динамически подключаемых библиотек и содержит в себе все необходимые зависимости для работы.

Сервис мониторинга внутри контейнер запускается от имени пользователя _netmon_. Часовая зона внутри контейнера
установлена в значение _Etc/UTC_.