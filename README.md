# GophKeeper

## Usage

### Users

#### SignUp

```bash
go run cmd/client/main.go signup -u admin -p qwerty
```

#### SignIn

```bash
go run cmd/client/main.go signin -u admin -p qwerty
```

### Bankcards

#### Create

```bash
go run cmd/client/main.go bankcard create mysecret --name "john smith" --number 0000111122223333 --expire-at "1234" --cvv 123
```
