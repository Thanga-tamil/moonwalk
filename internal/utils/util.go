package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"moonwalk/pkg"
	"strconv"

	log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	LogFile    = "moonwalk.log"
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
	return uuid.New().String()
}

func Filter(resources []pkg.Resources, filterType string) []pkg.Resources {
	result := make([]pkg.Resources, 0)

	for _, resource := range resources {
		if resource.Type == filterType {
			result = append(result, resource)
		}
	}

	return result
}

func PrettyPrint(title string, data interface{}) {
	data, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Errorf("failed to marshal %s: %v", title, err)
	}
	log.Infof("%s:\n%s", title, data)
}

func SortOrdersByCreatedAt(orders *[]pkg.Order) {
	for i := 0; i < len(*orders)-1; i++ {
		for j := 0; j < len(*orders)-i-1; j++ {
			if (*orders)[j].CreatedAt.After((*orders)[j+1].CreatedAt) {
				(*orders)[j], (*orders)[j+1] = (*orders)[j+1], (*orders)[j]
			}
		}
	}
}
