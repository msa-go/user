# User Service

MSA 구조의 사용자 서비스입니다. 회원가입, 로그인(JWT 발급), 사용자 정보 조회를 담당합니다.

## 기술 스택

| 구분 | 사용 기술 |
|------|-----------|
| 언어 | Go 1.26.5 |
| 웹 프레임워크 | [Gin](https://github.com/gin-gonic/gin) |
| ORM / DB | [GORM](https://gorm.io) + PostgreSQL |
| 캐시 | Redis (go-redis v9) |
| 인증 | JWT (golang-jwt/jwt v5), bcrypt |
| 설정 관리 | [Viper](https://github.com/spf13/viper) (YAML) + [godotenv](https://github.com/joho/godotenv) (.env) |
| 로깅 | [Logrus](https://github.com/sirupsen/logrus) |
| 트레이싱 | OpenTelemetry (OTLP HTTP) |

## 프로젝트 구조

레이어드 아키텍처를 따르며, 의존성은 항상 바깥쪽(HTTP)에서 안쪽(DB)으로 흐릅니다.

```
.
├── cmd/user/
│   ├── main.go            # 진입점 — 설정 로드, 의존성 조립, 서버 기동
│   ├── handler/           # HTTP 요청/응답 처리, 입력 유효성 검사
│   ├── usecase/           # 비즈니스 시나리오 (회원가입, 로그인, JWT 발급)
│   ├── service/           # 도메인 로직, repository 호출 조율
│   ├── repository/        # DB 질의 (GORM)
│   └── resource/          # DB / Redis 커넥션 초기화
├── config/                # 설정 구조체 정의 및 로딩 (Viper)
├── files/config/          # config.yaml (실제 설정값)
├── middleware/            # JWT 인증, 요청 로깅
├── models/                # 도메인 모델, 요청 DTO, 도메인 에러
├── routes/                # 라우팅 등록
├── infrastructure/log/    # 로거 초기화
├── trace/                 # OpenTelemetry 트레이서 초기화
├── utils/                 # 비밀번호 해싱 (bcrypt)
└── docker-compose.yml     # 로컬 개발용 PostgreSQL / Redis
```

### 요청 흐름

```
HTTP 요청
  → routes (라우팅)
  → middleware (요청 로깅 → JWT 인증)
  → handler   (요청 파싱, 유효성 검사)
  → usecase   (비즈니스 로직, 트랜잭션 경계)
  → service   (도메인 로직)
  → repository (GORM 질의)
  → PostgreSQL / Redis
```

각 계층은 바로 아래 계층만 알고 있습니다. 예를 들어 `handler`는 GORM의 존재를 모르며, DB에서 레코드를 찾지 못한 경우 `repository`가 `gorm.ErrRecordNotFound`를 도메인 에러인 `models.ErrUserNotFound`로 변환해 올려보냅니다.

## 요구 사항

- Go 1.26 이상
- Docker / Docker Compose (로컬 PostgreSQL, Redis 구동용)

## 초기 셋업

### 1. 의존성 설치

```bash
go mod download
```

### 2. `.env` 준비 (필수 — 컨테이너와 앱이 함께 사용)

```bash
cp .env.example .env
```

`docker-compose.yml`의 DB/Redis 비밀번호와 `config.LoadConfig()`가 읽는 앱 비밀값이 **같은 `.env` 파일**을 공유합니다. Docker Compose는 프로젝트 루트의 `.env`를 자동으로 읽어 `${VAR}` 자리를 치환하므로, 이 파일이 없으면 다음 단계의 `docker compose up`이 아예 실패합니다.

```console
$ docker compose config
error while interpolating services.postgres.environment.POSTGRES_PASSWORD: required variable USER_DATABASE_PASSWORD is missing a value: ...
```

값은 자세히 "5. 필수 환경변수"에서 다룹니다. 지금은 기본값(`password`) 그대로 두고 다음 단계로 넘어가도 됩니다.

### 3. 인프라 기동 (PostgreSQL + Redis)

```bash
docker compose up -d
```

| 서비스 | 이미지 | 포트 | 계정 |
|--------|--------|------|------|
| PostgreSQL | postgres:16-alpine | 5432 | `admin` / `.env`의 `USER_DATABASE_PASSWORD`, DB명 `user` |
| Redis | redis:7-alpine | 6379 | 비밀번호 `.env`의 `USER_REDIS_PASSWORD` |

`admin`, DB명 `user`는 `files/config/config.yaml`의 `database.user` / `database.name`과 맞춰야 하는 비밀 아닌 값이라 `docker-compose.yml`에 그대로 남아 있습니다.

컨테이너 상태 확인:

```bash
docker compose ps
```

### 4. 설정 확인

호스트/포트 등 비밀이 아닌 값은 `files/config/config.yaml`에서 읽습니다.

```yaml
app:
  port: 8080
  tokenexpiry: 24h        # 발급 토큰 유효 기간
  requesttimeout: 2s      # 요청당 컨텍스트 타임아웃

database:
  host: localhost
  user: admin
  name: user
  port: 5432

redis:
  host: 127.0.0.1
  port: 6379

observability:
  servicename: user
  otlpendpoint: localhost:4318

# 비밀값은 이 파일에 두지 않는다. 앞서 준비한 .env / "5. 필수 환경변수"로 주입한다.
```

`config.LoadConfig()`는 기동 시점에 `validate:"required"` 태그를 실제로 검사합니다. yaml/환경변수 어느 쪽이든 필수값이 비어 있으면 서버가 뜨지 않고 즉시 종료됩니다.

### 5. 필수 환경변수 (비밀값)

DB/Redis 비밀번호와 JWT 서명 키는 저장소에 커밋하지 않고 `USER_` 접두사 환경변수로만 주입받습니다. `AutomaticEnv`는 `.`을 `_`로 치환해 매칭하므로 (`config.SecretConfig.JWTSecret` → `secret.jwtsecret` → `USER_SECRET_JWTSECRET`), 아래 값을 실행 전에 반드시 설정해야 합니다. 같은 값을 "2. `.env` 준비"에서 만든 `.env`가 `docker-compose.yml`의 컨테이너 비밀번호에도 그대로 사용합니다.

| 환경변수 | 대응 설정 키 | 로컬 개발 값(`.env.example` 기준) |
|----------|--------------|-------------------------------------|
| `USER_DATABASE_PASSWORD` | `database.password` | `password` |
| `USER_REDIS_PASSWORD` | `redis.password` | `password` |
| `USER_SECRET_JWTSECRET` | `secret.jwtsecret` | 32자 이상 임의 문자열 (운영 값과 다르게) |

**로컬 개발 (`.env` 사용, 권장)**

`LoadConfig()`가 기동 시 `.env` 파일을 자동으로 읽어 프로세스 환경변수로 등록합니다([joho/godotenv](https://github.com/joho/godotenv)). `.env`는 `.gitignore`에 등록되어 있어 커밋되지 않습니다. "2. `.env` 준비"에서 이미 만들었다면 값만 필요 시 수정합니다.

```bash
go run ./cmd/user
```

이미 셸에 같은 이름의 환경변수가 설정되어 있으면 `.env` 값보다 그 실제 환경변수가 우선합니다(`godotenv.Load`는 기존 변수를 덮어쓰지 않습니다). `.env` 파일 자체가 없어도 에러 없이 넘어가며, 이 경우 운영 환경처럼 실제 환경변수만으로 동작합니다(단, `docker-compose.yml`은 `.env` 없이는 뜨지 않습니다).

**운영/CI (`.env` 없이 실제 환경변수만)**

```bash
export USER_DATABASE_PASSWORD=...
export USER_REDIS_PASSWORD=...
export USER_SECRET_JWTSECRET=...
go run ./cmd/user
```

두 방식 모두 `USER_SECRET_JWTSECRET`은 `min=32` 검증이 걸려 있어 32자 미만이면 기동이 거부됩니다.

### 6. 데이터베이스 스키마

자동 마이그레이션은 실행되지 않습니다. `users` 테이블은 서버 기동 전에 직접 생성해야 합니다.

## 실행 방법

"2. `.env` 준비"에서 `.env`를 만들었거나 "5. 필수 환경변수"의 값을 export 했다면 바로 실행합니다.

```bash
go run ./cmd/user
```

> **주의**: 반드시 프로젝트 루트에서 실행해야 합니다. 설정 로딩이 `./files/config` 상대 경로를 사용하기 때문에, `cmd/user` 디렉토리 안에서 `go run main.go`로 실행하면 설정 파일을 찾지 못하고 종료됩니다.

정상적으로 기동되면 다음 로그가 출력됩니다.

```
connected to DB
Connected to Redis
[GIN-debug] Listening and serving HTTP on :8080
```

동작 확인:

```bash
curl localhost:8080/ping
# {"status":"OK"}
```

### 빌드

```bash
go build -o bin/user ./cmd/user
./bin/user
```

## API 명세

기본 주소: `http://localhost:8080`

| 메서드 | 경로 | 인증 | 설명 |
|--------|------|------|------|
| GET | `/ping` | - | 헬스 체크 |
| POST | `/v1/register` | - | 회원가입 |
| POST | `/v1/login` | - | 로그인 (JWT 발급) |
| GET | `/api/v1/user_info` | Bearer | 내 정보 조회 |

### 회원가입

```bash
curl -X POST localhost:8080/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "홍길동",
    "email": "hong@example.com",
    "password": "password123",
    "confirm_password": "password123"
  }'
```

```json
{ "message": "회원가입 성공" }
```

**유효성 규칙**: 이메일 형식이어야 하며, 비밀번호는 8자 이상이고 `confirm_password`와 일치해야 합니다. 비밀번호는 bcrypt로 해싱되어 저장됩니다.

### 로그인

```bash
curl -X POST localhost:8080/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "hong@example.com",
    "password": "password123"
  }'
```

```json
{ "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }
```

발급된 토큰은 `user_id`, `exp` 클레임을 담고 있으며 `app.tokenexpiry`(기본 24시간) 후 만료됩니다. 인증 미들웨어는 HS256 서명만 허용하고, `exp`가 없는 토큰은 무조건 거부합니다.

이메일 미존재와 비밀번호 불일치는 계정 존재 여부가 노출되지 않도록 동일한 401 응답으로 처리됩니다.

### 내 정보 조회

```bash
curl localhost:8080/api/v1/user_info \
  -H "Authorization: Bearer <TOKEN>"
```

```json
{ "name": "홍길동", "email": "hong@example.com" }
```

### 에러 응답

| 상태 코드 | 상황 |
|-----------|------|
| 400 | 잘못된 요청 파라미터, 비밀번호 규칙 위반, `/api/v1/user_info`에서 존재하지 않는 사용자 |
| 401 | 로그인 실패(이메일/비밀번호 무관 동일 메시지), 토큰 누락 / 형식 오류 / 만료 / 알고리즘 불일치 |
| 500 | 서버 내부 오류 — 상세 원인은 클라이언트에 내려주지 않고 서버 로그(trace_id 포함)에만 남긴다 |

```json
{ "error_message": "이메일 또는 비밀번호가 올바르지 않습니다." }
```
```json
{ "error_message": "요청을 처리하는 중 오류가 발생했습니다." }
```

## 개발 시 참고

### 코드 검증

```bash
go build ./...
go vet ./...
```

### 인프라 정리

```bash
docker compose down       # 컨테이너만 중지
docker compose down -v    # 데이터 볼륨까지 삭제 (DB 초기화)
```

## 알려진 미구현 사항

- **분산 트레이싱은 배선되어 있으나 수집기가 필요**: `main.go`가 `trace.InitTracer(cfg.Observability.ServiceName, cfg.Observability.OTLPEndpoint)`를 호출해 활성화되어 있습니다. `observability.otlpendpoint`(기본 `localhost:4318`)에 OTLP 수집기(Jaeger 등)가 떠 있지 않으면 익스포트가 실패합니다.
- **Redis 미사용**: 커넥션은 초기화되어 `repository`에 주입되지만, 아직 캐싱 로직에서 사용하지 않습니다.
- **Graceful shutdown 미적용**: 종료 시그널 처리 없이 `router.Run()`으로 기동합니다.
- **테스트 없음**: 단위 테스트가 아직 작성되지 않았습니다.
- **DB 마이그레이션 미자동화**: `users` 테이블 생성/변경은 수동으로 관리해야 합니다.
