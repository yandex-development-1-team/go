package main

import (
	"context"
	"fmt"
	"github.com/yandex-development-1-team/go/internal/database"
)

func main() {
	fmt.Print("Start service")
	ctx := context.Background()

	clientDB := database.NewMockClient(ctx)
	clientDB.GetBoxes(ctx)
}
