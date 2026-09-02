package utils

import (
	"errors"
	"fmt"
	"moonwalk/pkg"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ServerAddr = "0.0.0.0:8080"
	LogFile = "moonwalk.log"
	ConfigFile = "config.json"
)

func Pagination(ctx *gin.Context) (int, int, error) {
	const (
		defaultPage = 1
		defaultSize = 10
	)

	page := defaultPage
	size := defaultSize

	if pageStr := ctx.Query("page"); pageStr != "" {
		var err error

		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid page %q: %w", pageStr, err)
		}

		if page < 1 {
			return 0, 0, errors.New("page must be greater than 0")
		}
	}

	if sizeStr := ctx.Query("size"); sizeStr != "" {
		var err error

		size, err = strconv.Atoi(sizeStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid size %q: %w", sizeStr, err)
		}

		if size < 1 {
			return 0, 0, errors.New("size must be greater than 0")
		}
	}

	return page, size, nil
}

func GetRandomUUID() string {
	uuid := uuid.New().String()
	fmt.Println("uuid:", uuid)
	return uuid
}

func Filter(items []pkg.Resources) ([]pkg.Resources, []pkg.Resources) {
	chefs := make([]pkg.Resources, 0)
	servers := make([]pkg.Resources, 0)

	for _, item := range items {
		if strings.ToLower(item.Type) == "chef" {
			chefs = append(chefs, item)
		} else {
			servers = append(servers, item)
		}
	}

	return chefs, servers
}

