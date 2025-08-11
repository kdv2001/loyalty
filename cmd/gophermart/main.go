package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/store/postgress/auth"
	"github.com/kdv2001/loyalty/internal/store/postgress/session"
	"github.com/kdv2001/loyalty/internal/usecases/user"
)

func main() {
	log.Fatal(initService())
}

func initService() error {
	ctx := context.Background()

	initValues, err := initFlags()
	if err != nil {
		return err
	}

	postgresConn, err := pgxpool.New(ctx, initValues.postgresDSN)
	if err != nil {
		return err
	}
	if err = postgresConn.Ping(ctx); err != nil {
		return err
	}

	authStore, err := auth.NewImplementation(ctx, postgresConn)
	if err != nil {
		return err
	}

	sessionStore, err := session.NewImplementation(ctx, postgresConn)
	if err != nil {
		return err
	}

	userUC := user.NewImplementation(authStore, sessionStore)
	t, err := userUC.RegisterAndLoginUser(ctx, domain.Auth{
		Login:    "rr",
		Password: "12345",
	})
	if err != nil {
		return err
	}
	fmt.Println(t)

	fmt.Println(userUC.AuthUser(ctx, domain.SessionToken{
		Token: "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOnsiSUQiOjF9fQ",
	}))

	return err
}
